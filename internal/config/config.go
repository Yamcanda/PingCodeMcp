package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

var (
	appConfig *Config
)

// Config 应用程序配置结构体
type Config struct {
	Version     string            `yaml:"version"`
	Server      ServerConfig      `yaml:"server"`
	Logging     LoggingConfig     `yaml:"logging"`
	EnvOverride EnvOverrideConfig `yaml:"env_override"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name            string `yaml:"name"`
	Version         string `yaml:"version"`
	Port            int    `yaml:"port"`
	Host            string `yaml:"host"`
	PingCodeBaseUrl string `yaml:"pingcode_base_url"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string         `yaml:"level"`
	Format   string         `yaml:"format"`
	Output   string         `yaml:"output"`
	File     FileLogConfig  `yaml:"file"`
	Rotation RotationConfig `yaml:"rotation"`
}

// FileLogConfig 文件日志配置
type FileLogConfig struct {
	Path       string `yaml:"path"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"`
	MaxAge     int    `yaml:"max_age"`
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// RotationConfig 日志轮转配置
type RotationConfig struct {
	Enabled bool `yaml:"enabled"`
	Daily   bool `yaml:"daily"`
}

// EnvOverrideConfig 环境变量覆盖配置
type EnvOverrideConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

// Load 加载配置文件
func Load() *Config {
	appConfig = LoadFromFile("config.yaml")
	return appConfig
}

// LoadFromFile 从指定文件加载配置
func LoadFromFile(filename string) *Config {
	config := &Config{}

	// 尝试加载YAML配置文件
	if err := loadYAMLConfig(filename, config); err != nil {
		return config
	}

	// 应用环境变量覆盖
	if config.EnvOverride.Enabled {
		applyEnvOverrides(config)
	}
	appConfig = config
	return config
}

// loadYAMLConfig 加载YAML配置文件
func loadYAMLConfig(filename string, config *Config) error {
	// 查找配置文件
	configPath := findConfigFile(filename)
	if configPath == "" {
		return fmt.Errorf("配置文件 %s 未找到", filename)
	}

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析YAML配置失败: %w", err)
	}

	return nil
}

// findConfigFile 查找配置文件
func findConfigFile(filename string) string {
	// 查找路径列表
	searchPaths := []string{
		filename,                           // 当前目录
		filepath.Join(".", filename),       // 当前目录
		filepath.Join("config", filename),  // config目录
		filepath.Join("configs", filename), // configs目录
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// applyEnvOverrides 应用环境变量覆盖
func applyEnvOverrides(config *Config) {
	prefix := config.EnvOverride.Prefix + "_"
	// 服务器配置覆盖
	if val := os.Getenv(prefix + "SERVER_NAME"); val != "" {
		config.Server.Name = val
	}
	if val := os.Getenv(prefix + "SERVER_VERSION"); val != "" {
		config.Server.Version = val
	}
	if val := os.Getenv(prefix + "SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.Server.Port = port
		}
	}
	if val := os.Getenv(prefix + "SERVER_HOST"); val != "" {
		config.Server.Host = val
	}
	// 日志配置覆盖
	if val := os.Getenv(prefix + "LOGGING_LEVEL"); val != "" {
		config.Logging.Level = val
	}
	if val := os.Getenv(prefix + "LOGGING_FORMAT"); val != "" {
		config.Logging.Format = val
	}
	if val := os.Getenv(prefix + "LOGGING_OUTPUT"); val != "" {
		config.Logging.Output = val
	}

}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量作为整数，如果不存在或转换失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetServerName 获取服务器名称 (向后兼容)
func (c *Config) GetServerName() string {
	return c.Server.Name
}

// GetServerVersion 获取服务器版本 (向后兼容)
func (c *Config) GetServerVersion() string {
	return c.Server.Version
}

// GetPort 获取端口号字符串 (向后兼容)
func (c *Config) GetPort() string {
	return strconv.Itoa(c.Server.Port)
}

// GetLogLevel 获取日志级别 (向后兼容)
func (c *Config) GetLogLevel() string {
	return c.Logging.Level
}

// LegacyConfig 为了向后兼容，保留原有的字段
type LegacyConfig struct {
	ServerName    string
	ServerVersion string
	Port          string
	LogLevel      string
}

// ToLegacy 转换为旧版配置结构 (向后兼容)
func (c *Config) ToLegacy() *LegacyConfig {
	return &LegacyConfig{
		ServerName:    c.Server.Name,
		ServerVersion: c.Server.Version,
		Port:          c.GetPort(),
		LogLevel:      c.Logging.Level,
	}
}

// GetLogConfig 获取日志配置
func (c *Config) GetLogConfig() *LogConfig {
	return &LogConfig{
		Level:      c.Logging.Level,
		Format:     c.Logging.Format,
		Output:     c.Logging.Output,
		Path:       c.Logging.File.Path,
		Filename:   c.Logging.File.Filename,
		MaxSize:    c.Logging.File.MaxSize,
		MaxAge:     c.Logging.File.MaxAge,
		MaxBackups: c.Logging.File.MaxBackups,
		Compress:   c.Logging.File.Compress,
		Daily:      c.Logging.Rotation.Daily,
	}
}

// LogConfig 日志配置结构体（为了避免循环导入）
type LogConfig struct {
	Level      string
	Format     string
	Output     string
	Path       string
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
	Daily      bool
}

// GetPingCodeBaseUrl 获取 PingCode 的基础 URL
func GetPingCodeBaseUrl() string {
	if appConfig == nil {
		return ""
	}
	return appConfig.Server.PingCodeBaseUrl
}
