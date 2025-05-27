package logger

import "time"

// Example demonstrates how to use the zap logger
func Example() {
	// 创建开发环境logger
	log := New()
	defer log.Sync()

	// 基本日志记录
	log.Info("这是一条信息日志")
	log.Error("这是一条错误日志")
	log.Debug("这是一条调试日志")
	log.Warn("这是一条警告日志")

	// 格式化日志
	log.Infof("用户 %s 登录成功，耗时 %d ms", "张三", 150)
	log.Errorf("连接数据库失败: %v", "connection timeout")

	// 结构化日志 - 添加上下文字段
	contextLogger := log.With(
		"user_id", 12345,
		"request_id", "req-abc-123",
		"ip", "192.168.1.1",
	)

	contextLogger.Info("用户执行了操作")
	contextLogger.Error("操作失败")

	// 记录性能指标
	start := time.Now()
	// 模拟一些操作
	time.Sleep(10 * time.Millisecond)
	duration := time.Since(start)

	log.With(
		"operation", "database_query",
		"duration_ms", duration.Milliseconds(),
		"success", true,
	).Info("数据库查询完成")

	// 记录HTTP请求
	log.With(
		"method", "POST",
		"path", "/api/users",
		"status_code", 201,
		"response_time_ms", 45,
		"user_agent", "Mozilla/5.0",
	).Info("HTTP请求处理完成")
}

// ExampleProduction demonstrates production logger usage
func ExampleProduction() {
	// 创建生产环境logger (JSON格式输出)
	log := NewProduction()
	defer log.Sync()

	log.Info("生产环境日志示例")

	// 生产环境通常使用结构化日志
	log.With(
		"service", "user-service",
		"version", "1.2.3",
		"environment", "production",
	).Info("服务启动")
}
