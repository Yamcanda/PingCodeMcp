package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"PingCodeMcp/internal/config"
	"PingCodeMcp/internal/server"
	"PingCodeMcp/pkg/logger"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 获取日志器
	log := logger.NewWithConfig(cfg)

	// 确保在程序退出前刷新日志缓冲区
	defer log.Sync()

	// 创建服务器
	s := server.NewMCPServer()

	// 获取配置好的 Mux
	mux := s.ServeHTTP()

	// 创建一个标准的 http.Server
	httpServer := &http.Server{
		Addr:    ":" + cfg.GetPort(),
		Handler: mux,
	}

	log.Infof("Starting %s v%s", cfg.GetServerName(), cfg.GetServerVersion())
	log.Infof("HTTP server listening on :%s", cfg.GetPort())

	// 使用 goroutine 异步启动服务器
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// 创建一个带超时的上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Info("Server exited")
}
