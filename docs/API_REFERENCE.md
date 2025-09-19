# 📋 API 参考文档

本文档提供 Runehammer 规则引擎的完整 API 参考，包括接口定义、配置选项、内置函数等。

## 📚 目录

- [核心接口](#核心接口)
- [配置选项](#配置选项)
- [内置函数参考](#内置函数参考)
- [错误处理](#错误处理)
- [类型定义](#类型定义)

## 🔌 核心接口

### Engine 接口

```go
type Engine[T any] interface {
    // 执行规则
    Exec(ctx context.Context, bizCode string, input any) (T, error)
    
    // 关闭引擎，释放资源
    Close() error
}
```

### BaseEngine 接口

```go
type BaseEngine interface {
    // 执行规则，返回通用map类型
    ExecRaw(ctx context.Context, bizCode string, input any) (map[string]interface{}, error)
    
    // 关闭引擎，释放资源
    Close() error
}
```

### DynamicEngine 接口

```go
type DynamicEngine[T any] interface {
    // 执行规则定义
    ExecuteRuleDefinition(ctx context.Context, rule interface{}, input any) (T, error)
    
    // 批量执行规则
    ExecuteBatch(ctx context.Context, rules []interface{}, input any) ([]T, error)
    
    // 注册自定义函数
    RegisterCustomFunction(name string, fn interface{})
    
    // 批量注册自定义函数
    RegisterCustomFunctions(functions map[string]interface{})
    
    // 获取缓存统计
    GetCacheStats() CacheStats
    
    // 清理缓存
    ClearCache()
}
```

## ⚙️ 配置选项

### 数据库引擎配置选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `WithDSN(dsn)` | 设置数据库连接字符串 | `WithDSN("mysql://user:pass@localhost/db")` |
| `WithCustomDB(db)` | 使用现有GORM数据库连接 | `WithCustomDB(gormDB)` |
| `WithAutoMigrate()` | 自动创建数据库表 | `WithAutoMigrate()` |

### 缓存配置选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `WithRedisCache(addr, pass, db)` | 配置Redis缓存 | `WithRedisCache("localhost:6379", "", 0)` |
| `WithMemoryCache(size)` | 配置内存缓存 | `WithMemoryCache(1000)` |
| `WithNoCache()` | 禁用缓存 | `WithNoCache()` |
| `WithCacheTTL(ttl)` | 设置缓存过期时间 | `WithCacheTTL(10*time.Minute)` |
| `WithMaxCacheSize(size)` | 设置最大缓存大小 | `WithMaxCacheSize(1000)` |

### 其他配置选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `WithCustomLogger(logger)` | 设置自定义日志器 | `WithCustomLogger(myLogger)` |
| `WithSyncInterval(interval)` | 设置同步间隔 | `WithSyncInterval(5*time.Minute)` |
| `WithCustomCache(cache)` | 使用自定义缓存实现 | `WithCustomCache(myCache)` |
| `WithCustomRuleMapper(mapper)` | 设置自定义规则映射器 | `WithCustomRuleMapper(myMapper)` |

### 动态引擎配置

```go
type DynamicEngineConfig struct {
    EnableCache       bool          // 是否启用缓存
    CacheTTL          time.Duration // 缓存过期时间
    MaxCacheSize      int           // 最大缓存大小
    StrictValidation  bool          // 是否严格验证
    ParallelExecution bool          // 是否支持并行执行批量规则
    DefaultTimeout    time.Duration // 默认超时时间
}
```

## 📊 内置函数参考

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

## 🎯 规则定义类型

### SimpleRule 简单规则

```go
type SimpleRule struct {
    When string            `json:"when"` // 条件表达式
    Then map[string]string `json:"then"` // 结果赋值
}
```

### MetricRule 指标规则

```go
type MetricRule struct {
    Name        string            `json:"name"`        // 指标名称
    Description string            `json:"description"` // 描述
    Formula     string            `json:"formula"`     // 计算公式
    Variables   map[string]string `json:"variables"`   // 变量定义
    Conditions  []string          `json:"conditions"`  // 前置条件
}
```

### StandardRule 标准规则

```go
type StandardRule struct {
    ID          string      `json:"id"`          // 规则ID
    Name        string      `json:"name"`        // 规则名称
    Description string      `json:"description"` // 规则描述
    Priority    int         `json:"priority"`    // 优先级
    Enabled     bool        `json:"enabled"`     // 是否启用
    Tags        []string    `json:"tags"`        // 标签
    Conditions  Condition   `json:"conditions"`  // 条件
    Actions     []Action    `json:"actions"`     // 动作
}
```

### Condition 条件定义

```go
type Condition struct {
    Type     ConditionType `json:"type"`     // 条件类型
    Left     string        `json:"left"`     // 左操作数
    Operator Operator      `json:"operator"` // 操作符
    Right    interface{}   `json:"right"`    // 右操作数
    Children []Condition   `json:"children"` // 子条件
}
```

### Action 动作定义

```go
type Action struct {
    Type   ActionType  `json:"type"`   // 动作类型
    Target string      `json:"target"` // 目标字段
    Value  interface{} `json:"value"`  // 值
}
```

## 🔤 枚举类型

### ConditionType 条件类型

```go
const (
    ConditionTypeSimple     ConditionType = "simple"     // 简单条件
    ConditionTypeComposite  ConditionType = "composite"  // 复合条件
    ConditionTypeExpression ConditionType = "expression" // 表达式条件
    ConditionTypeFunction   ConditionType = "function"   // 函数条件
    ConditionTypeAnd        ConditionType = "and"        // 逻辑与
    ConditionTypeOr         ConditionType = "or"         // 逻辑或
    ConditionTypeNot        ConditionType = "not"        // 逻辑非
)
```

### Operator 操作符

```go
const (
    // 比较操作符
    OpEqual              Operator = "=="
    OpNotEqual           Operator = "!="
    OpGreaterThan        Operator = ">"
    OpLessThan           Operator = "<"
    OpGreaterThanOrEqual Operator = ">="
    OpLessThanOrEqual    Operator = "<="
    
    // 逻辑操作符
    OpAnd                Operator = "and"
    OpOr                 Operator = "or"
    OpNot                Operator = "not"
    
    // 集合操作符
    OpIn                 Operator = "in"
    OpNotIn              Operator = "notIn"
    OpContains           Operator = "contains"
    OpMatches            Operator = "matches"
    OpBetween            Operator = "between"
)
```

### ActionType 动作类型

```go
const (
    ActionTypeAssign     ActionType = "assign"    // 赋值
    ActionTypeCalculate  ActionType = "calculate" // 计算
    ActionTypeInvoke     ActionType = "invoke"    // 调用函数
    ActionTypeAlert      ActionType = "alert"     // 告警
    ActionTypeLog        ActionType = "log"       // 记录日志
    ActionTypeStop       ActionType = "stop"      // 停止执行
)
```

## ⚠️ 错误处理

### 错误类型

```go
var (
    ErrNoRulesFound     = errors.New("no rules found")
    ErrCompileFailed    = errors.New("rule compile failed")
    ErrExecutionFailed  = errors.New("rule execution failed")
    ErrConfigInvalid    = errors.New("invalid configuration")
    ErrCacheTimeout     = errors.New("cache operation timeout")
)
```

### 错误处理示例

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
    case strings.Contains(err.Error(), "编译失败"):
        // 处理规则语法错误
        log.Printf("规则语法错误: %v", err)
    case strings.Contains(err.Error(), "执行失败"):
        // 处理规则执行错误
        log.Printf("规则执行错误: %v", err)
    default:
        // 其他错误
        log.Printf("未知错误: %v", err)
    }
}
```

## 📊 缓存统计

### CacheStats 结构

```go
type CacheStats struct {
    HitCount  int64   `json:"hit_count"`  // 缓存命中次数
    MissCount int64   `json:"miss_count"` // 缓存未命中次数
    HitRate   float64 `json:"hit_rate"`   // 缓存命中率
    Size      int64   `json:"size"`       // 当前缓存大小
}
```

### 获取缓存统计

```go
// 动态引擎
stats := dynamicEngine.GetCacheStats()
fmt.Printf("缓存命中率: %.2f%%", stats.HitRate*100)

// 清理缓存
dynamicEngine.ClearCache()
```

## 🔧 日志接口

### Logger 接口

```go
type Logger interface {
    Debugf(ctx context.Context, msg string, keyvals ...any)
    Infof(ctx context.Context, msg string, keyvals ...any)
    Warnf(ctx context.Context, msg string, keyvals ...any)
    Errorf(ctx context.Context, msg string, keyvals ...any)
}
```

### 自定义日志实现

```go
type MyLogger struct {
    logger *zap.Logger
}

func (l *MyLogger) Debugf(ctx context.Context, msg string, keyvals ...any) {
    l.logger.Debug(msg, zap.Any("data", keyvals))
}

func (l *MyLogger) Infof(ctx context.Context, msg string, keyvals ...any) {
    l.logger.Info(msg, zap.Any("data", keyvals))
}

func (l *MyLogger) Warnf(ctx context.Context, msg string, keyvals ...any) {
    l.logger.Warn(msg, zap.Any("data", keyvals))
}

func (l *MyLogger) Errorf(ctx context.Context, msg string, keyvals ...any) {
    l.logger.Error(msg, zap.Any("data", keyvals))
}
```

## 📋 变量访问规范

### 输入变量访问

| 输入数据类型 | 访问方式 | 示例 |
|-------------|----------|------|
| 结构体 | `Params.字段名` | `Params.User.Age >= 18` |
| 匿名结构体 | `Params.字段名` | `Params.Value`、`Params.Data` |
| 基础类型 | `Params` | `Params > 100`、`Params == "test"` |
| Map | `Params["key"]` | `Params["customer"]` |

### 输出变量访问

- **默认字段名**: `Result`（大写R开头）
- **访问方式**: `Result.字段名`（字段名使用大驼峰形式）
- **示例**: `Result.IsValid = true`, `Result.TotalScore = 85`

## 🎯 最佳实践

### 命名规范

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
```

### 性能优化

```go
// 启用合适的缓存策略
engine, _ := runehammer.New[ResultType](
    runehammer.WithRedisCache("localhost:6379", "", 0),
    runehammer.WithCacheTTL(30*time.Minute),
    runehammer.WithMaxCacheSize(1000),
)

// 动态引擎并发优化
dynamicEngine := engine.NewDynamicEngine[map[string]interface{}](
    engine.DynamicEngineConfig{
        EnableCache:       true,
        ParallelExecution: true,
        DefaultTimeout:    30 * time.Second,
    },
)
```

这份API参考文档提供了Runehammer规则引擎的完整接口和使用说明，可作为开发和集成的参考指南。