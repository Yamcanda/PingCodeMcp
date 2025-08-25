package tools

import (
	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/models"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	sprintTimeout = 15 * time.Second
)

func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewSprintListTool(),
			NewCreateSprintTool(),
		}
	})
}

// 定义了迭代的数据结构
type Sprint struct {
	ID          string  `json:"id"`
	URL         string  `json:"url"`
	Project     Project `json:"project"`
	Assignee    *User   `json:"assignee"`
	Name        string  `json:"name"`
	StartAt     int64   `json:"start_at"`
	EndAt       int64   `json:"end_at"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
	CreatedAt   int64   `json:"created_at"`
	CreatedBy   User    `json:"created_by"`
	UpdatedAt   int64   `json:"updated_at"`
	UpdatedBy   User    `json:"updated_by"`
}

// 定义了获取迭代列表 API 的响应结构
type SprintListResponse struct {
	PageIndex int      `json:"page_index"`
	PageSize  int      `json:"page_size"`
	Total     int      `json:"total"`
	Values    []Sprint `json:"values"`
}

// 定义了创建迭代 API 的请求结构
type CreateSprintRequest struct {
	Name         string   `json:"name"`
	AssigneeId   string   `json:"assignee_id"`
	StartAt      int64    `json:"start_at"`
	EndAt        int64    `json:"end_at"`
	Description  string   `json:"description,omitempty"`
	Status       string   `json:"status,omitempty"`
	Category_ids []string `json:"category_ids,omitempty"`
}

type SprintListTool struct {
	name        string
	description string
}

func NewSprintListTool() MCPTool {
	return &SprintListTool{
		name:        "get_sprint_list",
		description: "获取 PingCode 项目迭代列表 - 查看所有可访问的项目迭代信息，包括迭代名称、迭代唯一标识、状态等",
	}
}

func (t *SprintListTool) GetName() string {
	return t.name
}

func (t *SprintListTool) GetDescription() string {
	return t.description
}

func (t *SprintListTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("project_id", mcp.Description("项目ID - 必填，要查询迭代的项目唯一标识，可通过 get_project_list 工具获取"), mcp.Required()),
		mcp.WithString("name", mcp.Description("迭代名称 - 可选，要查询迭代的名称")),
		mcp.WithString("status", mcp.Description("迭代状态 - 可选，迭代的状态。允许值: pending, in_progress, completed")),
		mcp.WithString("created_between", mcp.Description("迭代开始时间 - 可选，创建时间介于的时间范围，通过','分割起始时间")),
		mcp.WithString("updated_between", mcp.Description("迭代更新时间 - 可选，更新时间介于的时间范围，通过','分割起始时间")),
		mcp.WithNumber("page_index", mcp.Description("页码 - 可选，从0开始，默认为为0时，表示第一页")),
		mcp.WithNumber("page_size", mcp.Description("每页数量 - 可选，默认30，最大100")),
	)
}

func (t *SprintListTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "get_sprint_list")
	requestLogger.Info("开始获取迭代列表")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	projectID, params, err := t.parseArguments(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doRequest(ctx, token, projectID, params, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("迭代列表API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp SprintListResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := t.formatResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *SprintListTool) parseArguments(args interface{}) (string, map[string]string, error) {
	params := map[string]string{
		"page_size": "100",
	}

	if args == nil {
		return "", nil, fmt.Errorf("missing required arguments")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("invalid arguments format")
	}

	// 解析字符串参数
	stringParams := []string{
		"name", "status", "created_between", "updated_between",
	}

	for _, param := range stringParams {
		if value, exists := argsMap[param]; exists {
			if str, ok := value.(string); ok && str != "" {
				params[param] = str
			}
		}
	}

	var projectID string
	if pid, exists := argsMap["project_id"]; exists {
		if pidStr, ok := pid.(string); ok && pidStr != "" {
			match, _ := regexp.MatchString("^[a-z0-9]+$", pidStr)
			if !match {
				return "", nil, fmt.Errorf("invalid project_id: must be a combination of lowercase letters and numbers")
			}
			projectID = pidStr
		} else {
			return "", nil, fmt.Errorf("project_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("project_id is required")
	}

	if page, exists := argsMap["page_index"]; exists {
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
			pageSizeInt := int(pageSizeFloat)
			if pageSizeInt > 0 {
				params["page_size"] = fmt.Sprintf("%d", pageSizeInt)
			}
		}
	}

	return projectID, params, nil
}

func (t *SprintListTool) doRequest(ctx context.Context, token string, projectID string, params map[string]string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, sprintTimeout)
	defer cancel()

	apiURL := fmt.Sprintf("%s/v1/project/projects/%s/sprints", GetBaseUrl(), projectID)

	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, params)
	if err != nil {
		return "", err
	}
	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("迭代列表API请求成功")

	return string(body), nil
}

func (t *SprintListTool) formatResponse(resp SprintListResponse) string {
	var result strings.Builder

	currentPage := resp.PageIndex + 1
	totalPages := (resp.Total + resp.PageSize - 1) / resp.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	result.WriteString(fmt.Sprintf("迭代列表 (第 %d 页，共 %d 页，总计 %d 个迭代):\n\n", currentPage, totalPages, resp.Total))

	if len(resp.Values) == 0 {
		result.WriteString("暂无迭代数据")
		return result.String()
	}

	for i, sprint := range resp.Values {
		result.WriteString(fmt.Sprintf("%d. %s (迭代ID[sprint_id]：%s)\n", i+1, sprint.Name, sprint.ID))
		result.WriteString(fmt.Sprintf("   状态: %s\n", sprint.Status))

		startAt := time.Unix(sprint.StartAt/1000, 0).Format("2006-01-02")
		endAt := time.Unix(sprint.EndAt/1000, 0).Format("2006-01-02")
		result.WriteString(fmt.Sprintf("   周期: %s ~ %s\n", startAt, endAt))

		if sprint.Assignee != nil {
			result.WriteString(fmt.Sprintf("   负责人: %s (%s)\n", sprint.Assignee.DisplayName, sprint.Assignee.Name))
		}

		result.WriteString(fmt.Sprintf("   所属项目: %s (项目ID[project_id] %s 项目标识 %s)\n", sprint.Project.Name, sprint.Project.ID, sprint.Project.Identifier))
		result.WriteString(fmt.Sprintf("   详情链接: %s\n\n", sprint.URL))
	}

	return result.String()
}

// 创建迭代工具结构体
type CreateSprintTool struct {
	name        string
	description string
}

// 创建迭代工具实例
func NewCreateSprintTool() MCPTool {
	return &CreateSprintTool{
		name:        "create_sprint",
		description: "创建新的迭代 - 在指定项目中创建一个包含目标和时间范围的新迭代",
	}
}

func (t *CreateSprintTool) GetName() string {
	return t.name
}

func (t *CreateSprintTool) GetDescription() string {
	return t.description
}

func (t *CreateSprintTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("project_id", mcp.Description("项目ID - 必填，迭代所属的项目"), mcp.Required()),
		mcp.WithString("name", mcp.Description("迭代名称 - 必填"), mcp.Required()),
		mcp.WithString("start_at", mcp.Description("开始日期 - 必填，格式 YYYY-MM-DD"), mcp.Required()),
		mcp.WithString("end_at", mcp.Description("结束日期 - 必填，格式 YYYY-MM-DD"), mcp.Required()),
		mcp.WithString("description", mcp.Description("迭代描述 - 可选")),
		mcp.WithString("status", mcp.Description("迭代状态 - 可选")),
		mcp.WithString("category_ids", mcp.Description("迭代类别id数组 - 可选")),
	)
}

func (t *CreateSprintTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", t.name)
	requestLogger.Info("开始创建迭代")

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	projectID, createReq, err := t.parseArguments(request.GetArguments())
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doRequest(ctx, token, projectID, createReq, requestLogger)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("创建迭代API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp Sprint
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to parse response: %v. Response string: %s", err, respStr)), nil
	}

	result := t.formatResponse(resp)
	return mcp.NewToolResultText(result), nil
}

func (t *CreateSprintTool) parseArguments(args interface{}) (string, *CreateSprintRequest, error) {
	if args == nil {
		return "", nil, fmt.Errorf("missing required arguments")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("invalid arguments format")
	}

	req := &CreateSprintRequest{}

	var projectID string
	if pid, exists := argsMap["project_id"]; exists {
		if pidStr, ok := pid.(string); ok && pidStr != "" {
			match, _ := regexp.MatchString("^[a-z0-9]+$", pidStr)
			if !match {
				return "", nil, fmt.Errorf("invalid project_id: must be a combination of lowercase letters and numbers")
			}
			projectID = pidStr
		} else {
			return "", nil, fmt.Errorf("project_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("project_id is required")
	}

	var startDateStr, endDateStr string

	if pid, exists := argsMap["assignee_id"]; exists {
		if pidStr, ok := pid.(string); ok && pidStr != "" {
			req.AssigneeId = pidStr
		} else {
			return "", nil, fmt.Errorf("assignee_id is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("assignee_id is required")
	}

	if name, exists := argsMap["name"]; exists {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			req.Name = nameStr
		} else {
			return "", nil, fmt.Errorf("name is required and must be a non-empty string")
		}
	} else {
		return "", nil, fmt.Errorf("name is required")
	}

	if sDate, exists := argsMap["start_at"]; exists {
		if dateStr, ok := sDate.(string); ok && dateStr != "" {
			startDateStr = dateStr
		} else {
			return "", nil, fmt.Errorf("start_at is required and must be in YYYY-MM-DD format")
		}
	} else {
		return "", nil, fmt.Errorf("start_at is required")
	}

	if eDate, exists := argsMap["end_at"]; exists {
		if dateStr, ok := eDate.(string); ok && dateStr != "" {
			endDateStr = dateStr
		} else {
			return "", nil, fmt.Errorf("end_at is required and must be in YYYY-MM-DD format")
		}
	} else {
		return "", nil, fmt.Errorf("end_at is required")
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
	sTime, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
	if err != nil {
		return "", nil, fmt.Errorf("invalid start_at format: %s. please use YYYY-MM-DD", startDateStr)
	}
	req.StartAt = sTime.Unix() * 1000

	eTime, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err != nil {
		return "", nil, fmt.Errorf("invalid end_at format: %s. please use YYYY-MM-DD", endDateStr)
	}
	req.EndAt = eTime.Unix() * 1000

	return projectID, req, nil
}

func (t *CreateSprintTool) doRequest(ctx context.Context, token string, projectID string, req *CreateSprintRequest, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Content-Type":  "application/json",
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, sprintTimeout)
	defer cancel()

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	apiURL := fmt.Sprintf("%s/v1/project/projects/%s/sprints", GetBaseUrl(), projectID)

	body, _, err := utils.DoPostJSON(timeoutCtx, apiURL, headers, reqBody)
	if err != nil {
		return "", err
	}
	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("创建迭代API请求成功")

	return string(body), nil
}

func (t *CreateSprintTool) formatResponse(sprint Sprint) string {
	var result strings.Builder

	result.WriteString("迭代创建成功:\n\n")
	result.WriteString(fmt.Sprintf("迭代名称: %s\n", sprint.Name))
	result.WriteString(fmt.Sprintf("迭代ID: %s\n", sprint.ID))
	result.WriteString(fmt.Sprintf("状态: %s\n", sprint.Status))

	startAt := time.Unix(sprint.StartAt/1000, 0).Format("2006-01-02")
	endAt := time.Unix(sprint.EndAt/1000, 0).Format("2006-01-02")
	result.WriteString(fmt.Sprintf("周期: %s ~ %s\n", startAt, endAt))

	result.WriteString(fmt.Sprintf("所属项目: %s (%s)\n", sprint.Project.Name, sprint.Project.Identifier))

	createdTime := time.Unix(sprint.CreatedAt/1000, 0).Format("2006-01-02 15:04:05")
	result.WriteString(fmt.Sprintf("创建时间: %s by %s\n", createdTime, sprint.CreatedBy.DisplayName))

	result.WriteString(fmt.Sprintf("详情链接: %s\n", sprint.URL))

	return result.String()
}
