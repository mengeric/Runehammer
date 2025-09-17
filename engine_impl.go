package runehammer

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitee.com/damengde/runehammer/config"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	"github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
	"github.com/robfig/cron/v3"
)

// ============================================================================
// 引擎实现结构体 - 内部实现细节
// ============================================================================

// engineImpl 规则引擎的具体实现 - 包含所有运行时状态和依赖
type engineImpl[T any] struct {
	// 核心配置和依赖
	config    *config.Config  // 引擎配置信息 - 使用config包的Config类型
	mapper    RuleMapper     // 规则数据访问接口
	cache     Cache           // 缓存接口（Redis或内存）
	cacheKeys CacheKeyBuilder // 缓存键构建器
	logger    Logger        // 日志接口 - 使用interface{}避免循环依赖

	// Grule引擎相关
	knowledgeLibrary *ast.KnowledgeLibrary // Grule知识库
	knowledgeBases   *sync.Map             // 编译后的知识库缓存

	// 系统状态管理
	cron   *cron.Cron   // 定时任务调度器
	closed bool         // 引擎是否已关闭
	mutex  sync.RWMutex // 读写锁保护
}

// NewEngineImpl 创建引擎实例
func NewEngineImpl[T any](
	cfg *config.Config,  // 使用config包的Config类型
	mapper RuleMapper,
	cache Cache,
	cacheKeys CacheKeyBuilder,
	logger Logger,
	knowledgeLibrary *ast.KnowledgeLibrary,
	knowledgeBases *sync.Map,
	cron *cron.Cron,
	closed bool,
) *engineImpl[T] {
	if knowledgeBases == nil {
		knowledgeBases = &sync.Map{}
	}
	
	return &engineImpl[T]{
		config:           cfg,  // 直接赋值config包的Config
		mapper:           mapper,
		cache:            cache,
		cacheKeys:        cacheKeys,
		logger:           logger,
		knowledgeLibrary: knowledgeLibrary,
		knowledgeBases:   knowledgeBases,
		cron:             cron,
		closed:           closed,
		mutex:            sync.RWMutex{},
	}
}

// Exec 规则执行器的核心方法 - 根据业务码执行对应的GRL规则集
func (e *engineImpl[T]) Exec(ctx context.Context, bizCode string, input any) (T, error) {
	var zero T

	// 1. 检查引擎状态
	e.mutex.RLock()
	if e.closed {
		e.mutex.RUnlock()
		return zero, fmt.Errorf("未定义错误: 引擎已关闭")
	}
	e.mutex.RUnlock()

	// 2. 参数验证
	if strings.TrimSpace(bizCode) == "" {
		return zero, fmt.Errorf("未定义错误: 无效的业务码")
	}
	if input == nil {
		return zero, fmt.Errorf("未定义错误: 输入参数为空")
	}

	// 3. 获取规则
	rules, err := e.getRules(ctx, bizCode)
	if err != nil {
		if e.logger != nil {
			e.logger.Errorf(ctx, "获取规则失败", "bizCode", bizCode, "error", err)
		}
		return zero, fmt.Errorf("未定义错误: 规则未找到")
	}

	if len(rules) == 0 {
		if e.logger != nil {
			e.logger.Warnf(ctx, "未找到有效规则", "bizCode", bizCode)
		}
		return zero, fmt.Errorf("未定义错误: 规则未找到")
	}

	// 4. 编译规则
	knowledgeBase, err := e.compileRules(bizCode, rules)
	if err != nil {
		if e.logger != nil {
			e.logger.Errorf(ctx, "规则编译失败", "bizCode", bizCode, "error", err)
		}
		return zero, fmt.Errorf("规则编译失败: %w", err)
	}

	// 5. 创建数据上下文和规则引擎
	dataCtx := ast.NewDataContext()
	ruleEngine := engine.NewGruleEngine()

	// 6. 注入输入数据
	if err := e.injectInputData(dataCtx, input); err != nil {
		if e.logger != nil {
			e.logger.Errorf(ctx, "数据注入失败", "bizCode", bizCode, "error", err)
		}
		return zero, fmt.Errorf("数据注入失败: %w", err)
	}

	// 7. 注入内置函数
	e.injectBuiltinFunctions(dataCtx)

	// 8. 执行规则
	if knowledgeBase == nil {
		if e.logger != nil {
			e.logger.Errorf(ctx, "知识库为空", "bizCode", bizCode)
		}
		return zero, fmt.Errorf("知识库为空")
	}
	
	if err := ruleEngine.Execute(dataCtx, knowledgeBase); err != nil {
		if e.logger != nil {
			e.logger.Errorf(ctx, "规则执行失败", "bizCode", bizCode, "error", err)
		}
		return zero, fmt.Errorf("规则执行失败: %w", err)
	}

	// 9. 提取结果
	result, err := e.extractResult(dataCtx)
	if err != nil {
		if e.logger != nil {
			e.logger.Errorf(ctx, "结果提取失败", "bizCode", bizCode, "error", err)
		}
		return zero, fmt.Errorf("结果提取失败: %w", err)
	}

	return result, nil
}

// ============================================================================
// 规则获取和缓存管理
// ============================================================================

// getRules 获取规则 - 支持缓存机制和数据库回退
func (e *engineImpl[T]) getRules(ctx context.Context, bizCode string) ([]*Rule, error) {
	// 1. 尝试从缓存获取
	if e.cache != nil {
		cacheKey := e.cacheKeys.RuleKey(bizCode)
		data, err := e.cache.Get(ctx, cacheKey)
		if err == nil {
			// 反序列化缓存数据
			var cacheItem RuleCacheItem
			if err := cacheItem.FromBytes(data); err == nil {
				if e.logger != nil {
					e.logger.Debugf(ctx, "从缓存获取规则成功", "bizCode", bizCode, "count", len(cacheItem.Rules))
				}
				return cacheItem.Rules, nil
			}
		}
	}

	// 2. 从数据库获取
	rules, err := e.mapper.FindByBizCode(ctx, bizCode)
	if err != nil {
		return nil, err
	}

	// 3. 更新缓存
	if e.cache != nil && len(rules) > 0 {
		cacheItem := RuleCacheItem{
			Rules:     rules,
			UpdatedAt: time.Now(),
			Version:   1,
		}
		if data, err := cacheItem.ToBytes(); err == nil {
			cacheKey := e.cacheKeys.RuleKey(bizCode)
			// 缓存1小时
			if err := e.cache.Set(ctx, cacheKey, data, time.Hour); err != nil && e.logger != nil {
				e.logger.Warnf(ctx, "规则缓存更新失败", "bizCode", bizCode, "error", err)
			}
		}
	}

	return rules, nil
}

// compileRules 编译规则 - 将GRL规则转换为可执行的知识库
func (e *engineImpl[T]) compileRules(bizCode string, rules []*Rule) (*ast.KnowledgeBase, error) {
	// 检查是否已编译缓存
	if kb, ok := e.knowledgeBases.Load(bizCode); ok {
		return kb.(*ast.KnowledgeBase), nil
	}

	// 使用互斥锁保护编译过程，防止并发编译同一个业务码的规则
	e.mutex.Lock()
	defer e.mutex.Unlock()
	
	// 双重检查，防止在等待锁的过程中其他协程已经编译完成
	if kb, ok := e.knowledgeBases.Load(bizCode); ok {
		return kb.(*ast.KnowledgeBase), nil
	}

	// 创建新的知识库
	if e.knowledgeLibrary == nil {
		return nil, fmt.Errorf("知识库库为空")
	}

	// 编译每个规则
	for _, rule := range rules {
		if !rule.Enabled {
			continue // 跳过禁用的规则
		}

		// 创建字节数组资源
		ruleBytes := pkg.NewBytesResource([]byte(rule.GRL))

		// 构建规则
		ruleBuilder := builder.NewRuleBuilder(e.knowledgeLibrary)
		if err := ruleBuilder.BuildRuleFromResource(bizCode, "1.0.0", ruleBytes); err != nil {
			return nil, fmt.Errorf("编译规则 %s 失败: %w", rule.Name, err)
		}
	}

	// 从knowledge library中获取构建好的知识库
	knowledgeBase, err := e.knowledgeLibrary.NewKnowledgeBaseInstance(bizCode, "1.0.0")
	if err != nil {
		return nil, fmt.Errorf("获取知识库实例失败: %w", err)
	}
	if knowledgeBase == nil {
		return nil, fmt.Errorf("知识库实例为空")
	}

	// 缓存编译结果
	e.knowledgeBases.Store(bizCode, knowledgeBase)

	return knowledgeBase, nil
}

// ============================================================================
// 引擎生命周期管理
// ============================================================================

// Close 关闭引擎 - 释放所有资源
func (e *engineImpl[T]) Close() error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if e.closed {
		return nil
	}

	// 停止定时任务
	if e.cron != nil {
		e.cron.Stop()
	}

	// 关闭缓存连接
	if e.cache != nil {
		if err := e.cache.Close(); err != nil && e.logger != nil {
			e.logger.Warnf(context.Background(), "关闭缓存连接失败", "error", err)
		}
	}

	e.closed = true

	if e.logger != nil {
		e.logger.Infof(context.Background(), "规则引擎已关闭")
	}

	return nil
}
