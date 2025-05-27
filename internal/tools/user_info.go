package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/config"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
)

type userInfoResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Phone      string `json:"phone"`
	Status     string `json:"status"`
	Role       string `json:"role"`
	Department string `json:"department"`
}

// UserInfoTool 用户信息工具结构体
type UserInfoTool struct {
	name        string
	description string
	config      *config.Config
}

// NewUserInfoTool 创建用户信息工具实例
func NewUserInfoTool(cfg *config.Config) MCPTool {
	return &UserInfoTool{
		name:        "get_user_info",
		description: "Get current user basic information from PingCode",
		config:      cfg,
	}
}

// GetName 返回工具名称
func (t *UserInfoTool) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *UserInfoTool) GetDescription() string {
	return t.description
}

// GetToolDefinition 返回工具定义
func (t *UserInfoTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
	)
}

// Handle 处理用户信息查询请求
func (t *UserInfoTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "get_user_info", "request_id", fmt.Sprintf("user_%d", time.Now().UnixNano()))

	requestLogger.Info("开始获取用户信息")

	// 从context中获取authkey
	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("获取用户信息失败：缺少认证令牌")
		return nil, fmt.Errorf("missing or empty authorization token: %v", err)
	}

	// 确保token格式正确
	if !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}

	respStr, err := t.doUserInfoRequest(ctx, token, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("用户信息API调用失败")
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	var resp userInfoResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("success", false, "error", "json_parse_failed").Error("解析用户信息响应失败")
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// 格式化返回结果
	result := fmt.Sprintf(`用户基本信息:ID: %s 姓名: %s 邮箱: %s 电话: %s 状态: %s 角色: %s 部门: %s 头像: %s`, resp.ID, resp.Name, resp.Email, resp.Phone, resp.Status, resp.Role, resp.Department, resp.Avatar)

	requestLogger.With("success", true, "user_id", resp.ID, "user_name", resp.Name).Info("用户信息获取成功")

	return mcp.NewToolResultText(result), nil
}

// doUserInfoRequest 执行用户信息API调用
func (t *UserInfoTool) doUserInfoRequest(ctx context.Context, token string, log *logger.Logger) (string, error) {
	headers := map[string]string{
		"Authorization": token,
		"Accept":        "application/json",
		"Content-Type":  "application/json",
	}

	// 从配置中获取API URL
	apiURL := t.config.API.UserInfo.URL

	log.With("url", apiURL, "method", "GET").Debug("发起用户信息API请求")

	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, t.config.API.UserInfo.Timeout)
	defer cancel()

	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, nil)
	if err != nil {
		log.With("url", apiURL, "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}

	log.With("url", apiURL, "response_size_bytes", len(body), "success", true).Debug("用户信息API请求成功")

	return string(body), nil
}

// HandleUserInfoTool 保持向后兼容的函数
func HandleUserInfoTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用默认配置创建工具实例
	cfg := config.Load()
	tool := NewUserInfoTool(cfg)
	return tool.Handle(ctx, request)
}
