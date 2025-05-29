package main

import (
	"PingCodeMcp/internal/config"
	"PingCodeMcp/pkg/logger"
	"time"
)

func main() {
	// 创建配置，启用每日日志轮转
	config := &config.Config{
		Server: config.ServerConfig{
			Name:    "daily-log-example",
			Version: "1.0.0",
		},
		Logging: config.LoggingConfig{
			Format: "json",
			Level:  "info",
			Output: "both", // 同时输出到控制台和文件
			Rotation: config.RotationConfig{
				Enabled: true,
				Daily:   true, // 启用每日轮转
			},
			File: config.FileLogConfig{
				Path:       "logs",
				Filename:   "example.log",
				MaxSize:    100,  // 100MB
				MaxAge:     30,   // 保留30天
				MaxBackups: 10,   // 最多保留10个备份文件
				Compress:   true, // 压缩旧文件
			},
		},
	}

	// 创建日志器
	log := logger.NewWithConfig(config)

	// 记录一些日志
	log.Info("应用程序启动")
	log.Infof("当前时间: %s", time.Now().Format("2006-01-02 15:04:05"))

	log.Debug("这是一条调试信息")
	log.Warn("这是一条警告信息")
	log.Error("这是一条错误信息")

	// 添加结构化字段
	log.With("user_id", 12345, "action", "login").Info("用户登录")
	log.With("request_id", "req-123", "duration", "150ms").Info("请求处理完成")

	log.Info("应用程序结束")

	// 确保日志写入
	log.Sync()
}
