package runehammer

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config 引擎配置 - 规则引擎的完整配置信息
type Config struct {
	// 数据库配置
	db          *gorm.DB     // GORM数据库实例，如果提供则优先使用，跳过DSN连接创建
	dsn         string       // MySQL数据库连接字符串，当db为nil时使用
	gormConfig  *gorm.Config // GORM配置选项，用于自定义数据库行为
	autoMigrate bool         // 是否自动迁移数据库表结构
	tableName   string       // 规则表名，默认为runehammer_rules

	// 缓存配置
	cache       Cache   // 缓存接口实例，可以是Redis或内存缓存
	enableCache bool          // 是否启用缓存功能
	cacheTTL    time.Duration // 缓存生存时间
	redisAddr   string        // Redis服务器地址
	redisPass   string        // Redis密码
	redisDB     int           // Redis数据库编号

	// 定时任务配置
	syncInterval time.Duration // 规则同步间隔，用于缓存清理和规则预热

	// 日志配置
	logger interface{} // 日志接口实现，使用interface{}避免循环依赖

	// 其他配置
	maxCacheSize int // 内存缓存最大条目数
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		tableName:    "runehammer_rules",
		cacheTTL:     10 * time.Minute,
		syncInterval: 5 * time.Minute,
		autoMigrate:  false,
		maxCacheSize: 1000,
		enableCache:  true,
		redisDB:      0,
	}
}

// GetDB 获取数据库连接
func (c *Config) GetDB() *gorm.DB {
	return c.db
}

// GetCache 获取缓存实例
func (c *Config) GetCache() Cache {
	return c.cache
}

// GetLogger 获取日志实例
func (c *Config) GetLogger() interface{} {
	return c.logger
}

// GetAutoMigrate 获取自动迁移配置
func (c *Config) GetAutoMigrate() bool {
	return c.autoMigrate
}

// GetSyncInterval 获取同步间隔
func (c *Config) GetSyncInterval() time.Duration {
	return c.syncInterval
}

// Validate 验证配置
func (c *Config) Validate() error {
	// 检查数据库配置
	if c.db == nil && c.dsn == "" {
		return ErrNoDatabaseConfig
	}
	return nil
}

// SetupDB 设置数据库
func (c *Config) SetupDB() error {
	// 如果已提供DB实例，直接使用（最高优先级）
	if c.db != nil {
		return nil
	}

	// 使用DSN创建新的数据库连接
	gormCfg := c.gormConfig
	if gormCfg == nil {
		gormCfg = &gorm.Config{}
	}

	db, err := gorm.Open(mysql.Open(c.dsn), gormCfg)
	if err != nil {
		return fmt.Errorf("创建MySQL连接失败: %w", err)
	}

	c.db = db
	return nil
}

// SetupCache 初始化缓存
func (c *Config) SetupCache() error {
	if !c.enableCache {
		return nil
	}

	if c.cache != nil {
		return nil
	}

	if c.redisAddr != "" {
		client := redis.NewClient(&redis.Options{
			Addr:     c.redisAddr,
			Password: c.redisPass,
			DB:       c.redisDB,
		})
		c.cache = NewRedisCache(client)
	} else {
		// 降级到内存缓存
		c.cache = NewMemoryCache(c.maxCacheSize)
	}

	return nil
}

// ApplyOption 应用配置选项
func (c *Config) ApplyOption(opt Option) {
	opt(c)
}

// Option 配置选项函数类型
type Option func(*Config)

// WithDB 配置GORM数据库实例
func WithDB(db *gorm.DB) Option {
	return func(c *Config) {
		c.db = db
	}
}

// WithDSN 配置MySQL数据源连接字符串
func WithDSN(dsn string) Option {
	return func(c *Config) {
		c.dsn = dsn
	}
}

// WithGormConfig 配置GORM选项
func WithGormConfig(cfg *gorm.Config) Option {
	return func(c *Config) {
		c.gormConfig = cfg
	}
}

// WithAutoMigrate 开启自动数据库迁移
func WithAutoMigrate() Option {
	return func(c *Config) {
		c.autoMigrate = true
	}
}

// WithTableName 自定义规则表名
func WithTableName(name string) Option {
	return func(c *Config) {
		c.tableName = name
	}
}

// WithCache 配置缓存实现
func WithCache(cacheImpl Cache) Option {
	return func(c *Config) {
		c.cache = cacheImpl
		c.enableCache = cacheImpl != nil
	}
}

// WithRedis 配置Redis连接参数
func WithRedis(addr, password string, db int) Option {
	return func(c *Config) {
		c.redisAddr = addr
		c.redisPass = password
		c.redisDB = db
	}
}

// WithCacheTTL 配置缓存生存时间
func WithCacheTTL(ttl time.Duration) Option {
	return func(c *Config) {
		c.cacheTTL = ttl
	}
}

// WithSyncInterval 配置同步间隔
func WithSyncInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.syncInterval = interval
	}
}

// WithLogger 配置日志接口
func WithLogger(logger interface{}) Option {
	return func(c *Config) {
		c.logger = logger
	}
}

// WithMaxCacheSize 配置最大缓存大小
func WithMaxCacheSize(size int) Option {
	return func(c *Config) {
		c.maxCacheSize = size
	}
}

// WithDisableCache 禁用缓存
func WithDisableCache() Option {
	return func(c *Config) {
		c.enableCache = false
	}
}