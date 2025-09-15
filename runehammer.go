package runehammer

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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
// 注意：对于需要支持多种返回类型的场景，推荐使用 BaseEngine + TypedEngine 的新方式
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

// ============================================================================
// 通用引擎接口 - 支持运行时泛型
// ============================================================================

// BaseEngine 通用引擎接口 - 不带泛型，返回原始map结果
//
// 核心功能:
//   - 启动时创建单个实例
//   - 执行任意业务规则
//   - 返回通用map结果
//   - 支持运行时类型转换
type BaseEngine interface {
	// ExecRaw 执行规则并返回原始结果
	//
	// 参数:
	//   ctx     - 上下文，用于超时控制和取消操作
	//   bizCode - 业务码，用于标识规则集合
	//   input   - 输入数据，支持map、结构体或其他类型
	//
	// 返回值:
	//   map[string]interface{} - 规则执行的原始结果
	//   error                  - 执行错误
	ExecRaw(ctx context.Context, bizCode string, input any) (map[string]interface{}, error)
	
	// Close 关闭引擎 - 释放所有资源
	Close() error
}

// TypedEngine 泛型包装器 - 将BaseEngine包装为强类型接口
//
// 泛型参数:
//   T - 目标结果类型
//
// 使用方式:
//   baseEngine := NewBaseEngine()
//   userEngine := &TypedEngine[UserResult]{base: baseEngine}
//   result, err := userEngine.Exec(ctx, "bizCode", input)
type TypedEngine[T any] struct {
	base BaseEngine
}

// Exec 执行规则并返回强类型结果
//
// 参数:
//   ctx     - 上下文，用于超时控制和取消操作  
//   bizCode - 业务码，用于标识规则集合
//   input   - 输入数据，支持map、结构体或其他类型
//
// 返回值:
//   T     - 强类型的规则执行结果
//   error - 执行错误
func (te *TypedEngine[T]) Exec(ctx context.Context, bizCode string, input any) (T, error) {
	var zero T
	
	// 1. 执行原始规则
	rawResult, err := te.base.ExecRaw(ctx, bizCode, input)
	if err != nil {
		return zero, err
	}
	
	// 2. 转换为目标类型
	return convertToType[T](rawResult)
}

// Close 关闭引擎
func (te *TypedEngine[T]) Close() error {
	return te.base.Close()
}

// NewTypedEngine 创建泛型包装器
//
// 泛型参数:
//   T - 目标结果类型
//
// 参数:
//   base - 基础引擎实例
//
// 返回值:
//   *TypedEngine[T] - 泛型包装器实例
func NewTypedEngine[T any](base BaseEngine) *TypedEngine[T] {
	return &TypedEngine[T]{base: base}
}

// ============================================================================
// 类型转换工具函数
// ============================================================================

// convertToType 将map[string]interface{}转换为指定类型
//
// 泛型参数:
//   T - 目标类型
//
// 参数:
//   rawResult - 原始map结果
//
// 返回值:
//   T     - 转换后的结果
//   error - 转换错误
func convertToType[T any](rawResult map[string]interface{}) (T, error) {
	var zero T
	
	// 1. 如果目标类型就是map[string]interface{}，直接返回
	if result, ok := any(rawResult).(T); ok {
		return result, nil
	}
	
	// 2. 如果目标类型是map[string]any，转换并返回
	if _, ok := any(zero).(map[string]any); ok {
		converted := make(map[string]any)
		for k, v := range rawResult {
			converted[k] = v
		}
		if result, ok := any(converted).(T); ok {
			return result, nil
		}
	}
	
	// 3. 尝试JSON序列化/反序列化进行结构体转换
	return convertMapToStruct[T](rawResult)
}

// convertMapToStruct 将map[string]interface{}转换为结构体
//
// 泛型参数:
//   T - 目标结构体类型
//
// 参数:
//   rawResult - 原始map结果
//
// 返回值:
//   T     - 转换后的结构体
//   error - 转换错误
func convertMapToStruct[T any](rawResult map[string]interface{}) (T, error) {
	var zero T
	
	// 1. 检查目标类型
	targetType := reflect.TypeOf(zero)
	if targetType == nil {
		return zero, fmt.Errorf("无法确定目标类型")
	}
	
	// 2. 如果是接口类型，直接返回map
	if targetType.Kind() == reflect.Interface {
		if result, ok := any(rawResult).(T); ok {
			return result, nil
		}
		return zero, fmt.Errorf("接口类型转换失败")
	}
	
	// 3. 使用JSON序列化/反序列化进行转换
	jsonData, err := json.Marshal(rawResult)
	if err != nil {
		return zero, fmt.Errorf("JSON序列化失败: %w", err)
	}
	
	var result T
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return zero, fmt.Errorf("JSON反序列化失败: %w", err)
	}
	
	return result, nil
}

// NewBaseEngine 创建通用基础引擎实例
//
// 参数:
//   opts - 配置选项，支持数据库、缓存、日志等配置
//
// 返回值:
//   BaseEngine - 基础引擎实例
//   error      - 创建过程中的错误
//
// 使用示例:
//   baseEngine, err := NewBaseEngine(
//       WithDB(db),
//       WithRedisCache("localhost:6379", 0),
//       WithLogger(logger),
//   )
func NewBaseEngine(opts ...Option) (BaseEngine, error) {
	// 使用map[string]interface{}作为内部类型创建引擎
	engine, err := New[map[string]interface{}](opts...)
	if err != nil {
		return nil, err
	}
	
	// 包装为BaseEngine接口
	return &baseEngineWrapper{engine: engine}, nil
}

// baseEngineWrapper BaseEngine接口的实现
type baseEngineWrapper struct {
	engine Engine[map[string]interface{}]
}

// ExecRaw 实现BaseEngine接口
func (w *baseEngineWrapper) ExecRaw(ctx context.Context, bizCode string, input any) (map[string]interface{}, error) {
	return w.engine.Exec(ctx, bizCode, input)
}

// Close 实现BaseEngine接口
func (w *baseEngineWrapper) Close() error {
	return w.engine.Close()
}

// New 创建规则引擎实例 - 工厂方法，支持选项模式配置
//
// 注意：对于需要支持多种返回类型的场景，推荐使用 NewBaseEngine + NewTypedEngine 的新方式
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
