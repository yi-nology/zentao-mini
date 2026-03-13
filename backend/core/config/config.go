package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Zentao   ZentaoConfig   `mapstructure:"zentao"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Security SecurityConfig `mapstructure:"security"`
	Log      LogConfig      `mapstructure:"log"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	// 应用类型: "wails" 或 "http"
	Type string `mapstructure:"type"`
	// HTTP服务器端口
	Port string `mapstructure:"port"`
	// 静态资源路径（仅用于HTTP模式）
	StaticPath string `mapstructure:"static_path"`
	// 读超时时间（秒）
	ReadTimeout int `mapstructure:"read_timeout"`
	// 写超时时间（秒）
	WriteTimeout int `mapstructure:"write_timeout"`
	// 优雅关闭超时时间（秒）
	ShutdownTimeout int `mapstructure:"shutdown_timeout"`
}

// ZentaoConfig 禅道配置
type ZentaoConfig struct {
	// 禅道服务器地址
	Server string `mapstructure:"server"`
	// 禅道账号
	Account string `mapstructure:"account"`
	// 禅道密码
	Password string `mapstructure:"password"`
	// Token刷新间隔（小时）
	TokenRefreshInterval int `mapstructure:"token_refresh_interval"`
	// 请求超时时间（秒）
	RequestTimeout int `mapstructure:"request_timeout"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	// 认证配置文件路径
	ConfigPath string `mapstructure:"config_path"`
	// 认证数据库路径
	DBPath string `mapstructure:"db_path"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 加密密钥
	EncryptionKey string `mapstructure:"encryption_key"`
	// CORS允许的域名（多个域名用逗号分隔）
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// LogConfig 日志配置
type LogConfig struct {
	// 日志级别: debug, info, warn, error
	Level string `mapstructure:"level"`
	// 日志格式: json, console
	Format string `mapstructure:"format"`
	// 日志输出路径（空表示标准输出）
	OutputPath string `mapstructure:"output_path"`
	// 是否启用调用者信息
	EnableCaller bool `mapstructure:"enable_caller"`
	// 是否启用堆栈跟踪
	EnableStacktrace bool `mapstructure:"enable_stacktrace"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 每分钟允许的请求数
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	// 超限后的封禁时长（分钟）
	BlockDurationMinutes int `mapstructure:"block_duration_minutes"`
}

// 全局配置实例
var globalConfig *Config

// Init 初始化配置
// configPath: 配置文件路径（可选）
// envPrefix: 环境变量前缀
func Init(configPath string, envPrefix string) error {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 设置环境变量前缀
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 尝试在多个位置查找配置文件
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./backend")
		v.AddConfigPath("./config")
	}

	// 尝试读取配置文件（如果存在）
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// 配置文件不存在是允许的，将使用环境变量和默认值
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := validate(&cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	globalConfig = &cfg
	return nil
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// 服务器配置
	v.SetDefault("server.type", "http")
	v.SetDefault("server.port", "12345")
	v.SetDefault("server.read_timeout", 120)
	v.SetDefault("server.write_timeout", 120)
	v.SetDefault("server.shutdown_timeout", 30)

	// 禅道配置
	v.SetDefault("zentao.token_refresh_interval", 12)
	v.SetDefault("zentao.request_timeout", 120)

	// 认证配置
	v.SetDefault("auth.config_path", "./auth_config.enc")
	v.SetDefault("auth.db_path", "./auth.db")

	// 日志配置
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.enable_caller", true)
	v.SetDefault("log.enable_stacktrace", false)

	// 限流配置
	v.SetDefault("rate_limit.requests_per_minute", 60)
	v.SetDefault("rate_limit.block_duration_minutes", 5)
}

// validate 验证配置
func validate(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Type != "http" && cfg.Server.Type != "wails" {
		return fmt.Errorf("invalid server type: %s, must be 'http' or 'wails'", cfg.Server.Type)
	}

	if cfg.Server.Type == "http" {
		if cfg.Server.Port == "" {
			return fmt.Errorf("server port is required for http mode")
		}
	}

	// 验证日志级别
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[cfg.Log.Level] {
		return fmt.Errorf("invalid log level: %s, must be one of: debug, info, warn, error", cfg.Log.Level)
	}

	// 验证日志格式
	if cfg.Log.Format != "json" && cfg.Log.Format != "console" {
		return fmt.Errorf("invalid log format: %s, must be 'json' or 'console'", cfg.Log.Format)
	}

	// 验证限流配置
	if cfg.RateLimit.RequestsPerMinute <= 0 {
		return fmt.Errorf("rate limit requests per minute must be positive")
	}

	if cfg.RateLimit.BlockDurationMinutes <= 0 {
		return fmt.Errorf("rate limit block duration must be positive")
	}

	// 验证超时配置
	if cfg.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server read timeout must be positive")
	}

	if cfg.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server write timeout must be positive")
	}

	if cfg.Server.ShutdownTimeout <= 0 {
		return fmt.Errorf("server shutdown timeout must be positive")
	}

	return nil
}

// Get 获取全局配置实例
func Get() *Config {
	if globalConfig == nil {
		panic("config not initialized, please call Init() first")
	}
	return globalConfig
}

// GetServerTimeout 获取服务器超时配置（转换为time.Duration）
func (c *ServerConfig) GetReadTimeout() time.Duration {
	return time.Duration(c.ReadTimeout) * time.Second
}

func (c *ServerConfig) GetWriteTimeout() time.Duration {
	return time.Duration(c.WriteTimeout) * time.Second
}

func (c *ServerConfig) GetShutdownTimeout() time.Duration {
	return time.Duration(c.ShutdownTimeout) * time.Second
}

// GetZentaoTimeout 获取禅道请求超时配置
func (c *ZentaoConfig) GetRequestTimeout() time.Duration {
	return time.Duration(c.RequestTimeout) * time.Second
}

// GetTokenRefreshInterval 获取Token刷新间隔
func (c *ZentaoConfig) GetTokenRefreshInterval() time.Duration {
	return time.Duration(c.TokenRefreshInterval) * time.Hour
}

// GetBlockDuration 获取限流封禁时长
func (c *RateLimitConfig) GetBlockDuration() time.Duration {
	return time.Duration(c.BlockDurationMinutes) * time.Minute
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Log.Format == "json"
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Log.Format == "console"
}
