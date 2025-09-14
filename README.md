# Runehammer

> **奥恩曾替古代弗雷尔卓德部族"把世界符文钉进现实"，用的就是"一把符文锻锤"**

Runehammer 是一个基于 [Grule](https://github.com/hyperjumptech/grule-rule-engine) 的通用规则引擎，专为业务规则与代码解耦、热更新和灵活扩展而设计。

## ✨ 核心特性

- 🔥 **热更新** - 规则存储在数据库，支持运行时动态更新
- 🏷️ **业务分组** - 通过业务码(bizCode)管理不同场景的规则集
- 🔀 **泛型支持** - 支持任意类型的规则执行结果
- ⚡ **高性能缓存** - 二级缓存机制(Redis + 内存)，自动失效与手动清理
- 📦 **版本管理** - 支持规则版本控制，便于灰度发布和回滚
- 🛠️ **简洁API** - 一行代码执行规则，开箱即用
- 🔌 **灵活扩展** - 支持自定义函数注入和多种缓存策略

## 🏗️ 软件架构

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
                       │ • 二级缓存策略   │
                       └──────────────────┘
```

### 执行流程

1. **调用** - 业务方调用 `engine.Exec(ctx, bizCode, input)`
2. **缓存** - 检查规则缓存，未命中则从数据库加载
3. **编译** - 将 GRL 规则编译为可执行的知识库
4. **执行** - 注入上下文数据，执行规则推理
5. **返回** - 收集执行结果，返回业务数据

## 🚀 快速开始

### 安装

```bash
go get gitee.com/damengde/runehammer
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
    created_by VARCHAR(100),                  -- 创建者
    updated_by VARCHAR(100),                  -- 更新者
    
    INDEX idx_biz_code (biz_code),
    INDEX idx_enabled (enabled)
);
```

### 最小化示例

```go
package main

import (
    "context"
    "fmt"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gitee.com/damengde/runehammer"
)

func main() {
    // 连接数据库
    db, _ := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/test?charset=utf8mb4"))
    
    // 创建规则引擎
    engine, err := runehammer.New[map[string]any](
        runehammer.WithDB(db),
        runehammer.WithAutoMigrate(),
    )
    if err != nil {
        panic(err)
    }
    defer engine.Close()
    
    // 准备输入数据
    input := map[string]any{
        "user": map[string]any{
            "age":  25,
            "vip":  true,
            "name": "Alice",
        },
        "order": map[string]any{
            "amount": 1000.0,
        },
    }
    
    // 执行规则
    result, err := engine.Exec(context.Background(), "user_discount", input)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("折扣结果: %+v\n", result)
}
```

对应的 GRL 规则（存储在数据库中）：

```grl
rule UserVipDiscount "VIP用户折扣规则" salience 100 {
    when
        user.vip == true && user.age >= 18 && order.amount >= 500
    then
        result["discount"] = 0.8;
        result["message"] = "VIP用户享受8折优惠";
        Retract("UserVipDiscount");
}

rule RegularDiscount "普通用户折扣规则" salience 50 {
    when
        result["discount"] == nil && order.amount >= 100
    then
        result["discount"] = 0.9;
        result["message"] = "满100元享受9折优惠";
        Retract("RegularDiscount");
}
```

## 📖 详细使用

### 配置选项

```go
engine, err := runehammer.New[YourResultType](
    // 数据库配置
    runehammer.WithDB(db),                                    // 使用现有数据库连接
    runehammer.WithDSN("user:pass@tcp(localhost:3306)/db"),  // 或使用连接字符串
    runehammer.WithAutoMigrate(),                             // 自动创建表结构
    runehammer.WithTableName("custom_rules"),                // 自定义表名
    
    // 缓存配置
    runehammer.WithRedis("localhost:6379", "", 0),           // Redis缓存
    runehammer.WithCache(customCache),                        // 自定义缓存实现
    runehammer.WithCacheTTL(10*time.Minute),                 // 缓存过期时间
    runehammer.WithMaxCacheSize(1000),                       // 内存缓存大小
    runehammer.WithDisableCache(),                            // 禁用缓存
    
    // 其他配置
    runehammer.WithLogger(logger),                           // 自定义日志器
    runehammer.WithSyncInterval(5*time.Minute),             // 同步间隔
)
```

### 业务场景示例

#### 1. 客户分级规则

```go
// 客户数据结构
type Customer struct {
    ID       string  `json:"id"`
    Age      int     `json:"age"`
    Income   float64 `json:"income"`
    CreditScore int  `json:"credit_score"`
}

// 执行客户分级
input := map[string]any{
    "customer": Customer{
        ID: "C001",
        Age: 35,
        Income: 80000,
        CreditScore: 750,
    },
}

result, err := engine.Exec(ctx, "customer_rating", input)
// result["level"] = "Gold"
// result["credit_limit"] = 50000
```

对应的 GRL 规则：

```grl
rule GoldCustomer "黄金客户评级" salience 100 {
    when
        customer.Age >= 25 && 
        customer.Income >= 50000 && 
        customer.CreditScore >= 700
    then
        result["level"] = "Gold";
        result["credit_limit"] = 50000;
        result["benefits"] = ["专属客服", "优先放款", "费率优惠"];
}

rule SilverCustomer "白银客户评级" salience 80 {
    when
        customer.Age >= 22 && 
        customer.Income >= 30000 && 
        customer.CreditScore >= 600
    then
        result["level"] = "Silver";
        result["credit_limit"] = 20000;
        result["benefits"] = ["在线客服", "标准放款"];
}
```

#### 2. 订单处理规则

```go
// 订单处理
input := map[string]any{
    "order": map[string]any{
        "amount":      1200.0,
        "customer_type": "VIP",
        "region":      "华东",
        "urgent":      true,
    },
    "inventory": map[string]any{
        "stock":    100,
        "reserved": 20,
    },
}

result, err := engine.Exec(ctx, "order_processing", input)
// result["processing_time"] = "2小时"
// result["shipping_cost"] = 0
// result["priority"] = "高"
```

## 📚 API 文档

### Engine 接口

```go
type Engine[T any] interface {
    // 执行规则
    Exec(ctx context.Context, bizCode string, input any) (T, error)
    
    // 关闭引擎，释放资源
    Close() error
}
```

### 配置选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `WithDB(db)` | 使用现有GORM数据库连接 | `WithDB(gormDB)` |
| `WithDSN(dsn)` | 使用数据库连接字符串 | `WithDSN("user:pass@tcp(host)/db")` |
| `WithAutoMigrate()` | 自动创建数据库表 | `WithAutoMigrate()` |
| `WithTableName(name)` | 自定义规则表名 | `WithTableName("my_rules")` |
| `WithRedis(addr, pass, db)` | 配置Redis缓存 | `WithRedis("localhost:6379", "", 0)` |
| `WithCache(cache)` | 使用自定义缓存实现 | `WithCache(myCache)` |
| `WithCacheTTL(ttl)` | 设置缓存过期时间 | `WithCacheTTL(10*time.Minute)` |
| `WithLogger(logger)` | 设置自定义日志器 | `WithLogger(myLogger)` |

### 错误处理

```go
result, err := engine.Exec(ctx, bizCode, input)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "规则未找到"):
        // 处理规则不存在的情况
    case strings.Contains(err.Error(), "编译失败"):
        // 处理规则语法错误
    case strings.Contains(err.Error(), "执行失败"):
        // 处理规则执行错误
    default:
        // 其他错误
    }
}
```

## 🔧 高级特性

### 缓存策略

```go
// 二级缓存：Redis + 内存
engine, _ := runehammer.New[ResultType](
    runehammer.WithRedis("localhost:6379", "", 0),  // 主缓存
    runehammer.WithMaxCacheSize(500),                // 备用内存缓存
    runehammer.WithCacheTTL(30*time.Minute),        // 30分钟过期
)

// 仅内存缓存
engine, _ := runehammer.New[ResultType](
    runehammer.WithMaxCacheSize(1000),
    runehammer.WithDisableCache(), // 先禁用默认缓存
)
```

### 日志集成

```go
// 实现Logger接口
type MyLogger struct {
    logger *zap.Logger
}

func (l *MyLogger) Debugf(ctx context.Context, msg string, keyvals ...any) {
    l.logger.Debug(msg, zap.Any("data", keyvals))
}
// ... 实现其他方法

// 使用自定义日志
engine, _ := runehammer.New[ResultType](
    runehammer.WithLogger(&MyLogger{logger: zapLogger}),
)
```

### 自定义函数注入

当前版本支持以下内置函数：

- `Now()` - 获取当前时间
- `DaysBetween(date1, date2)` - 计算日期差
- `Contains(str, substr)` - 字符串包含检查
- `Len(obj)` - 获取长度
- `Max(a, b)` / `Min(a, b)` - 最大值/最小值

在 GRL 规则中使用：

```grl
rule TimeBasedRule "基于时间的规则" {
    when
        DaysBetween(user.last_login, Now()) > 30
    then
        result["action"] = "send_recall_email";
}
```

## 💡 最佳实践

### 规则设计原则

1. **单一职责** - 每个规则专注解决一个特定问题
2. **优先级管理** - 使用 `salience` 控制规则执行顺序
3. **明确退出** - 使用 `Retract()` 避免重复执行
4. **输入验证** - 在规则中检查必要的输入参数

```grl
rule ValidateInput "输入验证" salience 1000 {
    when
        user == nil || user.id == nil
    then
        result["error"] = "用户信息不完整";
        result["valid"] = false;
        Retract("ValidateInput");
}
```

### 性能优化建议

1. **合理设置缓存时间** - 根据规则变更频率调整TTL
2. **规则分组** - 不同业务场景使用不同的 `bizCode`
3. **避免复杂计算** - 将重计算逻辑前置到输入准备阶段
4. **监控缓存命中率** - 定期检查缓存效果

### 版本管理策略

```sql
-- 发布新版本规则
UPDATE runehammer_rules 
SET version = version + 1, 
    grl = '新的规则内容',
    updated_at = NOW()
WHERE biz_code = 'user_discount';

-- 回滚到指定版本
UPDATE runehammer_rules 
SET enabled = false 
WHERE biz_code = 'user_discount' AND version > 2;
```

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
- 使用 GoConvey BDD 测试风格

## 📄 许可证

本项目采用 Apache 2.0 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [Grule 规则引擎](https://github.com/hyperjumptech/grule-rule-engine)
- [GRL 语法文档](https://hyperjumptech.github.io/grule-rule-engine/)

---

**"愿符文的力量与你同在"** ⚡