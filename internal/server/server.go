package server

import (
	"encoding/json"
	"net/http"

	"github.com/mark3labs/mcp-go/server"

	"PingCodeMcp/internal/auth"
	"PingCodeMcp/internal/config"
	"PingCodeMcp/internal/tools"
)

type MCPServer struct {
	server *server.MCPServer
	config *config.Config
}

// NewMCPServer 创建一个新的MCPServer实例
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
	registerTools(mcpServer)

	return &MCPServer{
		server: mcpServer,
		config: cfg,
	}
}

// ServeHTTP 创建并返回一个配置好所有路由的*http.ServeMux
func (s *MCPServer) ServeHTTP() *http.ServeMux {
	// 创建 mcp-go 的 HTTP 服务
	mcpHttpServer := server.NewStreamableHTTPServer(s.server,
		server.WithHTTPContextFunc(auth.AuthFromRequest),
	)

	// 创建一个新的 Mux
	mux := http.NewServeMux()

	// 注册健康检查路由
	mux.HandleFunc("/health", healthCheck)
	// 将 mcp-go 的处理器注册到根路径
	mux.Handle("/", mcpHttpServer)

	return mux
}

// healthCheck 是一个简单的健康检查处理函数
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// GetConfig 获取配置
func (s *MCPServer) GetConfig() *config.Config {
	return s.config
}

// registerTools 注册所有可用的工具
func registerTools(mcpServer *server.MCPServer) {
	// 使用新的接口方式注册所有工具
	tools.RegisterAllTools(mcpServer)
}
