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

// New creates a new zap logger instance (使用全局配置)
func New() *Logger {
	return GetGlobalLogger()
}

// GetGlobalLogger 获取全局日志器
func GetGlobalLogger() *Logger {
	if globalLogger == nil {
		// 如果没有初始化，使用默认配置
		defaultConfig := config.Config{
			Server: config.ServerConfig{
				Name:    "ping-code-mcp-server",
				Version: "1.0.0",
			},
			Logging: config.LoggingConfig{
				Format:   "json",
				Level:    "info",
				Output:   "both",
				Rotation: config.RotationConfig{Enabled: true, Daily: true},
				File: config.FileLogConfig{
					Compress:   true,
					Filename:   "PingCodeMcp.log",
					MaxAge:     7,
					MaxBackups: 3,
					MaxSize:    100,
				},
			},
		}
		globalLogger = NewWithConfig(&defaultConfig)
	}
	return globalLogger
}

// NewWithConfig 根据配置创建日志器
func NewWithConfig(config *config.Config) *Logger {
	// 确保日志目录存在
	if err := os.MkdirAll(config.GetLogConfig().Path, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// 创建日志文件路径
	var logFile string
	if config.GetLogConfig().Daily {
		// 按日期生成文件名
		timestamp := time.Now().Format("2006-01-02")
		filename := fmt.Sprintf("%s-%s.log",
			config.GetLogConfig().Filename[:len(config.GetLogConfig().Filename)-4], // 移除.log扩展名
			timestamp)
		logFile = filepath.Join(config.GetLogConfig().Path, filename)
	} else {
		logFile = filepath.Join(config.GetLogConfig().Path, config.GetLogConfig().Filename)
	}

	// 根据格式选择配置
	var zapConfig zap.Config
	if config.GetLogConfig().Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	level, err := zapcore.ParseLevel(config.GetLogConfig().Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// 设置输出路径
	switch config.GetLogConfig().Output {
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
	sugar = sugar.With("service", config.Server.Name)

	globalLogger = &Logger{
		SugaredLogger: sugar,
	}
	return globalLogger
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
