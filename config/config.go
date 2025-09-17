package config

import (
	"time"
)

// ============================================================================
// 纯配置定义 - 仅包含配置参数，不包含实例对象
// ============================================================================

// Config 纯配置结构 - 只包含配置参数，不持有任何实例
type Config struct {
	// 数据库配置参数
	DSN         string // 数据库连接字符串
	AutoMigrate bool   // 是否自动迁移数据库表结构
	TableName   string // 规则表名

	// 缓存配置参数
	EnableCache   bool          // 是否启用缓存功能
	CacheTTL      time.Duration // 缓存生存时间
	MaxCacheSize  int           // 内存缓存最大条目数
	RedisAddr     string        // Redis服务器地址
	RedisPassword string        // Redis密码
	RedisDB       int           // Redis数据库编号

	// 定时任务配置参数
	SyncInterval time.Duration // 规则同步间隔
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		TableName:    "runehammer_rules",
		CacheTTL:     10 * time.Minute,
		SyncInterval: 5 * time.Minute,
		AutoMigrate:  false,
		MaxCacheSize: 1000,
		EnableCache:  true,
		RedisDB:      0,
	}
}

// Validate 验证配置参数的合法性
func (c *Config) Validate() error {
	if c.DSN == "" {
		return &ConfigError{Message: "数据库DSN不能为空"}
	}
	if c.TableName == "" {
		return &ConfigError{Message: "表名不能为空"}
	}
	if c.EnableCache && c.MaxCacheSize <= 0 {
		return &ConfigError{Message: "启用缓存时，缓存大小必须大于0"}
	}
	return nil
}

// ConfigError 配置错误类型
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}