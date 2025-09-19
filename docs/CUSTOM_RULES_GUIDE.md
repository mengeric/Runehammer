# 🔧 Runehammer 自定义规则使用指南

## 📚 概述

Runehammer 规则引擎提供了三种使用方式，支持不同的业务场景和技术需求。本指南将介绍核心概念和快速入门方法。

## ⚠️ 重要：字段访问规范

### 核心规范
- **入参访问**: `Params.字段名`（大写P开头，字段名使用大驼峰）
- **返参访问**: `Result["字段名"]`（map访问形式）

### 示例
```go
type UserInput struct {
    Age        int     `json:"age"`         // JSON小写，规则中用大驼峰
    UserName   string  `json:"user_name"`   // JSON下划线，规则中用大驼峰  
    TotalScore float64 `json:"total_score"` // JSON下划线，规则中用大驼峰
}

// 规则中的正确访问方式：
"Params.Age >= 18"
"Params.UserName != ''"
"Result[\"IsValid\"] = true"
"Result[\"FinalScore\"] = Params.TotalScore * 1.2"
```

## 🎯 引擎类型对比

| 特性 | 传统引擎 | 通用引擎 | 动态引擎 |
|------|----------|----------|----------|
| 规则存储 | 数据库 | 数据库 | 运行时定义 |
| 返回类型 | 编译时指定 | 运行时灵活 | 运行时灵活 |
| 资源使用 | 多实例 | 单实例共享 | 轻量级 |
| 适用场景 | 固定业务 | 多样化需求 | 快速原型 |

## 🚀 快速开始

### 传统引擎
```go
// 适合：企业级应用，固定业务场景
userEngine, _ := runehammer.New[UserResult](
    runehammer.WithDSN("mysql://..."),
    runehammer.WithAutoMigrate(),
    runehammer.WithRedis("localhost:6379", "", 0),
)
result, _ := userEngine.Exec(ctx, "USER_VALIDATE", userData)
```

### 通用引擎
```go
// 适合：微服务架构，多种返回类型
baseEngine, _ := runehammer.NewBaseEngine(
    runehammer.WithDSN("mysql://..."),
    runehammer.WithRedis("localhost:6379", "", 0),
)
userEngine := runehammer.NewTypedEngine[UserResult](baseEngine)
orderEngine := runehammer.NewTypedEngine[OrderResult](baseEngine)
```

### 动态引擎
```go
// 适合：快速开发，临时规则
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache: true,
        ParallelExecution: true,
    },
)

simpleRule := rule.SimpleRule{
    When: "Params.Age >= 18",
    Then: map[string]string{
        "Result[\"Adult\"]": "true",
    },
}
result, _ := dynamicEngine.ExecuteRuleDefinition(ctx, simpleRule, input)
```

## 🆕 类型安全枚举系统

### 快速构建规则
```go
// 使用枚举类型（推荐）
rule := rule.NewStandardRule("user_validation", "用户验证规则").
    AddSimpleCondition("Params.Age", rule.OpGreaterThanOrEqual, 18).
    AddSimpleCondition("Params.Income", rule.OpGreaterThan, 50000).
    AddAction(rule.ActionTypeAssign, "Result[\"Eligible\"]", true)

// 主要枚举常量
rule.OpEqual, rule.OpGreaterThan, rule.OpLessThan        // 比较操作符
rule.OpIn, rule.OpContains, rule.OpBetween              // 集合操作符
rule.ActionTypeAssign, rule.ActionTypeCalculate         // 动作类型
rule.ConditionTypeAnd, rule.ConditionTypeOr             // 条件类型
```

## 🛠️ 基本规则语法

### GRL 语法结构
```grl
rule RuleName "规则描述" salience 100 {
    when 条件表达式
    then 
        Result["字段名"] = 值;
        其他操作;
}
```

### 常用表达式
```grl
// 条件判断
Params.Age >= 18 && Params.Income > 50000
Params.IsVip == true || Params.OrderAmount > 1000

// 结果赋值
Result["IsValid"] = true
Result["UserLevel"] = "premium"
Result["FinalAmount"] = Params.OrderAmount * 0.85

// 条件赋值
Result["StatusMessage"] = Params.Age >= 18 ? "成年人" : "未成年人"
```

## 💡 最佳实践

### 引擎选择指南
```go
// 企业级固定业务 → 传统引擎
engine, _ := runehammer.New[ResultType](options...)

// 微服务多业务 → 通用引擎  
baseEngine, _ := runehammer.NewBaseEngine(options...)
typedEngine := runehammer.NewTypedEngine[ResultType](baseEngine)

// 快速原型开发 → 动态引擎
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](config)
```

### 性能优化
```go
// 启用缓存
runehammer.WithRedis("localhost:6379", "", 0)
runehammer.WithCacheTTL(30*time.Minute)

// 批量执行
results, _ := dynamicEngine.ExecuteBatch(ctx, rules, input)

// 并发配置
config := engine.DynamicEngineConfig{
    ParallelExecution: true,
    DefaultTimeout:    5 * time.Second,
}
```

### 错误处理
```go
result, err := engine.Exec(ctx, bizCode, input)
if err != nil {
    switch {
    case errors.Is(err, runehammer.ErrNoRulesFound):
        // 处理规则不存在
    case errors.Is(err, context.DeadlineExceeded):
        // 处理超时
    default:
        // 其他错误
    }
}
```

## 📚 详细文档导航

| 文档 | 说明 | 适用场景 |
|------|------|----------|
| [引擎使用指南](./ENGINES_USAGE.md) | 三种引擎详细使用方法和选择策略 | 技术选型、实现参考 |
| [规则语法指南](./RULES_SYNTAX.md) | GRL语法、枚举系统、复杂条件构建 | 规则编写、语法学习 |  
| [最佳实践指南](./BEST_PRACTICES.md) | 性能优化、开发规范、监控调试 | 生产实践、代码质量 |
| [完整示例合集](./EXAMPLES.md) | 完整可运行的示例代码 | 参考实现、快速上手 |
| [API参考文档](./API_REFERENCE.md) | 完整的API文档和配置选项 | 开发参考、集成指南 |
| [性能优化指南](./PERFORMANCE.md) | 生产环境性能优化策略 | 生产部署、性能调优 |

## 🎯 核心规范总结

### 字段访问规范
1. **入参**: 必须使用 `Params.字段名`（大驼峰）
2. **返参**: 必须使用 `Result["字段名"]`（map形式）
3. **JSON标签**: 可用下划线，规则中仍用大驼峰访问

### 枚举类型使用
1. **推荐**: 使用类型安全的枚举常量
2. **兼容**: 支持显式类型转换 `rule.Operator(">=")` 
3. **禁止**: 直接字符串（编译错误）

### 引擎选择
- **传统引擎**: 固定业务 + 高性能要求
- **通用引擎**: 多业务场景 + 资源优化  
- **动态引擎**: 快速开发 + 灵活配置

选择合适的引擎类型和规则定义方式，遵循规范的字段访问方式，结合类型安全的枚举系统，可以大大提高开发效率和系统性能。
