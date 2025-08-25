package tools

import (
	"PingCodeMcp/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	workloadTimeout = 15 * time.Second
)

func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewCreateWorkloadTool(),
			NewGetWorkloadTypesTool(),
			NewDeleteWorkloadTool(),
			NewListWorkloadsTool(),
			NewUpdateWorkloadTool(),
		}
	})
}

// CreateWorkloadRequest 创建工时的请求结构
type CreateWorkloadRequest struct {
	// 工时主体的id
	PrincipalID string `json:"principal_id"`
	// 工时主体的类型。允许值: work_item
	PrincipalType string `json:"principal_type"`
	// 工时类型的id
	TypeID string `json:"type_id"`
	// 工时的时长。单位是小时，数值可以是为0-24之间，最多包含一位小数的正数
	Duration float64 `json:"duration"`
	// 工时的登记日期。该值为十位数字组成的时间戳，会被转换为该时间当天的零点零分零秒
	ReportAt int64 `json:"report_at"`
	// 工时的登记人，企业鉴权时必填。个人鉴权时不需要传递，即使传递了也会被忽略
	ReportByID *string `json:"report_by_id,omitempty"`
	// 工时的说明
	Description *string `json:"description,omitempty"`
}

// CreateWorkloadResponse 创建工时的响应结构
type CreateWorkloadResponse struct {
	ID          string  `json:"id"`
	Duration    float64 `json:"duration"`
	ReportAt    int64   `json:"report_at"`
	Description string  `json:"description"`
	CreatedAt   int64   `json:"created_at"`
	UpdatedAt   int64   `json:"updated_at"`

	Principal struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		URL      string `json:"url"`
		WorkItem *struct {
			ID         string `json:"id"`
			URL        string `json:"url"`
			Identifier string `json:"identifier"`
			Title      string `json:"title"`
			Type       string `json:"type"`
		} `json:"work_item,omitempty"`
	} `json:"principal"`

	Type struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"type"`

	ReportBy struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
		Avatar      string `json:"avatar"`
	} `json:"report_by"`

	CreatedBy struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
		Avatar      string `json:"avatar"`
	} `json:"created_by"`

	UpdatedBy struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
		URL         string `json:"url"`
		Avatar      string `json:"avatar"`
	} `json:"updated_by"`
}

// CreateWorkloadTool 创建工时工具
type CreateWorkloadTool struct {
	name        string
	description string
}

// NewCreateWorkloadTool 创建新的工时创建工具实例
func NewCreateWorkloadTool() MCPTool {
	return &CreateWorkloadTool{
		name:        "create_workload",
		description: "Create a workload",
	}
}

func (t *CreateWorkloadTool) GetName() string {
	return t.name
}

func (t *CreateWorkloadTool) GetDescription() string {
	return t.description
}

func (t *CreateWorkloadTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("创建工时记录 - 为工作项记录工作时长，支持设置工时类型、时长、登记日期等"),
		mcp.WithString("principal_id", mcp.Description("工时主体ID - 必填，通常是工作项ID，可通过list_work_items工具获取"), mcp.Required()),
		mcp.WithString("principal_type", mcp.Description("工时主体类型 - 必填，目前仅支持：work_item"), mcp.Required()),
		mcp.WithString("type_id", mcp.Description("工时类型ID - 必填，工时分类标识，可通过get_workload_types工具获取"), mcp.Required()),
		mcp.WithNumber("duration", mcp.Description("工时时长 - 必填，单位为小时，范围0-24，最多一位小数，如：8.5"), mcp.Required()),
		mcp.WithNumber("report_at", mcp.Description("登记日期 - 可选，Unix时间戳，默认为当天零点")),
		mcp.WithString("report_by_id", mcp.Description("登记人ID - 可选，企业鉴权时必填，个人鉴权时忽略")),
		mcp.WithString("description", mcp.Description("工时说明 - 可选，工时记录的详细描述")),
	)
}

func (t *CreateWorkloadTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "create_workload")
	requestLogger.Info("开始处理创建工时请求")

	// 从上下文中获取认证token
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.Error("获取认证token失败", "error", err)
		return mcp.NewToolResultError("认证失败: " + err.Error()), nil
	}

	// 解析请求参数
	workloadRequest, err := t.parseCreateWorkloadArguments(request.Params.Arguments)
	if err != nil {
		requestLogger.Error("解析请求参数失败", "error", err)
		return mcp.NewToolResultError("参数解析失败: " + err.Error()), nil
	}

	requestLogger.Info("解析请求参数成功",
		"principal_id", workloadRequest.PrincipalID,
		"principal_type", workloadRequest.PrincipalType,
		"type_id", workloadRequest.TypeID,
		"duration", workloadRequest.Duration,
		"report_at", workloadRequest.ReportAt)

	// 执行创建工时请求
	result, err := t.doCreateWorkloadRequest(ctx, token, workloadRequest, requestLogger)
	if err != nil {
		requestLogger.Error("创建工时请求失败", "error", err)
		return mcp.NewToolResultError("创建工时失败: " + err.Error()), nil
	}

	requestLogger.Info("创建工时成功")
	return mcp.NewToolResultText(result), nil
}

func (t *CreateWorkloadTool) parseCreateWorkloadArguments(args interface{}) (*CreateWorkloadRequest, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("参数格式错误")
	}

	request := &CreateWorkloadRequest{}

	// 解析必填参数
	if principalID, exists := argsMap["principal_id"]; exists {
		if str, ok := principalID.(string); ok {
			request.PrincipalID = str
		} else {
			return nil, fmt.Errorf("principal_id 必须是字符串")
		}
	} else {
		return nil, fmt.Errorf("缺少必填参数 principal_id")
	}

	if principalType, exists := argsMap["principal_type"]; exists {
		if str, ok := principalType.(string); ok {
			request.PrincipalType = str
		} else {
			return nil, fmt.Errorf("principal_type 必须是字符串")
		}
	} else {
		return nil, fmt.Errorf("缺少必填参数 principal_type")
	}

	if typeID, exists := argsMap["type_id"]; exists {
		if str, ok := typeID.(string); ok {
			request.TypeID = str
		} else {
			return nil, fmt.Errorf("type_id 必须是字符串")
		}
	} else {
		return nil, fmt.Errorf("缺少必填参数 type_id")
	}

	if duration, exists := argsMap["duration"]; exists {
		switch v := duration.(type) {
		case float64:
			if v < 0 || v > 24 {
				return nil, fmt.Errorf("duration 必须在0-24之间")
			}
			request.Duration = v
		case int:
			if v < 0 || v > 24 {
				return nil, fmt.Errorf("duration 必须在0-24之间")
			}
			request.Duration = float64(v)
		default:
			return nil, fmt.Errorf("duration 必须是数字")
		}
	} else {
		return nil, fmt.Errorf("缺少必填参数 duration")
	}

	// 解析 report_at 参数，如果为空则默认为当天
	if reportAt, exists := argsMap["report_at"]; exists {
		switch v := reportAt.(type) {
		case float64:
			request.ReportAt = int64(v)
		case int:
			request.ReportAt = int64(v)
		case int64:
			request.ReportAt = v
		default:
			return nil, fmt.Errorf("report_at 必须是时间戳")
		}
	} else {
		// 默认为当天零点的时间戳
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		request.ReportAt = today.Unix()
	}

	// 解析可选参数
	if reportByID, exists := argsMap["report_by_id"]; exists {
		if str, ok := reportByID.(string); ok && str != "" {
			request.ReportByID = &str
		}
	}

	if description, exists := argsMap["description"]; exists {
		if str, ok := description.(string); ok && str != "" {
			request.Description = &str
		}
	}

	return request, nil
}

func (t *CreateWorkloadTool) doCreateWorkloadRequest(ctx context.Context, token string, workloadRequest *CreateWorkloadRequest, log *logger.Logger) (string, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, workloadTimeout)
	defer cancel()

	// 序列化请求体
	requestBody, err := json.Marshal(workloadRequest)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 设置请求头
	headers := map[string]string{
		"Authorization": token,
	}

	// 发送HTTP请求
	apiURL := GetBaseUrl() + "/v1/workloads"
	body, _, err := utils.DoPostJSON(timeoutCtx, apiURL, headers, requestBody)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal(body, &errorRes); err == nil && errorRes.Code != "" {
		log.Error("创建工时失败", "error", err)
		return "", fmt.Errorf("创建工时失败: %s", errorRes.Message)
	}

	// 解析响应
	var workloadResp CreateWorkloadResponse
	if err := json.Unmarshal(body, &workloadResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 格式化响应
	return t.formatCreateWorkloadResponse(workloadResp), nil
}

func (t *CreateWorkloadTool) formatCreateWorkloadResponse(resp CreateWorkloadResponse) string {
	var result strings.Builder

	result.WriteString("✅ 工时创建成功\n\n")
	result.WriteString(fmt.Sprintf("🆔 工时ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("⏱️ 工时时长: %.1f 小时\n", resp.Duration))
	result.WriteString(fmt.Sprintf("📅 登记日期: %s\n", time.Unix(resp.ReportAt, 0).Format("2006-01-02")))

	if resp.Description != "" {
		result.WriteString(fmt.Sprintf("📝 说明: %s\n", resp.Description))
	}

	result.WriteString("\n📋 工时主体信息:\n")
	result.WriteString(fmt.Sprintf("  • 类型: %s\n", resp.Principal.Type))
	if resp.Principal.WorkItem != nil {
		result.WriteString(fmt.Sprintf("  • 工作项: %s (%s)\n", resp.Principal.WorkItem.Title, resp.Principal.WorkItem.Identifier))
		result.WriteString(fmt.Sprintf("  • 工作项链接: %s\n", resp.Principal.WorkItem.URL))
	}

	result.WriteString("\n🏷️ 工时类型:\n")
	result.WriteString(fmt.Sprintf("  • 名称: %s\n", resp.Type.Name))
	result.WriteString(fmt.Sprintf("  • 链接: %s\n", resp.Type.URL))

	result.WriteString("\n👤 登记人信息:\n")
	result.WriteString(fmt.Sprintf("  • 姓名: %s (%s)\n", resp.ReportBy.DisplayName, resp.ReportBy.Name))
	result.WriteString(fmt.Sprintf("  • 链接: %s\n", resp.ReportBy.URL))

	result.WriteString("\n📊 创建信息:\n")
	result.WriteString(fmt.Sprintf("  • 创建时间: %s\n", time.Unix(resp.CreatedAt, 0).Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("  • 创建人: %s (%s)\n", resp.CreatedBy.DisplayName, resp.CreatedBy.Name))

	return result.String()
}

// WorkloadType 工时类型结构
type WorkloadType struct {
	ID   string `json:"id"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

// GetWorkloadTypesResponse 获取工时类型列表的响应结构
type GetWorkloadTypesResponse struct {
	PageIndex int            `json:"page_index"`
	PageSize  int            `json:"page_size"`
	Total     int            `json:"total"`
	Values    []WorkloadType `json:"values"`
}

// GetWorkloadTypesTool 获取工时类型列表工具
type GetWorkloadTypesTool struct {
	name        string
	description string
}

// NewGetWorkloadTypesTool 创建新的获取工时类型列表工具实例
func NewGetWorkloadTypesTool() MCPTool {
	return &GetWorkloadTypesTool{
		name:        "get_workload_types",
		description: "Get a list of workload types",
	}
}

func (t *GetWorkloadTypesTool) GetName() string {
	return t.name
}

func (t *GetWorkloadTypesTool) GetDescription() string {
	return t.description
}

func (t *GetWorkloadTypesTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("获取工时类型列表 - 查看系统中所有可用的工时分类，用于创建工时记录时选择类型"),
	)
}

func (t *GetWorkloadTypesTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "get_workload_types")
	requestLogger.Info("开始处理获取工时类型列表请求")

	// 从上下文中获取认证token
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.Error("获取认证token失败", "error", err)
		return mcp.NewToolResultError("认证失败: " + err.Error()), nil
	}

	// 执行获取工时类型列表请求
	result, err := t.doGetWorkloadTypesRequest(ctx, token, requestLogger)
	if err != nil {
		requestLogger.Error("获取工时类型列表请求失败", "error", err)
		return mcp.NewToolResultError("获取工时类型列表失败: " + err.Error()), nil
	}

	requestLogger.Info("获取工时类型列表成功")
	return mcp.NewToolResultText(result), nil
}

func (t *GetWorkloadTypesTool) doGetWorkloadTypesRequest(ctx context.Context, token string, log *logger.Logger) (string, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, workloadTimeout)
	defer cancel()

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 设置请求头
	headers := map[string]string{
		"Authorization": token,
	}

	// 发送HTTP请求
	apiURL := GetBaseUrl() + "/v1/workload_types"
	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, nil)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal(body, &errorRes); err == nil && errorRes.Code != "" {
		log.Error("获取工时类型失败", "error", err)
		return "", fmt.Errorf("获取工时类型失败: %s", errorRes.Message)
	}

	// 解析响应
	var workloadTypesResp GetWorkloadTypesResponse
	if err := json.Unmarshal(body, &workloadTypesResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 格式化响应
	return t.formatGetWorkloadTypesResponse(workloadTypesResp), nil
}

func (t *GetWorkloadTypesTool) formatGetWorkloadTypesResponse(resp GetWorkloadTypesResponse) string {
	var result strings.Builder

	result.WriteString("📋 工时类型列表\n\n")
	result.WriteString(fmt.Sprintf("📊 总计: %d 个工时类型\n", resp.Total))
	result.WriteString(fmt.Sprintf("📄 当前页: %d, 每页: %d\n\n", resp.PageIndex+1, resp.PageSize))

	if len(resp.Values) == 0 {
		result.WriteString("暂无工时类型数据\n")
		return result.String()
	}

	result.WriteString("🏷️ 工时类型详情:\n")
	for i, workloadType := range resp.Values {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, workloadType.Name))
		result.WriteString(fmt.Sprintf("   • ID: %s\n", workloadType.ID))
		result.WriteString(fmt.Sprintf("   • 链接: %s\n", workloadType.URL))
		if i < len(resp.Values)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// DeleteWorkloadTool 删除工时工具
type DeleteWorkloadTool struct {
	name        string
	description string
}

// NewDeleteWorkloadTool 创建新的删除工时工具实例
func NewDeleteWorkloadTool() MCPTool {
	return &DeleteWorkloadTool{
		name:        "delete_workload",
		description: "Delete a workload",
	}
}

func (t *DeleteWorkloadTool) GetName() string {
	return t.name
}

func (t *DeleteWorkloadTool) GetDescription() string {
	return t.description
}

func (t *DeleteWorkloadTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("删除工时记录 - 删除指定的工时记录，操作不可逆"),
		mcp.WithString("workload_id", mcp.Description("工时记录ID - 必填，要删除的工时记录唯一标识，可通过list_workloads工具获取"), mcp.Required()),
	)
}

func (t *DeleteWorkloadTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "delete_workload")
	requestLogger.Info("开始处理删除工时请求")

	// 从上下文中获取认证token
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.Error("获取认证token失败", "error", err)
		return mcp.NewToolResultError("认证失败: " + err.Error()), nil
	}

	// 解析请求参数
	workloadID, err := t.parseDeleteWorkloadArguments(request.Params.Arguments)
	if err != nil {
		requestLogger.Error("解析请求参数失败", "error", err)
		return mcp.NewToolResultError("参数解析失败: " + err.Error()), nil
	}

	requestLogger.Info("解析请求参数成功", "workload_id", workloadID)

	// 执行删除工时请求
	result, err := t.doDeleteWorkloadRequest(ctx, token, workloadID, requestLogger)
	if err != nil {
		requestLogger.Error("删除工时请求失败", "error", err)
		return mcp.NewToolResultError("删除工时失败: " + err.Error()), nil
	}

	requestLogger.Info("删除工时成功")
	return mcp.NewToolResultText(result), nil
}

func (t *DeleteWorkloadTool) parseDeleteWorkloadArguments(args interface{}) (string, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("参数格式错误")
	}

	// 解析必填参数
	if workloadID, exists := argsMap["workload_id"]; exists {
		if str, ok := workloadID.(string); ok && str != "" {
			return str, nil
		} else {
			return "", fmt.Errorf("workload_id 必须是非空字符串")
		}
	} else {
		return "", fmt.Errorf("缺少必填参数 workload_id")
	}
}

func (t *DeleteWorkloadTool) doDeleteWorkloadRequest(ctx context.Context, token string, workloadID string, log *logger.Logger) (string, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, workloadTimeout)
	defer cancel()

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 设置请求头
	headers := map[string]string{
		"Authorization": token,
	}

	// 构建删除工时的URL
	deleteURL := fmt.Sprintf("%s/v1/workloads/%s", GetBaseUrl(), workloadID)

	// 发送HTTP DELETE请求
	_, resp, err := utils.DoDelete(timeoutCtx, deleteURL, headers, nil)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}

	// 检查状态码
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("删除工时失败，状态码: %d", resp.StatusCode)
	}

	// 格式化响应
	return t.formatDeleteWorkloadResponse(workloadID), nil
}

func (t *DeleteWorkloadTool) formatDeleteWorkloadResponse(workloadID string) string {
	var result strings.Builder

	result.WriteString("✅ 工时删除成功\n\n")
	result.WriteString(fmt.Sprintf("🆔 已删除工时ID: %s\n", workloadID))
	result.WriteString(fmt.Sprintf("⏰ 删除时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	result.WriteString("\n🗑️ 该工时记录已被永久删除")

	return result.String()
}

// ListWorkloadsRequest 获取工时列表的请求参数
type ListWorkloadsRequest struct {
	// 工时主体的类型。允许值: work_item
	PrincipalType string `json:"principal_type"`
	// 工时主体的id
	PrincipalID string `json:"principal_id"`
	// 登记日期查询的起始时间
	StartAt *int64 `json:"start_at,omitempty"`
	// 登记日期查询的结束时间
	EndAt *int64 `json:"end_at,omitempty"`
	// 登记人的id
	ReportByID *string `json:"report_by_id,omitempty"`
}

// ListWorkloadsResponse 获取工时列表的响应结构
type ListWorkloadsResponse struct {
	Values []CreateWorkloadResponse `json:"values"`
}

// ListWorkloadsTool 获取工时列表工具
type ListWorkloadsTool struct {
	name        string
	description string
}

// NewListWorkloadsTool 创建新的获取工时列表工具实例
func NewListWorkloadsTool() MCPTool {
	return &ListWorkloadsTool{
		name:        "list_workloads",
		description: "Get a list of workloads",
	}
}

func (t *ListWorkloadsTool) GetName() string {
	return t.name
}

func (t *ListWorkloadsTool) GetDescription() string {
	return t.description
}

func (t *ListWorkloadsTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("获取工时记录列表 - 查询指定工作项的工时记录，支持按时间范围和登记人筛选"),
		mcp.WithString("principal_type", mcp.Description("工时主体类型 - 可选，目前仅支持：work_item，默认为work_item")),
		mcp.WithString("principal_id", mcp.Description("工时主体ID - 必填，通常是工作项ID，可通过list_work_items工具获取"), mcp.Required()),
		mcp.WithNumber("start_at", mcp.Description("起始时间 - 可选，Unix时间戳，查询此时间之后的工时记录")),
		mcp.WithNumber("end_at", mcp.Description("结束时间 - 可选，Unix时间戳，查询此时间之前的工时记录")),
		mcp.WithString("report_by_id", mcp.Description("登记人ID - 可选，筛选指定用户登记的工时记录")),
	)
}

func (t *ListWorkloadsTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "list_workloads")
	requestLogger.Info("开始处理获取工时列表请求")

	// 从上下文中获取认证token
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.Error("获取认证token失败", "error", err)
		return mcp.NewToolResultError("认证失败: " + err.Error()), nil
	}

	// 解析请求参数
	listRequest, err := t.parseListWorkloadsArguments(request.Params.Arguments)
	if err != nil {
		requestLogger.Error("解析请求参数失败", "error", err)
		return mcp.NewToolResultError("参数解析失败: " + err.Error()), nil
	}

	requestLogger.Info("解析请求参数成功",
		"principal_type", listRequest.PrincipalType,
		"principal_id", listRequest.PrincipalID)

	// 执行获取工时列表请求
	result, err := t.doListWorkloadsRequest(ctx, token, listRequest, requestLogger)
	if err != nil {
		requestLogger.Error("获取工时列表请求失败", "error", err)
		return mcp.NewToolResultError("获取工时列表失败: " + err.Error()), nil
	}

	requestLogger.Info("获取工时列表成功")
	return mcp.NewToolResultText(result), nil
}

func (t *ListWorkloadsTool) parseListWorkloadsArguments(args interface{}) (*ListWorkloadsRequest, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("参数格式错误")
	}

	request := &ListWorkloadsRequest{}

	// 解析 principal_type 参数，如果为空则默认为 "work_item"
	if principalType, exists := argsMap["principal_type"]; exists {
		if str, ok := principalType.(string); ok && str != "" {
			request.PrincipalType = str
		} else {
			request.PrincipalType = "work_item"
		}
	} else {
		request.PrincipalType = "work_item"
	}

	if principalID, exists := argsMap["principal_id"]; exists {
		if str, ok := principalID.(string); ok {
			request.PrincipalID = str
		} else {
			return nil, fmt.Errorf("principal_id 必须是字符串")
		}
	} else {
		return nil, fmt.Errorf("缺少必填参数 principal_id")
	}

	// 解析可选参数
	if startAt, exists := argsMap["start_at"]; exists {
		switch v := startAt.(type) {
		case float64:
			val := int64(v)
			request.StartAt = &val
		case int:
			val := int64(v)
			request.StartAt = &val
		case int64:
			request.StartAt = &v
		default:
			return nil, fmt.Errorf("start_at 必须是时间戳")
		}
	}

	if endAt, exists := argsMap["end_at"]; exists {
		switch v := endAt.(type) {
		case float64:
			val := int64(v)
			request.EndAt = &val
		case int:
			val := int64(v)
			request.EndAt = &val
		case int64:
			request.EndAt = &v
		default:
			return nil, fmt.Errorf("end_at 必须是时间戳")
		}
	}

	if reportByID, exists := argsMap["report_by_id"]; exists {
		if str, ok := reportByID.(string); ok && str != "" {
			request.ReportByID = &str
		}
	}

	return request, nil
}

func (t *ListWorkloadsTool) doListWorkloadsRequest(ctx context.Context, token string, listRequest *ListWorkloadsRequest, log *logger.Logger) (string, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, workloadTimeout)
	defer cancel()

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 设置请求头
	headers := map[string]string{
		"Authorization": token,
	}

	// 构建查询参数
	queryParams := map[string]string{
		"principal_type": listRequest.PrincipalType,
		"principal_id":   listRequest.PrincipalID,
	}

	if listRequest.StartAt != nil {
		queryParams["start_at"] = fmt.Sprintf("%d", *listRequest.StartAt)
	}

	if listRequest.EndAt != nil {
		queryParams["end_at"] = fmt.Sprintf("%d", *listRequest.EndAt)
	}

	if listRequest.ReportByID != nil {
		queryParams["report_by_id"] = *listRequest.ReportByID
	}

	// 发送HTTP请求
	apiURL := GetBaseUrl() + "/v1/workloads"
	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, queryParams)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal(body, &errorRes); err == nil && errorRes.Code != "" {
		log.Error("获取工时列表失败", "error", err)
		return "", fmt.Errorf("获取工时列表失败: %s", errorRes.Message)
	}

	// 解析响应
	var workloadsResp ListWorkloadsResponse
	if err := json.Unmarshal(body, &workloadsResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 格式化响应
	return t.formatListWorkloadsResponse(workloadsResp), nil
}

func (t *ListWorkloadsTool) formatListWorkloadsResponse(resp ListWorkloadsResponse) string {
	var result strings.Builder

	result.WriteString("📋 工时列表\n\n")
	result.WriteString(fmt.Sprintf("📊 总计: %d 条工时记录\n\n", len(resp.Values)))

	if len(resp.Values) == 0 {
		result.WriteString("暂无工时记录\n")
		return result.String()
	}

	result.WriteString("⏱️ 工时记录详情:\n")
	for i, workload := range resp.Values {
		result.WriteString(fmt.Sprintf("%d. 工时ID: %s\n", i+1, workload.ID))
		result.WriteString(fmt.Sprintf("   • 时长: %.1f 小时\n", workload.Duration))
		result.WriteString(fmt.Sprintf("   • 登记日期: %s\n", time.Unix(workload.ReportAt, 0).Format("2006-01-02")))

		if workload.Description != "" {
			result.WriteString(fmt.Sprintf("   • 说明: %s\n", workload.Description))
		}

		if workload.Principal.WorkItem != nil {
			result.WriteString(fmt.Sprintf("   • 工作项: %s (%s)\n", workload.Principal.WorkItem.Title, workload.Principal.WorkItem.Identifier))
		}

		result.WriteString(fmt.Sprintf("   • 工时类型: %s\n", workload.Type.Name))
		result.WriteString(fmt.Sprintf("   • 登记人: %s (%s)\n", workload.ReportBy.DisplayName, workload.ReportBy.Name))
		result.WriteString(fmt.Sprintf("   • 创建时间: %s\n", time.Unix(workload.CreatedAt, 0).Format("2006-01-02 15:04:05")))

		if i < len(resp.Values)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// UpdateWorkloadRequest 部分更新工时的请求结构
type UpdateWorkloadRequest struct {
	// 工时类型的id
	TypeID *string `json:"type_id,omitempty"`
	// 工时的时长。单位是小时，数值可以是为0-24之间，最多包含一位小数的正数
	Duration *float64 `json:"duration,omitempty"`
	// 工时的登记日期。该值为十位数字组成的时间戳，会被转换为该时间当天的零点零分零秒
	ReportAt *int64 `json:"report_at,omitempty"`
	// 工时的登记人，企业鉴权时必填。个人鉴权时不需要传递，即使传递了也会被忽略
	ReportByID *string `json:"report_by_id,omitempty"`
	// 工时的说明
	Description *string `json:"description,omitempty"`
}

// UpdateWorkloadTool 部分更新工时工具
type UpdateWorkloadTool struct {
	name        string
	description string
}

// NewUpdateWorkloadTool 创建新的更新工时工具实例
func NewUpdateWorkloadTool() MCPTool {
	return &UpdateWorkloadTool{
		name:        "update_workload",
		description: "Partially update one workload",
	}
}

func (t *UpdateWorkloadTool) GetName() string {
	return t.name
}

func (t *UpdateWorkloadTool) GetDescription() string {
	return t.description
}

func (t *UpdateWorkloadTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("更新工时记录 - 修改已存在的工时记录信息，包括时长、类型、日期等"),
		mcp.WithString("workload_id", mcp.Description("工时记录ID - 必填，要更新的工时记录唯一标识，可通过list_workloads工具获取"), mcp.Required()),
		mcp.WithString("type_id", mcp.Description("工时类型ID - 可选，更改工时分类，可通过get_workload_types工具获取")),
		mcp.WithNumber("duration", mcp.Description("工时时长 - 可选，单位为小时，范围0-24，最多一位小数，如：8.5")),
		mcp.WithNumber("report_at", mcp.Description("登记日期 - 可选，Unix时间戳，更改工时登记日期")),
		mcp.WithString("report_by_id", mcp.Description("登记人ID - 可选，企业鉴权时可用，个人鉴权时忽略")),
		mcp.WithString("description", mcp.Description("工时说明 - 可选，更新工时记录的详细描述")),
	)
}

func (t *UpdateWorkloadTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log := logger.New()
	defer log.Sync()

	requestLogger := log.With("tool", "update_workload")
	requestLogger.Info("开始处理更新工时请求")

	// 从上下文中获取认证token
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.Error("获取认证token失败", "error", err)
		return mcp.NewToolResultError("认证失败: " + err.Error()), nil
	}

	// 解析请求参数
	workloadID, updateRequest, err := t.parseUpdateWorkloadArguments(request.Params.Arguments)
	if err != nil {
		requestLogger.Error("解析请求参数失败", "error", err)
		return mcp.NewToolResultError("参数解析失败: " + err.Error()), nil
	}

	requestLogger.Info("解析请求参数成功", "workload_id", workloadID)

	// 执行更新工时请求
	result, err := t.doUpdateWorkloadRequest(ctx, token, workloadID, updateRequest, requestLogger)
	if err != nil {
		requestLogger.Error("更新工时请求失败", "error", err)
		return mcp.NewToolResultError("更新工时失败: " + err.Error()), nil
	}

	requestLogger.Info("更新工时成功")
	return mcp.NewToolResultText(result), nil
}

func (t *UpdateWorkloadTool) parseUpdateWorkloadArguments(args interface{}) (string, *UpdateWorkloadRequest, error) {
	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("参数格式错误")
	}

	// 解析必填参数 workload_id
	var workloadID string
	if id, exists := argsMap["workload_id"]; exists {
		if str, ok := id.(string); ok && str != "" {
			workloadID = str
		} else {
			return "", nil, fmt.Errorf("workload_id 必须是非空字符串")
		}
	} else {
		return "", nil, fmt.Errorf("缺少必填参数 workload_id")
	}

	request := &UpdateWorkloadRequest{}

	// 解析可选参数
	if typeID, exists := argsMap["type_id"]; exists {
		if str, ok := typeID.(string); ok && str != "" {
			request.TypeID = &str
		}
	}

	if duration, exists := argsMap["duration"]; exists {
		switch v := duration.(type) {
		case float64:
			if v < 0 || v > 24 {
				return "", nil, fmt.Errorf("duration 必须在0-24之间")
			}
			request.Duration = &v
		case int:
			if v < 0 || v > 24 {
				return "", nil, fmt.Errorf("duration 必须在0-24之间")
			}
			val := float64(v)
			request.Duration = &val
		default:
			return "", nil, fmt.Errorf("duration 必须是数字")
		}
	}

	// 解析 report_at 参数，如果提供了值则使用，否则不更新（保持原值）
	if reportAt, exists := argsMap["report_at"]; exists {
		switch v := reportAt.(type) {
		case float64:
			val := int64(v)
			request.ReportAt = &val
		case int:
			val := int64(v)
			request.ReportAt = &val
		case int64:
			request.ReportAt = &v
		case nil:
			// 如果明确传入 null，则设置为当天
			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			val := today.Unix()
			request.ReportAt = &val
		default:
			return "", nil, fmt.Errorf("report_at 必须是时间戳")
		}
	}

	if reportByID, exists := argsMap["report_by_id"]; exists {
		if str, ok := reportByID.(string); ok && str != "" {
			request.ReportByID = &str
		}
	}

	if description, exists := argsMap["description"]; exists {
		if str, ok := description.(string); ok {
			request.Description = &str
		}
	}

	return workloadID, request, nil
}

func (t *UpdateWorkloadTool) doUpdateWorkloadRequest(ctx context.Context, token string, workloadID string, updateRequest *UpdateWorkloadRequest, log *logger.Logger) (string, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, workloadTimeout)
	defer cancel()

	// 序列化请求体
	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 设置请求头
	headers := map[string]string{
		"Authorization": token,
	}

	// 构建更新工时的URL
	updateURL := fmt.Sprintf("%s/v1/workloads/%s", GetBaseUrl(), workloadID)

	// 发送HTTP PATCH请求
	body, _, err := utils.DoPatch(timeoutCtx, updateURL, headers, requestBody)
	if err != nil {
		return "", fmt.Errorf("HTTP请求失败: %w", err)
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal(body, &errorRes); err == nil && errorRes.Code != "" {
		log.Error("更新工时失败", "error", err)
		return "", fmt.Errorf("更新工时失败: %s", errorRes.Message)
	}

	// 解析响应
	var workloadResp CreateWorkloadResponse
	if err := json.Unmarshal(body, &workloadResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	// 格式化响应
	return t.formatUpdateWorkloadResponse(workloadResp), nil
}

func (t *UpdateWorkloadTool) formatUpdateWorkloadResponse(resp CreateWorkloadResponse) string {
	var result strings.Builder

	result.WriteString("✅ 工时更新成功\n\n")
	result.WriteString(fmt.Sprintf("🆔 工时ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("⏱️ 工时时长: %.1f 小时\n", resp.Duration))
	result.WriteString(fmt.Sprintf("📅 登记日期: %s\n", time.Unix(resp.ReportAt, 0).Format("2006-01-02")))

	if resp.Description != "" {
		result.WriteString(fmt.Sprintf("📝 说明: %s\n", resp.Description))
	}

	result.WriteString("\n📋 工时主体信息:\n")
	result.WriteString(fmt.Sprintf("  • 类型: %s\n", resp.Principal.Type))
	if resp.Principal.WorkItem != nil {
		result.WriteString(fmt.Sprintf("  • 工作项: %s (%s)\n", resp.Principal.WorkItem.Title, resp.Principal.WorkItem.Identifier))
		result.WriteString(fmt.Sprintf("  • 工作项链接: %s\n", resp.Principal.WorkItem.URL))
	}

	result.WriteString("\n🏷️ 工时类型:\n")
	result.WriteString(fmt.Sprintf("  • 名称: %s\n", resp.Type.Name))
	result.WriteString(fmt.Sprintf("  • 链接: %s\n", resp.Type.URL))

	result.WriteString("\n👤 登记人信息:\n")
	result.WriteString(fmt.Sprintf("  • 姓名: %s (%s)\n", resp.ReportBy.DisplayName, resp.ReportBy.Name))
	result.WriteString(fmt.Sprintf("  • 链接: %s\n", resp.ReportBy.URL))

	result.WriteString("\n📊 更新信息:\n")
	result.WriteString(fmt.Sprintf("  • 更新时间: %s\n", time.Unix(resp.UpdatedAt, 0).Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("  • 更新人: %s (%s)\n", resp.UpdatedBy.DisplayName, resp.UpdatedBy.Name))
	result.WriteString(fmt.Sprintf("  • 创建时间: %s\n", time.Unix(resp.CreatedAt, 0).Format("2006-01-02 15:04:05")))
	result.WriteString(fmt.Sprintf("  • 创建人: %s (%s)\n", resp.CreatedBy.DisplayName, resp.CreatedBy.Name))

	return result.String()
}
