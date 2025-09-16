# 🔧 Runehammer 自定义规则使用指南

## 📚 概述

Runehammer规则引擎提供了多种自定义规则的使用方式，支持不同的业务场景和技术需求。本指南将详细介绍各种规则定义和使用方法。

## ⚠️ 重要说明

**所有Runehammer引擎都不支持 `map[string]interface{}` 作为输入数据**，因为底层的 grule-rule-engine 不支持 map 类型的解析。请始终使用结构体作为输入数据类型。返回值可以是 `map[string]interface{}` 类型。

## 🎯 规则引擎类型对比

| 特性 | 传统引擎 | 通用引擎 | 动态引擎 |
|------|----------|----------|----------|
| 规则存储 | 数据库 | 数据库 | 运行时定义 |
| 返回类型 | 编译时指定 | 运行时灵活 | 运行时灵活 |
| 资源使用 | 多实例 | 单实例共享 | 轻量级 |
| 适用场景 | 固定业务 | 多样化需求 | 快速原型 |

## 📖 使用方式详解

### 1. 🏛️ 传统引擎（Database-Based Rules）

#### 数据库规则定义

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
  when Params.age >= 18 && Params.income > 50000 
  then 
    Result.Adult = true; 
    Result.Eligible = true;
    Result.Level = "premium";
}', true);

-- 插入订单处理规则
INSERT INTO runehammer_rules (biz_code, name, grl, enabled) VALUES 
('ORDER_PROCESS', '订单处理规则', 
'rule OrderProcess "订单处理规则" { 
  when Params.amount > 1000 && Params.vip == true
  then 
    Result.Discount = 0.15; 
    Result.Priority = "high";
    Result.FreeShipping = true;
}', true);
```

#### Go代码实现

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "your-project/runehammer"
)

// 业务数据结构
type User struct {
    Age    int     `json:"age"`
    Income float64 `json:"income"`
    VIP    bool    `json:"vip"`
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
    // 创建传统引擎实例 - 每种返回类型需要独立实例
    userEngine, err := runehammer.New[ValidationResult](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
        runehammer.WithLogger(runehammer.NewConsoleLogger()),
        runehammer.WithRedisCache("localhost:6379", 0),
    )
    if err != nil {
        log.Fatal("创建用户引擎失败:", err)
    }
    defer userEngine.Close()
    
    // 创建订单引擎实例
    orderEngine, err := runehammer.New[OrderResult](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
        runehammer.WithLogger(runehammer.NewConsoleLogger()),
    )
    if err != nil {
        log.Fatal("创建订单引擎失败:", err)
    }
    defer orderEngine.Close()
    
    // 执行用户验证规则
    userData := User{
        Age:    25,
        Income: 80000.0,
        VIP:    true,
    }
    
    userResult, err := userEngine.Exec(context.Background(), "USER_VALIDATE", userData)
    if err != nil {
        log.Fatal("执行用户验证失败:", err)
    }
    
    fmt.Printf("用户验证结果: Adult=%v, Eligible=%v, Level=%s\\n", 
        userResult.Adult, userResult.Eligible, userResult.Level)
    
    // 执行订单处理规则
    orderData := Order{
        Amount: 1500.0,
        VIP:    true,
    }
    
    orderResult, err := orderEngine.Exec(context.Background(), "ORDER_PROCESS", orderData)
    if err != nil {
        log.Fatal("执行订单处理失败:", err)
    }
    
    fmt.Printf("订单处理结果: Discount=%.2f, Priority=%s, FreeShipping=%v\\n", 
        orderResult.Discount, orderResult.Priority, orderResult.FreeShipping)
}
```

### 2. 🚀 通用引擎（Universal Engine）

#### 核心优势
- **资源共享**: 一个BaseEngine实例支持多种返回类型
- **动态类型**: 运行时决定返回类型，无需编译时指定
- **统一管理**: 数据库连接、缓存、配置统一管理

#### 使用示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "your-project/runehammer"
)

func main() {
    // ============================================================================
    // 启动时创建单个BaseEngine实例
    // ============================================================================
    
    baseEngine, err := runehammer.NewBaseEngine(
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
        runehammer.WithLogger(runehammer.NewConsoleLogger()),
        runehammer.WithRedisCache("localhost:6379", 0),
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
    
    // 通用map引擎 - 返回灵活的map类型（注意：这里map作为返回类型，不是输入类型）
    mapEngine := runehammer.NewTypedEngine[map[string]interface{}](baseEngine)
    
    // ============================================================================
    // 测试数据
    // ============================================================================
    
    // 用户数据
    userData := User{
        Age:    25,
        Income: 80000.0,
        VIP:    true,
    }
    
    // 订单数据  
    orderData := Order{
        Amount: 1500.0,
        VIP:    true,
    }
    
    ctx := context.Background()
    
    // ============================================================================
    // 演示：同一个BaseEngine支持多种返回类型
    // ============================================================================
    
    // 用户验证 - 强类型结构体结果
    userResult, err := userEngine.Exec(ctx, "USER_VALIDATE", userData)
    if err != nil {
        log.Printf("用户验证失败: %v", err)
    } else {
        fmt.Printf("👤 用户验证结果: Adult=%v, Eligible=%v, Level=%s\\n", 
            userResult.Adult, userResult.Eligible, userResult.Level)
    }
    
    // 订单处理 - 强类型结构体结果
    orderResult, err := orderEngine.Exec(ctx, "ORDER_PROCESS", orderData)
    if err != nil {
        log.Printf("订单处理失败: %v", err)
    } else {
        fmt.Printf("🛒 订单处理结果: Discount=%.2f, Priority=%s, FreeShipping=%v\\n", 
            orderResult.Discount, orderResult.Priority, orderResult.FreeShipping)
    }
    
    // 通用map - 灵活的map结果
    mapResult, err := mapEngine.Exec(ctx, "USER_VALIDATE", userData)
    if err != nil {
        log.Printf("通用执行失败: %v", err)
    } else {
        fmt.Printf("🗂️  通用map结果: %+v\\n", mapResult)
    }
}
```

### 3. ⚡ 动态引擎（Dynamic Engine）

#### 特点
- **运行时定义**: 无需数据库存储，直接在代码中定义规则
- **快速原型**: 适合快速开发和测试
- **灵活配置**: 支持缓存、并发、超时等高级配置

#### 规则类型

##### 3.1 简单规则（SimpleRule）

```go
// 年龄数据结构
type AgeData struct {
    Age int `json:"age"`
}

// 年龄验证规则
ageRule := runehammer.SimpleRule{
    When: "agedata.Age >= 18", // 条件表达式
    Then: map[string]string{
        "Result.Adult":   "true",
        "Result.Message": "\\"符合年龄要求\\"",
    },
}

// 执行规则
ageData := AgeData{Age: 25}
result, err := engine.ExecuteRuleDefinition(context.Background(), ageRule, ageData)
```

##### 3.2 指标规则（MetricRule）

```go
// 客户评分计算
scoreRule := runehammer.MetricRule{
    Name:        "customer_score",
    Description: "客户综合评分",
    Formula:     "age_score + income_score + vip_score",
    Variables: map[string]string{
        "age_score":    "customer.Age * 0.1",
        "income_score": "customer.Income * 0.0001", 
        "vip_score":    "customer.VipLevel * 10",
    },
    Conditions: []string{
        "customer.Age >= 18",
        "customer.Income > 0",
    },
}

type Customer struct {
    Age      int     `json:"age"`
    Income   float64 `json:"income"`
    VipLevel int     `json:"vip_level"`
}

customer := Customer{Age: 30, Income: 80000, VipLevel: 3}
result, err := engine.ExecuteRuleDefinition(context.Background(), scoreRule, customer)
```

##### 3.3 标准规则（StandardRule）

```go
// 复杂业务规则
discountRule := runehammer.StandardRule{
    ID:          "order_discount",
    Name:        "订单折扣规则",
    Description: "根据客户等级和订单金额计算折扣",
    Priority:    100,
    Enabled:     true,
    Tags:        []string{"discount", "order"},
    Conditions: runehammer.Condition{
        Type: "and",
        Children: []runehammer.Condition{
            {
                Type:     "simple",
                Left:     "order.Amount",
                Operator: ">",
                Right:    500,
            },
            {
                Type:     "simple", 
                Left:     "customer.VipLevel",
                Operator: ">=",
                Right:    2,
            },
        },
    },
    Actions: []runehammer.Action{
        {
            Type:   "assign",
            Target: "Result.DiscountRate",
            Value:  0.15,
        },
        {
            Type:   "assign",
            Target: "Result.DiscountAmount", 
            Value:  "order.Amount * 0.15",
        },
    },
}
```

#### 自定义函数注册

```go
// 创建动态引擎（注意：这里map作为返回类型，不是输入类型）
engine := runehammer.NewDynamicEngine[map[string]interface{}](
    runehammer.DynamicEngineConfig{
        EnableCache:       true,
        CacheTTL:          5 * time.Minute,
        MaxCacheSize:      100,
        StrictValidation:  true,
        ParallelExecution: true,
        DefaultTimeout:    10 * time.Second,
    },
)

// 注册单个自定义函数
engine.RegisterCustomFunction("CalculateDiscount", func(amount float64, rate float64) float64 {
    return amount * rate
})

// 批量注册自定义函数
engine.RegisterCustomFunctions(map[string]interface{}{
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
customRule := runehammer.SimpleRule{
    When: "ValidateAge(customer.Age) && IsVip(customer.VipLevel)",
    Then: map[string]string{
        "Result.DiscountRate": "GetDiscountRate(customer.VipLevel, order.Amount)",
        "Result.DiscountAmount": "CalculateDiscount(order.Amount, GetDiscountRate(customer.VipLevel, order.Amount))",
    },
}
```

#### 批量规则执行

```go
// 订单客户数据结构
type OrderCustomer struct {
    Order struct {
        Amount float64 `json:"amount"`
    } `json:"order"`
    Customer struct {
        Age           int `json:"age"`
        PurchaseCount int `json:"purchase_count"`
    } `json:"customer"`
}

// 定义多个规则
batchRules := []interface{}{
    runehammer.SimpleRule{
        When: "ordercustomer.Order.Amount > 500",
        Then: map[string]string{
            "Result.FreeShipping": "true",
        },
    },
    runehammer.SimpleRule{
        When: "ordercustomer.Customer.Age > 60", 
        Then: map[string]string{
            "Result.SeniorDiscount": "0.05",
        },
    },
    runehammer.SimpleRule{
        When: "ordercustomer.Customer.PurchaseCount > 10",
        Then: map[string]string{
            "Result.LoyaltyBonus": "true",
        },
    },
}

// 输入数据
inputData := OrderCustomer{
    Order: struct {
        Amount float64 `json:"amount"`
    }{Amount: 600.0},
    Customer: struct {
        Age           int `json:"age"`
        PurchaseCount int `json:"purchase_count"`
    }{Age: 65, PurchaseCount: 15},
}

// 批量执行
results, err := engine.ExecuteBatch(context.Background(), batchRules, inputData)
if err != nil {
    log.Printf("批量执行失败: %v", err)
} else {
    for i, result := range results {
        fmt.Printf("规则%d结果: %+v\\n", i+1, result)
    }
}
```

## 🔍 规则语法说明

### GRL语法基础

```grl
rule RuleName "规则描述" {
    when 条件表达式
    then 
        结果赋值;
        其他操作;
}
```

### 变量访问方式

| 输入类型 | 访问方式 | 示例 |
|----------|----------|------|
| 结构体 | `结构体名小写.字段名` | `user.Age >= 18` |
| 基本类型 | `Params` | `Params >= 18` |

### 条件表达式

```grl
// 比较操作
Params.age >= 18
Params.income > 50000
Params.name == "张三"

// 逻辑操作
Params.age >= 18 && Params.income > 50000
Params.vip == true || Params.amount > 1000

// 函数调用
IsVip(Params.level)
CalculateScore(Params.age, Params.income)
```

### 结果赋值

```grl
// 基本赋值
Result.Adult = true
Result.Level = "premium"
Result.Discount = 0.15

// 计算赋值
Result.FinalAmount = Params.amount * 0.85
Result.Score = Params.age * 2 + Params.income * 0.001

// 条件赋值
Result.Message = Params.age >= 18 ? "成年人" : "未成年人"
```

## 🛠️ 最佳实践

### 1. 引擎选择指南

```go
// 场景1: 固定业务逻辑，性能要求高
// 推荐：传统引擎
userEngine, _ := runehammer.New[UserResult](options...)

// 场景2: 多样化业务需求，资源优化
// 推荐：通用引擎
baseEngine, _ := runehammer.NewBaseEngine(options...)
userEngine := runehammer.NewTypedEngine[UserResult](baseEngine)
orderEngine := runehammer.NewTypedEngine[OrderResult](baseEngine)

// 场景3: 快速原型，临时规则
// 推荐：动态引擎（注意：这里map作为返回类型，不是输入类型）
dynamicEngine := runehammer.NewDynamicEngine[map[string]interface{}](config)
```

### 2. 性能优化建议

```go
// 1. 启用缓存
runehammer.WithRedisCache("localhost:6379", 0)
runehammer.WithMemoryCache(1000, 10*time.Minute)

// 2. 连接池优化
runehammer.WithDB(dbInstance) // 复用数据库连接

// 3. 批量执行
results, err := engine.ExecuteBatch(ctx, rules, input)

// 4. 并发控制
config := runehammer.DynamicEngineConfig{
    ParallelExecution: true,
    DefaultTimeout:    5 * time.Second,
}
```

### 3. 错误处理

```go
result, err := engine.Exec(ctx, bizCode, input)
if err != nil {
    switch {
    case errors.Is(err, runehammer.ErrNoRulesFound):
        // 处理规则不存在
        log.Printf("规则不存在: %s", bizCode)
    case errors.Is(err, context.DeadlineExceeded):
        // 处理超时
        log.Printf("规则执行超时: %s", bizCode)
    default:
        // 其他错误
        log.Printf("规则执行失败: %v", err)
    }
}
```

### 4. 监控和调试

```go
// 启用详细日志
logger := runehammer.NewConsoleLogger()
logger.SetLevel(runehammer.LogLevelDebug)

// 获取缓存统计
stats := engine.GetCacheStats()
fmt.Printf("缓存命中率: %.2f%%", stats.HitRate*100)

// 清理缓存
engine.ClearCache()
```

## 📊 总结

Runehammer规则引擎提供了三种强大的自定义规则使用方式：

1. **传统引擎**: 适合固定业务场景，性能稳定
2. **通用引擎**: 适合多样化需求，资源优化
3. **动态引擎**: 适合快速开发，灵活配置

选择合适的引擎类型和规则定义方式，可以大大提高开发效率和系统性能。建议根据具体业务需求和技术架构来选择最适合的方案。