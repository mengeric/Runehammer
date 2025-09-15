package runehammer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
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
	cache       Cache         // 缓存接口实例，可以是Redis或内存缓存
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

	// 动态引擎配置
	dynamicConfig *DynamicConfig // 动态引擎配置选项
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
	
	// 根据DSN类型选择不同的驱动
	var db *gorm.DB
	var err error
	
	if strings.HasPrefix(c.dsn, "sqlite:") {
		// SQLite数据库
		sqliteDSN := strings.TrimPrefix(c.dsn, "sqlite:")
		db, err = gorm.Open(sqlite.Open(sqliteDSN), gormCfg)
		if err != nil {
			return fmt.Errorf("创建SQLite连接失败: %w", err)
		}
	} else {
		// 默认MySQL数据库
		db, err = gorm.Open(mysql.Open(c.dsn), gormCfg)
		if err != nil {
			return fmt.Errorf("创建MySQL连接失败: %w", err)
		}
	}
	
	c.db = db
	return nil
}

// SetupCache 初始化缓存 - 启动时确定唯一缓存策略
func (c *Config) SetupCache() error {
	if !c.enableCache {
		c.cache = nil
		return nil
	}

	// 如果已经手动设置了缓存实例，直接使用
	if c.cache != nil {
		return nil
	}

	// 启动时确定缓存策略，优先级：Redis > Memory > None
	if c.redisAddr != "" {
		// 尝试连接Redis
		client := redis.NewClient(&redis.Options{
			Addr:     c.redisAddr,
			Password: c.redisPass,
			DB:       c.redisDB,
		})

		// 测试Redis连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			// Redis连接失败，返回错误而不是降级
			return fmt.Errorf("Redis连接失败: %w", err)
		}

		c.cache = NewRedisCache(client)
		return nil
	}

	// 使用内存缓存
	c.cache = NewMemoryCache(c.maxCacheSize)
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

// ============================================================================
// 动态引擎配置 - 用于动态规则生成和转换
// ============================================================================

// DynamicConfig 动态引擎配置
type DynamicConfig struct {
	// 转换器配置
	ConverterConfig ConverterConfig `json:"converter_config" yaml:"converter_config"`

	// 表达式解析器配置
	ParserConfig ParserConfig `json:"parser_config" yaml:"parser_config"`

	// 缓存配置
	CacheConfig DynamicCacheConfig `json:"cache_config" yaml:"cache_config"`

	// 验证器配置
	ValidatorConfig ValidatorConfig `json:"validator_config" yaml:"validator_config"`

	// 执行配置
	ExecutionConfig ExecutionConfig `json:"execution_config" yaml:"execution_config"`

	// 自定义函数配置
	CustomFunctions map[string]interface{} `json:"custom_functions" yaml:"custom_functions"`
}

// ParserConfig 表达式解析器配置
type ParserConfig struct {
	// 默认语法类型
	DefaultSyntax SyntaxType `json:"default_syntax" yaml:"default_syntax"`

	// 支持的语法类型
	SupportedSyntax []SyntaxType `json:"supported_syntax" yaml:"supported_syntax"`

	// 自定义操作符映射
	CustomOperators map[string]string `json:"custom_operators" yaml:"custom_operators"`

	// 自定义函数映射
	CustomFunctionMappings map[string]string `json:"custom_function_mappings" yaml:"custom_function_mappings"`

	// 自定义关键字映射
	CustomKeywords map[string]string `json:"custom_keywords" yaml:"custom_keywords"`
}

// DynamicCacheConfig 动态缓存配置
type DynamicCacheConfig struct {
	// 是否启用缓存
	Enabled bool `json:"enabled" yaml:"enabled"`

	// 缓存大小限制
	MaxSize int `json:"max_size" yaml:"max_size"`

	// 缓存TTL
	TTL time.Duration `json:"ttl" yaml:"ttl"`

	// 是否启用LRU淘汰
	EnableLRU bool `json:"enable_lru" yaml:"enable_lru"`

	// 缓存清理间隔
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// ValidatorConfig 验证器配置
type ValidatorConfig struct {
	// 是否启用验证
	Enabled bool `json:"enabled" yaml:"enabled"`

	// 严格模式
	StrictMode bool `json:"strict_mode" yaml:"strict_mode"`

	// 自定义验证规则
	CustomValidators map[string]interface{} `json:"custom_validators" yaml:"custom_validators"`
}

// ExecutionConfig 执行配置
type ExecutionConfig struct {
	// 是否启用并行执行
	EnableParallel bool `json:"enable_parallel" yaml:"enable_parallel"`

	// 并发数量限制
	MaxConcurrency int `json:"max_concurrency" yaml:"max_concurrency"`

	// 执行超时时间
	ExecutionTimeout time.Duration `json:"execution_timeout" yaml:"execution_timeout"`

	// 是否启用规则优先级
	EnablePriority bool `json:"enable_priority" yaml:"enable_priority"`

	// 最大规则数量限制
	MaxRules int `json:"max_rules" yaml:"max_rules"`
}

// DefaultDynamicConfig 默认动态配置
func DefaultDynamicConfig() *DynamicConfig {
	return &DynamicConfig{
		ConverterConfig: ConverterConfig{
			VariablePrefix: map[string]string{
				"customer": "customer",
				"order":    "order",
				"user":     "user",
				"data":     "data",
				"result":   "result",
			},
			OperatorMapping: map[string]string{
				"==":       "==",
				"!=":       "!=",
				">":        ">",
				"<":        "<",
				">=":       ">=",
				"<=":       "<=",
				"and":      "&&",
				"or":       "||",
				"not":      "!",
				"in":       "Contains",
				"contains": "Contains",
				"matches":  "Matches",
				"between":  "BETWEEN",
			},
			FunctionMapping: map[string]string{
				"now":          "Now()",
				"today":        "Today()",
				"nowMillis":    "NowMillis()",
				"timeToMillis": "TimeToMillis",
				"millisToTime": "MillisToTime",
				"daysBetween":  "DaysBetween",
				"sum":          "Sum",
				"avg":          "Avg",
				"max":          "Max",
				"min":          "Min",
				"count":        "Count",
			},
			DefaultPriority: 50,
			StrictMode:      false,
		},
		ParserConfig: ParserConfig{
			DefaultSyntax: SyntaxTypeSQL,
			SupportedSyntax: []SyntaxType{
				SyntaxTypeSQL,
				SyntaxTypeJavaScript,
			},
			CustomOperators:        make(map[string]string),
			CustomFunctionMappings: make(map[string]string),
			CustomKeywords:         make(map[string]string),
		},
		CacheConfig: DynamicCacheConfig{
			Enabled:         true,
			MaxSize:         500,
			TTL:             5 * time.Minute,
			EnableLRU:       true,
			CleanupInterval: 1 * time.Minute,
		},
		ValidatorConfig: ValidatorConfig{
			Enabled:          true,
			StrictMode:       false,
			CustomValidators: make(map[string]interface{}),
		},
		ExecutionConfig: ExecutionConfig{
			EnableParallel:   true,
			MaxConcurrency:   10,
			ExecutionTimeout: 30 * time.Second,
			EnablePriority:   true,
			MaxRules:         100,
		},
		CustomFunctions: make(map[string]interface{}),
	}
}

// GetDynamicConfig 获取动态配置
func (c *Config) GetDynamicConfig() *DynamicConfig {
	if c.dynamicConfig == nil {
		c.dynamicConfig = DefaultDynamicConfig()
	}
	return c.dynamicConfig
}

// ============================================================================
// 动态配置选项函数
// ============================================================================

// WithDynamicConfig 设置动态配置
func WithDynamicConfig(config *DynamicConfig) Option {
	return func(c *Config) {
		c.dynamicConfig = config
	}
}

// WithConverterConfig 设置转换器配置
func WithConverterConfig(config ConverterConfig) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ConverterConfig = config
	}
}

// WithParserConfig 设置解析器配置
func WithParserConfig(config ParserConfig) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ParserConfig = config
	}
}

// WithDynamicCacheConfig 设置动态缓存配置
func WithDynamicCacheConfig(config DynamicCacheConfig) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.CacheConfig = config
	}
}

// WithValidatorConfig 设置验证器配置
func WithValidatorConfig(config ValidatorConfig) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ValidatorConfig = config
	}
}

// WithExecutionConfig 设置执行配置
func WithExecutionConfig(config ExecutionConfig) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ExecutionConfig = config
	}
}

// WithCustomFunctions 设置自定义函数
func WithCustomFunctions(functions map[string]interface{}) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.CustomFunctions = functions
	}
}

// WithDefaultSyntax 设置默认语法
func WithDefaultSyntax(syntax SyntaxType) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ParserConfig.DefaultSyntax = syntax
	}
}

// WithSupportedSyntax 设置支持的语法类型
func WithSupportedSyntax(syntaxes ...SyntaxType) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ParserConfig.SupportedSyntax = syntaxes
	}
}

// WithCustomOperators 设置自定义操作符
func WithCustomOperators(operators map[string]string) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ParserConfig.CustomOperators = operators
	}
}

// WithExecutionTimeout 设置执行超时时间
func WithExecutionTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ExecutionConfig.ExecutionTimeout = timeout
	}
}

// WithMaxConcurrency 设置最大并发数
func WithMaxConcurrency(max int) Option {
	return func(c *Config) {
		if c.dynamicConfig == nil {
			c.dynamicConfig = DefaultDynamicConfig()
		}
		c.dynamicConfig.ExecutionConfig.MaxConcurrency = max
	}
}
