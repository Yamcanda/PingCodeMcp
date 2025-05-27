package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"PingCodeMcp/internal/config"
	"PingCodeMcp/internal/utils"
	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	unknownAddr = "XX XX"
)

type ipSearchResponse struct {
	Status string `json:"status"`
	Data   struct {
		Country string `json:"country"`
		City    string `json:"city"`
	} `json:"data"`
}

// IpSearchTool IP查询工具结构体
type IpSearchTool struct {
	name        string
	description string
	config      *config.Config
}

// NewIpSearchTool 创建IP查询工具实例
func NewIpSearchTool(cfg *config.Config) MCPTool {
	return &IpSearchTool{
		name:        "ip_search",
		description: "Search IP location info",
		config:      cfg,
	}
}

// GetName 返回工具名称
func (t *IpSearchTool) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *IpSearchTool) GetDescription() string {
	return t.description
}

// GetToolDefinition 返回工具定义
func (t *IpSearchTool) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("ip",
			mcp.Description("IP address to search"),
			mcp.Required(),
		),
	)
}

// Handle 处理IP查询请求
func (t *IpSearchTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 记录请求开始时间，用于性能监控
	startTime := time.Now()

	ip, ok := request.GetArguments()["ip"].(string)
	if !ok || strings.TrimSpace(ip) == "" {
		log.With("duration_ms", time.Since(startTime).Milliseconds(), "success", false, "error", "missing_ip_parameter").Error("IP查询失败：缺少或空的IP参数")
		return nil, fmt.Errorf("missing or empty ip")
	}

	// 创建带有请求上下文的logger
	requestLogger := log.With("ip", ip, "tool", "ip_search", "request_id", fmt.Sprintf("ip_%d", time.Now().UnixNano()))

	requestLogger.Info("开始查询IP地理位置")

	address := unknownAddr
	respStr, err := t.doIpSearch(ctx, ip, requestLogger)
	if err != nil {
		requestLogger.With("duration_ms", time.Since(startTime).Milliseconds(), "success", false, "error", err.Error()).Error("IP查询API调用失败")
		return mcp.NewToolResultText(address), nil
	}
	var resp ipSearchResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		requestLogger.With("duration_ms", time.Since(startTime).Milliseconds(), "success", false, "error", "json_parse_failed", "response_preview", truncateString(respStr, 100)).Error("解析IP查询响应失败")
		return mcp.NewToolResultText(address), nil
	}
	if resp.Status == "Success" {
		region := resp.Data.Country
		city := resp.Data.City
		result := fmt.Sprintf("%s %s", region, city)
		requestLogger.With("duration_ms", time.Since(startTime).Milliseconds(), "success", true, "country", region, "city", city, "result", result).Info("IP查询成功")
		return mcp.NewToolResultText(result), nil
	}
	requestLogger.With("duration_ms", time.Since(startTime).Milliseconds(), "success", false, "api_status", resp.Status, "error", "api_returned_failure").Warn("IP查询返回非成功状态")
	return mcp.NewToolResultText(address), nil
}

// doIpSearch 执行IP查询API调用
func (t *IpSearchTool) doIpSearch(ctx context.Context, ip string, log *logger.Logger) (string, error) {
	query := map[string]string{"ip": ip}
	headers := map[string]string{"Accept": "application/json"}
	// 从配置中获取API URL
	apiURL := t.config.API.IPSearch.URL
	// 记录API调用开始
	apiStartTime := time.Now()
	log.With("url", apiURL, "method", "GET").Debug("发起IP查询API请求")
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, t.config.API.IPSearch.Timeout)
	defer cancel()
	body, _, err := utils.DoGet(timeoutCtx, apiURL, headers, query)
	apiDuration := time.Since(apiStartTime)
	if err != nil {
		log.With("url", apiURL, "api_duration_ms", apiDuration.Milliseconds(), "success", false, "error", err.Error()).Error("HTTP请求失败")
		return "", err
	}
	// 记录API调用成功
	log.With("url", apiURL, "api_duration_ms", apiDuration.Milliseconds(), "response_size_bytes", len(body), "success", true).Debug("IP查询API请求成功")
	return string(body), nil
}

// truncateString 截断字符串用于日志记录
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// HandleIpSearchTool 保持向后兼容的函数
func HandleIpSearchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 使用默认配置创建工具实例
	cfg := config.Load()
	tool := NewIpSearchTool(cfg)
	return tool.Handle(ctx, request)
}
