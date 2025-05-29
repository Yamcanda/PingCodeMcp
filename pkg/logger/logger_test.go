package logger

import (
	"PingCodeMcp/internal/config"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDailyLogRotation(t *testing.T) {
	// 创建临时测试目录
	tempDir := "test_logs"
	defer os.RemoveAll(tempDir)

	// 创建测试配置
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Name:    "test-server",
			Version: "1.0.0",
		},
		Logging: config.LoggingConfig{
			Format: "json",
			Level:  "info",
			Output: "file",
			Rotation: config.RotationConfig{
				Enabled: true,
				Daily:   true,
			},
			File: config.FileLogConfig{
				Path:       tempDir,
				Filename:   "test.log",
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   false,
			},
		},
	}

	// 创建日志器
	logger := NewWithConfig(testConfig)

	// 写入一些日志
	logger.Info("Test log message 1")
	logger.Error("Test error message")
	logger.Debug("Test debug message")

	// 检查日志文件是否创建
	expectedFilename := generateDailyFilename("test.log")
	logFilePath := filepath.Join(tempDir, expectedFilename)

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Expected log file %s was not created", logFilePath)
	}

	// 验证文件名格式
	today := time.Now().Format("2006-01-02")
	expectedName := "test-" + today + ".log"
	if expectedFilename != expectedName {
		t.Errorf("Expected filename %s, got %s", expectedName, expectedFilename)
	}
}

func TestGenerateDailyFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"app.log", "app-" + time.Now().Format("2006-01-02") + ".log"},
		{"server.txt", "server-" + time.Now().Format("2006-01-02") + ".txt"},
		{"test", "test-" + time.Now().Format("2006-01-02")},
	}

	for _, test := range tests {
		result := generateDailyFilename(test.input)
		if result != test.expected {
			t.Errorf("For input %s, expected %s, got %s", test.input, test.expected, result)
		}
	}
}

func TestLoggerWithoutDailyRotation(t *testing.T) {
	// 创建临时测试目录
	tempDir := "test_logs_no_rotation"
	defer os.RemoveAll(tempDir)

	// 创建测试配置（不启用每日轮转）
	testConfig := &config.Config{
		Server: config.ServerConfig{
			Name:    "test-server",
			Version: "1.0.0",
		},
		Logging: config.LoggingConfig{
			Format: "json",
			Level:  "info",
			Output: "file",
			Rotation: config.RotationConfig{
				Enabled: true,
				Daily:   false, // 不启用每日轮转
			},
			File: config.FileLogConfig{
				Path:       tempDir,
				Filename:   "test.log",
				MaxSize:    10,
				MaxAge:     7,
				MaxBackups: 3,
				Compress:   false,
			},
		},
	}

	// 创建日志器
	logger := NewWithConfig(testConfig)

	// 写入一些日志
	logger.Info("Test log message without daily rotation")

	// 检查日志文件是否创建（应该使用原始文件名）
	logFilePath := filepath.Join(tempDir, "test.log")

	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Expected log file %s was not created", logFilePath)
	}
}
