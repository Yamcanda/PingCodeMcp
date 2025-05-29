package logger

import (
	"PingCodeMcp/internal/config"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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
					Path:       "logs",
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
	logConfig := config.GetLogConfig()

	// 确保日志目录存在
	if err := os.MkdirAll(logConfig.Path, 0755); err != nil {
		panic("Failed to create logs directory: " + err.Error())
	}

	// 创建核心编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 根据格式选择编码器
	var encoder zapcore.Encoder
	if logConfig.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 设置日志级别
	level, err := zapcore.ParseLevel(logConfig.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	var cores []zapcore.Core

	// 控制台输出
	if logConfig.Output == "console" || logConfig.Output == "both" {
		consoleCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}

	// 文件输出
	if logConfig.Output == "file" || logConfig.Output == "both" {
		var fileWriter zapcore.WriteSyncer

		if logConfig.Daily {
			// 使用lumberjack实现每日轮转
			lumberJackLogger := &lumberjack.Logger{
				Filename: filepath.Join(logConfig.Path, generateDailyFilename(logConfig.Filename)),
				MaxSize:  logConfig.MaxSize, // MB
				MaxAge:   logConfig.MaxAge,  // days
				Compress: logConfig.Compress,
			}

			// 创建一个自定义的WriteSyncer来处理每日轮转
			fileWriter = &dailyRotateWriter{
				logger:   lumberJackLogger,
				config:   logConfig,
				lastDate: time.Now().Format("2006-01-02"),
			}
		} else {
			// 普通文件输出
			lumberJackLogger := &lumberjack.Logger{
				Filename:   filepath.Join(logConfig.Path, logConfig.Filename),
				MaxSize:    logConfig.MaxSize,
				MaxBackups: logConfig.MaxBackups,
				MaxAge:     logConfig.MaxAge,
				Compress:   logConfig.Compress,
			}
			fileWriter = zapcore.AddSync(lumberJackLogger)
		}

		fileCore := zapcore.NewCore(
			encoder,
			fileWriter,
			level,
		)
		cores = append(cores, fileCore)
	}

	// 创建多核心logger
	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	sugar := logger.Sugar()
	sugar = sugar.With("service", config.Server.Name)

	globalLogger = &Logger{
		SugaredLogger: sugar,
	}
	return globalLogger
}

// dailyRotateWriter 实现每日轮转的WriteSyncer
type dailyRotateWriter struct {
	logger   *lumberjack.Logger
	config   *config.LogConfig
	lastDate string
}

func (w *dailyRotateWriter) Write(p []byte) (n int, err error) {
	currentDate := time.Now().Format("2006-01-02")

	// 检查是否需要切换到新的日志文件
	if currentDate != w.lastDate {
		// 关闭当前文件
		w.logger.Close()

		// 更新文件名
		w.logger.Filename = filepath.Join(w.config.Path, generateDailyFilename(w.config.Filename))
		w.lastDate = currentDate
	}

	return w.logger.Write(p)
}

func (w *dailyRotateWriter) Sync() error {
	return nil // lumberjack doesn't support sync
}

// generateDailyFilename 生成带日期的文件名
func generateDailyFilename(originalFilename string) string {
	timestamp := time.Now().Format("2006-01-02")
	ext := filepath.Ext(originalFilename)
	nameWithoutExt := originalFilename[:len(originalFilename)-len(ext)]
	return fmt.Sprintf("%s-%s%s", nameWithoutExt, timestamp, ext)
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
