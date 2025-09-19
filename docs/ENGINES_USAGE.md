# 🏗️ Runehammer 引擎使用指南

## 📚 概述

Runehammer 提供三种类型的规则引擎，满足不同的业务场景和技术需求。本指南将详细介绍每种引擎的使用方法和适用场景。

## 🎯 引擎类型对比

| 特性 | 传统引擎 | 通用引擎 | 动态引擎 |
|------|----------|----------|----------|
| 规则存储 | 数据库 | 数据库 | 运行时定义 |
| 返回类型 | 编译时指定 | 运行时灵活 | 运行时灵活 |
| 资源使用 | 多实例 | 单实例共享 | 轻量级 |
| 适用场景 | 固定业务 | 多样化需求 | 快速原型 |
| 性能 | 最优 | 优秀 | 良好 |
| 复杂度 | 中等 | 较低 | 最低 |

## 1. 🏛️ 传统引擎（Database-Based Engine）

### 核心特点
- **数据库存储**: 规则存储在数据库中，支持热更新和版本管理
- **强类型**: 每个引擎实例绑定特定的返回类型
- **高性能**: 优化的缓存机制，适合高并发场景
- **企业级**: 完整的规则管理功能，适合企业级应用

### 数据库规则定义

```sql
-- 创建规则表
CREATE TABLE runehammer_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    biz_code VARCHAR(100) NOT NULL,
    name VARCHAR(200) NOT NULL,
    grl TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    version INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 插入用户验证规则
INSERT INTO runehammer_rules (biz_code, name, grl, enabled) VALUES 
('USER_VALIDATE', '用户验证规则', 
'rule UserValidation "用户验证规则" { 
  when Params.Age >= 18 && Params.Income > 50000 
  then 
    Result["Adult"] = true; 
    Result["Eligible"] = true;
    Result["Level"] = "premium";
}', true);

-- 插入订单处理规则
INSERT INTO runehammer_rules (biz_code, name, grl, enabled) VALUES 
('ORDER_PROCESS', '订单处理规则', 
'rule OrderProcess "订单处理规则" { 
  when Params.Amount > 1000 && Params.Vip == true
  then 
    Result["Discount"] = 0.15; 
    Result["Priority"] = "high";
    Result["FreeShipping"] = true;
}', true);
```

### Go代码实现

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "gitee.com/damengde/runehammer"
)

// 业务数据结构
type User struct {
    Age    int     `json:"age"`
    Income float64 `json:"income"`
    Vip    bool    `json:"vip"`
}

type Order struct {
    Amount float64 `json:"amount"`
    Vip    bool    `json:"vip"`
}

// 规则执行结果
type ValidationResult struct {
    Adult    bool   `json:"adult"`
    Eligible bool   `json:"eligible"`
    Level    string `json:"level"`
}

type OrderResult struct {
    Discount     float64 `json:"discount"`
    Priority     string  `json:"priority"`
    FreeShipping bool    `json:"free_shipping"`
}

func main() {
    // 创建用户验证引擎 - 每种返回类型需要独立实例
    userEngine, err := runehammer.New[ValidationResult](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
        runehammer.WithRedis("localhost:6379", "", 0),
    )
    if err != nil {
        log.Fatal("创建用户引擎失败:", err)
    }
    defer userEngine.Close()
    
    // 创建订单处理引擎
    orderEngine, err := runehammer.New[OrderResult](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
    )
    if err != nil {
        log.Fatal("创建订单引擎失败:", err)
    }
    defer orderEngine.Close()
    
    // 执行用户验证规则
    userData := User{
        Age:    25,
        Income: 80000.0,
        Vip:    true,
    }
    
    userResult, err := userEngine.Exec(context.Background(), "USER_VALIDATE", userData)
    if err != nil {
        log.Fatal("执行用户验证失败:", err)
    }
    
    fmt.Printf("用户验证结果: Adult=%v, Eligible=%v, Level=%s\n", 
        userResult.Adult, userResult.Eligible, userResult.Level)
    
    // 执行订单处理规则
    orderData := Order{
        Amount: 1500.0,
        Vip:    true,
    }
    
    orderResult, err := orderEngine.Exec(context.Background(), "ORDER_PROCESS", orderData)
    if err != nil {
        log.Fatal("执行订单处理失败:", err)
    }
    
    fmt.Printf("订单处理结果: Discount=%.2f, Priority=%s, FreeShipping=%v\n", 
        orderResult.Discount, orderResult.Priority, orderResult.FreeShipping)
}
```

### 适用场景
- ✅ 企业级应用，规则相对固定
- ✅ 高并发业务场景
- ✅ 需要规则版本管理和审计
- ✅ 对性能要求较高的场景

## 2. 🚀 通用引擎（Universal Engine）

### 核心优势
- **资源共享**: 一个BaseEngine实例支持多种返回类型
- **动态类型**: 运行时决定返回类型，无需编译时指定
- **统一管理**: 数据库连接、缓存、配置统一管理
- **灵活扩展**: 便于微服务架构和多业务场景

### 使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "gitee.com/damengde/runehammer"
)

func main() {
    // ============================================================================
    // 启动时创建单个BaseEngine实例
    // ============================================================================
    
    baseEngine, err := runehammer.NewBaseEngine(
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
        runehammer.WithRedis("localhost:6379", "", 0),
    )
    if err != nil {
        log.Fatal("创建BaseEngine失败:", err)
    }
    defer baseEngine.Close()
    
    // ============================================================================
    // 运行时创建不同类型的TypedEngine包装器
    // ============================================================================
    
    // 用户验证引擎 - 返回强类型结构体
    userEngine := runehammer.NewTypedEngine[ValidationResult](baseEngine)
    
    // 订单处理引擎 - 返回强类型结构体  
    orderEngine := runehammer.NewTypedEngine[OrderResult](baseEngine)
    
    // 通用map引擎 - 返回灵活的map类型
    mapEngine := runehammer.NewTypedEngine[map[string]interface{}](baseEngine)
    
    // ============================================================================
    // 测试数据
    // ============================================================================
    
    userData := User{Age: 25, Income: 80000.0, Vip: true}
    orderData := Order{Amount: 1500.0, Vip: true}
    
    ctx := context.Background()
    
    // ============================================================================
    // 演示：同一个BaseEngine支持多种返回类型
    // ============================================================================
    
    // 用户验证 - 强类型结构体结果
    userResult, err := userEngine.Exec(ctx, "USER_VALIDATE", userData)
    if err != nil {
        log.Printf("用户验证失败: %v", err)
    } else {
        fmt.Printf("👤 用户验证结果: Adult=%v, Eligible=%v, Level=%s\n", 
            userResult.Adult, userResult.Eligible, userResult.Level)
    }
    
    // 订单处理 - 强类型结构体结果
    orderResult, err := orderEngine.Exec(ctx, "ORDER_PROCESS", orderData)
    if err != nil {
        log.Printf("订单处理失败: %v", err)
    } else {
        fmt.Printf("🛒 订单处理结果: Discount=%.2f, Priority=%s, FreeShipping=%v\n", 
            orderResult.Discount, orderResult.Priority, orderResult.FreeShipping)
    }
    
    // 通用map - 灵活的map结果
    mapResult, err := mapEngine.Exec(ctx, "USER_VALIDATE", userData)
    if err != nil {
        log.Printf("通用执行失败: %v", err)
    } else {
        fmt.Printf("🗂️  通用map结果: %+v\n", mapResult)
    }
}
```

### 资源优化示例

```go
// 微服务架构中的引擎管理
type EngineManager struct {
    baseEngine  runehammer.BaseEngine
    userEngine  runehammer.Engine[ValidationResult]
    orderEngine runehammer.Engine[OrderResult]
    riskEngine  runehammer.Engine[RiskResult]
}

func NewEngineManager(dsn string) (*EngineManager, error) {
    // 创建共享的BaseEngine
    baseEngine, err := runehammer.NewBaseEngine(
        runehammer.WithDSN(dsn),
        runehammer.WithAutoMigrate(),
        runehammer.WithRedis("localhost:6379", "", 0),
        runehammer.WithCacheTTL(30*time.Minute),
    )
    if err != nil {
        return nil, err
    }
    
    return &EngineManager{
        baseEngine:  baseEngine,
        userEngine:  runehammer.NewTypedEngine[ValidationResult](baseEngine),
        orderEngine: runehammer.NewTypedEngine[OrderResult](baseEngine),
        riskEngine:  runehammer.NewTypedEngine[RiskResult](baseEngine),
    }, nil
}

func (em *EngineManager) ProcessUser(ctx context.Context, user User) (*ValidationResult, error) {
    return em.userEngine.Exec(ctx, "USER_VALIDATE", user)
}

func (em *EngineManager) ProcessOrder(ctx context.Context, order Order) (*OrderResult, error) {
    return em.orderEngine.Exec(ctx, "ORDER_PROCESS", order)
}

func (em *EngineManager) Close() {
    em.baseEngine.Close()
}
```

### 适用场景
- ✅ 微服务架构，多种业务类型
- ✅ 需要灵活的返回类型
- ✅ 资源利用率优化
- ✅ 统一的规则管理需求

## 3. ⚡ 动态引擎（Dynamic Engine）

### 核心特点
- **运行时定义**: 无需数据库存储，直接在代码中定义规则
- **快速原型**: 适合快速开发和测试
- **灵活配置**: 支持缓存、并发、超时等高级配置
- **多种规则格式**: 支持简单规则、指标规则、标准规则

### 基础使用

```go
import (
    "gitee.com/damengde/runehammer/engine"
    "gitee.com/damengde/runehammer/rule"
)

// 创建动态引擎
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache:       true,
        CacheTTL:          5 * time.Minute,
        MaxCacheSize:      100,
        StrictValidation:  true,
        ParallelExecution: true,
        DefaultTimeout:    10 * time.Second,
    },
)
```

### 简单规则（SimpleRule）

```go
type AgeData struct {
    Age int `json:"age"`
}

// 年龄验证规则
ageRule := rule.SimpleRule{
    When: "Params.Age >= 18",
    Then: map[string]string{
        "Result[\"Adult\"]":   "true",
        "Result[\"Message\"]": "\"符合年龄要求\"",
    },
}

// 执行规则
ageData := AgeData{Age: 25}
result, err := dynamicEngine.ExecuteRuleDefinition(context.Background(), ageRule, ageData)
```

### 指标规则（MetricRule）

```go
// 客户评分计算
scoreRule := rule.MetricRule{
    Name:        "customer_score",
    Description: "客户综合评分",
    Formula:     "age_score + income_score + vip_score",
    Variables: map[string]string{
        "age_score":    "Params.Age * 0.1",
        "income_score": "Params.Income * 0.0001", 
        "vip_score":    "Params.VipLevel * 10",
    },
    Conditions: []string{
        "Params.Age >= 18",
        "Params.Income > 0",
    },
}

type Customer struct {
    Age      int     `json:"age"`
    Income   float64 `json:"income"`
    VipLevel int     `json:"vip_level"`
}

customer := Customer{Age: 30, Income: 80000, VipLevel: 3}
result, err := dynamicEngine.ExecuteRuleDefinition(context.Background(), scoreRule, customer)
```

### 标准规则（StandardRule）

```go
// 订单折扣规则 - 使用枚举类型（类型安全）
discountRule := rule.StandardRule{
    ID:          "order_discount",
    Name:        "订单折扣规则",
    Description: "根据客户等级和订单金额计算折扣",
    Priority:    100,
    Enabled:     true,
    Tags:        []string{"discount", "order"},
    Conditions: rule.Condition{
        Type: rule.ConditionTypeAnd,
        Children: []rule.Condition{
            {
                Type:     rule.ConditionTypeSimple,
                Left:     "Params.Amount",
                Operator: rule.OpGreaterThan,
                Right:    500,
            },
            {
                Type:     rule.ConditionTypeSimple,
                Left:     "Params.VipLevel",
                Operator: rule.OpGreaterThanOrEqual,
                Right:    2,
            },
        },
    },
    Actions: []rule.Action{
        {
            Type:   rule.ActionTypeAssign,
            Target: "Result[\"DiscountRate\"]",
            Value:  0.15,
        },
        {
            Type:   rule.ActionTypeCalculate,
            Target: "Result[\"DiscountAmount\"]", 
            Value:  "Params.Amount * 0.15",
        },
    },
}
```

### 自定义函数注册

```go
// 注册单个自定义函数
dynamicEngine.RegisterCustomFunction("CalculateDiscount", func(amount float64, rate float64) float64 {
    return amount * rate
})

// 批量注册自定义函数
dynamicEngine.RegisterCustomFunctions(map[string]interface{}{
    "IsVip": func(level int) bool {
        return level >= 3
    },
    "GetDiscountRate": func(vipLevel int, amount float64) float64 {
        if vipLevel >= 5 {
            return 0.2
        } else if vipLevel >= 3 {
            return 0.15
        } else if amount > 1000 {
            return 0.1
        }
        return 0.05
    },
    "ValidateAge": func(age int) bool {
        return age >= 18 && age <= 120
    },
})

// 使用自定义函数的规则
customRule := rule.SimpleRule{
    When: "ValidateAge(Params.Age) && IsVip(Params.VipLevel)",
    Then: map[string]string{
        "Result[\"DiscountRate\"]": "GetDiscountRate(Params.VipLevel, Params.Amount)",
        "Result[\"DiscountAmount\"]": "CalculateDiscount(Params.Amount, GetDiscountRate(Params.VipLevel, Params.Amount))",
    },
}
```

### 批量规则执行

```go
type OrderCustomer struct {
    Amount        float64 `json:"amount"`
    Age           int     `json:"age"`
    PurchaseCount int     `json:"purchase_count"`
}

// 定义多个规则
batchRules := []interface{}{
    rule.SimpleRule{
        When: "Params.Amount > 500",
        Then: map[string]string{
            "Result[\"FreeShipping\"]": "true",
        },
    },
    rule.SimpleRule{
        When: "Params.Age > 60", 
        Then: map[string]string{
            "Result[\"SeniorDiscount\"]": "0.05",
        },
    },
    rule.SimpleRule{
        When: "Params.PurchaseCount > 10",
        Then: map[string]string{
            "Result[\"LoyaltyBonus\"]": "true",
        },
    },
}

// 输入数据
inputData := OrderCustomer{
    Amount:        600.0,
    Age:           65,
    PurchaseCount: 15,
}

// 批量执行
results, err := dynamicEngine.ExecuteBatch(context.Background(), batchRules, inputData)
if err != nil {
    log.Printf("批量执行失败: %v", err)
} else {
    for i, result := range results {
        fmt.Printf("规则%d结果: %+v\n", i+1, result)
    }
}
```

### 适用场景
- ✅ 快速原型开发和测试
- ✅ 临时性或实验性规则
- ✅ 规则逻辑相对简单
- ✅ 不需要复杂的规则管理功能

## 🎯 引擎选择指南

### 场景决策树

```
是否需要规则持久化存储？
├── 是 → 是否需要支持多种返回类型？
│   ├── 是 → 通用引擎（BaseEngine + TypedEngine）
│   └── 否 → 传统引擎（runehammer.New[T]）
└── 否 → 动态引擎（DynamicEngine）
```

### 性能对比

| 引擎类型 | 启动成本 | 内存占用 | 执行性能 | 扩展性 |
|---------|---------|---------|---------|-------|
| 传统引擎 | 高 | 中等 | 最优 | 中等 |
| 通用引擎 | 中等 | 低 | 优秀 | 最优 |
| 动态引擎 | 低 | 最低 | 良好 | 优秀 |

### 推荐搭配

```go
// 企业级应用推荐配置
userEngine, _ := runehammer.New[UserResult](
    runehammer.WithDSN("mysql://..."),
    runehammer.WithRedis("localhost:6379", "", 0),
    runehammer.WithCacheTTL(30*time.Minute),
)

// 微服务架构推荐配置
baseEngine, _ := runehammer.NewBaseEngine(
    runehammer.WithDSN("mysql://..."),
    runehammer.WithRedis("localhost:6379", "", 0),
)
userEngine := runehammer.NewTypedEngine[UserResult](baseEngine)
orderEngine := runehammer.NewTypedEngine[OrderResult](baseEngine)

// 快速开发推荐配置
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache:       true,
        ParallelExecution: true,
    },
)
```

## 📋 总结

选择合适的引擎类型是项目成功的关键：

- **传统引擎**: 适合固定业务场景，追求极致性能
- **通用引擎**: 适合多样化需求，平衡性能与灵活性  
- **动态引擎**: 适合快速开发，注重开发效率

根据您的具体业务需求、技术架构和团队能力来选择最合适的方案。更多详细信息请参考：

- [规则语法指南](./RULES_SYNTAX.md) - 详细的规则语法和枚举类型
- [最佳实践指南](./BEST_PRACTICES.md) - 性能优化和开发规范
- [完整示例合集](./EXAMPLES.md) - 更多实际使用示例