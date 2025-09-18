package config

import (
	"time"
)

// CacheType 缓存类型枚举
type CacheType string

const (
	CacheTypeMemory CacheType = "memory" // 内存缓存
	CacheTypeRedis  CacheType = "redis"  // Redis缓存
	CacheTypeNone   CacheType = "none"   // 禁用缓存
)

// ============================================================================
// 纯配置定义 - 仅包含配置参数，不包含实例对象
// ============================================================================

// Config 纯配置结构 - 只包含配置参数，不持有任何实例
type Config struct {
	// 数据库配置参数
	DSN         string // 数据库连接字符串
	AutoMigrate bool   // 是否自动迁移数据库表结构

	// 缓存配置参数
	CacheType     CacheType     // 缓存类型：memory、redis、none
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
		CacheTTL:     10 * time.Minute,
		SyncInterval: 5 * time.Minute,
		AutoMigrate:  false,
		MaxCacheSize: 1000,
		CacheType:    CacheTypeMemory, // 默认使用内存缓存
		RedisDB:      0,
	}
}

// Validate 验证配置参数的合法性
func (c *Config) Validate() error {
	if c.DSN == "" {
		return &ConfigError{Message: "数据库DSN不能为空"}
	}

	// 验证缓存类型
	if c.CacheType != CacheTypeMemory && c.CacheType != CacheTypeRedis && c.CacheType != CacheTypeNone {
		return &ConfigError{Message: "缓存类型必须是memory、redis或none"}
	}

	// 如果是Redis缓存，检查Redis配置
	if c.CacheType == CacheTypeRedis && c.RedisAddr == "" {
		return &ConfigError{Message: "使用Redis缓存时，Redis地址不能为空"}
	}

	// 如果是内存缓存，检查大小配置
	if c.CacheType == CacheTypeMemory && c.MaxCacheSize <= 0 {
		return &ConfigError{Message: "使用内存缓存时，缓存大小必须大于0"}
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
