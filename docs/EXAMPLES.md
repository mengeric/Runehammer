# 📖 完整示例合集

本文档包含 Runehammer 规则引擎的完整可运行示例，涵盖各种使用场景和配置方式。

## 📚 目录

- [动态引擎示例](#动态引擎示例)
- [结构体输入示例](#结构体输入示例)
- [通用引擎示例](#通用引擎示例)
- [传统引擎示例](#传统引擎示例)
- [高级功能示例](#高级功能示例)

## 🚀 动态引擎示例

### 基本类型输入示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "gitee.com/damengde/runehammer/engine"
    "gitee.com/damengde/runehammer/rule"
)

// 定义结果类型
type DynamicResult struct {
    Adult       bool   `json:"adult"`
    Message     string `json:"message"`
    Discount    int    `json:"discount"`
    IsVip       bool   `json:"is_vip"`
    Privilege   string `json:"privilege"`
    LargeAmount bool   `json:"large_amount"`
    SmallAmount bool   `json:"small_amount"`
}

func main() {
    fmt.Println("=== Runehammer 动态引擎示例 ===")
    
    // 创建动态引擎（动态引擎的返回类型建议使用 map[string]interface{}）
    dynEngine := engine.NewDynamicEngine[map[string]interface{}](
        engine.DynamicEngineConfig{
            EnableCache:       true,
            CacheTTL:          5 * time.Minute,
            MaxCacheSize:      100,
            StrictValidation:  false,
            ParallelExecution: false,
            DefaultTimeout:    10 * time.Second,
        },
    )
    
    // 示例1: 基本类型输入 - 年龄验证
    fmt.Println("\n--- 年龄验证规则 ---")
    ageRule := rule.SimpleRule{
        When: "Params >= 18", // 基本类型使用 Params 直接访问
        Then: map[string]string{
            "Result.Adult":   "true",
            "Result.Message": "\"符合年龄要求\"",
        },
    }
    
    result, err := dynEngine.ExecuteRuleDefinition(context.Background(), ageRule, 25)
    if err != nil {
        fmt.Printf("❌ 执行失败: %v\n", err)
    } else {
        fmt.Printf("✅ 年龄验证结果: %+v\n", result)
        // 输出: map[Adult:true Message:符合年龄要求]
    }
    
    // 示例2: 注册自定义函数
    fmt.Println("\n--- 自定义函数示例 ---")
    dynEngine.RegisterCustomFunction("IsAdult", func(age int) bool {
        return age >= 18
    })
    
    dynEngine.RegisterCustomFunction("CalculateDiscount", func(amount, rate float64) float64 {
        return amount * rate
    })
    
    customFuncRule := rule.SimpleRule{
        When: "IsAdult(Params)",
        Then: map[string]string{
            "Result.Adult":    "true",
            "Result.Discount": "CalculateDiscount(100.0, 0.1)",
        },
    }
    
    result, err = dynEngine.ExecuteRuleDefinition(context.Background(), customFuncRule, 25)
    if err != nil {
        fmt.Printf("❌ 执行失败: %v\n", err)
    } else {
        fmt.Printf("✅ 自定义函数结果: %+v\n", result)
        // 输出: map[Adult:true Discount:10]
    }
    
    // 示例3: 字符串输入
    fmt.Println("\n--- 字符串规则示例 ---")
    stringRule := rule.SimpleRule{
        When: "Params == \"VIP\"",
        Then: map[string]string{
            "Result.IsVip":    "true",
            "Result.Privilege": "\"高级权限\"",
        },
    }
    
    result, err = dynEngine.ExecuteRuleDefinition(context.Background(), stringRule, "VIP")
    if err != nil {
        fmt.Printf("❌ 执行失败: %v\n", err)
    } else {
        fmt.Printf("✅ 字符串规则结果: %+v\n", result)
        // 输出: map[IsVip:true Privilege:高级权限]
    }
    
    // 示例4: 批量规则执行
    fmt.Println("\n--- 批量规则执行示例 ---")
    batchRules := []interface{}{
        rule.SimpleRule{
            When: "Params > 100",
            Then: map[string]string{
                "Result.LargeAmount": "true",
            },
        },
        rule.SimpleRule{
            When: "Params <= 100",
            Then: map[string]string{
                "Result.SmallAmount": "true",
            },
        },
    }
    
    results, err := dynEngine.ExecuteBatch(context.Background(), batchRules, 150)
    if err != nil {
        fmt.Printf("❌ 批量执行失败: %v\n", err)
    } else {
        fmt.Println("✅ 批量执行结果:")
        for i, result := range results {
            fmt.Printf("   规则%d: %+v\n", i+1, result)
        }
        // 输出: 规则1: {LargeAmount: true}
        //      规则2: {}
    }
}
```

## 📊 结构体输入示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "gitee.com/damengde/runehammer/engine"
    "gitee.com/damengde/runehammer/rule"
)

// 定义业务数据结构
type CustomerOrder struct {
    Customer Customer `json:"customer"`
    Order    Order    `json:"order"`
}

type Customer struct {
    Age      int     `json:"age"`
    VipLevel int     `json:"vip_level"`
    Income   float64 `json:"income"`
}

type Order struct {
    Amount   float64 `json:"amount"`
    Quantity int     `json:"quantity"`
}

// 定义结果类型
type StructResult struct {
    Eligible        bool    `json:"eligible"`
    Discount        float64 `json:"discount"`
    CustomerScore   float64 `json:"customer_score"`
}

func main() {
    fmt.Println("=== 结构体输入示例 ===")
    
    // 创建动态引擎（返回 map[string]interface{} 更通用）
    dynEngine := engine.NewDynamicEngine[map[string]interface{}](
        engine.DynamicEngineConfig{
            EnableCache: true,
            CacheTTL:    5 * time.Minute,
        },
    )
    
    // 输入数据
    input := CustomerOrder{
        Customer: Customer{
            Age:      30,
            VipLevel: 3,
            Income:   80000,
        },
        Order: Order{
            Amount:   1200.0,
            Quantity: 2,
        },
    }
    
    // 简单规则示例
    eligibilityRule := rule.SimpleRule{
        When: "Params.Customer.Age >= 18 && Params.Order.Amount > 1000",
        Then: map[string]string{
            "Result.Eligible": "true",
            "Result.Discount": "0.1",
        },
    }
    
    result, err := dynEngine.ExecuteRuleDefinition(context.Background(), eligibilityRule, input)
    if err != nil {
        fmt.Printf("❌ 执行失败: %v\n", err)
    } else {
        fmt.Printf("✅ 资格验证结果: %+v\n", result)
        // 输出: map[Eligible:true Discount:0.1]
    }
    
    // 指标规则示例
    scoreRule := rule.MetricRule{
        Name:        "customer_score",
        Description: "客户综合评分",
        Formula:     "age_score + income_score + vip_score",
        Variables: map[string]string{
            "age_score":    "Params.Customer.Age * 0.1",
            "income_score": "Params.Customer.Income * 0.0001",
            "vip_score":    "Params.Customer.VipLevel * 10",
        },
        Conditions: []string{
            "Params.Customer.Age >= 18",
            "Params.Customer.Income > 0",
        },
    }
    
    result, err = dynEngine.ExecuteRuleDefinition(context.Background(), scoreRule, input)
    if err != nil {
        fmt.Printf("❌ 执行失败: %v\n", err)
    } else {
        fmt.Printf("✅ 评分计算结果: %+v\n", result)
        // 输出: map[customer_score:41]
    }
}
```

## 🌐 通用引擎示例

```go
package main

import (
    "context"
    "fmt"
    "gitee.com/damengde/runehammer"
    logger "gitee.com/damengde/runehammer/logger"
)

func main() {
    fmt.Println("=== 通用引擎使用示例 ===")
    
    // 创建 BaseEngine 实例（仅需一个）
    baseEngine, err := runehammer.NewBaseEngine(
        runehammer.WithDSN("sqlite:file:example.db?mode=memory&cache=shared&_fk=1"),
        runehammer.WithAutoMigrate(),
        runehammer.WithCustomLogger(logger.NewNoopLogger()),
    )
    if err != nil {
        fmt.Printf("❌ 创建BaseEngine失败: %v\n", err)
        return
    }
    defer baseEngine.Close()
    
    // 创建不同类型的 TypedEngine 包装器
    mapEngine := runehammer.NewTypedEngine[map[string]interface{}](baseEngine)
    
    // 测试数据
    testData := map[string]interface{}{
        "age":    25,
        "income": 80000.0,
        "vip":    true,
    }
    
    // 执行规则（注意：实际使用时规则需要在数据库中存在）
    result, err := mapEngine.Exec(context.Background(), "TEST_RULE", testData)
    if err != nil {
        fmt.Printf("❌ 执行规则失败（这是预期的，因为数据库中没有规则）: %v\n", err)
    } else {
        fmt.Printf("✅ 执行结果: %+v\n", result)
    }
    
    fmt.Println("✅ 通用引擎创建成功，可以通过TypedEngine包装器支持多种返回类型")
    fmt.Println("=== 通用引擎示例完成 ===")
}
```

## 🏛️ 传统引擎示例

```go
package main

import (
    "context"
    "fmt"
    "log"
    "gitee.com/damengde/runehammer"
    logger "gitee.com/damengde/runehammer/logger"
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
    // 创建传统引擎实例 - 每种返回类型需要独立实例
    userEngine, err := runehammer.New[ValidationResult](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
        runehammer.WithRedisCache("localhost:6379", "", 0),
        runehammer.WithCustomLogger(logger.NewConsoleLogger()),
    )
    if err != nil {
        log.Fatal("创建用户引擎失败:", err)
    }
    defer userEngine.Close()
    
    // 创建订单引擎实例
    orderEngine, err := runehammer.New[OrderResult](
        runehammer.WithDSN("mysql://user:pass@localhost:3306/ruledb"),
        runehammer.WithAutoMigrate(),
        runehammer.WithCustomLogger(logger.NewConsoleLogger()),
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

## 🔧 高级功能示例

### 自定义函数注册

```go
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
        "Result.DiscountRate": "GetDiscountRate(Params.VipLevel, Params.Amount)",
        "Result.DiscountAmount": "CalculateDiscount(Params.Amount, GetDiscountRate(Params.VipLevel, Params.Amount))",
    },
}
```

### 批量规则执行

```go
// 订单客户数据结构
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
            "Result.FreeShipping": "true",
        },
    },
    rule.SimpleRule{
        When: "Params.Age > 60", 
        Then: map[string]string{
            "Result.SeniorDiscount": "0.05",
        },
    },
    rule.SimpleRule{
        When: "Params.PurchaseCount > 10",
        Then: map[string]string{
            "Result.LoyaltyBonus": "true",
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

### 性能优化配置

```go
// 动态引擎性能优化配置
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache:       true,              // 启用缓存
        CacheTTL:          10 * time.Minute,  // 合理的缓存时间
        MaxCacheSize:      500,               // 足够的缓存空间
        ParallelExecution: true,              // 启用并行执行
        DefaultTimeout:    30 * time.Second,  // 合理的超时时间
    },
)

// 传统引擎性能优化配置
engine, err := runehammer.New[ResultType](
    runehammer.WithDSN("mysql://user:pass@localhost/db"),
    runehammer.WithRedisCache("localhost:6379", "", 0),
    runehammer.WithCacheTTL(30*time.Minute),
    runehammer.WithMaxCacheSize(1000),
)
```

## 🏗️ 软件架构示例

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   业务应用层     │───▶│   Runehammer     │───▶│   规则存储层     │
│                 │    │   Core Engine    │    │                 │
├─────────────────┤    ├──────────────────┤    ├─────────────────┤
│ • API调用       │    │ • 规则编译缓存   │    │ • MySQL/数据库  │
│ • 业务码标识     │    │ • 执行上下文管理 │    │ • 规则表结构    │
│ • 输入输出处理   │    │ • 结果收集处理   │    │ • 版本控制     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                       ┌──────────────────┐
                       │   缓存层(可选)    │
                       ├──────────────────┤
                       │ • Redis 分布式   │
                       │ • Memory 本地    │
                       └──────────────────┘
```

这些示例涵盖了 Runehammer 规则引擎的主要使用场景，从简单的快速开始到复杂的企业级应用都有涉及。你可以根据自己的需求选择合适的示例作为起点。