package tools

import (
	"PingCodeMcp/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	// API请求超时时间
	enterpriseUsersTimeout = 15 * time.Second
)

// 注册工具
func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewEnterpriseUsersTool(),
			NewCreateEnterpriseUserTool(),
		}
	})
}

// EnterpriseUser 企业成员信息结构体
type EnterpriseUser struct {
	ID             string `json:"id"`
	URL            string `json:"url"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Avatar         string `json:"avatar"`
	Email          string `json:"email"`
	Mobile         string `json:"mobile"`
	Status         string `json:"status"`
	Department     string `json:"department"`
	Job            string `json:"job"`
	EmployeeNumber string `json:"employee_number"`
}

// EnterpriseUsersResponse 企业成员列表响应结构体
type EnterpriseUsersResponse struct {
	PageIndex int              `json:"page_index"`
	PageSize  int              `json:"page_size"`
	Total     int              `json:"total"`
	Values    []EnterpriseUser `json:"values"`
}

// EnterpriseUsersTool 企业成员列表工具结构体
type EnterpriseUsersTool struct {
	name        string
	description string
}

// NewEnterpriseUsersTool 创建企业成员列表工具实例
func NewEnterpriseUsersTool() MCPTool {
	return &EnterpriseUsersTool{
		name:        "get_enterprise_users",
		description: "Get enterprise users list from PingCode",
	}
}

// GetName 返回工具名称
func (t *EnterpriseUsersTool) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *EnterpriseUsersTool) GetDescription() string {
	return t.description
}

// GetToolDefinition 返回工具定义
func (t *EnterpriseUsersTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("获取企业用户列表 - 查看企业内所有用户信息，支持按状态、部门、角色等条件筛选"),
		mcp.WithNumber("page",
			mcp.Description("页码 - 可选，从1开始，默认为1"),
		),
		mcp.WithNumber("page_size",
			mcp.Description("每页数量 - 可选，默认20，最大100"),
		),
		mcp.WithString("status",
			mcp.Description("用户状态过滤 - 可选，支持：active(活跃)、inactive(非活跃)、pending(待激活)"),
		),
		mcp.WithString("department",
			mcp.Description("部门过滤 - 可选，按部门名称筛选用户"),
		),
		mcp.WithString("role",
			mcp.Description("角色过滤 - 可选，按用户角色筛选"),
		),
	)
}

// Handle 处理企业成员列表查询请求
func (t *EnterpriseUsersTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "get_enterprise_users", "request_id", fmt.Sprintf("users_%d", time.Now().UnixNano()))

	requestLogger.Info("开始获取企业成员列表")

	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("获取企业成员列表失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 解析请求参数
	params, err := t.parseArguments(request.GetArguments())
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("解析请求参数失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doEnterpriseUsersRequest(ctx, token, params, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("企业成员列表API调用失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	var errorRes models.ErrorResponse
	if err := json.Unmarshal([]byte(respStr), &errorRes); err == nil && errorRes.Code != "" {
		requestLogger.With("success", false, "error", errorRes.Message).Error("企业成员列表API调用失败")
		return mcp.NewToolResultError(errorRes.Message), nil
	}

	var resp EnterpriseUsersResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("success", false, "error", "json_parse_failed").Error("解析企业成员列表响应失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 格式化返回结果
	result := t.formatResponse(resp)

	requestLogger.With("success", true, "total_users", resp.Total, "page", resp.PageIndex, "page_size", resp.PageSize).Info("企业成员列表获取成功")

	return mcp.NewToolResultText(result), nil
}

// parseArguments 解析请求参数
func (t *EnterpriseUsersTool) parseArguments(args interface{}) (map[string]string, error) {
	params := make(map[string]string)

	if args == nil {
		// 使用默认参数
		params["page_index"] = "0" // API使用page_index从0开始
		params["page_size"] = "20"
		return params, nil
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments format")
	}

	// 解析页码（用户传入的page从1开始，需要转换为page_index从0开始）
	if page, exists := argsMap["page"]; exists {
		if pageFloat, ok := page.(float64); ok {
			pageIndex := int(pageFloat) - 1
			if pageIndex < 0 {
				pageIndex = 0
			}
			params["page_index"] = fmt.Sprintf("%d", pageIndex)
		} else if pageStr, ok := page.(string); ok {
			// 尝试解析字符串为数字
			if pageInt, err := strconv.Atoi(pageStr); err == nil {
				pageIndex := pageInt - 1
				if pageIndex < 0 {
					pageIndex = 0
				}
				params["page_index"] = fmt.Sprintf("%d", pageIndex)
			} else {
				params["page_index"] = "0"
			}
		}
	} else {
		params["page_index"] = "0"
	}

	// 解析每页数量
	if pageSize, exists := argsMap["page_size"]; exists {
		if pageSizeFloat, ok := pageSize.(float64); ok {
			params["page_size"] = fmt.Sprintf("%.0f", pageSizeFloat)
		} else if pageSizeStr, ok := pageSize.(string); ok {
			params["page_size"] = pageSizeStr
		}
	} else {
		params["page_size"] = "20"
	}

	// 解析状态过滤
	if status, exists := argsMap["status"]; exists {
		if statusStr, ok := status.(string); ok && statusStr != "" {
			params["status"] = statusStr
		}
	}

	// 解析部门过滤
	if department, exists := argsMap["department"]; exists {
		if departmentStr, ok := department.(string); ok && departmentStr != "" {
			params["department"] = departmentStr
		}
	}

	// 解析角色过滤
	if role, exists := argsMap["role"]; exists {
		if roleStr, ok := role.(string); ok && roleStr != "" {
			params["role"] = roleStr
		}
	}

	return params, nil
}

// doEnterpriseUsersRequest 执行企业成员列表API调用
func (t *EnterpriseUsersTool) doEnterpriseUsersRequest(ctx context.Context, token string, params map[string]string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	apiURL := GetBaseUrl() + "/v1/directory/users"
	log.With("url", apiURL, "method", "GET", "params", params).Debug("发起企业成员列表API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, enterpriseUsersTimeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, params)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("企业成员列表API请求成功")

	return string(body), nil
}

// formatResponse 格式化响应结果
func (t *EnterpriseUsersTool) formatResponse(resp EnterpriseUsersResponse) string {
	var result strings.Builder

	// 计算当前页码（从1开始显示）和总页数
	currentPage := resp.PageIndex + 1
	totalPages := (resp.Total + resp.PageSize - 1) / resp.PageSize
	if totalPages == 0 {
		totalPages = 1
	}

	result.WriteString(fmt.Sprintf("企业成员列表 (第%d页，共%d页，总计%d人):\n\n", currentPage, totalPages, resp.Total))

	if len(resp.Values) == 0 {
		result.WriteString("暂无成员数据")
		return result.String()
	}

	for i, user := range resp.Values {
		result.WriteString(fmt.Sprintf("%d. %s\n", i+1, user.DisplayName))
		result.WriteString(fmt.Sprintf("   ID: %s\n", user.ID))
		result.WriteString(fmt.Sprintf("   用户名: %s\n", user.Name))
		result.WriteString(fmt.Sprintf("   邮箱: %s\n", user.Email))
		if user.Mobile != "" {
			result.WriteString(fmt.Sprintf("   手机: %s\n", user.Mobile))
		}
		result.WriteString(fmt.Sprintf("   状态: %s\n", user.Status))
		if user.Department != "" && user.Department != "null" {
			result.WriteString(fmt.Sprintf("   部门: %s\n", user.Department))
		}
		if user.Job != "" && user.Job != "null" {
			result.WriteString(fmt.Sprintf("   职位: %s\n", user.Job))
		}
		if user.EmployeeNumber != "" {
			result.WriteString(fmt.Sprintf("   员工编号: %s\n", user.EmployeeNumber))
		}
		if user.URL != "" {
			result.WriteString(fmt.Sprintf("   详情链接: %s\n", user.URL))
		}
		result.WriteString("\n")
	}

	return result.String()
}

// CreateEnterpriseUserRequest 创建企业成员请求结构体
type CreateEnterpriseUserRequest struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Email          string `json:"email"`
	Mobile         string `json:"mobile,omitempty"`
	Department     string `json:"department,omitempty"`
	Job            string `json:"job,omitempty"`
	EmployeeNumber string `json:"employee_number,omitempty"`
	Password       string `json:"password,omitempty"`
}

// CreateEnterpriseUserResponse 创建企业成员响应结构体
type CreateEnterpriseUserResponse struct {
	ID             string `json:"id"`
	URL            string `json:"url"`
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Avatar         string `json:"avatar"`
	Email          string `json:"email"`
	Mobile         string `json:"mobile"`
	Status         string `json:"status"`
	Department     string `json:"department"`
	Job            string `json:"job"`
	EmployeeNumber string `json:"employee_number"`
}

// CreateEnterpriseUserTool 创建企业成员工具结构体
type CreateEnterpriseUserTool struct {
	name        string
	description string
}

// NewCreateEnterpriseUserTool 创建企业成员工具实例
func NewCreateEnterpriseUserTool() MCPTool {
	return &CreateEnterpriseUserTool{
		name:        "create_enterprise_user",
		description: "Create a new enterprise user in PingCode",
	}
}

// GetName 返回工具名称
func (t *CreateEnterpriseUserTool) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *CreateEnterpriseUserTool) GetDescription() string {
	return t.description
}

// GetToolDefinition 返回工具定义
func (t *CreateEnterpriseUserTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription("创建企业用户 - 在企业中添加新用户，设置基本信息、部门、职位等"),
		mcp.WithString("name",
			mcp.Description("用户登录名 - 必填，用户的唯一登录标识"),
			mcp.Required(),
		),
		mcp.WithString("display_name",
			mcp.Description("用户显示名称 - 必填，用户的真实姓名或昵称"),
			mcp.Required(),
		),
		mcp.WithString("email",
			mcp.Description("用户邮箱地址 - 必填，用于登录和通知的邮箱"),
			mcp.Required(),
		),
		mcp.WithString("mobile",
			mcp.Description("用户手机号码 - 可选，联系电话"),
		),
		mcp.WithString("department",
			mcp.Description("用户所属部门 - 可选，用户的组织部门"),
		),
		mcp.WithString("job",
			mcp.Description("用户职位 - 可选，用户的工作职位或角色"),
		),
		mcp.WithString("employee_number",
			mcp.Description("员工编号 - 可选，企业内部员工编号"),
		),
		mcp.WithString("password",
			mcp.Description("用户密码 - 可选，初始登录密码，不填则系统自动生成"),
		),
	)
}

// Handle 处理创建企业成员请求
func (t *CreateEnterpriseUserTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "create_enterprise_user", "request_id", fmt.Sprintf("create_user_%d", time.Now().UnixNano()))

	requestLogger.Info("开始创建企业成员")

	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("创建企业成员失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	// 解析请求参数
	userRequest, err := t.parseCreateUserArguments(request.GetArguments())
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("解析请求参数失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	respStr, err := t.doCreateUserRequest(ctx, token, userRequest, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("创建企业成员API调用失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	var resp CreateEnterpriseUserResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("success", false, "error", "json_parse_failed").Error("解析创建企业成员响应失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 格式化返回结果
	result := t.formatCreateUserResponse(resp)

	requestLogger.With("success", true, "user_id", resp.ID, "display_name", resp.DisplayName).Info("企业成员创建成功")

	return mcp.NewToolResultText(result), nil
}

// parseCreateUserArguments 解析创建用户请求参数
func (t *CreateEnterpriseUserTool) parseCreateUserArguments(args interface{}) (*CreateEnterpriseUserRequest, error) {
	if args == nil {
		return nil, fmt.Errorf("missing required arguments")
	}

	argsMap, ok := args.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid arguments format")
	}

	userRequest := &CreateEnterpriseUserRequest{}

	// 必需参数
	if name, exists := argsMap["name"]; exists {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			userRequest.Name = nameStr
		} else {
			return nil, fmt.Errorf("name is required and must be a non-empty string")
		}
	} else {
		return nil, fmt.Errorf("name is required")
	}

	if displayName, exists := argsMap["display_name"]; exists {
		if displayNameStr, ok := displayName.(string); ok && displayNameStr != "" {
			userRequest.DisplayName = displayNameStr
		} else {
			return nil, fmt.Errorf("display_name is required and must be a non-empty string")
		}
	} else {
		return nil, fmt.Errorf("display_name is required")
	}

	if email, exists := argsMap["email"]; exists {
		if emailStr, ok := email.(string); ok && emailStr != "" {
			userRequest.Email = emailStr
		} else {
			return nil, fmt.Errorf("email is required and must be a non-empty string")
		}
	} else {
		return nil, fmt.Errorf("email is required")
	}

	// 可选参数
	if mobile, exists := argsMap["mobile"]; exists {
		if mobileStr, ok := mobile.(string); ok {
			userRequest.Mobile = mobileStr
		}
	}

	if department, exists := argsMap["department"]; exists {
		if departmentStr, ok := department.(string); ok {
			userRequest.Department = departmentStr
		}
	}

	if job, exists := argsMap["job"]; exists {
		if jobStr, ok := job.(string); ok {
			userRequest.Job = jobStr
		}
	}

	if employeeNumber, exists := argsMap["employee_number"]; exists {
		if employeeNumberStr, ok := employeeNumber.(string); ok {
			userRequest.EmployeeNumber = employeeNumberStr
		}
	}

	if password, exists := argsMap["password"]; exists {
		if passwordStr, ok := password.(string); ok {
			userRequest.Password = passwordStr
		}
	}

	return userRequest, nil
}

// doCreateUserRequest 执行创建企业成员API调用
func (t *CreateEnterpriseUserTool) doCreateUserRequest(ctx context.Context, token string, userRequest *CreateEnterpriseUserRequest, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	apiURL := GetBaseUrl() + "/v1/directory/users"
	log.With("url", apiURL, "method", "POST", "user_name", userRequest.Name).Debug("发起创建企业成员API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, enterpriseUsersTimeout)
	defer cancel()

	// 序列化请求体
	requestBody, err := json.Marshal(userRequest)
	if err != nil {
		log.With("error", err.Error()).Error("序列化请求体失败")
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	body, _, err := utils.DoPostJSON(timeoutCtx, apiURL, headers, requestBody)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("创建企业成员API请求成功")

	return string(body), nil
}

// formatCreateUserResponse 格式化创建用户响应结果
func (t *CreateEnterpriseUserTool) formatCreateUserResponse(resp CreateEnterpriseUserResponse) string {
	var result strings.Builder

	result.WriteString("企业成员创建成功:\n\n")
	result.WriteString(fmt.Sprintf("ID: %s\n", resp.ID))
	result.WriteString(fmt.Sprintf("用户名: %s\n", resp.Name))
	result.WriteString(fmt.Sprintf("显示名称: %s\n", resp.DisplayName))
	result.WriteString(fmt.Sprintf("邮箱: %s\n", resp.Email))

	if resp.Mobile != "" {
		result.WriteString(fmt.Sprintf("手机: %s\n", resp.Mobile))
	}

	result.WriteString(fmt.Sprintf("状态: %s\n", resp.Status))

	if resp.Department != "" && resp.Department != "null" {
		result.WriteString(fmt.Sprintf("部门: %s\n", resp.Department))
	}

	if resp.Job != "" && resp.Job != "null" {
		result.WriteString(fmt.Sprintf("职位: %s\n", resp.Job))
	}

	if resp.EmployeeNumber != "" {
		result.WriteString(fmt.Sprintf("员工编号: %s\n", resp.EmployeeNumber))
	}

	if resp.URL != "" {
		result.WriteString(fmt.Sprintf("详情链接: %s\n", resp.URL))
	}

	return result.String()
}
