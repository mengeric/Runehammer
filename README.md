# Runehammer

> **奥恩曾替古代弗雷尔卓德部族"把世界符文钉进现实"，用的就是"一把符文锻锤"**

Runehammer 是一个基于 [Grule](https://github.com/hyperjumptech/grule-rule-engine) 的通用规则引擎，专为业务规则与代码解耦、热更新和灵活扩展而设计。

## ⚠️ 重要：字段访问规范

为保证规则可读性与一致性，请遵循以下命名约定：

- **入参统一访问**: `Params.字段名`（字段名使用大驼峰）
- **返参统一访问**: `Result.字段名`（字段名使用大驼峰）

示例：`Params.User.Age >= 18`，`Result.IsValid = true`

## ✨ 核心特性

- 🔥 **热更新** - 规则存储在数据库，支持运行时动态更新
- 🏷️ **业务分组** - 通过业务码(bizCode)管理不同场景的规则集
- 🔀 **泛型支持** - 支持任意类型的规则执行结果
- ⚡ **高性能缓存** - 二级缓存机制(Redis + 内存)，自动失效与手动清理
- 📦 **版本管理** - 支持规则版本控制，便于灰度发布和回滚
- 🛠️ **简洁API** - 一行代码执行规则，开箱即用
- 🔌 **灵活扩展** - 支持自定义函数注入和多种缓存策略

## 🚀 快速开始

### 安装

```bash
go get gitee.com/damengde/runehammer
```

### 最小化示例

```go
package main

import (
    "context"
    "fmt"
    "gitee.com/damengde/runehammer"
)

// 定义输入数据结构
type UserInput struct {
    Age    int     `json:"age"`
    Income float64 `json:"income"`
    VIP    bool    `json:"vip"`
}

// 定义结果结构
type DiscountResult struct {
    Discount float64 `json:"discount"`
    Message  string  `json:"message"`
    Level    string  `json:"level"`
}

func main() {
    // 创建规则引擎
    engine, err := runehammer.New[DiscountResult](
        runehammer.WithDSN("sqlite:file:example.db?mode=memory&cache=shared&_fk=1"),
        runehammer.WithAutoMigrate(),
    )
    if err != nil {
        panic(err)
    }
    defer engine.Close()
    
    // 准备输入数据
    input := UserInput{
        Age:    25,
        Income: 80000.0,
        VIP:    true,
    }
    
    // 执行规则
    result, err := engine.Exec(context.Background(), "user_discount", input)
    if err != nil {
        fmt.Printf("执行规则失败: %v\n", err)
    } else {
        fmt.Printf("折扣结果: %+v\n", result)
    }
}
```

对应的 GRL 规则（存储在数据库中）：

```grl
rule UserVipDiscount "VIP用户折扣规则" salience 100 {
    when
        Params.VIP == true && Params.Age >= 18 && Params.Income >= 50000
    then
        Result.Discount = 0.8;
        Result.Message = "VIP用户享受8折优惠";
        Result.Level = "premium";
}
```

### 数据库表结构

```sql
CREATE TABLE runehammer_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    biz_code VARCHAR(100) NOT NULL,           -- 业务标识码
    name VARCHAR(200) NOT NULL,               -- 规则名称  
    grl TEXT NOT NULL,                        -- GRL规则内容
    version INT DEFAULT 1,                    -- 版本号
    enabled BOOLEAN NOT NULL DEFAULT true,   -- 是否启用
    description VARCHAR(500),                 -- 规则描述
    created_at DATETIME NOT NULL,             -- 创建时间
    updated_at DATETIME NOT NULL,             -- 更新时间
    
    INDEX idx_biz_code (biz_code),
    INDEX idx_enabled (enabled)
);
```

## 🎯 三种使用模式

### 1. 传统引擎（推荐企业应用）
- 规则存储在数据库，支持热更新和版本管理
- 每种返回类型需要独立的引擎实例
- 适合稳定的业务场景

```go
engine, err := runehammer.New[YourResultType](
    runehammer.WithDSN("mysql://user:pass@localhost/db"),
    runehammer.WithAutoMigrate(),
    runehammer.WithRedisCache("localhost:6379", "", 0),
)
```

### 2. 通用引擎（推荐微服务）
- 一个BaseEngine实例支持多种返回类型
- 资源共享，减少连接开销
- 运行时决定返回类型

```go
baseEngine, _ := runehammer.NewBaseEngine(options...)
userEngine := runehammer.NewTypedEngine[UserResult](baseEngine)
orderEngine := runehammer.NewTypedEngine[OrderResult](baseEngine)
```

### 3. 动态引擎（推荐快速开发）
- 无需数据库，规则即时定义即时执行
- 支持多种规则格式：简单规则、指标规则、标准规则
- 内置50+函数，支持自定义函数

```go
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache:       true,
        ParallelExecution: true,
    },
)
```

## 📋 配置选项

### 缓存配置
```go
// Redis缓存
runehammer.WithRedisCache("localhost:6379", "password", 0)

// 内存缓存
runehammer.WithMemoryCache(1000)

// 禁用缓存
runehammer.WithNoCache()
```

### 其他配置
```go
runehammer.WithDSN("mysql://user:pass@localhost/db")  // 数据库连接
runehammer.WithAutoMigrate()                           // 自动迁移表结构
runehammer.WithCustomLogger(logger)                   // 自定义日志
runehammer.WithSyncInterval(5*time.Minute)           // 同步间隔
```

## 📚 文档导航

| 文档 | 说明 | 适用场景 |
|------|------|----------|
| [README.md](./README.md) | 快速开始、基础配置 | 初次了解、快速上手 |
| [自定义规则使用指南](./docs/CUSTOM_RULES_GUIDE.md) | 概览介绍和核心概念 | 整体了解、快速入门 |
| [引擎使用指南](./docs/ENGINES_USAGE.md) | 三种引擎详细使用方法和选择策略 | 技术选型、实现参考 |
| [规则语法指南](./docs/RULES_SYNTAX.md) | GRL语法、枚举系统、复杂条件构建 | 规则编写、语法学习 |
| [最佳实践指南](./docs/BEST_PRACTICES.md) | 性能优化、开发规范、监控调试 | 生产实践、代码质量 |
| [完整示例合集](./docs/EXAMPLES.md) | 完整可运行的示例代码 | 参考实现、学习使用 |
| [API参考文档](./docs/API_REFERENCE.md) | 完整的API文档和配置选项 | 开发参考、集成指南 |
| [性能优化指南](./docs/PERFORMANCE.md) | 生产环境性能优化策略 | 生产部署、性能调优 |

## 🛠️ 变量访问规范

| 输入数据类型 | 访问方式 | 示例 |
|-------------|----------|------|
| 结构体 | `Params.字段名` | `Params.User.Age >= 18` |
| 基础类型 | `Params` | `Params >= 100` |
| 返回值 | `Result.字段名` | `Result.IsValid = true` |

**注意**：字段名必须使用大驼峰形式，即使Go结构体的JSON标签使用下划线

## 📊 内置函数

提供50+内置函数，涵盖各种常用场景：

### 数学函数
`Abs()`, `Max()`, `Min()`, `Round()`, `Pow()`, `Sqrt()`等

### 字符串函数  
`Contains()`, `HasPrefix()`, `ToUpper()`, `Split()`, `Join()`等

### 时间函数
`Now()`, `Today()`, `AddDays()`, `FormatTime()`等

### 验证函数
`IsEmail()`, `IsPhoneNumber()`, `Matches()`, `Between()`等

完整函数列表请参考 [API参考文档](./docs/API_REFERENCE.md#内置函数参考)

## 💡 最佳实践

### 规则设计
- **单一职责** - 每个规则专注解决一个特定问题
- **优先级管理** - 使用 `salience` 控制规则执行顺序
- **明确退出** - 使用 `Retract()` 避免重复执行

### 性能优化
- **启用缓存** - 根据规则变更频率调整TTL
- **规则分组** - 不同业务场景使用不同的 `bizCode`
- **批量执行** - 对于独立规则使用批量并行执行

### 引擎选择
| 使用场景 | 推荐引擎 | 理由 |
|----------|----------|------|
| 业务规则管理 | 传统引擎 | 支持热更新、版本控制 |
| 指标计算 | 动态引擎 | 实时计算、支持复杂公式 |
| 第三方集成 | 动态引擎 | 多格式支持、标准化接口 |
| 资源优化 | 通用引擎 | 单实例多用途 |

## 🤝 参与贡献

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 代码规范
- 遵循 Go 官方代码规范
- 添加详细的中文注释
- 确保测试覆盖率 ≥ 80%

## 📄 许可证

本项目采用 Apache 2.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [Grule 规则引擎](https://github.com/hyperjumptech/grule-rule-engine)
- [GRL 语法文档](https://hyperjumptech.github.io/grule-rule-engine/)

---

**"愿符文的力量与你同在"** ⚡