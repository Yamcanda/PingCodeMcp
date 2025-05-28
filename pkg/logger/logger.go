package logger

import (
	"PingCodeMcp/internal/config"
	"fmt"
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

var (
	globalLogger *Logger
)

// InitGlobalLogger 初始化全局日志器
func InitGlobalLogger(config *config.LogConfig) {
	globalLogger = NewWithConfig(config)
}

// GetGlobalLogger 获取全局日志器
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		defaultConfig := config.LogConfig{
			Level:    "info",
			Format:   "console",
			Output:   "both",
			Path:     "logs",
			Filename: "app.log",
			Daily:    true,
		}
		globalLogger = NewWithConfig(&defaultConfig)
	}
	return globalLogger
}

// NewWithConfig 根据配置创建日志器
func NewWithConfig(config *config.LogConfig) *Logger {
	// 确保日志目录存在
	if err := os.MkdirAll(config.Path, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// 创建日志文件路径
	var logFile string
	if config.Daily {
		// 按日期生成文件名
		timestamp := time.Now().Format("2006-01-02")
		filename := fmt.Sprintf("%s-%s.log",
			config.Filename[:len(config.Filename)-4], // 移除.log扩展名
			timestamp)
		logFile = filepath.Join(config.Path, filename)
	} else {
		logFile = filepath.Join(config.Path, config.Filename)
	}

	// 根据格式选择配置
	var zapConfig zap.Config
	if config.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// 设置输出路径
	switch config.Output {
	case "console":
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	case "file":
		zapConfig.OutputPaths = []string{logFile}
		zapConfig.ErrorOutputPaths = []string{logFile}
	case "both":
		zapConfig.OutputPaths = []string{"stdout", logFile}
		zapConfig.ErrorOutputPaths = []string{"stderr", logFile}
	default:
		zapConfig.OutputPaths = []string{"stdout", logFile}
		zapConfig.ErrorOutputPaths = []string{"stderr", logFile}
	}

	// 自定义编码器配置
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.EncoderConfig.LevelKey = "level"
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.CallerKey = "caller"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 创建logger
	logger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	sugar := logger.Sugar()
	sugar = sugar.With("service", "PINGCODE-MCP")

	return &Logger{
		SugaredLogger: sugar,
	}
}

// New creates a new zap logger instance (使用全局配置)
func New() *Logger {
	return GetGlobalLogger()
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

	// 关键配置：增加调用栈跳过层数，显示真实调用位置
	logger, err := config.Build(zap.AddCallerSkip(1))
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

	// 关键配置：增加调用栈跳过层数，显示真实调用位置
	logger, err := config.Build(zap.AddCallerSkip(1))
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

	// 关键配置：增加调用栈跳过层数，显示真实调用位置
	logger, err := config.Build(zap.AddCallerSkip(1))
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
