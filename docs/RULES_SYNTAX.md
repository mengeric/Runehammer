# 🔍 Runehammer 规则语法指南

## 📚 概述

Runehammer 规则引擎基于 GRL（Grule Rule Language）语法，并提供了类型安全的枚举系统。本指南将详细介绍规则语法、枚举类型系统和复杂条件构建方法。

## ⚠️ 重要：字段访问规范

### 字段访问规则

Runehammer 规则引擎有严格的字段访问规范，必须遵循以下规则：

#### 入参字段访问
- **默认字段名**: `Params`（大写P开头）
- **访问方式**: `Params.字段名`（字段名使用大驼峰形式）
- **示例**: `Params.Age`, `Params.UserName`, `Params.OrderAmount`

#### 返参字段访问  
- **默认字段名**: `Result`（大写R开头）
- **访问方式**: `Result["字段名"]`（使用map访问形式）
- **示例**: `Result["IsValid"]`, `Result["TotalScore"]`, `Result["DiscountRate"]`

### 字段命名规范

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
"Result[\"IsValid\"] = true"
"Result[\"FinalScore\"] = Params.TotalScore * 1.2"

// ❌ 错误的访问方式
// "params.age >= 18"        // params 小写
// "Params.age >= 18"        // age 小写
// "user.Age >= 18"          // 错误：应使用 Params.Age
// "result.isValid = true"   // result 小写
```

### JSON 标签 vs 规则访问

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

## 🔧 GRL 语法基础

### 基本语法结构

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
| 输出赋值 | `Result["字段名"]` | `Result["IsValid"] = true`, `Result["Score"] = 85` | 使用map访问形式 |

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
// 基本赋值（注意：返回字段使用map形式）
Result["IsAdult"] = true
Result["UserLevel"] = "premium"
Result["DiscountRate"] = 0.15

// 计算赋值
Result["FinalAmount"] = Params.OrderAmount * 0.85
Result["TotalScore"] = Params.Age * 2 + Params.Income * 0.001

// 条件赋值
Result["StatusMessage"] = Params.Age >= 18 ? "成年人" : "未成年人"
```

### 规则优先级和控制

```grl
// 使用 salience 控制执行顺序，数值越大优先级越高
rule InputValidation "输入验证" salience 1000 {
    when Params == nil || Params.UserId == ""
    then 
        Result["Error"] = "输入数据无效";
        Retract("InputValidation");  // 验证失败后退出
}

rule VipUserRule "VIP用户规则" salience 500 {
    when Params.VipLevel >= 3
    then Result["Discount"] = 0.2;
}

rule RegularUserRule "普通用户规则" salience 100 {
    when Result["Discount"] == nil
    then Result["Discount"] = 0.05;
}
```

## 🎯 类型安全的枚举系统

### 枚举类型概述

Runehammer v1.0 引入了类型安全的枚举系统，提供更好的开发体验：

- **编译时检查**: 避免拼写错误
- **IDE 支持**: 自动补全、类型提示、重构安全
- **向后兼容**: 支持传统字符串写法的显式转换

### 可用的枚举常量

#### 条件类型枚举 (ConditionType)
```go
rule.ConditionTypeSimple     // "simple"     - 简单条件
rule.ConditionTypeComposite  // "composite"  - 复合条件
rule.ConditionTypeExpression // "expression" - 表达式条件
rule.ConditionTypeFunction   // "function"   - 函数条件
rule.ConditionTypeAnd        // "and"        - 逻辑与
rule.ConditionTypeOr         // "or"         - 逻辑或
rule.ConditionTypeNot        // "not"        - 逻辑非
```

#### 操作符枚举 (Operator)
```go
// 比较操作符
rule.OpEqual              // "=="
rule.OpNotEqual           // "!="
rule.OpGreaterThan        // ">"
rule.OpLessThan           // "<"
rule.OpGreaterThanOrEqual // ">="
rule.OpLessThanOrEqual    // "<="

// 逻辑操作符
rule.OpAnd                // "and"
rule.OpOr                 // "or"
rule.OpNot                // "not"

// 集合操作符
rule.OpIn                 // "in"
rule.OpNotIn              // "notIn"
rule.OpContains           // "contains"
rule.OpMatches            // "matches"
rule.OpBetween            // "between"
```

#### 动作类型枚举 (ActionType)
```go
rule.ActionTypeAssign     // "assign"    - 赋值
rule.ActionTypeCalculate  // "calculate" - 计算
rule.ActionTypeInvoke     // "invoke"    - 调用函数
rule.ActionTypeAlert      // "alert"     - 告警
rule.ActionTypeLog        // "log"       - 记录日志
rule.ActionTypeStop       // "stop"      - 停止执行
```

### 快速构建规则（推荐方式）

```go
// 使用工厂方法和链式调用
rule := rule.NewStandardRule("user_validation", "用户验证规则").
    AddSimpleCondition("Params.Age", rule.OpGreaterThanOrEqual, 18).
    AddSimpleCondition("Params.Income", rule.OpGreaterThan, 50000).
    AddSimpleCondition("Params.Status", rule.OpEqual, "active").
    AddAction(rule.ActionTypeAssign, "Result[\"Eligible\"]", true).
    AddAction(rule.ActionTypeCalculate, "Result[\"Score\"]", "Params.Age * 2 + Params.Income * 0.001")
```

### 完整可运行示例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "gitee.com/damengde/runehammer/engine"
    "gitee.com/damengde/runehammer/rule"
)

func main() {
    fmt.Println("=== 枚举类型使用完整示例 ===")
    
    // 创建动态引擎
    dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
        engine.DynamicEngineConfig{
            EnableCache: true,
            CacheTTL:    5 * time.Minute,
        },
    )
    
    // 使用枚举类型快速构建规则
    rule := rule.NewStandardRule("advanced_validation", "高级验证规则").
        AddSimpleCondition("Params.Age", rule.OpGreaterThanOrEqual, 18).
        AddSimpleCondition("Params.Income", rule.OpGreaterThan, 50000).
        AddSimpleCondition("Params.CreditScore", rule.OpBetween, []int{600, 850}).
        AddAction(rule.ActionTypeAssign, "Result[\"Approved\"]", true).
        AddAction(rule.ActionTypeCalculate, "Result[\"CreditLimit\"]", "Params.Income * 5").
        AddAction(rule.ActionTypeLog, "Result[\"Message\"]", "User approved for premium service")
    
    // 输入数据
    input := map[string]interface{}{
        "Age":         30,
        "Income":      80000.0,
        "CreditScore": 750,
    }
    
    // 执行规则
    result, err := dynamicEngine.ExecuteRuleDefinition(context.Background(), *rule, input)
    if err != nil {
        fmt.Printf("❌ 执行失败: %v\n", err)
    } else {
        fmt.Printf("✅ 执行结果: %+v\n", result)
        // 输出: {Approved: true, CreditLimit: 400000, Message: "User approved for premium service"}
    }
    
    fmt.Println("=== 示例完成 ===")
}
```

## 🔄 向后兼容性

```go
// ✅ 新枚举写法（推荐）
rule.AddSimpleCondition("field", rule.OpGreaterThan, 100)

// ✅ 传统字符串写法（兼容）
rule.AddSimpleCondition("field", rule.Operator(">"), 100)

// ❌ 直接字符串（编译错误）
// rule.AddSimpleCondition("field", ">", 100)  // 不再支持
```

## 📊 复杂条件构建

### 嵌套逻辑条件

```go
// 构建复杂的嵌套条件：(年龄>=18) AND ((收入>50000) OR (VIP等级>=3)) AND (状态为活跃)
complexRule := rule.StandardRule{
    ID:          "complex_validation",
    Name:        "复杂验证规则",
    Description: "演示复杂条件构建",
    Conditions: rule.Condition{
        Type: rule.ConditionTypeAnd,  // 主条件：逻辑与
        Children: []rule.Condition{
            {
                Type:     rule.ConditionTypeSimple,
                Left:     "Params.Age",
                Operator: rule.OpGreaterThanOrEqual,
                Right:    18,
            },
            {
                Type: rule.ConditionTypeOr,  // 嵌套条件：逻辑或
                Children: []rule.Condition{
                    {
                        Type:     rule.ConditionTypeSimple,
                        Left:     "Params.Income",
                        Operator: rule.OpGreaterThan,
                        Right:    50000,
                    },
                    {
                        Type:     rule.ConditionTypeSimple,
                        Left:     "Params.VipLevel",
                        Operator: rule.OpGreaterThanOrEqual,
                        Right:    3,
                    },
                },
            },
            {
                Type:     rule.ConditionTypeSimple,
                Left:     "Params.Status",
                Operator: rule.OpIn,
                Right:    []string{"active", "premium"},
            },
        },
    },
    Actions: []rule.Action{
        {
            Type:   rule.ActionTypeAssign,
            Target: "Result[\"Approved\"]",
            Value:  true,
        },
        {
            Type:   rule.ActionTypeCalculate,
            Target: "Result[\"Rating\"]",
            Value:  "Params.Income * 0.001 + Params.VipLevel * 10",
        },
        {
            Type:   rule.ActionTypeLog,
            Target: "audit.log",
            Value:  "User validation completed",
        },
    },
}
```

### 条件表达式类型

```go
// 1. 简单条件 - 单个字段比较
simpleCondition := rule.Condition{
    Type:     rule.ConditionTypeSimple,
    Left:     "Params.Age",
    Operator: rule.OpGreaterThanOrEqual,
    Right:    18,
}

// 2. 表达式条件 - 复杂表达式计算
expressionCondition := rule.Condition{
    Type:       rule.ConditionTypeExpression,
    Expression: "Params.Income * 12 > 100000",
}

// 3. 函数条件 - 自定义函数调用
functionCondition := rule.Condition{
    Type:         rule.ConditionTypeFunction,
    FunctionName: "ValidateEmail",
    Arguments:    []interface{}{"Params.Email"},
}

// 4. 复合条件 - 嵌套逻辑
compositeCondition := rule.Condition{
    Type: rule.ConditionTypeAnd,
    Children: []rule.Condition{simpleCondition, expressionCondition},
}
```

## 🎭 动作类型详解

### 赋值动作 (ActionTypeAssign)

```go
assignAction := rule.Action{
    Type:   rule.ActionTypeAssign,
    Target: "Result[\"UserLevel\"]",
    Value:  "premium",
}
```

### 计算动作 (ActionTypeCalculate)

```go
calculateAction := rule.Action{
    Type:   rule.ActionTypeCalculate,
    Target: "Result[\"FinalScore\"]",
    Value:  "Params.BaseScore * 1.2 + Params.BonusPoints",
}
```

### 函数调用动作 (ActionTypeInvoke)

```go
invokeAction := rule.Action{
    Type:   rule.ActionTypeInvoke,
    Target: "SendNotification",
    Value:  []interface{}{"Params.UserId", "Welcome message"},
}
```

### 日志记录动作 (ActionTypeLog)

```go
logAction := rule.Action{
    Type:   rule.ActionTypeLog,
    Target: "audit.log",
    Value:  "User validation completed successfully",
}
```

## 🚀 IDE 支持和开发体验

### 自动补全

使用枚举类型时，IDE 会提供：
- **自动补全**: 输入 `rule.Op` 时自动提示所有操作符
- **类型检查**: 编译时检查类型匹配，避免拼写错误
- **重构支持**: 安全地重命名和重构代码
- **文档提示**: 悬停显示枚举值的含义

### 代码示例

```go
// IDE 会在输入时提供自动补全
rule := rule.NewStandardRule("example", "示例规则").
    AddSimpleCondition("Params.Amount", rule.Op... /* 这里会显示所有可用操作符 */)
    
// 类型安全 - 编译时错误检查
// rule.AddSimpleCondition("field", "invalid_operator", 100) // ❌ 编译错误
```

## 🔍 高级语法特性

### 内置函数支持

```go
// 数学函数
"Result[\"AbsValue\"] = Abs(Params.Amount)"
"Result[\"MaxValue\"] = Max(Params.Score1, Params.Score2)"
"Result[\"RoundedValue\"] = Round(Params.DecimalValue, 2)"

// 字符串函数
"Result[\"ContainsKeyword\"] = Contains(Params.Description, \"special\")"
"Result[\"UpperCaseName\"] = ToUpper(Params.Name)"
"Result[\"EmailValid\"] = IsEmail(Params.EmailAddress)"

// 时间函数
"Result[\"CurrentTime\"] = Now()"
"Result[\"IsWeekend\"] = IsWeekend(Today())"
"Result[\"DaysFromNow\"] = DaysBetween(Today(), Params.TargetDate)"
```

### 条件表达式高级用法

```go
// 范围检查
"Params.Age >= 18 && Params.Age <= 65"
"Between(Params.Score, 60, 100)"

// 集合检查
"Params.Category in [\"premium\", \"vip\", \"gold\"]"
"Contains([\"admin\", \"manager\"], Params.Role)"

// 正则匹配
"Matches(Params.PhoneNumber, \"^1[3-9]\\\\d{9}$\")"
"Matches(Params.Email, \"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\\\.[a-zA-Z]{2,}$\")"

// 复合条件
"(Params.Age >= 18 && Params.Income > 50000) || Params.VipLevel >= 5"
```

## 📋 语法最佳实践

### 1. 字段命名一致性

```go
// ✅ 推荐：统一使用大驼峰
type UserInput struct {
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    EmailAddress string `json:"email_address"`
}

// 规则中统一访问
"Params.FirstName != ''"
"Params.LastName != ''"  
"IsEmail(Params.EmailAddress)"

// ❌ 避免：混合命名风格
type BadInput struct {
    firstName    string // 小写
    Last_Name    string // 下划线
    emailaddress string // 全小写
}
```

### 2. 规则优先级管理

```go
// 高优先级：数据验证规则
rule DataValidation "数据验证" salience 1000 {
    when Params.UserId == "" || Params.Amount <= 0
    then 
        Result["Error"] = "数据验证失败";
        Retract("DataValidation");
}

// 中优先级：业务逻辑规则  
rule BusinessLogic "业务逻辑" salience 500 {
    when Params.Amount > 1000 && Params.VipLevel >= 3
    then Result["DiscountRate"] = 0.15;
}

// 低优先级：默认处理规则
rule DefaultRule "默认规则" salience 100 {
    when Result["DiscountRate"] == nil
    then Result["DiscountRate"] = 0.05;
}
```

### 3. 错误处理和退出机制

```go
// 使用 Retract 避免重复执行
rule ProcessOrder "订单处理" salience 100 {
    when Params.Status == "pending"
    then 
        Result["ProcessStatus"] = "processing";
        Retract("ProcessOrder");  // 处理后立即退出
}

// 错误处理规则
rule ErrorHandler "错误处理" salience 999 {
    when Params == nil
    then 
        Result["Error"] = "输入参数为空";
        Result["Success"] = false;
        Complete();  // 完全停止规则执行
}
```

## 📊 总结

Runehammer 的规则语法系统提供了：

### 🆕 v1.0 新特性
- **类型安全枚举**: 编译时检查，避免错误
- **IDE 友好**: 自动补全和类型提示
- **链式构建**: 快速构建复杂规则
- **向后兼容**: 支持传统字符串用法

### 核心语法规范
1. **字段访问**: `Params.字段名`（大驼峰） + `Result["字段名"]`
2. **条件表达式**: 支持比较、逻辑、函数、集合操作
3. **枚举类型**: 类型安全的条件类型、操作符、动作类型
4. **优先级控制**: 使用 `salience` 和 `Retract()`

### 最佳实践
- 统一使用枚举常量而非字符串
- 合理设置规则优先级  
- 添加错误处理和退出机制
- 保持字段命名一致性

更多实际使用示例请参考：
- [引擎使用指南](./ENGINES_USAGE.md) - 详细的引擎使用方法
- [最佳实践指南](./BEST_PRACTICES.md) - 性能优化和开发规范
- [完整示例合集](./EXAMPLES.md) - 更多实际使用案例