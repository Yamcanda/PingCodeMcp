package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger with additional functionality
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new zap logger instance (开发环境，输出到控制台和文件)
func New() *Logger {
	// 确保logs目录存在
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// 开发环境配置
	config := zap.NewDevelopmentConfig()

	// 同时输出到控制台和文件
	logFile := filepath.Join(logsDir, "app.log")
	config.OutputPaths = []string{"stdout", logFile}
	config.ErrorOutputPaths = []string{"stderr", logFile}

	// 自定义编码器配置
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	sugar := logger.Sugar()
	sugar = sugar.With("service", "PINGCODE-MCP")

	return &Logger{
		SugaredLogger: sugar,
	}
}

// NewProduction creates a new production zap logger instance (生产环境，JSON格式)
func NewProduction() *Logger {
	// 确保logs目录存在
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// 生产环境配置
	config := zap.NewProductionConfig()

	// 同时输出到控制台和文件
	logFile := filepath.Join(logsDir, "app.log")
	config.OutputPaths = []string{"stdout", logFile}
	config.ErrorOutputPaths = []string{"stderr", logFile}

	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize production logger: " + err.Error())
	}

	sugar := logger.Sugar()
	sugar = sugar.With("service", "PINGCODE-MCP")

	return &Logger{
		SugaredLogger: sugar,
	}
}

// NewFileOnly creates a logger that only writes to files (不输出到控制台)
func NewFileOnly() *Logger {
	// 确保logs目录存在
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// 生产环境配置，只输出到文件
	config := zap.NewProductionConfig()

	// 只输出到文件
	logFile := filepath.Join(logsDir, "app.log")
	config.OutputPaths = []string{logFile}
	config.ErrorOutputPaths = []string{logFile}

	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize file-only logger: " + err.Error())
	}

	sugar := logger.Sugar()
	sugar = sugar.With("service", "PINGCODE-MCP")

	return &Logger{
		SugaredLogger: sugar,
	}
}

// NewWithRotation creates a logger with log rotation (日志轮转)
func NewWithRotation() *Logger {
	// 确保logs目录存在
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// 创建带时间戳的日志文件名
	timestamp := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logsDir, "app-"+timestamp+".log")

	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", logFile}
	config.ErrorOutputPaths = []string{"stderr", logFile}

	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize rotation logger: " + err.Error())
	}

	sugar := logger.Sugar()
	sugar = sugar.With("service", "PINGCODE-MCP")

	return &Logger{
		SugaredLogger: sugar,
	}
}

// Info logs an info message
func (l *Logger) Info(args ...interface{}) {
	l.SugaredLogger.Info(args...)
}

// Infof logs an info message with format
func (l *Logger) Infof(template string, args ...interface{}) {
	l.SugaredLogger.Infof(template, args...)
}

// Error logs an error message
func (l *Logger) Error(args ...interface{}) {
	l.SugaredLogger.Error(args...)
}

// Errorf logs an error message with format
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.SugaredLogger.Errorf(template, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(args ...interface{}) {
	l.SugaredLogger.Debug(args...)
}

// Debugf logs a debug message with format
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.SugaredLogger.Debugf(template, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(args ...interface{}) {
	l.SugaredLogger.Warn(args...)
}

// Warnf logs a warning message with format
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.SugaredLogger.Warnf(template, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(args ...interface{}) {
	l.SugaredLogger.Fatal(args...)
}

// Fatalf logs a fatal message with format and exits
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.SugaredLogger.Fatalf(template, args...)
}

// With adds structured context to the logger
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		SugaredLogger: l.SugaredLogger.With(args...),
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.SugaredLogger.Sync()
}
