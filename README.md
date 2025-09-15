# Runehammer

> **奥恩曾替古代弗雷尔卓德部族"把世界符文钉进现实"，用的就是"一把符文锻锤"**

Runehammer 是一个基于 [Grule](https://github.com/hyperjumptech/grule-rule-engine) 的通用规则引擎，专为业务规则与代码解耦、热更新和灵活扩展而设计。

## ✨ 核心特性

### 基础功能
- 🔥 **热更新** - 规则存储在数据库，支持运行时动态更新
- 🏷️ **业务分组** - 通过业务码(bizCode)管理不同场景的规则集
- 🔀 **泛型支持** - 支持任意类型的规则执行结果
- ⚡ **高性能缓存** - 二级缓存机制(Redis + 内存)，自动失效与手动清理
- 📦 **版本管理** - 支持规则版本控制，便于灰度发布和回滚
- 🛠️ **简洁API** - 一行代码执行规则，开箱即用
- 🔌 **灵活扩展** - 支持自定义函数注入和多种缓存策略

### 动态规则引擎
- 🚀 **动态规则生成** - 支持实时生成和执行规则，无需数据库存储
- 🔄 **多格式转换** - 支持多种规则格式互相转换（标准规则、简单规则、指标规则）
- 🌐 **多语法支持** - 支持 SQL、JavaScript 等表达式语法
- 📊 **内置函数库** - 50+ 内置函数，涵盖数学、字符串、时间、验证等功能
- 🔀 **并行执行** - 支持批量规则并行处理，提升执行效率
- 🎯 **第三方集成** - 标准化规则定义格式，便于第三方系统接入

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

// 定义输入数据结构
type UserDiscountInput struct {
    User  User  `json:"user"`
    Order Order `json:"order"`
}

type User struct {
    Age  int    `json:"age"`
    VIP  bool   `json:"vip"`
    Name string `json:"name"`
}

type Order struct {
    Amount float64 `json:"amount"`
}

// 定义结果结构
type DiscountResult struct {
    Discount float64 `json:"discount"`
    Message  string  `json:"message"`
}

func main() {
    // 连接数据库
    db, _ := gorm.Open(mysql.Open("user:pass@tcp(localhost:3306)/test?charset=utf8mb4"))
    
    // 创建规则引擎
    engine, err := runehammer.New[DiscountResult](
        runehammer.WithDB(db),
        runehammer.WithAutoMigrate(),
    )
    if err != nil {
        panic(err)
    }
    defer engine.Close()
    
    // 准备输入数据
    input := UserDiscountInput{
        User: User{
            Age:  25,
            VIP:  true,
            Name: "Alice",
        },
        Order: Order{
            Amount: 1000.0,
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
        userdiscountinput.User.VIP == true && userdiscountinput.User.Age >= 18 && userdiscountinput.Order.Amount >= 500
    then
        result["discount"] = 0.8;
        result["message"] = "VIP用户享受8折优惠";
        Retract("UserVipDiscount");
}

rule RegularDiscount "普通用户折扣规则" salience 50 {
    when
        result["discount"] == nil && userdiscountinput.Order.Amount >= 100
    then
        result["discount"] = 0.9;
        result["message"] = "满100元享受9折优惠";
        Retract("RegularDiscount");
}
```

## 🚀 动态规则引擎

除了传统的数据库存储规则方式，Runehammer 还提供了动态规则引擎，支持实时生成和执行规则，无需预先存储。这对于指标计算、临时规则、第三方系统集成等场景特别有用。

### 核心优势

- **实时执行** - 无需预先存储，规则即时生成即时执行
- **多格式支持** - 支持简单规则、标准规则、指标规则等多种格式
- **语法转换** - 支持将多种表达式语法转换为 GRL
- **内存缓存** - 自动缓存编译结果，提升重复执行效率

### 基本用法

```go
// 定义输入数据结构
type CustomerOrder struct {
    Customer Customer `json:"customer"`
    Order    Order    `json:"order"`
}

type Customer struct {
    Age int `json:"age"`
}

type Order struct {
    Amount float64 `json:"amount"`
}

// 定义结果结构
type EligibilityResult struct {
    Eligible bool    `json:"eligible"`
    Discount float64 `json:"discount"`
}

// 创建动态引擎
dynamicEngine := runehammer.NewDynamicEngine[EligibilityResult](
    runehammer.DynamicEngineConfig{
        EnableCache:       true,
        CacheTTL:          5 * time.Minute,
        StrictValidation:  true,
        ParallelExecution: true,
    },
)

// 执行简单规则
simpleRule := runehammer.SimpleRule{
    When: "customerorder.Customer.Age >= 18 && customerorder.Order.Amount > 100",
    Then: map[string]string{
        "result.Eligible": "true",
        "result.Discount": "0.1",
    },
}

input := CustomerOrder{
    Customer: Customer{Age: 25},
    Order:    Order{Amount: 150.0},
}

result, err := dynamicEngine.ExecuteRuleDefinition(ctx, simpleRule, input)
// result.Eligible = true
// result.Discount = 0.1
```

### 规则类型

#### 1. 简单规则 (SimpleRule)
适用于快速定义简单的条件-结果规则：

```go
rule := runehammer.SimpleRule{
    When: "userinput.User.VIP == true && userinput.Order.Amount > 500",
    Then: map[string]string{
        "result.Priority":     "\"high\"",
        "result.FreeShipping": "true",
    },
}
```

#### 2. 指标规则 (MetricRule)
专门用于指标计算和数据分析：

```go
metricRule := runehammer.MetricRule{
    Name:        "customer_score",
    Description: "客户评分计算",
    Formula:     "age_score + income_score + credit_score",
    Variables: map[string]string{
        "age_score":    "customerdata.Age * 0.1",
        "income_score": "customerdata.Income * 0.0001",
        "credit_score": "customerdata.Credit / 10",
    },
    Conditions: []string{
        "customerdata.Age >= 18",
        "customerdata.Income > 0",
    },
}

result, err := dynamicEngine.ExecuteRuleDefinition(ctx, metricRule, input)
// result.CustomerScore = 计算后的评分
```

#### 3. 标准规则 (StandardRule)
完整的规则定义格式，支持复杂条件和多种动作：

```go
standardRule := runehammer.StandardRule{
    ID:          "loan_approval",
    Name:        "贷款审批规则",
    Description: "根据客户信息进行贷款审批",
    Priority:    100,
    Enabled:     true,
    Tags:        []string{"loan", "approval"},
    Conditions: runehammer.Condition{
        Type:     "composite",
        Operator: "and",
        Children: []runehammer.Condition{
            {
                Type:     "simple",
                Left:     "loanapplication.Customer.Age",
                Operator: ">=",
                Right:    22,
            },
            {
                Type:     "simple",
                Left:     "loanapplication.Customer.CreditScore",
                Operator: ">=",
                Right:    650,
            },
        },
    },
    Actions: []runehammer.Action{
        {
            Type:   "assign",
            Target: "result.approved",
            Value:  true,
        },
        {
            Type:       "calculate",
            Target:     "result.LoanAmount",
            Expression: "loanapplication.Customer.Income * 5",
        },
    },
}
```

### 多语法表达式解析

动态引擎支持多种表达式语法，可以根据来源系统选择合适的语法：

#### SQL-like 语法
```go
parser := runehammer.NewExpressionParser(runehammer.SyntaxTypeSQL)
// "age >= 18 AND income > 30000"
// 转换为: "age >= 18 && income > 30000"
```

#### JavaScript-like 语法
```go
parser := runehammer.NewExpressionParser(runehammer.SyntaxTypeJavaScript)
// "orders.filter(o => o.amount > 100).length > 0"
// 转换为: "Count(Filter(orders, \"amount > 100\")) > 0"
```

### 批量执行

支持批量执行多个规则，提升处理效率：

```go
rules := []interface{}{
    runehammer.SimpleRule{
        When: "order.amount > 100",
        Then: map[string]string{"result.discount": "0.05"},
    },
    runehammer.SimpleRule{
        When: "customer.vip == true",
        Then: map[string]string{"result.vip_bonus": "50"},
    },
}

results, err := dynamicEngine.ExecuteBatch(ctx, rules, input)
// results[0] = 第一个规则的结果
// results[1] = 第二个规则的结果
```

### 自定义函数

动态引擎支持注册自定义函数：

```go
// 注册单个函数
dynamicEngine.RegisterCustomFunction("CalculateDiscount", func(amount float64, rate float64) float64 {
    return amount * rate
})

// 批量注册函数
dynamicEngine.RegisterCustomFunctions(map[string]interface{}{
    "ValidateEmail": func(email string) bool {
        // 邮箱验证逻辑
        return true
    },
    "GetRegionCode": func(address string) string {
        // 地区编码获取逻辑
        return "CN-GD"
    },
})

// 在规则中使用
rule := runehammer.SimpleRule{
    When: "ValidateEmail(Params.customer.email) && Params.order.amount > 0",
    Then: map[string]string{
        "result.discount": "CalculateDiscount(Params.order.amount, 0.1)",
        "result.region":   "GetRegionCode(Params.customer.address)",
    },
}
```

## 📖 详细使用

### 配置选项

#### 数据库引擎配置
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

#### 动态引擎配置
```go
dynamicEngine := runehammer.NewDynamicEngine[ResultType](
    runehammer.DynamicEngineConfig{
        // 基础配置
        EnableCache:       true,              // 启用缓存
        CacheTTL:          5 * time.Minute,   // 缓存过期时间
        MaxCacheSize:      100,               // 最大缓存大小
        StrictValidation:  true,              // 严格验证
        ParallelExecution: true,              // 支持并行执行
        DefaultTimeout:    30 * time.Second,  // 默认超时时间
    },
)

// 或使用完整配置选项
engine, err := runehammer.New[YourResultType](
    // 动态配置
    runehammer.WithDynamicConfig(&runehammer.DynamicConfig{
        // 转换器配置
        ConverterConfig: runehammer.ConverterConfig{
            StrictMode:      false,
            DefaultPriority: 50,
            VariablePrefix: map[string]string{
                "user":     "user",
                "order":    "order",
                "customer": "customer",
            },
        },
        
        // 解析器配置
        ParserConfig: runehammer.ParserConfig{
            DefaultSyntax: runehammer.SyntaxTypeSQL,
            SupportedSyntax: []runehammer.SyntaxType{
                runehammer.SyntaxTypeSQL,
                runehammer.SyntaxTypeJavaScript,
            },
        },
        
        // 执行配置
        ExecutionConfig: runehammer.ExecutionConfig{
            EnableParallel:   true,
            MaxConcurrency:   10,
            ExecutionTimeout: 30 * time.Second,
            MaxRules:         100,
        },
    }),
    
    // 快速配置选项
    runehammer.WithDefaultSyntax(runehammer.SyntaxTypeSQL),
    runehammer.WithMaxConcurrency(10),
    runehammer.WithExecutionTimeout(30 * time.Second),
    runehammer.WithCustomFunctions(map[string]interface{}{
        "CustomFunc": func(x int) int { return x * 2 },
    }),
)
```

### 业务场景示例

#### 1. 客户分级规则

```go
// 客户数据结构
type CustomerRatingInput struct {
    Customer Customer `json:"customer"`
}

type Customer struct {
    ID          string  `json:"id"`
    Age         int     `json:"age"`
    Income      float64 `json:"income"`
    CreditScore int     `json:"credit_score"`
}

// 执行客户分级
input := CustomerRatingInput{
    Customer: Customer{
        ID:          "C001",
        Age:         35,
        Income:      80000,
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
        customerratinginput.Customer.Age >= 25 && 
        customerratinginput.Customer.Income >= 50000 && 
        customerratinginput.Customer.CreditScore >= 700
    then
        result["level"] = "Gold";
        result["credit_limit"] = 50000;
        result["benefits"] = ["专属客服", "优先放款", "费率优惠"];
}

rule SilverCustomer "白银客户评级" salience 80 {
    when
        customerratinginput.Customer.Age >= 22 && 
        customerratinginput.Customer.Income >= 30000 && 
        customerratinginput.Customer.CreditScore >= 600
    then
        result["level"] = "Silver";
        result["credit_limit"] = 20000;
        result["benefits"] = ["在线客服", "标准放款"];
}
```

#### 2. 订单处理规则

```go
// 订单处理结构
type OrderProcessingInput struct {
    Order     Order     `json:"order"`
    Inventory Inventory `json:"inventory"`
}

type Order struct {
    Amount       float64 `json:"amount"`
    CustomerType string  `json:"customer_type"`
    Region       string  `json:"region"`
    Urgent       bool    `json:"urgent"`
}

type Inventory struct {
    Stock    int `json:"stock"`
    Reserved int `json:"reserved"`
}

// 订单处理
input := OrderProcessingInput{
    Order: Order{
        Amount:       1200.0,
        CustomerType: "VIP",
        Region:       "华东",
        Urgent:       true,
    },
    Inventory: Inventory{
        Stock:    100,
        Reserved: 20,
    },
}

result, err := engine.Exec(ctx, "order_processing", input)
// result["processing_time"] = "2小时"
// result["shipping_cost"] = 0
// result["priority"] = "高"
```

## 📊 内置函数参考

Runehammer 提供了 50+ 内置函数，涵盖各种常用场景：

### 数学函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Abs(x)` | 绝对值 | `Abs(-10)` → `10` |
| `Max(a, b)` | 最大值 | `Max(5, 10)` → `10` |
| `Min(a, b)` | 最小值 | `Min(5, 10)` → `5` |
| `Round(x)` | 四舍五入 | `Round(3.7)` → `4` |
| `Floor(x)` | 向下取整 | `Floor(3.7)` → `3` |
| `Ceil(x)` | 向上取整 | `Ceil(3.2)` → `4` |
| `Pow(x, y)` | 幂运算 | `Pow(2, 3)` → `8` |
| `Sqrt(x)` | 平方根 | `Sqrt(16)` → `4` |
| `Sin(x)` | 正弦 | `Sin(0)` → `0` |
| `Cos(x)` | 余弦 | `Cos(0)` → `1` |
| `Tan(x)` | 正切 | `Tan(0)` → `0` |
| `Log(x)` | 自然对数 | `Log(2.718)` → `1` |
| `Log10(x)` | 以10为底的对数 | `Log10(100)` → `2` |

### 统计函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Sum(values)` | 求和 | `Sum([1,2,3,4])` → `10` |
| `Avg(values)` | 平均值 | `Avg([1,2,3,4])` → `2.5` |
| `MaxSlice(values)` | 数组最大值 | `MaxSlice([1,5,3])` → `5` |
| `MinSlice(values)` | 数组最小值 | `MinSlice([1,5,3])` → `1` |

### 字符串函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Contains(s, substr)` | 包含检查 | `Contains("hello", "ell")` → `true` |
| `HasPrefix(s, prefix)` | 前缀检查 | `HasPrefix("hello", "he")` → `true` |
| `HasSuffix(s, suffix)` | 后缀检查 | `HasSuffix("hello", "lo")` → `true` |
| `Len(s)` | 字符串长度 | `Len("hello")` → `5` |
| `ToUpper(s)` | 转大写 | `ToUpper("hello")` → `"HELLO"` |
| `ToLower(s)` | 转小写 | `ToLower("HELLO")` → `"hello"` |
| `Split(s, sep)` | 字符串分割 | `Split("a,b,c", ",")` → `["a","b","c"]` |
| `Join(elems, sep)` | 字符串连接 | `Join(["a","b"], ",")` → `"a,b"` |
| `Replace(s, old, new, n)` | 字符串替换 | `Replace("hello", "l", "L", 1)` → `"heLlo"` |
| `TrimSpace(s)` | 去除空白 | `TrimSpace(" hello ")` → `"hello"` |

### 时间函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Now()` | 当前时间 | `Now()` |
| `Today()` | 今天开始时间 | `Today()` |
| `NowMillis()` | 当前毫秒时间戳 | `NowMillis()` |
| `TimeToMillis(t)` | 时间转毫秒时间戳 | `TimeToMillis(Now())` |
| `MillisToTime(millis)` | 毫秒时间戳转时间 | `MillisToTime(1699123200000)` |
| `FormatTime(t, layout)` | 格式化时间 | `FormatTime(Now(), "2006-01-02")` |
| `ParseTime(layout, value)` | 解析时间 | `ParseTime("2006-01-02", "2023-12-01")` |
| `AddDays(t, days)` | 加减天数 | `AddDays(Today(), 7)` |
| `AddHours(t, hours)` | 加减小时 | `AddHours(Now(), -2)` |

### 验证函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `Matches(s, pattern)` | 正则匹配 | `Matches("abc123", "\\d+")` → `true` |
| `IsEmail(email)` | 邮箱验证 | `IsEmail("test@example.com")` → `true` |
| `IsPhoneNumber(phone)` | 手机号验证 | `IsPhoneNumber("13800138000")` → `true` |
| `IsIDCard(id)` | 身份证验证 | `IsIDCard("110101199001011234")` → `true` |
| `Between(value, min, max)` | 范围检查 | `Between(5, 1, 10)` → `true` |
| `LengthBetween(s, min, max)` | 长度检查 | `LengthBetween("hello", 3, 10)` → `true` |

### 类型转换函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `ToString(v)` | 转字符串 | `ToString(123)` → `"123"` |
| `ToInt(s)` | 转整数 | `ToInt("123")` → `123` |
| `ToFloat(s)` | 转浮点数 | `ToFloat("3.14")` → `3.14` |
| `ToBool(s)` | 转布尔值 | `ToBool("true")` → `true` |

### 工具函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `IsEmpty(v)` | 空值检查 | `IsEmpty("")` → `true` |
| `IsNotEmpty(v)` | 非空检查 | `IsNotEmpty("hello")` → `true` |
| `IF(condition, trueVal, falseVal)` | 条件表达式 | `IF(age >= 18, "成年", "未成年")` |

### 集合函数

| 函数 | 说明 | 示例 |
|------|------|------|
| `ContainsSlice(slice, item)` | 数组包含 | `ContainsSlice([1,2,3], 2)` → `true` |
| `Count(slice)` | 数组长度 | `Count([1,2,3])` → `3` |
| `Unique(slice)` | 数组去重 | `Unique([1,2,2,3])` → `[1,2,3]` |

### 在规则中使用内置函数

```grl
rule MathExample "数学函数示例" salience 100 {
    when
        Abs(customerinput.Customer.Balance) > 1000 &&
        Between(customerinput.Customer.Age, 18, 65)
    then
        result["credit_score"] = Round(customerinput.Customer.Income * 0.001);
        result["risk_level"] = IF(customerinput.Customer.DebtRatio > 0.5, "高", "低");
}

rule StringExample "字符串函数示例" salience 90 {
    when
        Contains(customerinput.Customer.Email, "@") &&
        LengthBetween(customerinput.Customer.Name, 2, 50)
    then
        result["email_valid"] = IsEmail(customerinput.Customer.Email);
        result["name_upper"] = ToUpper(customerinput.Customer.Name);
}

rule TimeExample "时间函数示例" salience 80 {
    when
        customerinput.Customer.LastLogin != nil
    then
        result["days_inactive"] = (Now().Unix() - customerinput.Customer.LastLogin.Unix()) / 86400;
        result["is_active"] = result["days_inactive"] <= 30;
        result["current_millis"] = NowMillis();
        result["login_millis"] = TimeToMillis(customerinput.Customer.LastLogin);
}
```

## 📋 变量注入规则

Runehammer在执行规则时，会将输入数据按照以下规则注入到规则执行上下文中：

### 🔤 注入规则

| 输入数据类型 | 注入方式 | 规则中访问方式 | 示例 |
|-------------|----------|---------------|------|
| **Map类型（仅数据库引擎）** | 作为整体注入为`Params` | `Params.键名` | `Params.customer.age`、`Params.order.amount` |
| **结构体类型** | 使用类型名（小写）| `类型名.字段` | `user.name`、`product.price` |
| **匿名结构体** | 统一使用`Params` | `Params.字段` | `Params.value`、`Params.data` |
| **其他类型** | 统一使用`Params` | 直接访问`Params` | `Params > 100`、`Params == "test"` |

### 🔍 详细说明

#### 1. Map类型数据注入（仅数据库引擎支持）
```go
// 数据库引擎map类型示例（动态引擎不支持map类型）
type CustomerOrderInput struct {
    Customer CustomerInfo `json:"customer"`
    Order    OrderInfo    `json:"order"`
}

type CustomerInfo struct {
    Age int  `json:"age"`
    VIP bool `json:"vip"`
}

type OrderInfo struct {
    Amount int    `json:"amount"`
    Status string `json:"status"`
}

// 使用结构体作为输入
input := CustomerOrderInput{
    Customer: CustomerInfo{Age: 25, VIP: true},
    Order:    OrderInfo{Amount: 1500, Status: "paid"},
}

// 规则中访问结构体字段
rule CustomerVip "VIP客户判断" {
    when
        customerorderinput.Customer.Age >= 18 && customerorderinput.Customer.VIP == true && customerorderinput.Order.Amount > 1000
    then
        result["level"] = "VIP";
}
```

#### 2. 结构体类型数据注入
```go
// 定义结构体
type Customer struct {
    Age  int    `json:"age"`
    Name string `json:"name"`
    VIP  bool   `json:"vip"`
}

// 输入数据
customer := Customer{Age: 25, Name: "张三", VIP: true}

// 规则中使用类型名（小写）访问
rule CheckCustomer "检查客户" {
    when
        customer.Age >= 18 && customer.VIP == true
    then
        result["eligible"] = true;
        result["message"] = customer.Name + " 符合条件";
}
```

#### 3. 匿名结构体和其他类型
```go
// 匿名结构体
input := struct {
    Value int
    Flag  bool
}{Value: 100, Flag: true}

// 或者基本类型
input := 100

// 规则中使用Params访问
rule CheckValue "检查值" {
    when
        Params.Value > 50 && Params.Flag == true
        // 或对于基本类型: Params > 50
    then
        result["valid"] = true;
}
```

### ⚠️ 注意事项

1. **变量名大小写**：结构体类型名会转换为小写作为变量名
2. **统一访问方式**：除结构体类型外，所有其他类型都通过`Params`访问
3. **保留字段**：避免使用`result`作为输入字段名，这是输出结果的保留字段
4. **类型安全**：在规则中访问字段时，请确保字段存在，否则可能导致执行错误

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

#### 数据库引擎配置选项
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

#### 动态引擎配置选项
| 选项 | 说明 | 示例 |
|------|------|------|
| `WithDynamicConfig(config)` | 设置完整动态配置 | `WithDynamicConfig(dynamicConfig)` |
| `WithConverterConfig(config)` | 设置转换器配置 | `WithConverterConfig(converterConfig)` |
| `WithParserConfig(config)` | 设置解析器配置 | `WithParserConfig(parserConfig)` |
| `WithExecutionConfig(config)` | 设置执行配置 | `WithExecutionConfig(execConfig)` |
| `WithDefaultSyntax(syntax)` | 设置默认语法类型 | `WithDefaultSyntax(SyntaxTypeSQL)` |
| `WithSupportedSyntax(syntaxes...)` | 设置支持的语法类型 | `WithSupportedSyntax(SQL, JS)` |
| `WithCustomOperators(ops)` | 设置自定义操作符 | `WithCustomOperators(operators)` |
| `WithCustomFunctions(funcs)` | 设置自定义函数 | `WithCustomFunctions(functions)` |
| `WithMaxConcurrency(max)` | 设置最大并发数 | `WithMaxConcurrency(10)` |
| `WithExecutionTimeout(timeout)` | 设置执行超时时间 | `WithExecutionTimeout(30*time.Second)` |

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
// 单一缓存策略 - 启动时确定
engine, _ := runehammer.New[ResultType](
    // 选择Redis缓存
    runehammer.WithRedis("localhost:6379", "", 0),
    runehammer.WithCacheTTL(30*time.Minute),        // 30分钟过期
)

// 或选择内存缓存
engine, _ := runehammer.New[ResultType](
    runehammer.WithMaxCacheSize(1000),              // 最大1000条规则
    runehammer.WithCacheTTL(30*time.Minute),
)

// 或完全禁用缓存
engine, _ := runehammer.New[ResultType](
    runehammer.WithDisableCache(),
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

#### 数据库引擎
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

#### 动态引擎
动态引擎支持 50+ 内置函数，并且可以注册自定义函数：

```go
// 创建动态引擎
dynamicEngine := runehammer.NewDynamicEngine[CustomResult](
    runehammer.DynamicEngineConfig{
        EnableCache: true,
        CacheTTL:    5 * time.Minute,
    },
)

// 注册自定义函数
dynamicEngine.RegisterCustomFunction("CalculateScore", func(income, age, credit int) float64 {
    return float64(income)*0.001 + float64(age)*0.1 + float64(credit)*0.01
})

dynamicEngine.RegisterCustomFunctions(map[string]interface{}{
    "ValidateIDCard": func(id string) bool {
        // 身份证验证逻辑
        return len(id) == 18
    },
    "GetCityCode": func(address string) string {
        // 根据地址获取城市代码
        if strings.Contains(address, "北京") {
            return "010"
        }
        return "000"
    },
})

metricRule := runehammer.MetricRule{
    Name:        "comprehensive_score",
    Description: "综合评分计算",
    Formula:     "CalculateScore(customerscoeinput.Customer.Income, customerscoeinput.Customer.Age, customerscoeinput.Customer.Credit)",
    Conditions: []string{
        "ValidateIDCard(customerscoeinput.Customer.IDCard)",
        "customerscoeinput.Customer.Income > 0",
    },
}

// 输入数据结构
type CustomerScoreInput struct {
    Customer CustomerDetails `json:"customer"`
}

type CustomerDetails struct {
    Income  int    `json:"income"`
    Age     int    `json:"age"`
    Credit  int    `json:"credit"`
    IDCard  string `json:"id_card"`
    Address string `json:"address"`
}

input := CustomerScoreInput{
    Customer: CustomerDetails{
        Income:  80000,
        Age:     30,
        Credit:  750,
        IDCard:  "110101199001011234",
        Address: "北京市朝阳区",
    },
}

result, err := dynamicEngine.ExecuteRuleDefinition(ctx, rule, input)
// result.ComprehensiveScore = 88.5 (计算结果)
```

### 规则转换器

动态引擎内置规则转换器，支持多种格式互转：

```go
converter := runehammer.NewGRLConverter()

// 从 JSON 转换为结构体的示例
jsonRule := `{
    "when": "customerinput.Customer.Age >= 18 && orderinput.Order.Amount > 100",
    "then": {
        "result.eligible": "true",
        "result.discount": "0.1"
    }
}`

grl, err := converter.ConvertToGRL(jsonRule)
// 生成标准的 GRL 规则

// 从标准规则转换
standardRule := runehammer.StandardRule{
    ID:          "approval_rule",
    Name:        "审批规则",
    Description: "自动审批逻辑",
    Priority:    100,
    Conditions: runehammer.Condition{
        Type:     "simple",
        Left:     "applicationinput.Application.Score",
        Operator: ">=",
        Right:    700,
    },
    Actions: []runehammer.Action{
        {
            Type:   "assign",
            Target: "result.approved",
            Value:  true,
        },
    },
}

grl, err = converter.ConvertRule(standardRule, runehammer.Definitions{})
```

### 多语法支持示例

```go
parser := runehammer.NewExpressionParser()

// SQL 语法转换 - 使用结构体字段
parser.SetSyntax(runehammer.SyntaxTypeSQL)
condition, _ := parser.ParseCondition("userinput.User.Age >= 18 AND userinput.User.Income BETWEEN 30000 AND 100000")
// 输出: "userinput.User.Age >= 18 && userinput.User.Income >= 30000 && userinput.User.Income <= 100000"

// JavaScript 语法转换 - 使用结构体字段
parser.SetSyntax(runehammer.SyntaxTypeJavaScript)
condition, _ = parser.ParseCondition("orderinput.Orders.filter(o => o.amount > 100).length > 0")
// 输出: "Count(Filter(orderinput.Orders, \"amount > 100\")) > 0"
```

### 批量规则执行

```go
// 定义多个不同类型的规则
rules := []interface{}{
    // 简单规则
    runehammer.SimpleRule{
        When: "batchexampleinput.Order.Amount > 500",
        Then: map[string]string{
            "result.FreeShipping": "true",
        },
    },
    
    // 指标规则
    runehammer.MetricRule{
        Name:    "loyalty_score",
        Formula: "purchase_count * 10 + total_amount * 0.01",
        Variables: map[string]string{
            "purchase_count": "batchexampleinput.Customer.PurchaseCount",
            "total_amount":   "batchexampleinput.Customer.TotalAmount",
        },
    },
    
    // 标准规则
    runehammer.StandardRule{
        ID:   "vip_check",
        Name: "VIP检查",
        Conditions: runehammer.Condition{
            Type:     "simple",
            Left:     "batchexampleinput.Customer.VipLevel",
            Operator: ">=",
            Right:    3,
        },
        Actions: []runehammer.Action{
            {
                Type:   "assign",
                Target: "result.IsVip",
                Value:  true,
            },
        },
    },
}

// 输入数据结构
type BatchExampleInput struct {
    Customer BatchCustomer `json:"customer"`
    Order    BatchOrder    `json:"order"`
}

type BatchCustomer struct {
    PurchaseCount int     `json:"purchase_count"`
    TotalAmount   float64 `json:"total_amount"`
    VipLevel      int     `json:"vip_level"`
}

type BatchOrder struct {
    Amount float64 `json:"amount"`
}

input := BatchExampleInput{
    Customer: BatchCustomer{
        PurchaseCount: 50,
        TotalAmount:   25000.0,
        VipLevel:      4,
    },
    Order: BatchOrder{
        Amount: 600.0,
    },
}

// 批量执行所有规则
results, err := dynamicEngine.ExecuteBatch(ctx, rules, input)
if err != nil {
    log.Fatal(err)
}

// 处理每个规则的执行结果
for i, result := range results {
    fmt.Printf("规则 %d 执行结果: %+v\n", i, result)
}
// 结果:
// 规则 0 执行结果: {FreeShipping: true}
// 规则 1 执行结果: {LoyaltyScore: 750}
// 规则 2 执行结果: {IsVip: true}
```

## 💡 最佳实践

### 规则设计原则

#### 数据库存储规则
1. **单一职责** - 每个规则专注解决一个特定问题
2. **优先级管理** - 使用 `salience` 控制规则执行顺序
3. **明确退出** - 使用 `Retract()` 避免重复执行
4. **输入验证** - 在规则中检查必要的输入参数

```grl
rule ValidateInput "输入验证" salience 1000 {
    when
        userinput == nil || userinput.User.ID == nil
    then
        result["error"] = "用户信息不完整";
        result["valid"] = false;
        Retract("ValidateInput");
}
```

#### 动态规则
1. **选择合适的规则类型**
   - **SimpleRule**: 适用于简单的条件-结果映射
   - **MetricRule**: 适用于指标计算和数据分析
   - **StandardRule**: 适用于复杂的业务逻辑

2. **语法选择**
   - **SQL语法**: 适合数据库背景的开发人员
   - **JavaScript语法**: 适合前端开发人员，支持常用的JS表达式语法

3. **函数使用**
   - 优先使用内置函数，性能更好
   - 自定义函数按需注册，避免过度复杂化
   - 验证函数放在条件中，计算函数放在动作中

### 性能优化建议

#### 数据库引擎优化
1. **合理设置缓存时间** - 根据规则变更频率调整TTL
2. **规则分组** - 不同业务场景使用不同的 `bizCode`
3. **避免复杂计算** - 将重计算逻辑前置到输入准备阶段
4. **监控缓存命中率** - 定期检查缓存效果

#### 动态引擎优化
1. **启用缓存** - 对于重复执行的规则，启用内存缓存
2. **并行执行** - 对于独立的规则，使用批量并行执行
3. **合理设置并发数** - 根据系统资源设置 `MaxConcurrency`
4. **避免深层嵌套** - 复杂条件可以拆分为多个简单规则

```go
// 性能优化配置示例
dynamicEngine := runehammer.NewDynamicEngine[OptimizedResult](
    runehammer.DynamicEngineConfig{
        EnableCache:       true,              // 启用缓存
        CacheTTL:          10 * time.Minute,  // 合理的缓存时间
        MaxCacheSize:      500,               // 足够的缓存空间
        ParallelExecution: true,              // 启用并行执行
        DefaultTimeout:    30 * time.Second,  // 合理的超时时间
    },
)
```

### 引擎选择指南

| 使用场景 | 推荐引擎 | 理由 |
|----------|----------|------|
| 业务规则管理 | 数据库引擎 | 支持热更新、版本控制、持久化存储 |
| 指标计算 | 动态引擎 | 实时计算、无需存储、支持复杂公式 |
| 第三方集成 | 动态引擎 | 多格式支持、语法转换、标准化接口 |
| 临时规则 | 动态引擎 | 快速执行、无需管理、即用即弃 |
| 批量处理 | 动态引擎 | 并行执行、高性能、支持批量操作 |
| 配置化规则 | 数据库引擎 | 界面配置、规则管理、权限控制 |

### 混合使用策略

在实际项目中，可以结合两种引擎的优势：

```go
// 初始化两个引擎
dbEngine, _ := runehammer.New[BusinessResult](
    runehammer.WithDB(db),
    runehammer.WithRedis("localhost:6379", "", 0),
)

dynamicEngine := runehammer.NewDynamicEngine[MetricResult](
    runehammer.DynamicEngineConfig{
        EnableCache:       true,
        ParallelExecution: true,
    },
)

// 业务规则使用数据库引擎
businessResult, err := dbEngine.Exec(ctx, "user_level_check", input)

// 指标计算使用动态引擎
metricRule := runehammer.MetricRule{
    Name:    "risk_score",
    Formula: "income_score * 0.4 + credit_score * 0.6",
        Variables: map[string]string{
            "income_score": "mixedinput.Customer.Income / 10000",
            "credit_score": "mixedinput.Customer.Credit / 10",
        },
}

metricResult, err := dynamicEngine.ExecuteRuleDefinition(ctx, metricRule, input)
```

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