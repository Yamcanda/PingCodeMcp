package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/models"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mozillazg/go-pinyin"
)

const (
	projectTimeout = 15 * time.Second
)

func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewProjectListTool(),
			NewCreateProjectTool(),
			NewAddProjectMemberTool(),
			NewRemoveProjectMemberTool(),
			NewGetProjectMembersTool(),
		}
	})
}

type User struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar"`
}

type Member struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Type string `json:"type"`
	User User   `json:"user"`
	Role *Role  `json:"role,omitempty"` // 角色信息，可选
}

type ProjectState struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Project struct {
	ID          string       `json:"id"`
	URL         string       `json:"url"`
	Visibility  string       `json:"visibility"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Identifier  string       `json:"identifier"`
	Color       string       `json:"color"`
	Description string       `json:"description"`
	Members     []Member     `json:"members"`
	State       ProjectState `json:"state"`
	Assignee    *User        `json:"assignee"`
	ScopeType   string       `json:"scope_type"`
	StartAt     *int64       `json:"start_at"`
	EndAt       *int64       `json:"end_at"`
	CreatedAt   int64        `json:"created_at"`
	CreatedBy   User         `json:"created_by"`
	UpdatedAt   int64        `json:"updated_at"`
	UpdatedBy   User         `json:"updated_by"`
	IsArchived  int          `json:"is_archived"`
	IsDeleted   int          `json:"is_deleted"`
}

type ProjectListResponse struct {
	PageIndex int       `json:"page_index"`
	PageSize  int       `json:"page_size"`
	Total     int       `json:"total"`
	Values    []Project `json:"values"`
}

type ProjectListTool struct {
	name        string
	description string
}

func NewProjectListTool() MCPTool {
	return &ProjectListTool{
		name:        "get_project_list",
		description: "获取 PingCode 项目列表 - 查看所有可访问的项目信息，包括项目名称、项目标识、类型、状态、成员等",
	}
}

func (t *ProjectListTool) GetName() string {
	return t.name
}

func (t *ProjectListTool) GetDescription() string {
	return t.description
}

func (t *ProjectListTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("identifier", mcp.Description("项目的标识 - 可选，项目的唯一标识")),
		mcp.WithString("type", mcp.Description("项目的类型 - 允许值: scrum, kanban, waterfall, hybrid")),
		mcp.WithBoolean("include_deleted", mcp.Description("是否查询已删除的项目 - 可选，该值默认为false")),
		mcp.WithBoolean("include_archived", mcp.Description("是否查询已归档的项目 - 可选，该值默认为false")),
		mcp.WithNumber("page_index", mcp.Description("页码 - 可选，从0开始，默认为为0时，表示第一页")),
		mcp.WithNumber("page_size", mcp.Description("每页数量 - 可选，默认30，最大100")),
	)
}

func (t *ProjectListTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "get_project_list")
	requestLogger.Info("开始获取项目列表")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	params := t.parseArguments(request.GetArguments())
	respStr, err := t.doRequest(ctx, token, params, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("项目列表API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp ProjectListResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := t.formatResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *ProjectListTool) parseArguments(args interface{}) map[string]string {
	params := map[string]string{
		"page_size": "100",
	}

	if args == nil {
		return params
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return params
	}

	// 解析字符串参数
	stringParams := []string{
		"identifier", "type", "include_deleted", "include_archived",
	}

	for _, param := range stringParams {
		if value, exists := argsMap[param]; exists {
			if str, ok := value.(string); ok && str != "" {
				params[param] = str
			}
		}
	}

	if pageIndex, exists := argsMap["page_index"]; exists {
		if pageIndexFloat, ok := pageIndex.(float64); ok {
			pageIndexInt := int(pageIndexFloat) - 1
			if pageIndexInt < 0 {
				pageIndexInt = 0
			}
			params["page_index"] = fmt.Sprintf("%d", pageIndexInt)
		}
	}

	if pageSize, exists := argsMap["page_size"]; exists {
		if pageSizeFloat, ok := pageSize.(float64); ok {
			params["page_size"] = fmt.Sprintf("%.0f", pageSizeFloat)
		}
	}

	return params
}

func (t *ProjectListTool) doRequest(ctx context.Context, token string, params map[string]string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, projectTimeout)
	defer cancel()

	apiURL := GetBaseUrl() + "/v1/project/projects"
	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, params)
	if err != nil {
		return "", err
	}
	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("项目列表API请求成功")

	return string(body), nil
}

func (t *ProjectListTool) formatResponse(resp ProjectListResponse) string {
	var result strings.Builder

	// 计算当前页码（从1开始显示）和总页数
	currentPage := resp.PageIndex + 1
	totalPages := (resp.Total + resp.PageSize - 1) / resp.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	result.WriteString(fmt.Sprintf("📋 项目列表 (第 %d 页，共 %d 页，总计 %d 个项目):\n\n", currentPage, totalPages, resp.Total))

	if len(resp.Values) == 0 {
		result.WriteString("暂无项目数据")
		return result.String()
	}

	for i, project := range resp.Values {
		result.WriteString(fmt.Sprintf("📂 %d. %s (项目标识：%s)\n", i+1, project.Name, project.Identifier))
		result.WriteString(fmt.Sprintf("🆔 项目ID(project_id): %s\n", project.ID))
		result.WriteString(fmt.Sprintf("类型: %s\n", project.Type))
		result.WriteString(fmt.Sprintf("状态: %s\n", project.State))
		result.WriteString(fmt.Sprintf("📦 可见性: %s\n", project.Visibility))

		if project.Color != "" {
			result.WriteString(fmt.Sprintf("🌈 颜色: %s\n", project.Color))
		}

		if project.Description != "" {
			result.WriteString(fmt.Sprintf("📝 描述: %s\n", project.Description))
		}

		// 显示成员信息
		if len(project.Members) > 0 {
			result.WriteString(fmt.Sprintf("   成员数量: %d\n", len(project.Members)))
			result.WriteString("   成员列表:\n")
			for _, member := range project.Members {
				result.WriteString(fmt.Sprintf("     - %s (%s)\n", member.User.DisplayName, member.User.Name))
			}
		}

		// 显示负责人
		if project.Assignee != nil {
			result.WriteString(fmt.Sprintf("   负责人: %s (%s)\n", project.Assignee.DisplayName, project.Assignee.Name))
		}

		// 显示时间信息
		if project.StartAt != nil {
			startTime := time.Unix(*project.StartAt, 0).Format("2006-01-02")
			result.WriteString(fmt.Sprintf("   开始时间: %s\n", startTime))
		}

		if project.EndAt != nil {
			endTime := time.Unix(*project.EndAt, 0).Format("2006-01-02")
			result.WriteString(fmt.Sprintf("   结束时间: %s\n", endTime))
		}

		// 显示创建信息
		createdTime := time.Unix(project.CreatedAt, 0).Format("2006-01-02 15:04:05")
		result.WriteString(fmt.Sprintf("   创建时间: %s\n", createdTime))
		result.WriteString(fmt.Sprintf("   创建者: %s (%s)\n", project.CreatedBy.DisplayName, project.CreatedBy.Name))

		// 显示归档状态
		if project.IsArchived == 1 {
			result.WriteString("   状态: 已归档\n")
		}

		result.WriteString(fmt.Sprintf("   详情链接: %s\n\n", project.URL))
	}

	return result.String()
}

// CreateProjectRequest 创建项目请求结构体
type CreateProjectRequest struct {
	Name                string `json:"name"`
	Identifier          string `json:"identifier"`
	Type                string `json:"type"`
	Visibility          string `json:"visibility,omitempty"`
	Description         string `json:"description,omitempty"`
	Color               string `json:"color,omitempty"`
	IdentifierGenerated bool   `json:"-"` // 标记 identifier 是否自动生成，不发送到 API
}

// generateProjectIdentifier 根据项目名称生成项目标识符
// 如果项目名称包含中文，转换为拼音首字母；如果是英文，取首字母
func generateProjectIdentifier(projectName string) string {
	if projectName == "" {
		return ""
	}

	var result strings.Builder

	// 遍历项目名称的每个字符
	for _, char := range projectName {
		if unicode.Is(unicode.Han, char) {
			// 中文字符，转换为拼音首字母
			pinyinSlice := pinyin.Pinyin(string(char), pinyin.NewArgs())
			if len(pinyinSlice) > 0 && len(pinyinSlice[0]) > 0 {
				// 取拼音的首字母并转为大写
				firstLetter := strings.ToUpper(string(pinyinSlice[0][0][0]))
				result.WriteString(firstLetter)
			}
		} else if unicode.IsLetter(char) {
			// 英文字母，直接取首字母并转为大写
			result.WriteString(strings.ToUpper(string(char)))
		}
		// 忽略数字、空格和其他特殊字符
	}

	identifier := result.String()

	// 如果生成的标识符为空，使用默认值
	if identifier == "" {
		identifier = "PROJ"
	}

	// 限制长度，避免过长
	if len(identifier) > 10 {
		identifier = identifier[:10]
	}

	return identifier
}

// 创建项目工具结构体
type CreateProjectTool struct {
	name        string
	description string
}

// 创建项目工具实例
func NewCreateProjectTool() MCPTool {
	return &CreateProjectTool{
		name:        "create_project",
		description: "Create a new project in PingCode",
	}
}

// 返回工具名称
func (t *CreateProjectTool) GetName() string {
	return t.name
}

// 返回工具描述
func (t *CreateProjectTool) GetDescription() string {
	return t.description
}

// 返回工具定义
func (t *CreateProjectTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("创建新的 PingCode 项目 - 支持 Scrum、Kanban、瀑布等项目类型，可设置项目名称、标识符、可见性等"),
		mcp.WithString("name", mcp.Description("项目名称 - 必填，项目的显示名称"), mcp.Required()),
		mcp.WithString("identifier", mcp.Description("项目的标识 - 可选，项目的唯一标识，如果为空将根据项目名称自动生成")),
		mcp.WithString("type", mcp.Description("项目类型 - 必填，支持：scrum(敏捷)、kanban(看板)、waterfall(瀑布)"), mcp.Required()),
		mcp.WithString("visibility", mcp.Description("项目可见性 - 可选，支持：public(公开)、private(私有)，默认private")),
		mcp.WithString("description", mcp.Description("项目描述 - 可选，项目的详细描述信息")),
		mcp.WithString("color", mcp.Description("项目颜色 - 可选，十六进制格式，如：#56ABFB")),
	)
}

// Handle 处理创建项目请求
func (t *CreateProjectTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "create_project", "request_id", fmt.Sprintf("create_project_%d", time.Now().UnixNano()))

	requestLogger.Info("开始创建项目")

	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("创建项目失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 解析请求参数
	projectRequest, err := t.parseCreateProjectArguments(request.GetArguments())
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("解析请求参数失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doCreateProjectRequest(ctx, token, projectRequest, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("创建项目API调用失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("创建项目API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp Project
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("success", false, "error", "json_parse_failed").Error("解析创建项目响应失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 格式化返回结果
	result := t.formatCreateProjectResponse(resp, projectRequest.IdentifierGenerated)

	requestLogger.With("success", true, "project_id", resp.ID, "project_name", resp.Name).Info("项目创建成功")

	return mcp.NewToolResultText(result), nil
}

// parseCreateProjectArguments 解析创建项目请求参数
func (t *CreateProjectTool) parseCreateProjectArguments(args interface{}) (*CreateProjectRequest, error) {
	if args == nil {
		return nil, fmt.Errorf("missing required arguments")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments format")
	}

	projectRequest := &CreateProjectRequest{}

	// 必需参数
	if name, exists := argsMap["name"]; exists {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			projectRequest.Name = nameStr
		} else {
			return nil, fmt.Errorf("name is required and must be a non-empty string")
		}
	} else {
		return nil, fmt.Errorf("name is required")
	}

	// identifier 参数处理：如果为空或不存在，根据项目名称自动生成
	if identifier, exists := argsMap["identifier"]; exists {
		if identifierStr, ok := identifier.(string); ok && identifierStr != "" {
			projectRequest.Identifier = identifierStr
			projectRequest.IdentifierGenerated = false
		} else {
			// identifier 存在但为空，自动生成
			projectRequest.Identifier = generateProjectIdentifier(projectRequest.Name)
			projectRequest.IdentifierGenerated = true
		}
	} else {
		// identifier 不存在，自动生成
		projectRequest.Identifier = generateProjectIdentifier(projectRequest.Name)
		projectRequest.IdentifierGenerated = true
	}

	if projectType, exists := argsMap["type"]; exists {
		if typeStr, ok := projectType.(string); ok && typeStr != "" {
			// 验证项目类型
			validTypes := map[string]bool{
				"scrum":     true,
				"kanban":    true,
				"waterfall": true,
			}
			if !validTypes[typeStr] {
				return nil, fmt.Errorf("invalid project type: %s. Valid types are: scrum, kanban, waterfall", typeStr)
			}
			projectRequest.Type = typeStr
		} else {
			return nil, fmt.Errorf("type is required and must be a non-empty string")
		}
	} else {
		return nil, fmt.Errorf("type is required")
	}

	// 可选参数
	if visibility, exists := argsMap["visibility"]; exists {
		if visibilityStr, ok := visibility.(string); ok && visibilityStr != "" {
			// 验证可见性
			validVisibilities := map[string]bool{
				"public":  true,
				"private": true,
			}
			if !validVisibilities[visibilityStr] {
				return nil, fmt.Errorf("invalid visibility: %s. Valid values are: public, private", visibilityStr)
			}
			projectRequest.Visibility = visibilityStr
		}
	}

	if description, exists := argsMap["description"]; exists {
		if descriptionStr, ok := description.(string); ok {
			projectRequest.Description = descriptionStr
		}
	}

	if color, exists := argsMap["color"]; exists {
		if colorStr, ok := color.(string); ok && colorStr != "" {
			// 简单验证颜色格式
			if !strings.HasPrefix(colorStr, "#") || len(colorStr) != 7 {
				return nil, fmt.Errorf("invalid color format: %s. Expected format: #RRGGBB", colorStr)
			}
			projectRequest.Color = colorStr
		}
	}

	return projectRequest, nil
}

// doCreateProjectRequest 执行创建项目API调用
func (t *CreateProjectTool) doCreateProjectRequest(ctx context.Context, token string, projectRequest *CreateProjectRequest, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}
	apiURL := GetBaseUrl() + "/v1/project/projects"
	log.With("url", apiURL, "method", "POST", "project_name", projectRequest.Name).Debug("发起创建项目API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, projectTimeout)
	defer cancel()

	// 序列化请求体
	requestBody, err := json.Marshal(projectRequest)
	if err != nil {
		log.With("error", err.Error()).Error("序列化请求体失败")
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	body, _, err := utils.DoPostJSON(timeoutCtx, apiURL, headers, requestBody)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("创建项目API请求成功")

	return string(body), nil
}

// formatCreateProjectResponse 格式化创建项目响应结果
func (t *CreateProjectTool) formatCreateProjectResponse(project Project, identifierGenerated bool) string {
	var result strings.Builder

	result.WriteString("项目创建成功:\n\n")
	result.WriteString(fmt.Sprintf("项目名称: %s\n", project.Name))

	if identifierGenerated {
		result.WriteString(fmt.Sprintf("项目标识符: %s (根据项目名称自动生成)\n", project.Identifier))
	} else {
		result.WriteString(fmt.Sprintf("项目标识符: %s\n", project.Identifier))
	}

	result.WriteString(fmt.Sprintf("项目ID: %s\n", project.ID))
	result.WriteString(fmt.Sprintf("项目类型: %s\n", project.Type))
	result.WriteString(fmt.Sprintf("可见性: %s\n", project.Visibility))

	if project.Description != "" {
		result.WriteString(fmt.Sprintf("描述: %s\n", project.Description))
	}

	if project.Color != "" {
		result.WriteString(fmt.Sprintf("颜色: %s\n", project.Color))
	}

	result.WriteString(fmt.Sprintf("状态: %s\n", project.State))

	// 显示创建信息
	createdTime := time.Unix(project.CreatedAt, 0).Format("2006-01-02 15:04:05")
	result.WriteString(fmt.Sprintf("创建时间: %s\n", createdTime))
	result.WriteString(fmt.Sprintf("创建者: %s (%s)\n", project.CreatedBy.DisplayName, project.CreatedBy.Name))

	result.WriteString(fmt.Sprintf("详情链接: %s\n", project.URL))

	return result.String()
}

// 添加项目成员请求结构体
type AddProjectMemberRequest struct {
	UserID string `json:"user_id"`
	Type   string `json:"type,omitempty"` // 成员类型，可选
}

// 添加项目成员响应结构体
type AddProjectMemberResponse struct {
	ID      string  `json:"id"`
	URL     string  `json:"url"`
	Project Project `json:"project"`
	Type    string  `json:"type"`
	User    User    `json:"user"`
	Role    Role    `json:"role"`
}

// 添加项目成员工具结构体
type AddProjectMemberTool struct {
	name        string
	description string
}

// 创建添加项目成员工具实例
func NewAddProjectMemberTool() MCPTool {
	return &AddProjectMemberTool{
		name:        "add_project_member",
		description: "Add a member to a PingCode project",
	}
}

// 返回工具名称
func (t *AddProjectMemberTool) GetName() string {
	return t.name
}

// 返回工具描述
func (t *AddProjectMemberTool) GetDescription() string {
	return t.description
}

// 返回工具定义
func (t *AddProjectMemberTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("添加项目成员 - 将指定用户添加到项目中，可设置成员类型和角色"),
		mcp.WithString("project_id", mcp.Description("项目ID - 必填，要添加成员的项目唯一标识，可通过get_project_list工具获取"), mcp.Required()),
		mcp.WithString("user_id", mcp.Description("用户ID - 必填，要添加的用户唯一标识，可通过企业用户列表工具获取"), mcp.Required()),
		mcp.WithString("type", mcp.Description("成员类型 - 可选，指定成员在项目中的类型")),
	)
}

// 处理添加项目成员请求
func (t *AddProjectMemberTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "add_project_member", "request_id", fmt.Sprintf("add_member_%d", time.Now().UnixNano()))

	requestLogger.Info("开始添加项目成员")

	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("添加项目成员失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 解析请求参数
	projectID, memberRequest, err := t.parseAddMemberArguments(request.GetArguments())
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("解析请求参数失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doAddMemberRequest(ctx, token, projectID, memberRequest, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("添加项目成员API调用失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("添加项目成员API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp AddProjectMemberResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("success", false, "error", "json_parse_failed").Error("解析添加成员响应失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 格式化返回结果
	result := t.formatAddMemberResponse(resp, projectID)

	requestLogger.With("success", true, "project_id", projectID, "user_id", memberRequest.UserID).Info("项目成员添加成功")

	return mcp.NewToolResultText(result), nil
}

// parseAddMemberArguments 解析添加项目成员请求参数
func (t *AddProjectMemberTool) parseAddMemberArguments(args interface{}) (string, *AddProjectMemberRequest, error) {
	if args == nil {
		return "", nil, fmt.Errorf("missing required arguments")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("invalid arguments format")
	}

	var projectID string
	memberRequest := &AddProjectMemberRequest{}

	// 必需参数：项目ID
	if pid, exists := argsMap["project_id"]; exists {
		if pidStr, ok := pid.(string); ok && pidStr != "" {
			projectID = pidStr
		} else {
			return "", nil, fmt.Errorf("project_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("project_id is required")
	}

	// 必需参数：用户ID
	if uid, exists := argsMap["user_id"]; exists {
		if uidStr, ok := uid.(string); ok && uidStr != "" {
			memberRequest.UserID = uidStr
		} else {
			return "", nil, fmt.Errorf("user_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("user_id is required")
	}

	// 可选参数：成员类型
	if memberType, exists := argsMap["type"]; exists {
		if typeStr, ok := memberType.(string); ok && typeStr != "" {
			memberRequest.Type = typeStr
		}
	}

	return projectID, memberRequest, nil
}

// doAddMemberRequest 执行添加项目成员API调用
func (t *AddProjectMemberTool) doAddMemberRequest(ctx context.Context, token string, projectID string, memberRequest *AddProjectMemberRequest, log *logger.Logger) (string, error) {
	// 构建API URL
	apiURL := fmt.Sprintf(GetBaseUrl()+"/v1/project/projects/%s/members", projectID)

	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	log.With("url", apiURL, "method", "POST", "project_id", projectID, "user_id", memberRequest.UserID).Debug("发起添加项目成员API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, projectTimeout)
	defer cancel()

	// 序列化请求体
	requestBody, err := json.Marshal(memberRequest)
	if err != nil {
		log.With("error", err.Error()).Error("序列化请求体失败")
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	body, _, err := utils.DoPostJSON(timeoutCtx, apiURL, headers, requestBody)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("添加项目成员API请求成功")

	return string(body), nil
}

// formatAddMemberResponse 格式化添加项目成员响应结果
func (t *AddProjectMemberTool) formatAddMemberResponse(resp AddProjectMemberResponse, projectID string) string {
	var result strings.Builder

	result.WriteString("项目成员添加成功:\n\n")
	result.WriteString(fmt.Sprintf("项目ID: %s\n", projectID))
	result.WriteString(fmt.Sprintf("项目名称: %s (%s)\n", resp.Project.Name, resp.Project.Identifier))
	result.WriteString(fmt.Sprintf("成员ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("成员类型: %s\n", resp.Type))

	// 显示角色信息
	result.WriteString(fmt.Sprintf("角色: %s (ID: %s)\n", resp.Role.Name, resp.Role.ID))

	result.WriteString("用户信息:\n")
	result.WriteString(fmt.Sprintf("  - 用户ID: %s\n", resp.User.ID))
	result.WriteString(fmt.Sprintf("  - 用户名: %s\n", resp.User.Name))
	result.WriteString(fmt.Sprintf("  - 显示名称: %s\n", resp.User.DisplayName))

	if resp.User.Avatar != "" {
		result.WriteString(fmt.Sprintf("  - 头像: %s\n", resp.User.Avatar))
	}

	result.WriteString(fmt.Sprintf("  - 详情链接: %s\n", resp.User.URL))
	result.WriteString(fmt.Sprintf("成员详情链接: %s\n", resp.URL))

	return result.String()
}

// RemoveProjectMemberRequest 移除项目成员请求结构体
type RemoveProjectMemberRequest struct {
	UserID string `json:"user_id"`
}

// RemoveProjectMemberResponse 移除项目成员响应结构体
type RemoveProjectMemberResponse struct {
	ID      string  `json:"id"`
	URL     string  `json:"url"`
	Type    string  `json:"type"`
	User    User    `json:"user"`
	Role    Role    `json:"role"`
	Project Project `json:"project"`
}

// 角色结构体
type Role struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

// 移除项目成员工具结构体
type RemoveProjectMemberTool struct {
	name        string
	description string
}

// 创建移除项目成员工具实例
func NewRemoveProjectMemberTool() MCPTool {
	return &RemoveProjectMemberTool{
		name:        "remove_project_member",
		description: "Remove a member from a PingCode project",
	}
}

// 返回工具名称
func (t *RemoveProjectMemberTool) GetName() string {
	return t.name
}

// 返回工具描述
func (t *RemoveProjectMemberTool) GetDescription() string {
	return t.description
}

// 返回工具定义
func (t *RemoveProjectMemberTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("移除项目成员 - 从指定项目中移除用户，取消其项目访问权限"),
		mcp.WithString("project_id", mcp.Description("项目ID - 必填，要移除成员的项目唯一标识，可通过 get_project_list 工具获取"), mcp.Required()),
		mcp.WithString("user_id", mcp.Description("用户ID - 必填，要移除的用户唯一标识，可通过 get_project_members 工具获取"), mcp.Required()),
	)
}

// 处理移除项目成员请求
func (t *RemoveProjectMemberTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "remove_project_member", "request_id", fmt.Sprintf("remove_member_%d", time.Now().UnixNano()))

	requestLogger.Info("开始移除项目成员")

	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("移除项目成员失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 解析请求参数
	projectID, memberRequest, err := t.parseRemoveMemberArguments(request.GetArguments())
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("解析请求参数失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doRemoveMemberRequest(ctx, token, projectID, memberRequest, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("移除项目成员API调用失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("移除项目成员API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp RemoveProjectMemberResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("success", false, "error", "json_parse_failed").Error("解析移除成员响应失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 格式化返回结果
	result := t.formatRemoveMemberResponse(resp, projectID)

	requestLogger.With("success", true, "project_id", projectID, "user_id", memberRequest.UserID).Info("项目成员移除成功")

	return mcp.NewToolResultText(result), nil
}

// 解析移除项目成员请求参数
func (t *RemoveProjectMemberTool) parseRemoveMemberArguments(args interface{}) (string, *RemoveProjectMemberRequest, error) {
	if args == nil {
		return "", nil, fmt.Errorf("missing required arguments")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("invalid arguments format")
	}

	var projectID string
	memberRequest := &RemoveProjectMemberRequest{}

	// 必需参数：项目ID
	if pid, exists := argsMap["project_id"]; exists {
		if pidStr, ok := pid.(string); ok && pidStr != "" {
			projectID = pidStr
		} else {
			return "", nil, fmt.Errorf("project_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("project_id is required")
	}

	// 必需参数：用户ID
	if uid, exists := argsMap["user_id"]; exists {
		if uidStr, ok := uid.(string); ok && uidStr != "" {
			memberRequest.UserID = uidStr
		} else {
			return "", nil, fmt.Errorf("user_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("user_id is required")
	}

	return projectID, memberRequest, nil
}

// 执行移除项目成员API调用
func (t *RemoveProjectMemberTool) doRemoveMemberRequest(ctx context.Context, token string, projectID string, memberRequest *RemoveProjectMemberRequest, log *logger.Logger) (string, error) {
	// 构建API URL - 使用 projectRemoveMemberAPIURL 并包含用户ID
	apiURL := fmt.Sprintf(GetBaseUrl()+"/v1/project/projects/%s/members/%s", projectID, memberRequest.UserID)

	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	log.With("url", apiURL, "method", "DELETE", "project_id", projectID, "user_id", memberRequest.UserID).Debug("发起移除项目成员API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, projectTimeout)
	defer cancel()

	// 对于DELETE请求，通常不需要请求体，直接调用DoDelete
	body, _, err := utils.DoDelete(timeoutCtx, apiURL, headers, nil)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("移除项目成员API请求成功")

	return string(body), nil
}

// 格式化移除项目成员响应结果
func (t *RemoveProjectMemberTool) formatRemoveMemberResponse(resp RemoveProjectMemberResponse, projectID string) string {
	var result strings.Builder

	result.WriteString("项目成员移除成功:\n\n")
	result.WriteString(fmt.Sprintf("项目ID: %s\n", projectID))
	result.WriteString(fmt.Sprintf("项目名称: %s (%s)\n", resp.Project.Name, resp.Project.Identifier))
	result.WriteString("移除的成员信息:\n")
	result.WriteString(fmt.Sprintf("  - 成员ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("  - 成员类型: %s\n", resp.Type))
	result.WriteString(fmt.Sprintf("  - 用户ID: %s\n", resp.User.ID))
	result.WriteString(fmt.Sprintf("  - 用户名: %s\n", resp.User.Name))
	result.WriteString(fmt.Sprintf("  - 显示名称: %s\n", resp.User.DisplayName))

	if resp.User.Avatar != "" {
		result.WriteString(fmt.Sprintf("  - 头像: %s\n", resp.User.Avatar))
	}

	result.WriteString(fmt.Sprintf("  - 角色: %s (ID: %s)\n", resp.Role.Name, resp.Role.ID))
	result.WriteString(fmt.Sprintf("  - 用户详情: %s\n", resp.User.URL))
	result.WriteString(fmt.Sprintf("  - 成员详情: %s\n", resp.URL))

	return result.String()
}

// 获取项目成员列表响应结构体
type ProjectMembersResponse struct {
	PageIndex int      `json:"page_index"`
	PageSize  int      `json:"page_size"`
	Total     int      `json:"total"`
	Values    []Member `json:"values"`
}

// 获取项目成员列表工具结构体
type GetProjectMembersTool struct {
	name        string
	description string
}

// 创建获取项目成员列表工具实例
func NewGetProjectMembersTool() MCPTool {
	return &GetProjectMembersTool{
		name:        "get_project_members",
		description: "Get the list of members in a PingCode project",
	}
}

// 返回工具名称
func (t *GetProjectMembersTool) GetName() string {
	return t.name
}

// 返回工具描述
func (t *GetProjectMembersTool) GetDescription() string {
	return t.description
}

// 返回工具定义
func (t *GetProjectMembersTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("获取项目成员列表 - 查看指定项目的所有成员信息，包括用户详情、角色、权限等"),
		mcp.WithString("project_id", mcp.Description("项目ID - 必填，要查询成员的项目唯一标识，可通过 get_project_list 工具获取"), mcp.Required()),
		mcp.WithNumber("page_index", mcp.Description("页码 - 可选，从0开始，默认为为0时，表示第一页")),
		mcp.WithNumber("page_size", mcp.Description("每页数量 - 可选，默认30，最大100")),
	)
}

// Handle 处理获取项目成员列表请求
func (t *GetProjectMembersTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "get_project_members", "request_id", fmt.Sprintf("get_members_%d", time.Now().UnixNano()))

	requestLogger.Info("开始获取项目成员列表")

	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("获取项目成员列表失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 解析请求参数
	projectID, params, err := t.parseGetMembersArguments(request.GetArguments())
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("解析请求参数失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doGetMembersRequest(ctx, token, projectID, params, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("获取项目成员列表API调用失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("获取项目成员列表API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp ProjectMembersResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("success", false, "error", "json_parse_failed").Error("解析获取成员列表响应失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 格式化返回结果
	result := t.formatGetMembersResponse(resp, projectID)

	requestLogger.With("success", true, "project_id", projectID, "member_count", len(resp.Values)).Info("项目成员列表获取成功")

	return mcp.NewToolResultText(result), nil
}

// 解析获取项目成员列表请求参数
func (t *GetProjectMembersTool) parseGetMembersArguments(args interface{}) (string, map[string]string, error) {
	if args == nil {
		return "", nil, fmt.Errorf("missing required arguments")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("invalid arguments format")
	}

	var projectID string
	params := map[string]string{
		"page_index": "0",
		"page_size":  "20",
	}

	// 必需参数：项目ID
	if pid, exists := argsMap["project_id"]; exists {
		if pidStr, ok := pid.(string); ok && pidStr != "" {
			projectID = pidStr
		} else {
			return "", nil, fmt.Errorf("project_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("project_id is required")
	}

	// 可选参数：页码
	if page, exists := argsMap["page"]; exists {
		if pageFloat, ok := page.(float64); ok {
			pageIndex := int(pageFloat) - 1
			if pageIndex < 0 {
				pageIndex = 0
			}
			params["page_index"] = fmt.Sprintf("%d", pageIndex)
		}
	}

	// 可选参数：每页数量
	if pageSize, exists := argsMap["page_size"]; exists {
		if pageSizeFloat, ok := pageSize.(float64); ok {
			params["page_size"] = fmt.Sprintf("%.0f", pageSizeFloat)
		}
	}

	return projectID, params, nil
}

// 执行获取项目成员列表API调用
func (t *GetProjectMembersTool) doGetMembersRequest(ctx context.Context, token string, projectID string, params map[string]string, log *logger.Logger) (string, error) {
	// 构建API URL
	apiURL := fmt.Sprintf(GetBaseUrl()+"/v1/project/projects/%s/members", projectID)

	headers := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	log.With("url", apiURL, "method", "GET", "project_id", projectID).Debug("发起获取项目成员列表API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, projectTimeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, params)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("获取项目成员列表API请求成功")

	return string(body), nil
}

// 格式化获取项目成员列表响应结果
func (t *GetProjectMembersTool) formatGetMembersResponse(resp ProjectMembersResponse, projectID string) string {
	var result strings.Builder

	// 计算当前页码（从1开始显示）和总页数
	currentPage := resp.PageIndex + 1
	totalPages := (resp.Total + resp.PageSize - 1) / resp.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	result.WriteString(fmt.Sprintf("项目成员列表 (项目ID: %s)\n", projectID))
	result.WriteString(fmt.Sprintf("第%d页，共%d页，总计%d个成员:\n\n", currentPage, totalPages, resp.Total))

	if len(resp.Values) == 0 {
		result.WriteString("该项目暂无成员")
		return result.String()
	}

	for i, member := range resp.Values {
		result.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, member.User.DisplayName, member.User.Name))
		result.WriteString(fmt.Sprintf("   成员ID: %s\n", member.ID))
		result.WriteString(fmt.Sprintf("   用户ID: %s\n", member.User.ID))
		result.WriteString(fmt.Sprintf("   成员类型: %s\n", member.Type))

		// 显示角色信息
		if member.Role != nil {
			result.WriteString(fmt.Sprintf("   角色: %s (ID: %s)\n", member.Role.Name, member.Role.ID))
		}

		if member.User.Avatar != "" {
			result.WriteString(fmt.Sprintf("   头像: %s\n", member.User.Avatar))
		}

		result.WriteString(fmt.Sprintf("   用户详情: %s\n", member.User.URL))
		result.WriteString(fmt.Sprintf("   成员详情: %s\n\n", member.URL))
	}

	return result.String()
}
