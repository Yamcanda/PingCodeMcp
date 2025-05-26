package main

import (
	"fmt"

	"PingCodeMcp/internal/server"
	"PingCodeMcp/pkg/logger"
)

func main() {
	// 初始化日志
	log := logger.New()

	// 创建服务器
	s := server.NewMCPServer()
	cfg := s.GetConfig()

	// 启动 HTTP 服务器
	httpServer := s.ServeHTTP()

	log.Info(fmt.Sprintf("Starting %s v%s", cfg.ServerName, cfg.ServerVersion))
	log.Info(fmt.Sprintf("HTTP server listening on :%s", cfg.Port))

	if err := httpServer.Start(":" + cfg.Port); err != nil {
		log.Fatal("Server error:", err)
	}
}
