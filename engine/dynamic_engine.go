package engine

//go:generate mockgen -source=dynamic_engine.go -destination=dynamic_engine_mock.go -package=engine

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	logger "gitee.com/damengde/runehammer/logger"
	"gitee.com/damengde/runehammer/rule"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/hyperjumptech/grule-rule-engine/builder"
	grengine "github.com/hyperjumptech/grule-rule-engine/engine"
	"github.com/hyperjumptech/grule-rule-engine/pkg"
)

// ============================================================================
// 动态规则引擎 - 实时生成和执行规则，不依赖数据库
// ============================================================================

// DynamicEngine 动态规则引擎
type DynamicEngine[T any] struct {
	converter        rule.RuleConverter     // 规则转换器
	knowledgeLibrary *ast.KnowledgeLibrary  // Grule知识库
	customFunctions  map[string]interface{} // 自定义函数库
	validators       []RuleValidator        // 规则验证器
	logger           logger.Logger          // 日志记录器
	cache            *DynamicRuleCache      // 规则缓存（可选）
	config           DynamicEngineConfig    // 引擎配置
}

// DynamicEngineConfig 动态引擎配置
type DynamicEngineConfig struct {
	EnableCache       bool          // 是否启用缓存
	CacheTTL          time.Duration // 缓存过期时间
	MaxCacheSize      int           // 最大缓存大小
	StrictValidation  bool          // 是否严格验证
	ParallelExecution bool          // 是否支持并行执行
	DefaultTimeout    time.Duration // 默认超时时间
}

// RuleValidator 规则验证器接口
type RuleValidator interface {
	Validate(definition interface{}) []rule.ValidationError
}

// DynamicRuleCache 动态规则缓存
type DynamicRuleCache struct {
	cache   map[string]*CachedRule
	mu      sync.RWMutex
	ttl     time.Duration
	maxSize int
	size    int
}

// CachedRule 缓存的规则
type CachedRule struct {
	KB        *ast.KnowledgeBase
	Hash      string
	CreatedAt time.Time
	HitCount  int64
}

// NewDynamicEngine 创建动态规则引擎
func NewDynamicEngine[T any](config ...DynamicEngineConfig) *DynamicEngine[T] {
	// 默认配置
	defaultConfig := DynamicEngineConfig{
		EnableCache:       true,
		CacheTTL:          30 * time.Minute,
		MaxCacheSize:      1000,
		StrictValidation:  false,
		ParallelExecution: true,
		DefaultTimeout:    30 * time.Second,
	}

	if len(config) > 0 {
		defaultConfig = config[0]
	}

	engine := &DynamicEngine[T]{
		converter:        rule.NewGRLConverter(),
		knowledgeLibrary: ast.NewKnowledgeLibrary(),
		customFunctions:  make(map[string]interface{}),
		validators:       []RuleValidator{},
		config:           defaultConfig,
	}

	// 初始化缓存
	if defaultConfig.EnableCache {
		engine.cache = NewDynamicRuleCache(defaultConfig.CacheTTL, defaultConfig.MaxCacheSize)
	}

	return engine
}

// ExecuteRuleDefinition 执行规则定义
func (e *DynamicEngine[T]) ExecuteRuleDefinition(
	ctx context.Context,
	definition interface{},
	input any,
) (T, error) {
	var zero T

	// 1. 验证规则定义
	if e.config.StrictValidation {
		if err := e.validateRuleDefinition(definition); err != nil {
			return zero, fmt.Errorf("规则验证失败: %w", err)
		}
	}

	// 2. 生成规则hash用于缓存
	ruleHash := e.calculateRuleHash(definition)

	// 3. 检查缓存
	var knowledgeBase *ast.KnowledgeBase
	var err error

	if e.cache != nil {
		if cached := e.cache.Get(ruleHash); cached != nil {
			knowledgeBase = cached.KB
			cached.HitCount++
			if e.logger != nil {
				e.logger.Debugf(ctx, "使用缓存的规则", "hash", ruleHash, "hitCount", cached.HitCount)
			}
		}
	}

	// 4. 如果缓存未命中，编译规则
	if knowledgeBase == nil {
		// 转换为GRL
		grl, convErr := e.converter.ConvertToGRL(definition)
		if convErr != nil {
			return zero, fmt.Errorf("规则转换失败: %w", convErr)
		}

		// 编译GRL
		knowledgeBase, err = e.compileGRL(grl, ruleHash)
		if err != nil {
			return zero, fmt.Errorf("规则编译失败: %w", err)
		}

		// 存入缓存
		if e.cache != nil {
			e.cache.Set(ruleHash, &CachedRule{
				KB:        knowledgeBase,
				Hash:      ruleHash,
				CreatedAt: time.Now(),
				HitCount:  1,
			})
		}
	}

	// 5. 执行规则
	return e.executeWithKnowledgeBase(ctx, knowledgeBase, input)
}

// ExecuteBatch 批量执行多个规则
func (e *DynamicEngine[T]) ExecuteBatch(
	ctx context.Context,
	definitions []interface{},
	input any,
) ([]T, error) {
	if !e.config.ParallelExecution {
		return e.executeBatchSequential(ctx, definitions, input)
	}

	return e.executeBatchParallel(ctx, definitions, input)
}

// ExecuteWithTimeout 带超时的规则执行
func (e *DynamicEngine[T]) ExecuteWithTimeout(
	ctx context.Context,
	definition interface{},
	input any,
	timeout time.Duration,
) (T, error) {
	if timeout == 0 {
		timeout = e.config.DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return e.ExecuteRuleDefinition(ctx, definition, input)
}

// RegisterCustomFunction 注册自定义函数
func (e *DynamicEngine[T]) RegisterCustomFunction(name string, fn interface{}) {
	e.customFunctions[name] = fn
}

// RegisterCustomFunctions 批量注册自定义函数
func (e *DynamicEngine[T]) RegisterCustomFunctions(functions map[string]interface{}) {
	for name, fn := range functions {
		e.customFunctions[name] = fn
	}
}

// RegisterValidator 注册验证器
func (e *DynamicEngine[T]) RegisterValidator(validator RuleValidator) {
	e.validators = append(e.validators, validator)
}

// SetLogger 设置日志记录器
func (e *DynamicEngine[T]) SetLogger(logger logger.Logger) {
	e.logger = logger
}

// GetCacheStats 获取缓存统计信息
func (e *DynamicEngine[T]) GetCacheStats() CacheStats {
	if e.cache == nil {
		return CacheStats{}
	}

	return e.cache.GetStats()
}

// ClearCache 清空缓存
func (e *DynamicEngine[T]) ClearCache() {
	if e.cache != nil {
		e.cache.Clear()
	}
}

// ============================================================================
// 内部实现方法
// ============================================================================

// compileGRL 编译GRL规则
func (e *DynamicEngine[T]) compileGRL(grl, ruleID string) (*ast.KnowledgeBase, error) {
	// 创建规则资源
	ruleBytes := pkg.NewBytesResource([]byte(grl))

	// 构建规则到知识库
	ruleBuilder := builder.NewRuleBuilder(e.knowledgeLibrary)
	if err := ruleBuilder.BuildRuleFromResource(ruleID, "1.0.0", ruleBytes); err != nil {
		return nil, fmt.Errorf("构建规则失败: %w", err)
	}

	// 获取知识库实例
	knowledgeBase, err := e.knowledgeLibrary.NewKnowledgeBaseInstance(ruleID, "1.0.0")
	if err != nil {
		return nil, fmt.Errorf("创建知识库实例失败: %w", err)
	}

	return knowledgeBase, nil
}

// executeWithKnowledgeBase 使用知识库执行规则
func (e *DynamicEngine[T]) executeWithKnowledgeBase(
	ctx context.Context,
	knowledgeBase *ast.KnowledgeBase,
	input any,
) (T, error) {
	var zero T

	// 创建数据上下文
	dataCtx := ast.NewDataContext()

	// 注入输入数据
	if err := e.injectInputData(dataCtx, input); err != nil {
		return zero, fmt.Errorf("数据注入失败: %w", err)
	}

	// 注入内置函数
	e.injectBuiltinFunctions(dataCtx)

	// 注入自定义函数
	e.injectCustomFunctions(dataCtx)

	// 创建规则引擎
    ruleEngine := grengine.NewGruleEngine()

	// 验证知识库不为空
	if knowledgeBase == nil {
		return zero, fmt.Errorf("知识库为空")
	}

	// 执行规则
	if err := ruleEngine.Execute(dataCtx, knowledgeBase); err != nil {
		return zero, fmt.Errorf("规则执行失败: %w", err)
	}

	// 提取结果
	return e.extractResult(dataCtx)
}

// executeBatchSequential 顺序批量执行
func (e *DynamicEngine[T]) executeBatchSequential(
	ctx context.Context,
	definitions []interface{},
	input any,
) ([]T, error) {
	var results []T

	for i, def := range definitions {
		result, err := e.ExecuteRuleDefinition(ctx, def, input)
		if err != nil {
			if e.logger != nil {
				e.logger.Warnf(ctx, "规则执行失败，跳过", "index", i, "error", err)
			}
			var zero T
			results = append(results, zero)
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// executeBatchParallel 并行批量执行
func (e *DynamicEngine[T]) executeBatchParallel(
	ctx context.Context,
	definitions []interface{},
	input any,
) ([]T, error) {
	var wg sync.WaitGroup
	results := make([]T, len(definitions))
	errors := make([]error, len(definitions))

	for i, def := range definitions {
		wg.Add(1)
		go func(idx int, definition interface{}) {
			defer wg.Done()
			results[idx], errors[idx] = e.ExecuteRuleDefinition(ctx, definition, input)
		}(i, def)
	}

	wg.Wait()

	// 记录错误
	for i, err := range errors {
		if err != nil && e.logger != nil {
			e.logger.Warnf(ctx, "并行规则执行失败", "index", i, "error", err)
		}
	}

	return results, nil
}

// injectInputData 注入输入数据 - 将各种类型的输入数据注入到执行上下文
//
// 变量注入规则:
//  1. 结构体类型：作为单个对象注入，使用类型名（小写）作为变量名
//  2. 匿名结构体和其他类型：统一以"Params"名称注入
//
// 注意：不支持 map[string]interface{} 类型，请使用结构体替代
//
// 参数:
//
//	dataCtx - Grule数据上下文
//	input   - 输入数据，支持结构体和基本类型
//
// 返回值:
//
//	error - 注入过程中的错误
func (e *DynamicEngine[T]) injectInputData(dataCtx ast.IDataContext, input any) error {
	// 首先初始化Result变量作为一个空的map
	result := make(map[string]any)
	if err := dataCtx.Add("Result", result); err != nil {
		return fmt.Errorf("注入Result变量失败: %w", err)
	}

	v := reflect.ValueOf(input)
	t := reflect.TypeOf(input)

	// 处理指针类型，获取实际的值和类型
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		return fmt.Errorf("不支持 map 类型，请使用结构体替代")
	case reflect.Struct:
		return e.injectStructData(dataCtx, input, t)
	default:
		return e.injectDefaultData(dataCtx, input)
	}
}

// injectStructData 注入结构体数据 - 将整个结构体作为单个对象注入
func (e *DynamicEngine[T]) injectStructData(dataCtx ast.IDataContext, input any, t reflect.Type) error {
	// 统一使用Params作为输入变量名，保持与引擎一致
	inputName := "Params"

	if err := dataCtx.Add(inputName, input); err != nil {
		return fmt.Errorf("注入结构体 %s 失败: %w", inputName, err)
	}

	return nil
}

// injectDefaultData 注入其他类型数据 - 直接以Params名称注入
func (e *DynamicEngine[T]) injectDefaultData(dataCtx ast.IDataContext, input any) error {
	if err := dataCtx.Add("Params", input); err != nil {
		return fmt.Errorf("注入Params变量失败: %w", err)
	}
	return nil
}

// injectBuiltinFunctions 注入内置函数
func (e *DynamicEngine[T]) injectBuiltinFunctions(dataCtx ast.IDataContext) {
	// 注入时间函数
	dataCtx.Add("Now", func() time.Time {
		return time.Now()
	})

	dataCtx.Add("Today", func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	})

	// 注入数学函数
	dataCtx.Add("Max", func(a, b float64) float64 {
		if a > b {
			return a
		}
		return b
	})

	dataCtx.Add("Min", func(a, b float64) float64 {
		if a < b {
			return a
		}
		return b
	})

	// 注入字符串函数
	dataCtx.Add("Contains", func(s, substr string) bool {
		return strings.Contains(s, substr)
	})

	// 注入集合函数
	dataCtx.Add("Len", func(obj interface{}) int {
		switch v := obj.(type) {
		case string:
			return len(v)
		case []interface{}:
			return len(v)
		case map[string]interface{}:
			return len(v)
		default:
			return 0
		}
	})
}

// injectCustomFunctions 注入自定义函数
func (e *DynamicEngine[T]) injectCustomFunctions(dataCtx ast.IDataContext) {
	for name, fn := range e.customFunctions {
		dataCtx.Add(name, fn)
	}
}

// extractResult 提取结果
func (e *DynamicEngine[T]) extractResult(dataCtx ast.IDataContext) (T, error) {
	var zero T

	// 获取结果变量
	resultVar := dataCtx.Get("Result")
	if resultVar == nil {
		return zero, fmt.Errorf("未找到结果变量")
	}

	// 获取结果值
	resultValue, err := resultVar.GetValue()
	if err != nil {
		return zero, fmt.Errorf("获取结果值失败: %w", err)
	}

	// 获取实际的interface{}值
	actualData := resultValue.Interface()

	// 根据泛型类型进行相应的转换
	var result T
	resultType := reflect.TypeOf(result)

	if resultType.Kind() == reflect.Map {
		// 如果目标类型是map，直接转换
		if mapResult, ok := actualData.(map[string]interface{}); ok {
			if typedResult, ok := interface{}(mapResult).(T); ok {
				return typedResult, nil
			}
		}
	}

	// 尝试直接类型转换
	if typedResult, ok := actualData.(T); ok {
		return typedResult, nil
	}

	return zero, fmt.Errorf("结果类型转换失败，期望类型 %T，实际类型 %T", zero, actualData)
}

// validateRuleDefinition 验证规则定义
func (e *DynamicEngine[T]) validateRuleDefinition(definition interface{}) error {
	for _, validator := range e.validators {
		errors := validator.Validate(definition)
		if len(errors) > 0 {
			return fmt.Errorf("验证失败: %s", errors[0].Message)
		}
	}

	return e.converter.Validate(definition)
}

// calculateRuleHash 计算规则hash
func (e *DynamicEngine[T]) calculateRuleHash(definition interface{}) string {
	// 将规则定义序列化为字符串
	var data string

	switch def := definition.(type) {
	case string:
		data = def
	default:
		// 尝试JSON序列化
		if jsonData, err := json.Marshal(definition); err == nil {
			data = string(jsonData)
		} else {
			data = fmt.Sprintf("%+v", definition)
		}
	}

	// 计算SHA256哈希
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ============================================================================
// 缓存实现
// ============================================================================

// NewDynamicRuleCache 创建动态规则缓存
func NewDynamicRuleCache(ttl time.Duration, maxSize int) *DynamicRuleCache {
	return &DynamicRuleCache{
		cache:   make(map[string]*CachedRule),
		ttl:     ttl,
		maxSize: maxSize,
		size:    0,
	}
}

// Get 获取缓存的规则
func (c *DynamicRuleCache) Get(hash string) *CachedRule {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, ok := c.cache[hash]
	if !ok {
		return nil
	}

	// 检查是否过期
	if time.Since(cached.CreatedAt) > c.ttl {
		// 异步清理过期项
		go func() {
			c.mu.Lock()
			delete(c.cache, hash)
			c.size--
			c.mu.Unlock()
		}()
		return nil
	}

	return cached
}

// Set 设置缓存
func (c *DynamicRuleCache) Set(hash string, rule *CachedRule) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查容量
	if c.size >= c.maxSize {
		c.evictLRU()
	}

	c.cache[hash] = rule
	c.size++
}

// Clear 清空缓存
func (c *DynamicRuleCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CachedRule)
	c.size = 0
}

// GetStats 获取缓存统计
func (c *DynamicRuleCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var totalHits int64
	activeRules := 0
	expiredRules := 0

	now := time.Now()
	for _, cached := range c.cache {
		totalHits += cached.HitCount
		if now.Sub(cached.CreatedAt) > c.ttl {
			expiredRules++
		} else {
			activeRules++
		}
	}

	return CacheStats{
		Size:         c.size,
		MaxSize:      c.maxSize,
		ActiveRules:  activeRules,
		ExpiredRules: expiredRules,
		TotalHits:    totalHits,
		HitRate:      float64(totalHits) / float64(c.size+1), // 避免除零
	}
}

// evictLRU 淘汰最少使用的缓存项
func (c *DynamicRuleCache) evictLRU() {
	var oldestHash string
	var oldestTime time.Time
	var minHitCount int64 = -1

	// 找到最老且使用次数最少的项
	for hash, cached := range c.cache {
		if minHitCount == -1 || cached.HitCount < minHitCount ||
			(cached.HitCount == minHitCount && cached.CreatedAt.Before(oldestTime)) {
			oldestHash = hash
			oldestTime = cached.CreatedAt
			minHitCount = cached.HitCount
		}
	}

	if oldestHash != "" {
		delete(c.cache, oldestHash)
		c.size--
	}
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Size         int     `json:"size"`         // 当前缓存大小
	MaxSize      int     `json:"maxSize"`      // 最大缓存大小
	ActiveRules  int     `json:"activeRules"`  // 活跃规则数
	ExpiredRules int     `json:"expiredRules"` // 过期规则数
	TotalHits    int64   `json:"totalHits"`    // 总命中次数
	HitRate      float64 `json:"hitRate"`      // 命中率
}
