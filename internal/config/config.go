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
	Version     string            `yaml:"version"`
	Server      ServerConfig      `yaml:"server"`
	Logging     LoggingConfig     `yaml:"logging"`
	API         APIConfig         `yaml:"api"`
	Auth        AuthConfig        `yaml:"auth"`
	Performance PerformanceConfig `yaml:"performance"`
	Monitoring  MonitoringConfig  `yaml:"monitoring"`
	Development DevelopmentConfig `yaml:"development"`
	Production  ProductionConfig  `yaml:"production"`
	EnvOverride EnvOverrideConfig `yaml:"env_override"`
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

// Load 加载配置文件
func Load() *Config {
	return LoadFromFile("config.yaml")
}

// LoadFromFile 从指定文件加载配置
func LoadFromFile(filename string) *Config {
	config := &Config{}

	// 尝试加载YAML配置文件
	if err := loadYAMLConfig(filename, config); err != nil {
		// 如果加载失败，使用默认配置
		config = getDefaultConfig()
	}

	// 应用环境变量覆盖
	if config.EnvOverride.Enabled {
		applyEnvOverrides(config)
	}

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

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		Version: "1.0.0",
		Server: ServerConfig{
			Name:    getEnv("SERVER_NAME", "pingcode-mcp-server"),
			Version: getEnv("SERVER_VERSION", "1.0.0"),
			Port:    getEnvAsInt("PORT", 8080),
			Host:    getEnv("HOST", "0.0.0.0"),
			Timeout: TimeoutConfig{
				Read:     30 * time.Second,
				Write:    30 * time.Second,
				Idle:     60 * time.Second,
				Shutdown: 10 * time.Second,
			},
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "both"),
			File: FileLogConfig{
				Path:       "logs",
				Filename:   "app.log",
				MaxSize:    100,
				MaxAge:     7,
				MaxBackups: 10,
				Compress:   true,
			},
			Rotation: RotationConfig{
				Enabled: true,
				Daily:   true,
			},
		},
		API: APIConfig{
			IPSearch: IPSearchConfig{
				URL:        "https://whois.suyun.store/query",
				Timeout:    10 * time.Second,
				RetryCount: 3,
				RetryDelay: 1 * time.Second,
			},
			UserInfo: UserInfoConfig{
				URL:        "https://api.pingcode.com/v1/user/info",
				Timeout:    15 * time.Second,
				RetryCount: 2,
				RetryDelay: 2 * time.Second,
			},
			EnterpriseUsers: EnterpriseUsersConfig{
				URL:        "https://api.pingcode.com/v1/enterprise/users",
				Timeout:    15 * time.Second,
				RetryCount: 2,
				RetryDelay: 2 * time.Second,
			},
		},
		Auth: AuthConfig{
			Token: TokenConfig{
				HeaderName: "Authorization",
				Prefix:     "Bearer ",
			},
			Session: SessionConfig{
				Timeout:          3600 * time.Second,
				RefreshThreshold: 300 * time.Second,
			},
		},
		Performance: PerformanceConfig{
			HTTPClient: HTTPClientConfig{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
			Concurrency: ConcurrencyConfig{
				MaxWorkers: 50,
				QueueSize:  1000,
			},
		},
		Monitoring: MonitoringConfig{
			HealthCheck: HealthCheckConfig{
				Enabled:  true,
				Path:     "/health",
				Interval: 30 * time.Second,
			},
			Metrics: MetricsConfig{
				Enabled: true,
				Path:    "/metrics",
			},
		},
		Development: DevelopmentConfig{
			Debug:     false,
			HotReload: false,
			Profiling: false,
		},
		Production: ProductionConfig{
			Debug:     false,
			Profiling: false,
		},
		EnvOverride: EnvOverrideConfig{
			Enabled: true,
			Prefix:  "PINGCODE_MCP",
		},
	}
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
