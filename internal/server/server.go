package server

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/config"
	"PingCodeMcp/internal/tools"
)

type MCPServer struct {
	server *server.MCPServer
	config *config.Config
}

func NewMCPServer() *MCPServer {
	cfg := config.Load()

	mcpServer := server.NewMCPServer(
		cfg.ServerName,
		cfg.ServerVersion,
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)

	// 注册工具
	registerTools(mcpServer)

	return &MCPServer{
		server: mcpServer,
		config: cfg,
	}
}

func (s *MCPServer) ServeHTTP() *server.StreamableHTTPServer {
	return server.NewStreamableHTTPServer(s.server,
		server.WithHTTPContextFunc(auth.AuthFromRequest),
	)
}

func (s *MCPServer) GetConfig() *config.Config {
	return s.config
}

// registerTools 注册所有可用的工具
func registerTools(mcpServer *server.MCPServer) {

	// 注册认证请求工具
	mcpServer.AddTool(mcp.NewTool("make_authenticated_request",
		mcp.WithDescription("Makes an authenticated request"),
		mcp.WithString("message",
			mcp.Description("Message to echo"),
			mcp.Required(),
		),
	), tools.HandleMakeAuthenticatedRequestTool)

	// 注册 IP 查询工具
	mcpServer.AddTool(mcp.NewTool("ip_search",
		mcp.WithDescription("Search IP location info"),
		mcp.WithString("ip",
			mcp.Description("IP address to search"),
			mcp.Required(),
		),
	), tools.HandleIpSearchTool)

	// 注册用户信息工具
	mcpServer.AddTool(mcp.NewTool("get_user_info",
		mcp.WithDescription("Get current user basic information from PingCode"),
	), tools.HandleUserInfoTool)
}
