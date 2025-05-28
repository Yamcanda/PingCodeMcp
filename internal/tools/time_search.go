package tools

import (
	"context"
	"fmt"
	"time"

	"PingCodeMcp/pkg/logger"

	"github.com/mark3labs/mcp-go/mcp"
)

// 注册工具
func init() {
	toolsFns = append(toolsFns, func() *[]MCPTool {
		return &[]MCPTool{
			NewTimeSearch(),
		}
	})
}

// TimeSearch 时间查询工具
type TimeSearch struct {
	name        string
	description string
}

// NewTimeSearch 创建新的时间工具实例
func NewTimeSearch() MCPTool {
	return &TimeSearch{
		name:        "get_current_time",
		description: "Get the current time in multiple formats and time zones",
	}
}

// GetName 返回工具名称
func (t *TimeSearch) GetName() string {
	return t.name
}

// GetDescription 返回工具描述
func (t *TimeSearch) GetDescription() string {
	return t.description
}

// GetToolDefinition 返回工具定义
func (t *TimeSearch) GetToolDefinition() mcp.Tool {
	return mcp.NewTool(t.name,
		mcp.WithDescription(t.description),
		mcp.WithString("format",
			mcp.Description("时间格式类型 (default, iso, unix, rfc3339, custom)"),
		),
		mcp.WithString("timezone",
			mcp.Description("时区 (如: Asia/Shanghai, UTC, America/New_York)"),
		),
		mcp.WithString("custom_format",
			mcp.Description("自定义时间格式 (当format为custom时使用，Go时间格式)"),
		),
	)
}

// Handle 处理时间查询请求
func (t *TimeSearch) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 创建logger实例
	log := logger.New()
	defer log.Sync()

	// 创建带有请求上下文的logger
	requestLogger := log.With("tool", "get_current_time", "request_id", fmt.Sprintf("time_%d", time.Now().UnixNano()))

	requestLogger.Info("开始获取当前时间")

	// 解析参数
	format := "default"
	timezone := "Asia/Shanghai"
	customFormat := ""

	args := request.GetArguments()
	if args != nil {
		// 修复类型断言问题
		if f, exists := args["format"]; exists {
			if str, ok := f.(string); ok && str != "" {
				format = str
			}
		}
		if tz, exists := args["timezone"]; exists {
			if str, ok := tz.(string); ok && str != "" {
				timezone = str
			}
		}
		if cf, exists := args["custom_format"]; exists {
			if str, ok := cf.(string); ok && str != "" {
				customFormat = str
			}
		}
	}

	requestLogger.With("format", format, "timezone", timezone).Debug("解析时间查询参数")

	// 加载时区
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		requestLogger.With("success", false, "error", err.Error()).Error("加载时区失败")
		// 如果时区加载失败，使用本地时区
		loc = time.Local
		requestLogger.With("fallback_timezone", "Local").Warn("使用本地时区作为备选")
	}

	// 获取当前时间
	now := time.Now().In(loc)

	// 根据格式类型格式化时间
	var result string
	switch format {
	case "iso":
		result = fmt.Sprintf("当前时间 (ISO 8601): %s", now.Format("2006-01-02T15:04:05Z07:00"))
	case "unix":
		result = fmt.Sprintf("当前时间:\n- Unix 时间戳: %d\n- Unix 毫秒时间戳: %d\n- 可读格式: %s",
			now.Unix(), now.UnixMilli(), now.Format("2006-01-02 15:04:05"))
	case "rfc3339":
		result = fmt.Sprintf("当前时间 (RFC3339): %s", now.Format(time.RFC3339))
	case "custom":
		if customFormat == "" {
			customFormat = "2006-01-02 15:04:05"
		}
		result = fmt.Sprintf("当前时间 (自定义格式): %s", now.Format(customFormat))
	default: // "default"
		result = fmt.Sprintf(`当前时间信息:
			- 标准格式: %s
			- 日期: %s
			- 时间: %s
			- 时区: %s
			- 星期: %s
			- Unix时间戳: %d`,
			now.Format("2006-01-02 15:04:05"),
			now.Format("2006-01-02"),
			now.Format("15:04:05"),
			now.Location().String(),
			now.Weekday().String(),
			now.Unix())
	}

	requestLogger.With("success", true, "format", format, "timezone", timezone).Info("时间获取成功")

	return mcp.NewToolResultText(result), nil
}
