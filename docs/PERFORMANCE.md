# 🚀 性能优化指南

本文档提供 Runehammer 规则引擎的性能优化策略和最佳实践，帮助您在生产环境中获得最佳性能。

## 📚 目录

- [缓存策略优化](#缓存策略优化)
- [数据库优化](#数据库优化)
- [规则设计优化](#规则设计优化)
- [引擎选择策略](#引擎选择策略)
- [监控和调试](#监控和调试)
- [部署建议](#部署建议)

## 💾 缓存策略优化

### 缓存类型选择

```go
// 生产环境推荐：Redis 分布式缓存
engine, _ := runehammer.New[ResultType](
    runehammer.WithRedisCache("localhost:6379", "", 0),
    runehammer.WithCacheTTL(30*time.Minute),  // 根据规则变更频率调整
)

// 单机应用：内存缓存
engine, _ := runehammer.New[ResultType](
    runehammer.WithMemoryCache(1000),          // 适当的缓存大小
    runehammer.WithCacheTTL(15*time.Minute),
)

// 开发环境：禁用缓存
engine, _ := runehammer.New[ResultType](
    runehammer.WithNoCache(),
)
```

### 缓存TTL策略

| 规则变更频率 | 推荐TTL | 说明 |
|------------|---------|------|
| 极少变更（月级） | 2-4小时 | 长期稳定的业务规则 |
| 偶尔变更（周级） | 30-60分钟 | 一般业务规则 |
| 频繁变更（天级） | 5-15分钟 | 活动规则、促销规则 |
| 实时变更 | 1-5分钟 | 实时风控规则 |

### 缓存预热策略

```go
// 系统启动时预热热点规则
func preloadHotRules(engine runehammer.Engine[ResultType]) {
    hotBizCodes := []string{"user_level", "order_discount", "risk_check"}
    
    for _, bizCode := range hotBizCodes {
        // 使用虚拟数据触发规则编译和缓存
        dummyInput := createDummyInput()
        _, _ = engine.Exec(context.Background(), bizCode, dummyInput)
    }
}
```

## 🗄️ 数据库优化

### 索引优化

```sql
-- 基础索引（必须）
CREATE INDEX idx_biz_code ON runehammer_rules (biz_code);
CREATE INDEX idx_enabled ON runehammer_rules (enabled);

-- 复合索引（推荐）
CREATE INDEX idx_biz_enabled_version ON runehammer_rules (biz_code, enabled, version DESC);

-- 覆盖索引（高性能）
CREATE INDEX idx_covering ON runehammer_rules (biz_code, enabled, version DESC) 
INCLUDE (id, name, grl, updated_at);
```

### 连接池配置

```go
// GORM 连接池优化
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
sqlDB, _ := db.DB()

// 设置连接池参数
sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间

// 使用连接池
engine, _ := runehammer.New[ResultType](
    runehammer.WithCustomDB(db),
)
```

### 规则表分区（大数据量）

```sql
-- 按业务码分区
CREATE TABLE runehammer_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    biz_code VARCHAR(100) NOT NULL,
    -- 其他字段...
) PARTITION BY HASH(CRC32(biz_code)) PARTITIONS 8;

-- 按时间分区（历史数据归档）
CREATE TABLE runehammer_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    created_at DATETIME NOT NULL,
    -- 其他字段...
) PARTITION BY RANGE (YEAR(created_at)) (
    PARTITION p2023 VALUES LESS THAN (2024),
    PARTITION p2024 VALUES LESS THAN (2025),
    PARTITION p_future VALUES LESS THAN MAXVALUE
);
```

## 📋 规则设计优化

### 规则粒度控制

```go
// ✅ 好的规则设计：单一职责
rule1 := `
rule AgeCheck "年龄检查" salience 100 {
    when Params.Age >= 18
    then Result.AdultStatus = true;
}
`

rule2 := `
rule IncomeCheck "收入检查" salience 90 {
    when Params.Income > 50000
    then Result.HighIncome = true;
}
`

// ❌ 避免的设计：职责过重
badRule := `
rule ComplexCheck "复杂检查" {
    when Params.Age >= 18 && Params.Income > 50000 && 
         Params.CreditScore > 700 && Params.HasJob == true
    then 
        Result.EligibleForLoan = true;
        Result.MaxLoanAmount = Params.Income * 5;
        Result.InterestRate = 0.05;
        // ... 更多复杂逻辑
}
`
```

### 规则优先级策略

```grl
-- 使用 salience 控制执行顺序，数值越大优先级越高
rule InputValidation "输入验证" salience 1000 {
    when Params == nil || Params.UserId == ""
    then 
        Result.Error = "输入数据无效";
        Retract("InputValidation");  -- 验证失败后退出
}

rule VipUserRule "VIP用户规则" salience 500 {
    when Params.VipLevel >= 3
    then Result.Discount = 0.2;
}

rule RegularUserRule "普通用户规则" salience 100 {
    when Result.Discount == nil
    then Result.Discount = 0.05;
}
```

### 规则退出机制

```grl
-- 使用 Retract 避免重复执行
rule ProcessOrder "订单处理" salience 100 {
    when Params.Status == "pending"
    then 
        Result.ProcessStatus = "processing";
        Retract("ProcessOrder");  -- 处理后立即退出
}
```

## 🎯 引擎选择策略

### 性能对比

| 引擎类型 | 启动成本 | 内存使用 | 执行性能 | 适用场景 |
|---------|---------|---------|---------|----------|
| 传统引擎 | 高 | 高 | 最优 | 固定业务，高并发 |
| 通用引擎 | 中 | 中 | 优 | 多样化业务，中等并发 |
| 动态引擎 | 低 | 低 | 良 | 快速开发，低并发 |

### 场景选择指南

```go
// 高并发固定业务场景
userEngine, _ := runehammer.New[UserResult](
    runehammer.WithDSN("mysql://..."),
    runehammer.WithRedisCache("localhost:6379", "", 0),
    runehammer.WithCacheTTL(30*time.Minute),
)

// 多业务共享资源场景
baseEngine, _ := runehammer.NewBaseEngine(
    runehammer.WithDSN("mysql://..."),
    runehammer.WithRedisCache("localhost:6379", "", 0),
)
userEngine := runehammer.NewTypedEngine[UserResult](baseEngine)
orderEngine := runehammer.NewTypedEngine[OrderResult](baseEngine)

// 快速开发测试场景
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache:       true,
        ParallelExecution: true,
    },
)
```

## 🔄 批量处理优化

### 动态引擎批量执行

```go
// 并行批量执行规则
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache:       true,
        ParallelExecution: true,        // 启用并行执行
        DefaultTimeout:    10 * time.Second,
    },
)

// 批量规则定义
rules := []interface{}{
    rule.SimpleRule{
        When: "Params.Amount > 1000",
        Then: map[string]string{"Result.HighValue": "true"},
    },
    rule.SimpleRule{
        When: "Params.VipLevel >= 3", 
        Then: map[string]string{"Result.VipDiscount": "0.1"},
    },
    // ... 更多规则
}

// 批量执行
results, err := dynamicEngine.ExecuteBatch(ctx, rules, input)
```

### 数据库批量查询优化

```go
// 批量预加载规则
func preloadRules(bizCodes []string) {
    // 使用 IN 查询替代多次单独查询
    sql := `
        SELECT biz_code, grl, version 
        FROM runehammer_rules 
        WHERE biz_code IN (?) AND enabled = true
        ORDER BY biz_code, version DESC
    `
    // 执行批量查询...
}
```

## 📊 监控和调试

### 性能指标监控

```go
type PerformanceMonitor struct {
    execCount    int64
    totalTime    time.Duration
    cacheHits    int64
    cacheMisses  int64
}

func (m *PerformanceMonitor) RecordExecution(duration time.Duration, cacheHit bool) {
    atomic.AddInt64(&m.execCount, 1)
    atomic.AddInt64((*int64)(&m.totalTime), int64(duration))
    
    if cacheHit {
        atomic.AddInt64(&m.cacheHits, 1)
    } else {
        atomic.AddInt64(&m.cacheMisses, 1)
    }
}

func (m *PerformanceMonitor) GetStats() map[string]interface{} {
    return map[string]interface{}{
        "exec_count":   atomic.LoadInt64(&m.execCount),
        "avg_time_ms":  float64(atomic.LoadInt64((*int64)(&m.totalTime))) / float64(time.Millisecond) / float64(atomic.LoadInt64(&m.execCount)),
        "cache_hit_rate": float64(atomic.LoadInt64(&m.cacheHits)) / float64(atomic.LoadInt64(&m.cacheHits) + atomic.LoadInt64(&m.cacheMisses)),
    }
}
```

### 慢查询监控

```go
// 添加执行时间监控
func monitoredExec[T any](engine runehammer.Engine[T], ctx context.Context, bizCode string, input any) (T, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 100*time.Millisecond {
            log.Printf("慢规则执行: bizCode=%s, duration=%v", bizCode, duration)
        }
    }()
    
    return engine.Exec(ctx, bizCode, input)
}
```

### 缓存统计

```go
// 动态引擎缓存统计
stats := dynamicEngine.GetCacheStats()
fmt.Printf("缓存命中率: %.2f%%", stats.HitRate*100)
fmt.Printf("缓存大小: %d", stats.Size)

// 定期清理缓存
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        if stats := dynamicEngine.GetCacheStats(); stats.HitRate < 0.3 {
            dynamicEngine.ClearCache() // 命中率低时清理缓存
        }
    }
}()
```

## 🚀 部署建议

### 生产环境配置

```go
// 生产环境推荐配置
engine, err := runehammer.New[ResultType](
    // 数据库配置
    runehammer.WithDSN("mysql://user:pass@localhost:3306/prod_db"),
    
    // 缓存配置
    runehammer.WithRedisCache("redis-cluster:6379", "password", 0),
    runehammer.WithCacheTTL(30*time.Minute),
    runehammer.WithMaxCacheSize(5000),
    
    // 其他配置
    runehammer.WithSyncInterval(5*time.Minute),
    runehammer.WithCustomLogger(productionLogger),
)
```

### 容器化部署

```dockerfile
# Dockerfile 优化
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### 健康检查

```go
// 添加健康检查端点
func healthCheck(engine runehammer.Engine[map[string]interface{}]) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 快速健康检查
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        // 使用轻量级规则检查引擎状态
        dummyInput := map[string]interface{}{"test": true}
        _, err := engine.Exec(ctx, "health_check", dummyInput)
        
        if err != nil {
            http.Error(w, "Engine unhealthy", http.StatusServiceUnavailable)
            return
        }
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }
}
```

## 📈 性能基准测试

### 基准测试代码

```go
func BenchmarkEngineExecution(b *testing.B) {
    engine, _ := runehammer.New[TestResult](
        runehammer.WithDSN("sqlite::memory:"),
        runehammer.WithMemoryCache(1000),
    )
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
```

### 性能目标

| 指标 | 目标值 | 备注 |
|------|--------|------|
| 单次执行延迟 | < 10ms | 缓存命中时 |
| 首次执行延迟 | < 50ms | 缓存未命中时 |
| 并发处理能力 | > 1000 QPS | 单实例 |
| 缓存命中率 | > 80% | 稳定业务 |
| 内存使用 | < 500MB | 单实例 |

## 💡 性能优化检查清单

### 配置优化

- [ ] 选择合适的缓存策略（Redis vs 内存）
- [ ] 设置合理的缓存TTL
- [ ] 配置数据库连接池
- [ ] 添加必要的数据库索引

### 规则优化

- [ ] 规则职责单一化
- [ ] 设置合理的优先级
- [ ] 使用Retract避免重复执行
- [ ] 避免复杂的嵌套条件

### 代码优化

- [ ] 复用引擎实例
- [ ] 使用批量执行
- [ ] 添加超时控制
- [ ] 实现优雅关闭

### 监控优化

- [ ] 添加性能指标监控
- [ ] 设置慢查询告警
- [ ] 监控缓存命中率
- [ ] 添加健康检查

通过遵循这些性能优化策略，您可以在生产环境中获得Runehammer规则引擎的最佳性能表现。