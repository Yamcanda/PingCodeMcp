package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPTool 定义MCP工具接口
type MCPTool interface {
	// GetToolDefinition 返回工具的定义信息
	GetToolDefinition() mcp.Tool

	// Handle 处理工具调用请求
	Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	// GetName 返回工具名称
	GetName() string

	// GetDescription 返回工具描述
	GetDescription() string
}

// RegisterTool 注册单个工具到MCP服务器
func RegisterTool(mcpServer *server.MCPServer, tool MCPTool) {
	mcpServer.AddTool(tool.GetToolDefinition(), tool.Handle)
}

// RegisterAllTools 注册所有工具到MCP服务器
func RegisterAllTools(mcpServer *server.MCPServer) {
	tools := []MCPTool{
		NewIpSearchTool(),
		NewUserInfoTool(),
		NewAuthenticatedRequestTool(),
	}

	for _, tool := range tools {
		RegisterTool(mcpServer, tool)
	}
}
