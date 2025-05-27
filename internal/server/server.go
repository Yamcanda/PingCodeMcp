package server

import (
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
		cfg.GetServerName(),
		cfg.GetServerVersion(),
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
	)

	// 注册工具
	registerTools(mcpServer, cfg)

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
func registerTools(mcpServer *server.MCPServer, cfg *config.Config) {
	// 使用新的接口方式注册所有工具
	tools.RegisterAllTools(mcpServer, cfg)
}
