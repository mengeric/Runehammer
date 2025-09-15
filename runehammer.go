package runehammer

import (
	"context"
	"fmt"
	"sync"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/robfig/cron/v3"
)

// ============================================================================
// 重新导出配置选项 - 对外暴露配置接口
// ============================================================================

// 配置选项已经在同一包中定义，无需重新导出

// Engine 规则引擎接口 - 提供规则执行的核心能力
//
// 泛型参数:
//
//	T - 规则执行结果的类型，支持任意类型
//
// 核心功能:
//   - 基于业务码执行规则
//   - 支持泛型结果类型
//   - 自动缓存和同步
//   - 上下文传递和超时控制
type Engine[T any] interface {
	// Exec 执行规则 - 根据业务码执行对应的规则集
	//
	// 参数:
	//   ctx     - 上下文，用于超时控制和取消操作
	//   bizCode - 业务码，用于标识规则集合
	//   input   - 输入数据，支持map、结构体或其他类型
	//
	// 返回值:
	//   T     - 规则执行结果，类型由泛型参数决定
	//   error - 执行错误
	//
	// 使用示例:
	//   engine := New[MyResult]()
	//   result, err := engine.Exec(ctx, "USER_VALIDATE", userInput)
	Exec(ctx context.Context, bizCode string, input any) (T, error)

	// Close 关闭引擎 - 释放所有资源
	//
	// 返回值:
	//   error - 关闭过程中的错误
	Close() error
}

// New 创建规则引擎实例 - 工厂方法，支持选项模式配置
//
// 泛型参数:
//
//	T - 规则执行结果的类型
//
// 参数:
//
//	opts - 配置选项，支持数据库、缓存、日志等配置
//
// 返回值:
//
//	Engine[T] - 规则引擎实例
//	error     - 创建过程中的错误
//
// 使用示例:
//
//	engine, err := New[MyResult](
//	    WithDB(db),
//	    WithRedisCache("localhost:6379", 0),
//	    WithLogger(logger),
//	)
func New[T any](opts ...Option) (Engine[T], error) {
	// 1. 初始化默认配置
	cfg := DefaultConfig()
	for _, opt := range opts {
		cfg.ApplyOption(opt)
	}

	// 2. 配置验证 - 确保必要的配置项已设置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	// 3. 设置数据库连接 - 支持DSN字符串或直接传入GORM实例
	if err := cfg.SetupDB(); err != nil {
		return nil, fmt.Errorf("数据库初始化失败: %w", err)
	}

	// 4. 设置缓存系统 - 支持Redis或内存缓存
	if err := cfg.SetupCache(); err != nil {
		return nil, fmt.Errorf("缓存初始化失败: %w", err)
	}

	// 5. 执行数据库迁移 - 自动创建规则表结构
	if cfg.GetAutoMigrate() {
		if err := cfg.GetDB().AutoMigrate(&Rule{}); err != nil {
			return nil, fmt.Errorf("数据库迁移失败: %w", err)
		}
	}

	// 6. 创建规则映射器 - 负责规则的数据库操作
	ruleMapper := NewRuleMapper(cfg.GetDB())

	// 7. 创建引擎实例
	eng := NewEngineImpl[T](
		cfg, // 直接传递Config，它现在实现了ConfigInterface
		ruleMapper,
		cfg.GetCache(),
		CacheKeyBuilder{},
		cfg.GetLogger().(Logger), // 类型断言从interface{}转换为Logger
		ast.NewKnowledgeLibrary(),
		&sync.Map{},
		cron.New(),
		false,
	)

	// 8. 启动定时同步任务 - 用于缓存清理和规则预热
	if err := eng.StartSync(); err != nil {
		return nil, fmt.Errorf("启动同步任务失败: %w", err)
	}

	return eng, nil
}
