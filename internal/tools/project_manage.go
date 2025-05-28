package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mozillazg/go-pinyin"
)

const (
	projectAPIURL  = baseUrl + "/v1/project/projects"
	projectTimeout = 15 * time.Second
)

func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewProjectListTool(),
			NewCreateProjectTool(),
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
		description: "Get the list of PingCode projects",
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
		mcp.WithNumber("page", mcp.Description("页码，从1开始")),
		mcp.WithNumber("page_size", mcp.Description("每页数量，最大100")),
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

	var resp ProjectListResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := t.formatResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *ProjectListTool) parseArguments(args interface{}) map[string]string {
	params := map[string]string{
		"page_index": "0",
		"page_size":  "20",
	}

	if args == nil {
		return params
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return params
	}

	if page, exists := argsMap["page"]; exists {
		if pageFloat, ok := page.(float64); ok {
			pageIndex := int(pageFloat) - 1
			if pageIndex < 0 {
				pageIndex = 0
			}
			params["page_index"] = fmt.Sprintf("%d", pageIndex)
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

	body, _, err := utils.DoGet(timeoutCtx, projectAPIURL, headers, params)
	if err != nil {
		return "", err
	}
	log.With("url", projectAPIURL, "response_size_bytes", len(body), "success", true).Debug("项目列表API请求成功")

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

	result.WriteString(fmt.Sprintf("项目列表 (第%d页，共%d页，总计%d个项目):\n\n", currentPage, totalPages, resp.Total))

	if len(resp.Values) == 0 {
		result.WriteString("暂无项目数据")
		return result.String()
	}

	for i, project := range resp.Values {
		result.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, project.Name, project.Identifier))
		result.WriteString(fmt.Sprintf("   项目ID: %s\n", project.ID))
		result.WriteString(fmt.Sprintf("   类型: %s\n", project.Type))
		result.WriteString(fmt.Sprintf("   状态: %s (%s)\n", project.State.Name, project.State.Type))
		result.WriteString(fmt.Sprintf("   可见性: %s\n", project.Visibility))

		if project.Color != "" {
			result.WriteString(fmt.Sprintf("   颜色: %s\n", project.Color))
		}

		if project.Description != "" {
			result.WriteString(fmt.Sprintf("   描述: %s\n", project.Description))
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
	IdentifierGenerated bool   `json:"-"` // 标记identifier是否自动生成，不发送到API
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

// CreateProjectTool 创建项目工具结构体
type CreateProjectTool struct {
	name        string
	description string
}

// NewCreateProjectTool 创建项目工具实例
func NewCreateProjectTool() MCPTool {
	return &CreateProjectTool{
		name:        "create_project",
		description: "Create a new project in PingCode",
	}
}

// GetName 返回工具名称
func (t *CreateProjectTool) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *CreateProjectTool) GetDescription() string {
	return t.description
}

// GetToolDefinition 返回工具定义
func (t *CreateProjectTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("name",
			mcp.Description("项目名称"),
			mcp.Required(),
		),
		mcp.WithString("identifier",
			mcp.Description("项目标识符（可选，如果为空将根据项目名称自动生成）"),
		),
		mcp.WithString("type",
			mcp.Description("项目类型 (scrum, kanban, waterfall)"),
			mcp.Required(),
		),
		mcp.WithString("visibility",
			mcp.Description("项目可见性 (public, private)"),
		),
		mcp.WithString("description",
			mcp.Description("项目描述"),
		),
		mcp.WithString("color",
			mcp.Description("项目颜色（十六进制格式，如 #56ABFB）"),
		),
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

	// identifier参数处理：如果为空或不存在，根据项目名称自动生成
	if identifier, exists := argsMap["identifier"]; exists {
		if identifierStr, ok := identifier.(string); ok && identifierStr != "" {
			projectRequest.Identifier = identifierStr
			projectRequest.IdentifierGenerated = false
		} else {
			// identifier存在但为空，自动生成
			projectRequest.Identifier = generateProjectIdentifier(projectRequest.Name)
			projectRequest.IdentifierGenerated = true
		}
	} else {
		// identifier不存在，自动生成
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

	log.With("url", projectAPIURL, "method", "POST", "project_name", projectRequest.Name).Debug("发起创建项目API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, projectTimeout)
	defer cancel()

	// 序列化请求体
	requestBody, err := json.Marshal(projectRequest)
	if err != nil {
		log.With("error", err.Error()).Error("序列化请求体失败")
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	body, _, err := utils.DoPostJSON(timeoutCtx, projectAPIURL, headers, requestBody)
	if err != nil {
		log.With("url", projectAPIURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", projectAPIURL, "response_size_bytes", len(body), "success", true).Debug("创建项目API请求成功")

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

	result.WriteString(fmt.Sprintf("状态: %s (%s)\n", project.State.Name, project.State.Type))

	// 显示创建信息
	createdTime := time.Unix(project.CreatedAt, 0).Format("2006-01-02 15:04:05")
	result.WriteString(fmt.Sprintf("创建时间: %s\n", createdTime))
	result.WriteString(fmt.Sprintf("创建者: %s (%s)\n", project.CreatedBy.DisplayName, project.CreatedBy.Name))

	result.WriteString(fmt.Sprintf("详情链接: %s\n", project.URL))

	return result.String()
}
