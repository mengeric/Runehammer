package runehammer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gitee.com/damengde/runehammer/cache"
	"gitee.com/damengde/runehammer/config"
	logger "gitee.com/damengde/runehammer/logger"
	"gitee.com/damengde/runehammer/rule"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ============================================================================
// 运行时上下文 - 管理所有实例对象
// ============================================================================

// RuntimeContext 运行时上下文 - 持有所有运行时实例对象
type RuntimeContext struct {
	// 实例对象
	DB     *gorm.DB      // 数据库连接实例
	Cache  cache.Cache   // 缓存实例
	Logger logger.Logger // 日志实例

	// 组件对象
	RuleMapper rule.RuleMapper // 规则映射器

	// 配置
	config *config.Config
}

// NewRuntimeContext 创建运行时上下文
func NewRuntimeContext(cfg *config.Config) (*RuntimeContext, error) {
	ctx := newRuntimeContext(cfg)
	if err := ctx.initialize(); err != nil {
		return nil, err
	}
	return ctx, nil
}

func newRuntimeContext(cfg *config.Config) *RuntimeContext {
	return &RuntimeContext{config: cfg}
}

func (ctx *RuntimeContext) initialize() error {
	// 初始化数据库
	if ctx.DB == nil {
		if err := ctx.setupDatabase(); err != nil {
			return fmt.Errorf("数据库初始化失败: %w", err)
		}
	}

	// 初始化缓存
	if ctx.Cache == nil {
		if err := ctx.setupCache(); err != nil {
			return fmt.Errorf("缓存初始化失败: %w", err)
		}
	}

	// 初始化日志
	if ctx.Logger == nil {
		ctx.Logger = logger.NewNoopLogger()
	}

	// 初始化规则映射器
	if ctx.RuleMapper == nil {
		ctx.RuleMapper = rule.NewRuleMapper(ctx.DB)
	}

	// 执行自动迁移
	if ctx.config.AutoMigrate {
		if err := ctx.DB.AutoMigrate(&rule.Rule{}); err != nil {
			return fmt.Errorf("数据库迁移失败: %w", err)
		}
	}

	return nil
}

// setupDatabase 初始化数据库连接
func (ctx *RuntimeContext) setupDatabase() error {
	config := ctx.config

	var db *gorm.DB
	var err error

	if strings.HasPrefix(config.DSN, "sqlite:") {
		// SQLite数据库
		sqliteDSN := strings.TrimPrefix(config.DSN, "sqlite:")
		db, err = gorm.Open(sqlite.Open(sqliteDSN), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("创建SQLite连接失败: %w", err)
		}
	} else {
		// 默认MySQL数据库
		db, err = gorm.Open(mysql.Open(config.DSN), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("创建MySQL连接失败: %w", err)
		}
	}

	ctx.DB = db
	return nil
}

// setupCache 初始化缓存系统
func (ctx *RuntimeContext) setupCache() error {
	cf := ctx.config

	switch cf.CacheType {
	case config.CacheTypeRedis:
		// 创建Redis缓存
		client := redis.NewClient(&redis.Options{
			Addr:     cf.RedisAddr,
			Password: cf.RedisPassword,
			DB:       cf.RedisDB,
		})

		// 测试Redis连接
		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(pingCtx).Err(); err != nil {
			return fmt.Errorf("Redis连接失败: %w", err)
		}

		ctx.Cache = cache.NewRedisCache(client)
		return nil

	case config.CacheTypeMemory:
		// 创建内存缓存
		ctx.Cache = cache.NewMemoryCache(cf.MaxCacheSize)
		return nil

	case config.CacheTypeNone:
		// 禁用缓存，不设置Cache（保持为nil）
		ctx.Cache = nil
		return nil

	default:
		return fmt.Errorf("不支持的缓存类型: %s", cf.CacheType)
	}
}

// Close 关闭上下文中的所有资源
func (ctx *RuntimeContext) Close() error {
	var errors []error

	// 关闭缓存
	if ctx.Cache != nil {
		if err := ctx.Cache.Close(); err != nil {
			errors = append(errors, fmt.Errorf("关闭缓存失败: %w", err))
		}
	}

	// 关闭数据库连接
	if ctx.DB != nil {
		if sqlDB, err := ctx.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("关闭数据库失败: %w", err))
			}
		}
	}

	// 如果有多个错误，返回第一个
	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// GetConfig 获取配置（只读）
func (ctx *RuntimeContext) GetConfig() *config.Config {
	return ctx.config
}
