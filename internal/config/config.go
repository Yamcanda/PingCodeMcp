package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用程序配置结构体
type Config struct {
	Version       string            `yaml:"version"`
	Server        ServerConfig      `yaml:"server"`
	Logging       LoggingConfig     `yaml:"logging"`
	APIConfigFile string            `yaml:"api_config_file"` // API配置文件路径
	API           APIConfig         `yaml:"api,omitempty"`   // 可选：直接配置或从文件加载
	Auth          AuthConfig        `yaml:"auth,omitempty"`  // 可选：直接配置或从文件加载
	Performance   PerformanceConfig `yaml:"performance"`
	Monitoring    MonitoringConfig  `yaml:"monitoring"`
	Development   DevelopmentConfig `yaml:"development"`
	Production    ProductionConfig  `yaml:"production"`
	EnvOverride   EnvOverrideConfig `yaml:"env_override"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name    string        `yaml:"name"`
	Version string        `yaml:"version"`
	Port    int           `yaml:"port"`
	Host    string        `yaml:"host"`
	Timeout TimeoutConfig `yaml:"timeout"`
}

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	Read     time.Duration `yaml:"read"`
	Write    time.Duration `yaml:"write"`
	Idle     time.Duration `yaml:"idle"`
	Shutdown time.Duration `yaml:"shutdown"`
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

// APIConfig API配置
type APIConfig struct {
	IPSearch        IPSearchConfig        `yaml:"ip_search"`
	UserInfo        UserInfoConfig        `yaml:"user_info"`
	EnterpriseUsers EnterpriseUsersConfig `yaml:"enterprise_users"`
}

// IPSearchConfig IP查询服务配置
type IPSearchConfig struct {
	URL        string        `yaml:"url"`
	Timeout    time.Duration `yaml:"timeout"`
	RetryCount int           `yaml:"retry_count"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// UserInfoConfig 用户信息服务配置
type UserInfoConfig struct {
	URL        string        `yaml:"url"`
	Timeout    time.Duration `yaml:"timeout"`
	RetryCount int           `yaml:"retry_count"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// EnterpriseUsersConfig 企业成员列表服务配置
type EnterpriseUsersConfig struct {
	URL        string        `yaml:"url"`
	Timeout    time.Duration `yaml:"timeout"`
	RetryCount int           `yaml:"retry_count"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Token   TokenConfig   `yaml:"token"`
	Session SessionConfig `yaml:"session"`
}

// TokenConfig Token配置
type TokenConfig struct {
	HeaderName string `yaml:"header_name"`
	Prefix     string `yaml:"prefix"`
}

// SessionConfig 会话配置
type SessionConfig struct {
	Timeout          time.Duration `yaml:"timeout"`
	RefreshThreshold time.Duration `yaml:"refresh_threshold"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	HTTPClient  HTTPClientConfig  `yaml:"http_client"`
	Concurrency ConcurrencyConfig `yaml:"concurrency"`
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	MaxIdleConns        int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int           `yaml:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `yaml:"idle_conn_timeout"`
}

// ConcurrencyConfig 并发配置
type ConcurrencyConfig struct {
	MaxWorkers int `yaml:"max_workers"`
	QueueSize  int `yaml:"queue_size"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Metrics     MetricsConfig     `yaml:"metrics"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Path     string        `yaml:"path"`
	Interval time.Duration `yaml:"interval"`
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

// DevelopmentConfig 开发环境配置
type DevelopmentConfig struct {
	Debug     bool `yaml:"debug"`
	HotReload bool `yaml:"hot_reload"`
	Profiling bool `yaml:"profiling"`
}

// ProductionConfig 生产环境配置
type ProductionConfig struct {
	Debug     bool `yaml:"debug"`
	Profiling bool `yaml:"profiling"`
}

// EnvOverrideConfig 环境变量覆盖配置
type EnvOverrideConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

// APIConfigFile API配置文件结构
type APIConfigFile struct {
	Version string     `yaml:"version"`
	API     APIConfig  `yaml:"api"`
	Auth    AuthConfig `yaml:"auth"`
}

// Load 加载配置文件
func Load() *Config {
	return LoadWithAPIConfig()
}

// LoadWithAPIConfig 加载主配置并合并 API 配置
func LoadWithAPIConfig() *Config {
	// 加载主配置
	config := LoadFromFile("config.yaml")

	// 检查是否指定了 API 配置文件
	if config.APIConfigFile != "" {
		if apiConfig, err := LoadAPIConfig(config.APIConfigFile); err == nil {
			// 合并 API 配置
			config.API = apiConfig.API
			config.Auth = apiConfig.Auth
		}
	}

	return config
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

	return config
}

// LoadAPIConfig 加载 API 配置文件
func LoadAPIConfig(filename string) (*APIConfigFile, error) {
	apiConfig := &APIConfigFile{}

	// 查找配置文件
	configPath := findConfigFile(filename)
	if configPath == "" {
		return nil, fmt.Errorf("API配置文件 %s 未找到", filename)
	}

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取API配置文件失败: %w", err)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, apiConfig); err != nil {
		return nil, fmt.Errorf("解析API配置YAML失败: %w", err)
	}

	return apiConfig, nil
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
	// API配置文件路径覆盖
	if val := os.Getenv(prefix + "API_CONFIG_FILE"); val != "" {
		config.APIConfigFile = val
	}
	// API配置覆盖
	if val := os.Getenv(prefix + "API_IP_SEARCH_URL"); val != "" {
		config.API.IPSearch.URL = val
	}
	if val := os.Getenv(prefix + "API_USER_INFO_URL"); val != "" {
		config.API.UserInfo.URL = val
	}
	if val := os.Getenv(prefix + "API_ENTERPRISE_USERS_URL"); val != "" {
		config.API.EnterpriseUsers.URL = val
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
