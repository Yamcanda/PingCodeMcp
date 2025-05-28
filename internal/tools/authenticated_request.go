package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/models"
	"PingCodeMcp/pkg/logger"
)

// 注册工具
func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewAuthenticatedRequestTool(),
		}
	})
}

// AuthenticatedRequestTool 认证请求工具结构体
type AuthenticatedRequestTool struct {
	name        string
	description string
}

// NewAuthenticatedRequestTool 创建认证请求工具实例
func NewAuthenticatedRequestTool() MCPTool {
	return &AuthenticatedRequestTool{
		name:        "make_authenticated_request",
		description: "Makes an authenticated request",
	}
}

// GetName 返回工具名称
func (t *AuthenticatedRequestTool) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *AuthenticatedRequestTool) GetDescription() string {
	return t.description
}

// GetToolDefinition 返回工具定义
func (t *AuthenticatedRequestTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("message", mcp.Description("Message to echo"), mcp.Required()),
	)
}

// Handle 处理认证请求
func (t *AuthenticatedRequestTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "make_authenticated_request", "request_id", fmt.Sprintf("auth_%d", time.Now().UnixNano()))

	requestLogger.Info("开始处理认证请求")

	message, ok := request.GetArguments()["message"].(string)
	if !ok {
		requestLogger.With("success", false, "error", "missing_message").Error("认证请求失败：缺少消息参数")
		return mcp.NewToolResultError(errors.New("missing or empty message").Error()), nil
	}

	token, err := auth.TokenFromContext(ctx)
	if err != nil {
		requestLogger.With("success", false, "error", "missing_token").Error("认证请求失败：缺少认证令牌")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Now our tool can make a request with the token, irrespective of where it came from.
	resp, err := t.makeRequest(ctx, message, token, requestLogger)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("认证请求执行失败")
		return mcp.NewToolResultError(err.Error()), nil
	}

	requestLogger.With("success", true, "message", message).Info("认证请求执行成功")

	return mcp.NewToolResultText(fmt.Sprintf("%+v", resp)), nil
}

// makeRequest makes a request to httpbin.org including the auth token in the request
// headers and the message in the query string.
func (t *AuthenticatedRequestTool) makeRequest(ctx context.Context, message, token string, log *logger.Logger) (*models.Response, error) {
	log.With("url", "https://httpbin.org/anything", "method", "GET", "message", message).Debug("发起认证请求")
	req, err := http.NewRequestWithContext(ctx, "GET", "https://httpbin.org/anything", nil)
	if err != nil {
		log.With("success", false, "error", err.Error()).Error("创建HTTP请求失败")
		return nil, err
	}
	req.Header.Set("Authorization", token)
	query := req.URL.Query()
	query.Add("message", message)
	req.URL.RawQuery = query.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.With("success", false, "error", err.Error()).Error("HTTP请求执行失败")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.With("success", false, "error", err.Error()).Error("读取响应体失败")
		return nil, err
	}

	var r *models.Response
	if err := json.Unmarshal(body, &r); err != nil {
		log.With("success", false, "error", err.Error()).Error("解析响应JSON失败")
		return nil, err
	}

	log.With("success", true, "response_size_bytes", len(body), "status_code", resp.StatusCode).Debug("认证请求执行成功")

	return r, nil
}
