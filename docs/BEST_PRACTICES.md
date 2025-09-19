# 💡 Runehammer 最佳实践指南

## 📚 概述

本指南提供 Runehammer 规则引擎的最佳实践、开发规范和常见问题解决方案，帮助您编写高质量、高性能的规则代码。

## 🎯 字段命名最佳实践

### 结构体设计规范

```go
// ✅ 推荐的结构体定义
type UserValidationInput struct {
    Age           int     `json:"age"`
    UserName      string  `json:"user_name"`
    Email         string  `json:"email"`
    PhoneNumber   string  `json:"phone_number"`
    AnnualIncome  float64 `json:"annual_income"`
    IsVipMember   bool    `json:"is_vip_member"`
    AccountLevel  int     `json:"account_level"`
    CreatedAt     time.Time `json:"created_at"`
}

type ValidationResult struct {
    IsValid         bool    `json:"is_valid"`
    ErrorMessage    string  `json:"error_message"`
    UserLevel       string  `json:"user_level"`
    DiscountRate    float64 `json:"discount_rate"`
    RecommendLevel  string  `json:"recommend_level"`
    ProcessedAt     time.Time `json:"processed_at"`
}

// 对应的规则表达式：
"Params.Age >= 18 && Params.UserName != ''"
"Result[\"IsValid\"] = Params.Age >= 18"
"Result[\"UserLevel\"] = Params.IsVipMember ? 'premium' : 'standard'"
"Result[\"DiscountRate\"] = Params.AccountLevel >= 3 ? 0.15 : 0.05"

// ❌ 避免的命名方式
type BadExample struct {
    age       int    `json:"age"`        // 小写字段名
    user_name string `json:"user_name"`  // 下划线字段名
    isVIP     bool   `json:"is_vip"`     // 不规范的大小写混合
}

// ❌ 错误的规则访问
// "Params.age >= 18"           // 小写
// "Params.user_name != ''"     // 下划线
// "Params.isVIP == true"       // 不规范大小写
```

### 字段命名约定

| 场景 | Go字段名 | JSON标签 | 规则访问 |
|------|---------|---------|----------|
| 用户ID | `UserId` | `"user_id"` | `Params.UserId` |
| 订单金额 | `OrderAmount` | `"order_amount"` | `Params.OrderAmount` |
| 是否VIP | `IsVipMember` | `"is_vip_member"` | `Params.IsVipMember` |
| 创建时间 | `CreatedAt` | `"created_at"` | `Params.CreatedAt` |

## 🏗️ 引擎选择最佳实践

### 场景决策矩阵

```go
// 场景1: 企业级固定业务 - 传统引擎
func createEnterpriseEngine[T any]() (runehammer.Engine[T], error) {
    return runehammer.New[T](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/prod_db"),
        runehammer.WithAutoMigrate(),
        runehammer.WithRedis("redis-cluster:6379", "password", 0),
        runehammer.WithCacheTTL(30*time.Minute),
    )
}

// 场景2: 微服务架构 - 通用引擎
func createMicroServiceEngine() (runehammer.BaseEngine, error) {
    return runehammer.NewBaseEngine(
        runehammer.WithDSN("mysql://user:pass@localhost:3306/microservice_db"),
        runehammer.WithAutoMigrate(),
        runehammer.WithRedis("localhost:6379", "", 0),
    )
}

// 场景3: 快速原型开发 - 动态引擎
func createPrototypeEngine() engine.DynamicEngine[map[string]interface{}] {
    return engine.NewDynamicEngine[map[string]interface{}](
        engine.DynamicEngineConfig{
            EnableCache:       true,
            ParallelExecution: true,
            DefaultTimeout:    5 * time.Second,
        },
    )
}
```

### 引擎资源管理

```go
// 推荐：引擎管理器模式
type RuleEngineManager struct {
    baseEngine   runehammer.BaseEngine
    userEngine   runehammer.Engine[UserResult]
    orderEngine  runehammer.Engine[OrderResult]
    riskEngine   runehammer.Engine[RiskResult]
    mu           sync.RWMutex
    isShutdown   bool
}

func NewRuleEngineManager(dsn string) (*RuleEngineManager, error) {
    baseEngine, err := runehammer.NewBaseEngine(
        runehammer.WithDSN(dsn),
        runehammer.WithAutoMigrate(),
        runehammer.WithRedis("localhost:6379", "", 0),
        runehammer.WithCacheTTL(30*time.Minute),
    )
    if err != nil {
        return nil, fmt.Errorf("创建BaseEngine失败: %w", err)
    }
    
    return &RuleEngineManager{
        baseEngine:  baseEngine,
        userEngine:  runehammer.NewTypedEngine[UserResult](baseEngine),
        orderEngine: runehammer.NewTypedEngine[OrderResult](baseEngine),
        riskEngine:  runehammer.NewTypedEngine[RiskResult](baseEngine),
    }, nil
}

func (rem *RuleEngineManager) ProcessUser(ctx context.Context, bizCode string, user UserInput) (*UserResult, error) {
    rem.mu.RLock()
    if rem.isShutdown {
        rem.mu.RUnlock()
        return nil, errors.New("引擎管理器已关闭")
    }
    rem.mu.RUnlock()
    
    return rem.userEngine.Exec(ctx, bizCode, user)
}

func (rem *RuleEngineManager) Shutdown() {
    rem.mu.Lock()
    defer rem.mu.Unlock()
    
    if !rem.isShutdown {
        rem.baseEngine.Close()
        rem.isShutdown = true
    }
}
```

## ⚡ 性能优化最佳实践

### 缓存策略优化

```go
// 1. 分层缓存配置
func createOptimizedEngine[T any]() (runehammer.Engine[T], error) {
    return runehammer.New[T](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/db"),
        runehammer.WithAutoMigrate(),
        
        // Redis L1 缓存 - 分布式共享
        runehammer.WithRedis("localhost:6379", "", 0),
        runehammer.WithCacheTTL(30*time.Minute),
        
        // 内存 L2 缓存 - 本地热数据
        runehammer.WithMemory(500, 5*time.Minute),
    )
}

// 2. 智能缓存预热
func preloadHotRules(engine runehammer.Engine[any]) {
    hotBizCodes := []string{
        "user_level_check",
        "order_discount", 
        "risk_assessment",
        "vip_validation",
    }
    
    // 并发预热
    var wg sync.WaitGroup
    for _, bizCode := range hotBizCodes {
        wg.Add(1)
        go func(code string) {
            defer wg.Done()
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            // 使用虚拟数据触发规则编译和缓存
            dummyInput := createDummyInput(code)
            _, _ = engine.Exec(ctx, code, dummyInput)
        }(bizCode)
    }
    wg.Wait()
}

// 3. 缓存监控和清理
type CacheMonitor struct {
    engine   runehammer.Engine[any]
    ticker   *time.Ticker
    stopCh   chan struct{}
}

func (cm *CacheMonitor) Start() {
    cm.ticker = time.NewTicker(1 * time.Hour)
    cm.stopCh = make(chan struct{})
    
    go func() {
        for {
            select {
            case <-cm.ticker.C:
                cm.checkAndCleanCache()
            case <-cm.stopCh:
                return
            }
        }
    }()
}

func (cm *CacheMonitor) checkAndCleanCache() {
    // 获取缓存统计信息
    if stats := cm.engine.GetCacheStats(); stats != nil {
        hitRate := float64(stats.Hits) / float64(stats.Hits + stats.Misses)
        
        // 命中率过低时清理缓存
        if hitRate < 0.3 {
            cm.engine.ClearCache()
            log.Printf("缓存命中率过低(%.2f%%)，已清理缓存", hitRate*100)
        }
    }
}
```

### 数据库优化

```go
// 1. 连接池优化
func createOptimizedDB() (*gorm.DB, error) {
    dsn := "mysql://user:pass@localhost:3306/db?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
    sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
    sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间
    
    return db, nil
}

// 2. 索引优化建议
const createIndexesSQL = `
-- 基础索引（必须）
CREATE INDEX idx_biz_code ON runehammer_rules (biz_code);
CREATE INDEX idx_enabled ON runehammer_rules (enabled);

-- 复合索引（推荐）
CREATE INDEX idx_biz_enabled_version ON runehammer_rules (biz_code, enabled, version DESC);

-- 覆盖索引（高性能）
CREATE INDEX idx_covering ON runehammer_rules (biz_code, enabled, version DESC) 
INCLUDE (id, name, grl, updated_at);
`
```

### 批量处理优化

```go
// 1. 动态引擎批量执行
func processBatchRules(engine engine.DynamicEngine[map[string]interface{}], orders []OrderData) ([]map[string]interface{}, error) {
    // 构建批量规则
    batchRules := []interface{}{
        rule.SimpleRule{
            When: "Params.Amount > 500",
            Then: map[string]string{
                "Result[\"FreeShipping\"]": "true",
            },
        },
        rule.SimpleRule{
            When: "Params.VipLevel >= 3",
            Then: map[string]string{
                "Result[\"VipDiscount\"]": "0.1",
            },
        },
        rule.SimpleRule{
            When: "Params.Amount > 1000 && Params.VipLevel >= 5",
            Then: map[string]string{
                "Result[\"PremiumDiscount\"]": "0.2",
            },
        },
    }
    
    // 并发批量处理
    results := make([]map[string]interface{}, len(orders))
    var wg sync.WaitGroup
    errCh := make(chan error, len(orders))
    
    for i, order := range orders {
        wg.Add(1)
        go func(index int, orderData OrderData) {
            defer wg.Done()
            
            ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
            defer cancel()
            
            batchResults, err := engine.ExecuteBatch(ctx, batchRules, orderData)
            if err != nil {
                errCh <- fmt.Errorf("批量执行失败[%d]: %w", index, err)
                return
            }
            
            // 合并批量结果
            mergedResult := make(map[string]interface{})
            for _, result := range batchResults {
                for k, v := range result {
                    mergedResult[k] = v
                }
            }
            results[index] = mergedResult
        }(i, order)
    }
    
    wg.Wait()
    close(errCh)
    
    // 检查错误
    if err := <-errCh; err != nil {
        return nil, err
    }
    
    return results, nil
}
```

## 🔧 错误处理最佳实践

### 分层错误处理

```go
// 1. 自定义错误类型
type RuleEngineError struct {
    Code    string
    Message string
    BizCode string
    Err     error
}

func (e *RuleEngineError) Error() string {
    return fmt.Sprintf("[%s] %s (bizCode: %s): %v", e.Code, e.Message, e.BizCode, e.Err)
}

var (
    ErrRuleNotFound     = &RuleEngineError{Code: "RULE_NOT_FOUND", Message: "规则不存在"}
    ErrRuleExecTimeout  = &RuleEngineError{Code: "RULE_EXEC_TIMEOUT", Message: "规则执行超时"}
    ErrInvalidInput     = &RuleEngineError{Code: "INVALID_INPUT", Message: "输入数据无效"}
    ErrRuleCompileError = &RuleEngineError{Code: "RULE_COMPILE_ERROR", Message: "规则编译失败"}
)

// 2. 错误处理包装器
func safeExecRule[T any](engine runehammer.Engine[T], ctx context.Context, bizCode string, input any) (result T, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = &RuleEngineError{
                Code:    "RULE_PANIC",
                Message: "规则执行发生panic",
                BizCode: bizCode,
                Err:     fmt.Errorf("%v", r),
            }
        }
    }()
    
    // 输入验证
    if input == nil {
        return result, &RuleEngineError{
            Code:    "INVALID_INPUT",
            Message: "输入数据为空",
            BizCode: bizCode,
        }
    }
    
    // 执行规则
    result, err = engine.Exec(ctx, bizCode, input)
    if err != nil {
        switch {
        case errors.Is(err, runehammer.ErrNoRulesFound):
            return result, &RuleEngineError{
                Code:    "RULE_NOT_FOUND",
                Message: "未找到对应规则",
                BizCode: bizCode,
                Err:     err,
            }
        case errors.Is(err, context.DeadlineExceeded):
            return result, &RuleEngineError{
                Code:    "RULE_EXEC_TIMEOUT",
                Message: "规则执行超时",
                BizCode: bizCode,
                Err:     err,
            }
        default:
            return result, &RuleEngineError{
                Code:    "RULE_EXEC_ERROR",
                Message: "规则执行失败",
                BizCode: bizCode,
                Err:     err,
            }
        }
    }
    
    return result, nil
}

// 3. 业务层错误处理
func processUserValidation(ctx context.Context, engine runehammer.Engine[UserResult], user UserInput) (*UserResult, error) {
    result, err := safeExecRule(engine, ctx, "USER_VALIDATE", user)
    if err != nil {
        var ruleErr *RuleEngineError
        if errors.As(err, &ruleErr) {
            switch ruleErr.Code {
            case "RULE_NOT_FOUND":
                // 记录日志并返回默认结果
                log.Printf("规则不存在，使用默认验证逻辑: %s", ruleErr.BizCode)
                return defaultUserValidation(user), nil
            case "RULE_EXEC_TIMEOUT":
                // 超时处理
                log.Printf("规则执行超时: %s", ruleErr.BizCode)
                return nil, fmt.Errorf("用户验证超时，请稍后重试")
            case "INVALID_INPUT":
                // 输入验证错误
                return nil, fmt.Errorf("用户输入数据无效: %s", ruleErr.Message)
            default:
                // 其他错误
                log.Printf("规则执行失败: %s", ruleErr.Error())
                return nil, fmt.Errorf("用户验证失败，请联系管理员")
            }
        }
        return nil, fmt.Errorf("未知错误: %w", err)
    }
    
    return &result, nil
}
```

## 📊 监控和调试最佳实践

### 性能监控

```go
// 1. 性能指标收集
type PerformanceMonitor struct {
    execCount    int64
    totalTime    time.Duration
    cacheHits    int64
    cacheMisses  int64
    errors       int64
    mu           sync.RWMutex
}

func (pm *PerformanceMonitor) RecordExecution(duration time.Duration, cacheHit bool, hasError bool) {
    atomic.AddInt64(&pm.execCount, 1)
    atomic.AddInt64((*int64)(&pm.totalTime), int64(duration))
    
    if hasError {
        atomic.AddInt64(&pm.errors, 1)
    }
    
    if cacheHit {
        atomic.AddInt64(&pm.cacheHits, 1)
    } else {
        atomic.AddInt64(&pm.cacheMisses, 1)
    }
}

func (pm *PerformanceMonitor) GetStats() map[string]interface{} {
    execCount := atomic.LoadInt64(&pm.execCount)
    totalTime := time.Duration(atomic.LoadInt64((*int64)(&pm.totalTime)))
    cacheHits := atomic.LoadInt64(&pm.cacheHits)
    cacheMisses := atomic.LoadInt64(&pm.cacheMisses)
    errors := atomic.LoadInt64(&pm.errors)
    
    avgTime := float64(0)
    if execCount > 0 {
        avgTime = float64(totalTime) / float64(time.Millisecond) / float64(execCount)
    }
    
    cacheHitRate := float64(0)
    if cacheHits+cacheMisses > 0 {
        cacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses)
    }
    
    errorRate := float64(0)
    if execCount > 0 {
        errorRate = float64(errors) / float64(execCount)
    }
    
    return map[string]interface{}{
        "exec_count":      execCount,
        "avg_time_ms":     avgTime,
        "cache_hit_rate":  cacheHitRate,
        "error_rate":      errorRate,
        "total_errors":    errors,
    }
}

// 2. 监控装饰器
func monitoredExec[T any](engine runehammer.Engine[T], monitor *PerformanceMonitor) func(context.Context, string, any) (T, error) {
    return func(ctx context.Context, bizCode string, input any) (T, error) {
        start := time.Now()
        var result T
        var cacheHit bool
        
        // 执行规则
        result, err := engine.Exec(ctx, bizCode, input)
        
        // 记录指标
        duration := time.Since(start)
        hasError := err != nil
        
        // 检查是否命中缓存（简化示例）
        cacheHit = duration < 10*time.Millisecond
        
        monitor.RecordExecution(duration, cacheHit, hasError)
        
        // 慢查询日志
        if duration > 100*time.Millisecond {
            log.Printf("慢规则执行: bizCode=%s, duration=%v, error=%v", bizCode, duration, err)
        }
        
        return result, err
    }
}
```

### 调试和日志

```go
// 1. 结构化日志
type RuleLogger struct {
    logger *logrus.Logger
}

func NewRuleLogger() *RuleLogger {
    logger := logrus.New()
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetLevel(logrus.InfoLevel)
    
    return &RuleLogger{logger: logger}
}

func (rl *RuleLogger) LogRuleExecution(bizCode string, input interface{}, result interface{}, duration time.Duration, err error) {
    fields := logrus.Fields{
        "biz_code":  bizCode,
        "duration":  duration.Milliseconds(),
        "timestamp": time.Now().Unix(),
    }
    
    if err != nil {
        fields["error"] = err.Error()
        rl.logger.WithFields(fields).Error("规则执行失败")
    } else {
        fields["success"] = true
        rl.logger.WithFields(fields).Info("规则执行成功")
    }
}

// 2. 调试模式
type DebugEngine[T any] struct {
    engine runehammer.Engine[T]
    logger *RuleLogger
    debug  bool
}

func NewDebugEngine[T any](engine runehammer.Engine[T], debug bool) *DebugEngine[T] {
    return &DebugEngine[T]{
        engine: engine,
        logger: NewRuleLogger(),
        debug:  debug,
    }
}

func (de *DebugEngine[T]) Exec(ctx context.Context, bizCode string, input any) (T, error) {
    if de.debug {
        inputJson, _ := json.Marshal(input)
        log.Printf("DEBUG: 执行规则 %s，输入: %s", bizCode, string(inputJson))
    }
    
    start := time.Now()
    result, err := de.engine.Exec(ctx, bizCode, input)
    duration := time.Since(start)
    
    if de.debug {
        if err != nil {
            log.Printf("DEBUG: 规则执行失败 %s，耗时: %v，错误: %v", bizCode, duration, err)
        } else {
            resultJson, _ := json.Marshal(result)
            log.Printf("DEBUG: 规则执行成功 %s，耗时: %v，结果: %s", bizCode, duration, string(resultJson))
        }
    }
    
    de.logger.LogRuleExecution(bizCode, input, result, duration, err)
    return result, err
}
```

## 🔄 规则版本管理最佳实践

### 规则版本控制

```go
// 1. 规则版本管理器
type RuleVersionManager struct {
    db *gorm.DB
    mu sync.RWMutex
}

func (rvm *RuleVersionManager) DeployRule(bizCode, ruleName, grlContent string) error {
    rvm.mu.Lock()
    defer rvm.mu.Unlock()
    
    // 获取当前最大版本号
    var currentVersion int
    rvm.db.Model(&runehammer.Rule{}).
        Where("biz_code = ?", bizCode).
        Select("COALESCE(MAX(version), 0)").
        Scan(&currentVersion)
    
    // 创建新版本规则
    newRule := &runehammer.Rule{
        BizCode: bizCode,
        Name:    ruleName,
        GRL:     grlContent,
        Version: currentVersion + 1,
        Enabled: false, // 新版本默认禁用
    }
    
    return rvm.db.Create(newRule).Error
}

func (rvm *RuleVersionManager) EnableRuleVersion(bizCode string, version int) error {
    rvm.mu.Lock()
    defer rvm.mu.Unlock()
    
    tx := rvm.db.Begin()
    defer tx.Rollback()
    
    // 禁用所有版本
    if err := tx.Model(&runehammer.Rule{}).
        Where("biz_code = ?", bizCode).
        Update("enabled", false).Error; err != nil {
        return err
    }
    
    // 启用指定版本
    if err := tx.Model(&runehammer.Rule{}).
        Where("biz_code = ? AND version = ?", bizCode, version).
        Update("enabled", true).Error; err != nil {
        return err
    }
    
    return tx.Commit().Error
}

func (rvm *RuleVersionManager) RollbackToVersion(bizCode string, version int) error {
    return rvm.EnableRuleVersion(bizCode, version)
}
```

### 灰度发布

```go
// 2. 灰度发布管理器
type GrayReleaseManager struct {
    rvm     *RuleVersionManager
    engine  runehammer.Engine[any]
    config  *GrayReleaseConfig
}

type GrayReleaseConfig struct {
    GrayPercent int      // 灰度流量百分比
    WhiteList   []string // 白名单用户
    BlackList   []string // 黑名单用户
}

func (grm *GrayReleaseManager) ShouldUseGrayVersion(userID string) bool {
    // 黑名单用户不使用灰度版本
    for _, blackUser := range grm.config.BlackList {
        if blackUser == userID {
            return false
        }
    }
    
    // 白名单用户强制使用灰度版本
    for _, whiteUser := range grm.config.WhiteList {
        if whiteUser == userID {
            return true
        }
    }
    
    // 根据百分比随机决定
    hash := fnv.New32a()
    hash.Write([]byte(userID))
    return int(hash.Sum32()%100) < grm.config.GrayPercent
}

func (grm *GrayReleaseManager) ExecWithGray(ctx context.Context, bizCode string, userID string, input any) (any, error) {
    // 决定使用哪个版本
    useGray := grm.ShouldUseGrayVersion(userID)
    
    // 执行对应版本的规则
    if useGray {
        return grm.engine.Exec(ctx, bizCode+"_gray", input)
    } else {
        return grm.engine.Exec(ctx, bizCode, input)
    }
}
```

## 🧪 测试最佳实践

### 单元测试

```go
// 1. 规则引擎测试套件
type RuleEngineTestSuite struct {
    suite.Suite
    engine runehammer.Engine[TestResult]
    db     *gorm.DB
}

func (suite *RuleEngineTestSuite) SetupSuite() {
    // 创建测试数据库
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    suite.Require().NoError(err)
    
    // 自动迁移
    err = db.AutoMigrate(&runehammer.Rule{})
    suite.Require().NoError(err)
    
    suite.db = db
    
    // 创建测试引擎
    suite.engine, err = runehammer.New[TestResult](
        runehammer.WithCustomDB(db),
        runehammer.WithMemoryCache(100),
    )
    suite.Require().NoError(err)
}

func (suite *RuleEngineTestSuite) TearDownSuite() {
    suite.engine.Close()
}

func (suite *RuleEngineTestSuite) TestUserValidation() {
    // 准备测试规则
    testRule := &runehammer.Rule{
        BizCode: "test_user_validation",
        Name:    "测试用户验证规则",
        GRL: `rule TestUserValidation "测试用户验证" {
            when Params.Age >= 18 && Params.Income > 50000
            then 
                Result["IsValid"] = true;
                Result["Level"] = "premium";
        }`,
        Enabled: true,
        Version: 1,
    }
    
    err := suite.db.Create(testRule).Error
    suite.Require().NoError(err)
    
    // 测试用例
    testCases := []struct {
        name     string
        input    TestInput
        expected TestResult
        hasError bool
    }{
        {
            name: "有效用户",
            input: TestInput{
                Age:    25,
                Income: 80000,
            },
            expected: TestResult{
                IsValid: true,
                Level:   "premium",
            },
            hasError: false,
        },
        {
            name: "年龄不足",
            input: TestInput{
                Age:    17,
                Income: 80000,
            },
            expected: TestResult{},
            hasError: false,
        },
        {
            name: "收入不足",
            input: TestInput{
                Age:    25,
                Income: 30000,
            },
            expected: TestResult{},
            hasError: false,
        },
    }
    
    for _, tc := range testCases {
        suite.Run(tc.name, func() {
            result, err := suite.engine.Exec(context.Background(), "test_user_validation", tc.input)
            
            if tc.hasError {
                suite.Error(err)
            } else {
                suite.NoError(err)
                suite.Equal(tc.expected.IsValid, result.IsValid)
                suite.Equal(tc.expected.Level, result.Level)
            }
        })
    }
}

func TestRuleEngineTestSuite(t *testing.T) {
    suite.Run(t, new(RuleEngineTestSuite))
}
```

### 基准测试

```go
// 2. 性能基准测试
func BenchmarkRuleEngineExecution(b *testing.B) {
    engine, err := runehammer.New[TestResult](
        runehammer.WithDSN("sqlite::memory:"),
        runehammer.WithMemoryCache(1000),
    )
    if err != nil {
        b.Fatal(err)
    }
    defer engine.Close()
    
    input := TestInput{Age: 25, Income: 80000}
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := engine.Exec(context.Background(), "test_rule", input)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}

func BenchmarkCacheHitRate(b *testing.B) {
    engine, _ := runehammer.New[TestResult](
        runehammer.WithDSN("sqlite::memory:"),
        runehammer.WithRedisCache("localhost:6379", "", 0),
        runehammer.WithCacheTTL(5*time.Minute),
    )
    defer engine.Close()
    
    input := TestInput{Age: 25, Income: 80000}
    
    // 预热缓存
    engine.Exec(context.Background(), "test_rule", input)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        engine.Exec(context.Background(), "test_rule", input)
    }
}
```

## 📋 开发规范检查清单

### 代码质量检查

- [ ] **字段命名规范**: Go字段使用大驼峰，规则中使用`Params.字段名`访问
- [ ] **返回值规范**: 使用`Result["字段名"]`形式赋值
- [ ] **枚举类型使用**: 优先使用类型安全的枚举常量
- [ ] **错误处理**: 实现分层错误处理和异常恢复
- [ ] **资源管理**: 正确关闭引擎和数据库连接

### 性能优化检查

- [ ] **缓存配置**: 根据业务特点选择合适的缓存策略
- [ ] **连接池设置**: 配置合理的数据库连接池参数
- [ ] **批量处理**: 对独立规则使用批量并行执行
- [ ] **监控指标**: 添加性能监控和慢查询日志

### 测试覆盖检查

- [ ] **单元测试**: 规则逻辑测试覆盖率 ≥ 80%
- [ ] **集成测试**: 端到端业务场景测试
- [ ] **性能测试**: 并发和基准测试
- [ ] **边界测试**: 异常输入和极限情况测试

## 📊 总结

遵循这些最佳实践可以帮助您：

### 🎯 提高代码质量
- 统一的命名规范和字段访问方式
- 类型安全的枚举系统使用
- 完善的错误处理和异常恢复

### ⚡ 优化系统性能
- 智能缓存策略和预热机制
- 数据库连接池和索引优化
- 批量处理和并发执行

### 🔧 简化开发维护
- 规则版本管理和灰度发布
- 性能监控和调试工具
- 完整的测试覆盖

更多详细信息请参考：
- [引擎使用指南](./ENGINES_USAGE.md) - 选择合适的引擎类型
- [规则语法指南](./RULES_SYNTAX.md) - 掌握规则语法和枚举系统
- [性能优化指南](./PERFORMANCE.md) - 深入的性能优化策略