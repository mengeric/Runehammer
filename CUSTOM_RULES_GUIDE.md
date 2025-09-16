# 🔧 Runehammer 自定义规则使用指南

## 📚 概述

Runehammer规则引擎提供了多种自定义规则的使用方式，支持不同的业务场景和技术需求。本指南将详细介绍各种规则定义和使用方法。

## ⚠️ 重要说明

**所有Runehammer引擎都不支持 `map[string]interface{}` 作为输入数据**，因为底层的 grule-rule-engine 不支持 map 类型的解析。请始终使用结构体作为输入数据类型。返回值可以是 `map[string]interface{}` 类型。

### 🎯 字段访问规范（重要）

Runehammer 规则引擎有严格的字段访问规范，必须遵循以下规则：

#### 入参字段访问
- **默认字段名**: `Params`（大写P开头）
- **访问方式**: `Params.字段名`（字段名使用大驼峰形式）
- **示例**: `Params.Age`, `Params.UserName`, `Params.OrderAmount`

#### 返参字段访问  
- **默认字段名**: `Result`（大写R开头）
- **访问方式**: `Result.字段名`（字段名使用大驼峰形式）
- **示例**: `Result.IsValid`, `Result.TotalScore`, `Result.DiscountRate`

#### 字段命名规范

```go
// ✅ 正确的字段命名和访问方式
type UserInput struct {
    Age        int     `json:"age"`         // JSON用小写，但规则中用大驼峰
    UserName   string  `json:"user_name"`   // JSON用下划线，但规则中用大驼峰  
    TotalScore float64 `json:"total_score"` // JSON用下划线，但规则中用大驼峰
}

// 在规则中的正确访问方式：
"Params.Age >= 18"
"Params.UserName != ''"
"Params.TotalScore > 80.0"
"Result.IsValid = true"
"Result.FinalScore = Params.TotalScore * 1.2"

// ❌ 错误的访问方式
// "params.age >= 18"        // params 小写
// "Params.age >= 18"        // age 小写
// "user.Age >= 18"          // 错误：应使用 Params.Age
// "result.isValid = true"   // result 小写
```

#### JSON 标签 vs 规则访问

虽然 Go 结构体的 JSON 标签可以使用小写或下划线命名，但在规则表达式中必须使用大驼峰形式：

```go
type OrderData struct {
    OrderAmount    float64 `json:"order_amount"`     // JSON: order_amount
    CustomerLevel  int     `json:"customer_level"`   // JSON: customer_level  
    ShippingAddr   string  `json:"shipping_addr"`    // JSON: shipping_addr
}

// 规则中的访问（注意大驼峰）：
"Params.OrderAmount > 100"      // 不是 order_amount
"Params.CustomerLevel >= 3"     // 不是 customer_level
"Params.ShippingAddr != ''"     // 不是 shipping_addr
```

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
    When: "Params.Age >= 18", // 条件表达式
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
result, err := engine.ExecuteRuleDefinition(context.Background(), scoreRule, customer)
```

##### 3.3 标准规则（StandardRule）

**⚠️ 重要更新：自 v1.0 起，Condition 和 Action 支持类型安全的枚举类型**

```go
// 🎯 推荐写法：使用枚举类型（类型安全）
discountRule := runehammer.StandardRule{
    ID:          "order_discount",
    Name:        "订单折扣规则",
    Description: "根据客户等级和订单金额计算折扣",
    Priority:    100,
    Enabled:     true,
    Tags:        []string{"discount", "order"},
    Conditions: runehammer.Condition{
        Type: runehammer.ConditionTypeAnd, // 🆕 使用枚举类型
        Children: []runehammer.Condition{
            {
                Type:     runehammer.ConditionTypeSimple, // 🆕 使用枚举类型
                Left:     "Params.Amount",
                Operator: runehammer.OpGreaterThan,       // 🆕 使用枚举类型
                Right:    500,
            },
            {
                Type:     runehammer.ConditionTypeSimple,    // 🆕 使用枚举类型
                Left:     "Params.VipLevel",
                Operator: runehammer.OpGreaterThanOrEqual,  // 🆕 使用枚举类型
                Right:    2,
            },
        },
    },
    Actions: []runehammer.Action{
        {
            Type:   runehammer.ActionTypeAssign,    // 🆕 使用枚举类型
            Target: "Result.DiscountRate",
            Value:  0.15,
        },
        {
            Type:   runehammer.ActionTypeCalculate, // 🆕 使用枚举类型
            Target: "Result.DiscountAmount", 
            Value:  "Params.Amount * 0.15",
        },
    },
}

// 📝 也支持传统字符串写法（向后兼容）
legacyRule := runehammer.StandardRule{
    ID:          "legacy_rule",
    Name:        "传统写法示例",
    Description: "演示字符串类型的向后兼容性",
    Conditions: runehammer.Condition{
        Type:     runehammer.ConditionType("simple"),  // 显式转换
        Left:     "Params.Age",
        Operator: runehammer.Operator(">="),           // 显式转换
        Right:    18,
    },
    Actions: []runehammer.Action{
        {
            Type:   runehammer.ActionType("assign"),    // 显式转换
            Target: "Result.IsAdult",
            Value:  true,
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
    When: "ValidateAge(Params.Age) && IsVip(Params.VipLevel)",
    Then: map[string]string{
        "Result.DiscountRate": "GetDiscountRate(Params.VipLevel, Params.Amount)",
        "Result.DiscountAmount": "CalculateDiscount(Params.Amount, GetDiscountRate(Params.VipLevel, Params.Amount))",
    },
}
```

#### 批量规则执行

```go
// 订单客户数据结构
type OrderCustomer struct {
    Amount        float64 `json:"amount"`
    Age           int     `json:"age"`
    PurchaseCount int     `json:"purchase_count"`
}

// 定义多个规则
batchRules := []interface{}{
    runehammer.SimpleRule{
        When: "Params.Amount > 500",
        Then: map[string]string{
            "Result.FreeShipping": "true",
        },
    },
    runehammer.SimpleRule{
        When: "Params.Age > 60", 
        Then: map[string]string{
            "Result.SeniorDiscount": "0.05",
        },
    },
    runehammer.SimpleRule{
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
results, err := engine.ExecuteBatch(context.Background(), batchRules, inputData)
if err != nil {
    log.Printf("批量执行失败: %v", err)
} else {
    for i, result := range results {
        fmt.Printf("规则%d结果: %+v\\n", i+1, result)
    }
}
```

## 🎯 枚举类型使用指南

### 类型安全的枚举系统

Runehammer v1.0 引入了类型安全的枚举系统，提供更好的开发体验：

#### 🔧 快速构建规则（推荐方式）

```go
// 使用工厂方法和链式调用
rule := runehammer.NewStandardRule("user_validation", "用户验证规则").
    AddSimpleCondition("Params.Age", runehammer.OpGreaterThanOrEqual, 18).
    AddSimpleCondition("Params.Income", runehammer.OpGreaterThan, 50000).
    AddSimpleCondition("Params.Status", runehammer.OpEqual, "active").
    AddAction(runehammer.ActionTypeAssign, "Result.Eligible", true).
    AddAction(runehammer.ActionTypeCalculate, "Result.Score", "Params.Age * 2 + Params.Income * 0.001")
```

#### 🎯 完整可运行示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "gitee.com/damengde/runehammer"
)

func main() {
    fmt.Println("=== 枚举类型使用完整示例 ===")
    
    // 创建动态引擎
    engine := runehammer.NewDynamicEngine[map[string]interface{}](
        runehammer.DynamicEngineConfig{
            EnableCache: true,
            CacheTTL:    5 * time.Minute,
        },
    )
    
    // 使用枚举类型快速构建规则
    rule := runehammer.NewStandardRule("advanced_validation", "高级验证规则").
        AddSimpleCondition("Params.Age", runehammer.OpGreaterThanOrEqual, 18).
        AddSimpleCondition("Params.Income", runehammer.OpGreaterThan, 50000).
        AddSimpleCondition("Params.CreditScore", runehammer.OpBetween, []int{600, 850}).
        AddAction(runehammer.ActionTypeAssign, "Result.Approved", true).
        AddAction(runehammer.ActionTypeCalculate, "Result.CreditLimit", "Params.Income * 5").
        AddAction(runehammer.ActionTypeLog, "Result.Message", "User approved for premium service")
    
    // 输入数据
    input := map[string]interface{}{
        "Age":         30,
        "Income":      80000.0,
        "CreditScore": 750,
    }
    
    // 执行规则
    result, err := engine.ExecuteRuleDefinition(context.Background(), *rule, input)
    if err != nil {
        fmt.Printf("❌ 执行失败: %v\n", err)
    } else {
        fmt.Printf("✅ 执行结果: %+v\n", result)
        // 输出: {Approved: true, CreditLimit: 400000, Message: "User approved for premium service"}
    }
    
    // 演示类型安全的好处
    fmt.Println("\n--- 类型安全演示 ---")
    
    // ✅ 正确用法 - 使用枚举常量
    safeRule := runehammer.NewStandardRule("safe_rule", "类型安全规则").
        AddSimpleCondition("Params.Amount", runehammer.OpGreaterThan, 1000).
        AddAction(runehammer.ActionTypeAssign, "Result.Eligible", true)
    
    fmt.Println("✅ 类型安全的规则构建完成")
    
    // 如果尝试使用错误的字符串，编译时就会出错：
    // rule.AddSimpleCondition("Params.Amount", ">", 1000)  // ❌ 编译失败
    
    fmt.Println("=== 示例完成 ===")
}
```

#### 📖 可用的枚举常量

```go
// 条件类型枚举 (ConditionType)
runehammer.ConditionTypeSimple     // "simple"     - 简单条件
runehammer.ConditionTypeComposite  // "composite"  - 复合条件
runehammer.ConditionTypeExpression // "expression" - 表达式条件
runehammer.ConditionTypeFunction   // "function"   - 函数条件
runehammer.ConditionTypeAnd        // "and"        - 逻辑与
runehammer.ConditionTypeOr         // "or"         - 逻辑或
runehammer.ConditionTypeNot        // "not"        - 逻辑非

// 操作符枚举 (Operator)
// 比较操作符
runehammer.OpEqual              // "=="
runehammer.OpNotEqual           // "!="
runehammer.OpGreaterThan        // ">"
runehammer.OpLessThan           // "<"
runehammer.OpGreaterThanOrEqual // ">="
runehammer.OpLessThanOrEqual    // "<="

// 逻辑操作符
runehammer.OpAnd                // "and"
runehammer.OpOr                 // "or"
runehammer.OpNot                // "not"

// 集合操作符
runehammer.OpIn                 // "in"
runehammer.OpNotIn              // "notIn"
runehammer.OpContains           // "contains"
runehammer.OpMatches            // "matches"
runehammer.OpBetween            // "between"

// 动作类型枚举 (ActionType)
runehammer.ActionTypeAssign     // "assign"    - 赋值
runehammer.ActionTypeCalculate  // "calculate" - 计算
runehammer.ActionTypeInvoke     // "invoke"    - 调用函数
runehammer.ActionTypeAlert      // "alert"     - 告警
runehammer.ActionTypeLog        // "log"       - 记录日志
runehammer.ActionTypeStop       // "stop"      - 停止执行
```

#### 🚀 IDE 支持和自动补全

使用枚举类型时，IDE 会提供：
- **自动补全**: 输入 `runehammer.Op` 时自动提示所有操作符
- **类型检查**: 编译时检查类型匹配，避免拼写错误
- **重构支持**: 安全地重命名和重构代码
- **文档提示**: 悬停显示枚举值的含义

#### 🔄 向后兼容性

```go
// ✅ 新枚举写法（推荐）
rule.AddSimpleCondition("field", runehammer.OpGreaterThan, 100)

// ✅ 传统字符串写法（兼容）
rule.AddSimpleCondition("field", runehammer.Operator(">"), 100)

// ❌ 直接字符串（编译错误）
// rule.AddSimpleCondition("field", ">", 100)  // 不再支持
```

#### 📊 复杂条件构建示例

```go
// 构建复杂的嵌套条件
complexRule := runehammer.StandardRule{
    ID:          "complex_validation",
    Name:        "复杂验证规则",
    Description: "演示复杂条件构建",
    Conditions: runehammer.Condition{
        Type: runehammer.ConditionTypeAnd,  // 主条件：逻辑与
        Children: []runehammer.Condition{
            {
                Type:     runehammer.ConditionTypeSimple,
                Left:     "Params.Age",
                Operator: runehammer.OpGreaterThanOrEqual,
                Right:    18,
            },
            {
                Type: runehammer.ConditionTypeOr,  // 嵌套条件：逻辑或
                Children: []runehammer.Condition{
                    {
                        Type:     runehammer.ConditionTypeSimple,
                        Left:     "Params.Income",
                        Operator: runehammer.OpGreaterThan,
                        Right:    50000,
                    },
                    {
                        Type:     runehammer.ConditionTypeSimple,
                        Left:     "Params.VipLevel",
                        Operator: runehammer.OpGreaterThanOrEqual,
                        Right:    3,
                    },
                },
            },
            {
                Type:     runehammer.ConditionTypeSimple,
                Left:     "Params.Status",
                Operator: runehammer.OpIn,
                Right:    []string{"active", "premium"},
            },
        },
    },
    Actions: []runehammer.Action{
        {
            Type:   runehammer.ActionTypeAssign,
            Target: "Result.Approved",
            Value:  true,
        },
        {
            Type:   runehammer.ActionTypeCalculate,
            Target: "Result.Rating",
            Value:  "Params.Income * 0.001 + Params.VipLevel * 10",
        },
        {
            Type:   runehammer.ActionTypeLog,
            Target: "audit.log",
            Value:  "User validation completed",
        },
    },
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

| 输入类型 | 访问方式 | 示例 | 说明 |
|----------|----------|------|------|
| 结构体 | `Params.字段名` | `Params.Age >= 18`, `Params.UserName != ""` | 字段名必须使用大驼峰形式 |
| 基本类型 | `Params` | `Params >= 18` | 直接访问基本类型值 |
| 输出赋值 | `Result.字段名` | `Result.IsValid = true`, `Result.Score = 85` | 返回字段名也必须使用大驼峰形式 |

### 条件表达式

```grl
// 比较操作（注意：字段名使用大驼峰）
Params.Age >= 18
Params.Income > 50000
Params.UserName == "张三"

// 逻辑操作
Params.Age >= 18 && Params.Income > 50000
Params.IsVip == true || Params.OrderAmount > 1000

// 函数调用
IsVip(Params.Level)
CalculateScore(Params.Age, Params.Income)
```

### 结果赋值

```grl
// 基本赋值（注意：返回字段使用大驼峰）
Result.IsAdult = true
Result.UserLevel = "premium"
Result.DiscountRate = 0.15

// 计算赋值
Result.FinalAmount = Params.OrderAmount * 0.85
Result.TotalScore = Params.Age * 2 + Params.Income * 0.001

// 条件赋值
Result.StatusMessage = Params.Age >= 18 ? "成年人" : "未成年人"
```

## 🛠️ 最佳实践

### 1. 字段命名规范建议

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
}

type ValidationResult struct {
    IsValid         bool    `json:"is_valid"`
    ErrorMessage    string  `json:"error_message"`
    UserLevel       string  `json:"user_level"`
    DiscountRate    float64 `json:"discount_rate"`
    RecommendLevel  string  `json:"recommend_level"`
}

// 对应的规则表达式：
"Params.Age >= 18 && Params.UserName != ''"
"Result.IsValid = Params.Age >= 18"
"Result.UserLevel = Params.IsVipMember ? 'premium' : 'standard'"
"Result.DiscountRate = Params.AccountLevel >= 3 ? 0.15 : 0.05"

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

### 2. 引擎选择指南

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

### 4. 枚举类型最佳实践

```go
// ✅ 推荐：使用枚举常量
rule := runehammer.NewStandardRule("discount", "折扣规则").
    AddSimpleCondition("Params.Amount", runehammer.OpGreaterThan, 100).
    AddSimpleCondition("Params.Level", runehammer.OpIn, []string{"vip", "premium"}).
    AddAction(runehammer.ActionTypeCalculate, "Result.Discount", "Params.Amount * 0.1")

// ✅ 允许：显式类型转换（向后兼容）
legacyOperator := runehammer.Operator(">=")
rule.AddSimpleCondition("Params.Age", legacyOperator, 18)

// ❌ 避免：直接使用字符串（编译错误）
// rule.AddSimpleCondition("Params.Age", ">=", 18) // 不再支持

// 🔍 枚举值查看
fmt.Println("等于操作符:", string(runehammer.OpEqual))           // 输出: ==
fmt.Println("赋值动作:", string(runehammer.ActionTypeAssign))     // 输出: assign
fmt.Println("简单条件:", string(runehammer.ConditionTypeSimple)) // 输出: simple
```

### 5. 监控和调试

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

### 🆕 v1.0 新特性：类型安全枚举系统

- **类型安全**: 编译时检查，避免拼写错误
- **IDE 支持**: 自动补全、类型提示、重构安全
- **向后兼容**: 支持传统字符串写法的显式转换
- **开发体验**: 链式调用 API，快速构建复杂规则

### 选择建议

| 场景 | 推荐方案 | 枚举类型使用 | 字段命名规范 |
|------|----------|-------------|-------------|
| 企业级应用 | 传统引擎 + 枚举类型 | 必须使用，提高代码质量 | 严格遵循 Params/Result + 大驼峰 |
| 微服务架构 | 通用引擎 + 枚举类型 | 推荐使用，统一标准 | 严格遵循 Params/Result + 大驼峰 |
| 快速原型 | 动态引擎 + 枚举类型 | 可选使用，便于后期重构 | 建议遵循 Params/Result + 大驼峰 |

### 🎯 核心规范总结

1. **入参访问**: 必须使用 `Params.字段名`（大驼峰）
2. **返参访问**: 必须使用 `Result.字段名`（大驼峰）  
3. **枚举类型**: 推荐使用类型安全的枚举常量
4. **结构体定义**: Go 字段名用大驼峰，JSON 标签可用下划线

选择合适的引擎类型和规则定义方式，结合类型安全的枚举系统和规范的字段访问方式，可以大大提高开发效率和系统性能。建议根据具体业务需求和技术架构来选择最适合的方案。