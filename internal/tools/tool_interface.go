package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// baseUrl = "https://open.pingcode.com"
	baseUrl = "http://192.168.0.170/open"
)

var (
	toolsFns []func() *[]MCPTool
)

// 定义 MCP 工具接口
type MCPTool interface {

	// 返回工具的定义信息
	GetToolDefinition() mcp.Tool

	// 处理工具调用请求
	Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// 返回工具名称
	GetName() string

	// 返回工具描述
	GetDescription() string
}

// 注册单个工具到 MCP 服务器
func RegisterTool(mcpServer *server.MCPServer, tool MCPTool) {
	mcpServer.AddTool(tool.GetToolDefinition(), tool.Handle)
}

// 注册所有工具到 MCP 服务器
func RegisterAllTools(mcpServer *server.MCPServer) {
	// 遍历所有注册的工具创建函数
	for _, toolFn := range toolsFns {
		// 调用工具创建函数获取工具列表
		tools := toolFn()
		if tools != nil {
			// 注册每个工具到 MCP 服务器
			for _, tool := range *tools {
				RegisterTool(mcpServer, tool)
			}
		}
	}
}
