package engine

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
)

// ============================================================================
// 内置函数管理 - 为规则提供常用的内置函数
// ============================================================================

// injectBuiltinFunctions 注入内置函数 - 为规则引擎提供常用的内置函数
//
// 内置函数分类:
//   - 时间函数：当前时间、格式化等
//   - 字符串函数：包含、前缀、后缀等
//   - 数学函数：最大值、最小值等
//   - 工具函数：长度、空值检查等
//
// 参数:
//   dataCtx - Grule数据上下文
func (e *engineImpl[T]) injectBuiltinFunctions(dataCtx ast.IDataContext) {
	// 注入时间相关函数
	e.injectTimeFunctions(dataCtx)
	
	// 注入字符串相关函数
	e.injectStringFunctions(dataCtx)
	
	// 注入数学相关函数
	e.injectMathFunctions(dataCtx)
	
	// 注入工具函数
	e.injectUtilFunctions(dataCtx)
	
	// 注入集合函数
	e.injectCollectionFunctions(dataCtx)
	
	// 注入验证函数
	e.injectValidationFunctions(dataCtx)
}

// injectTimeFunctions 注入时间函数
func (e *engineImpl[T]) injectTimeFunctions(dataCtx ast.IDataContext) {
	// 获取当前时间
	dataCtx.Add("Now", func() time.Time {
		return time.Now()
	})
	
	// 获取今天的开始时间（00:00:00）
	dataCtx.Add("Today", func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	})
	
	// 格式化时间
	dataCtx.Add("FormatTime", func(t time.Time, layout string) string {
		return t.Format(layout)
	})
	
	// 解析时间字符串
	dataCtx.Add("ParseTime", func(layout, value string) (time.Time, error) {
		return time.Parse(layout, value)
	})
	
	// 时间加减
	dataCtx.Add("AddDays", func(t time.Time, days int) time.Time {
		return t.AddDate(0, 0, days)
	})
	
	dataCtx.Add("AddHours", func(t time.Time, hours int) time.Time {
		return t.Add(time.Duration(hours) * time.Hour)
	})
	
	// 毫秒时间戳相关函数
	dataCtx.Add("NowMillis", func() int64 {
		return time.Now().UnixMilli()
	})
	
	dataCtx.Add("TimeToMillis", func(t time.Time) int64 {
		return t.UnixMilli()
	})
	
	dataCtx.Add("MillisToTime", func(millis int64) time.Time {
		return time.UnixMilli(millis)
	})
}

// injectStringFunctions 注入字符串函数
func (e *engineImpl[T]) injectStringFunctions(dataCtx ast.IDataContext) {
	// 字符串包含检查
	dataCtx.Add("Contains", func(s, substr string) bool {
		return strings.Contains(s, substr)
	})
	
	// 前缀检查
	dataCtx.Add("HasPrefix", func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	})
	
	// 后缀检查  
	dataCtx.Add("HasSuffix", func(s, suffix string) bool {
		return strings.HasSuffix(s, suffix)
	})
	
	// 字符串长度
	dataCtx.Add("Len", func(s string) int {
		return len(s)
	})
	
	// 字符串转大写
	dataCtx.Add("ToUpper", func(s string) string {
		return strings.ToUpper(s)
	})
	
	// 字符串转小写
	dataCtx.Add("ToLower", func(s string) string {
		return strings.ToLower(s)
	})
	
	// 字符串分割
	dataCtx.Add("Split", func(s, sep string) []string {
		return strings.Split(s, sep)
	})
	
	// 字符串连接
	dataCtx.Add("Join", func(elems []string, sep string) string {
		return strings.Join(elems, sep)
	})
	
	// 字符串替换
	dataCtx.Add("Replace", func(s, old, new string, n int) string {
		return strings.Replace(s, old, new, n)
	})
	
	// 去除空白字符
	dataCtx.Add("TrimSpace", func(s string) string {
		return strings.TrimSpace(s)
	})
}

// injectMathFunctions 注入数学函数
func (e *engineImpl[T]) injectMathFunctions(dataCtx ast.IDataContext) {
	// 基础数学函数
	dataCtx.Add("Abs", func(x float64) float64 {
		return math.Abs(x)
	})
	
	dataCtx.Add("Max", func(x, y float64) float64 {
		return math.Max(x, y)
	})
	
	dataCtx.Add("Min", func(x, y float64) float64 {
		return math.Min(x, y)
	})
	
	dataCtx.Add("Round", func(x float64) float64 {
		return math.Round(x)
	})
	
	dataCtx.Add("Floor", func(x float64) float64 {
		return math.Floor(x)
	})
	
	dataCtx.Add("Ceil", func(x float64) float64 {
		return math.Ceil(x)
	})
	
	dataCtx.Add("Pow", func(x, y float64) float64 {
		return math.Pow(x, y)
	})
	
	dataCtx.Add("Sqrt", func(x float64) float64 {
		return math.Sqrt(x)
	})
	
	// 三角函数
	dataCtx.Add("Sin", func(x float64) float64 {
		return math.Sin(x)
	})
	
	dataCtx.Add("Cos", func(x float64) float64 {
		return math.Cos(x)
	})
	
	dataCtx.Add("Tan", func(x float64) float64 {
		return math.Tan(x)
	})
	
	// 对数函数
	dataCtx.Add("Log", func(x float64) float64 {
		return math.Log(x)
	})
	
	dataCtx.Add("Log10", func(x float64) float64 {
		return math.Log10(x)
	})
	
	// 统计函数 - 支持切片
	dataCtx.Add("Sum", func(values []float64) float64 {
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
	})
	
	dataCtx.Add("Avg", func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))
	})
	
	dataCtx.Add("MaxSlice", func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		max := values[0]
		for _, v := range values {
			if v > max {
				max = v
			}
		}
		return max
	})
	
	dataCtx.Add("MinSlice", func(values []float64) float64 {
		if len(values) == 0 {
			return 0
		}
		min := values[0]
		for _, v := range values {
			if v < min {
				min = v
			}
		}
		return min
	})
}

// injectUtilFunctions 注入工具函数
func (e *engineImpl[T]) injectUtilFunctions(dataCtx ast.IDataContext) {
	// 类型转换函数
	dataCtx.Add("ToString", func(v interface{}) string {
		switch val := v.(type) {
		case string:
			return val
		case int:
			return strconv.Itoa(val)
		case int64:
			return strconv.FormatInt(val, 10)
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(val)
		default:
			return ""
		}
	})
	
	dataCtx.Add("ToInt", func(s string) (int, error) {
		return strconv.Atoi(s)
	})
	
	dataCtx.Add("ToFloat", func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
	
	dataCtx.Add("ToBool", func(s string) (bool, error) {
		return strconv.ParseBool(s)
	})
	
	// 空值检查
	dataCtx.Add("IsEmpty", func(v interface{}) bool {
		if v == nil {
			return true
		}
		switch val := v.(type) {
		case string:
			return val == ""
		case []interface{}:
			return len(val) == 0
		default:
			return false
		}
	})
	
	dataCtx.Add("IsNotEmpty", func(v interface{}) bool {
		return !e.isEmpty(v)
	})
	
	// 条件函数
	dataCtx.Add("IF", func(condition bool, trueValue, falseValue interface{}) interface{} {
		if condition {
			return trueValue
		}
		return falseValue
	})
}

// injectCollectionFunctions 注入集合函数
func (e *engineImpl[T]) injectCollectionFunctions(dataCtx ast.IDataContext) {
	// 数组包含检查
	dataCtx.Add("ContainsSlice", func(slice []interface{}, item interface{}) bool {
		for _, v := range slice {
			if v == item {
				return true
			}
		}
		return false
	})
	
	// 数组长度
	dataCtx.Add("Count", func(slice []interface{}) int {
		return len(slice)
	})
	
	// 数组过滤（简化版）
	dataCtx.Add("Filter", func(slice []interface{}, predicate string) []interface{} {
		// 这里是简化实现，实际可以更复杂
		var result []interface{}
		// TODO: 实现复杂的过滤逻辑
		return result
	})
	
	// 数组映射
	dataCtx.Add("Map", func(slice []interface{}, mapper string) []interface{} {
		// 这里是简化实现，实际可以更复杂
		var result []interface{}
		// TODO: 实现复杂的映射逻辑
		return result
	})
	
	// 数组去重
	dataCtx.Add("Unique", func(slice []interface{}) []interface{} {
		seen := make(map[interface{}]bool)
		var result []interface{}
		for _, item := range slice {
			if !seen[item] {
				seen[item] = true
				result = append(result, item)
			}
		}
		return result
	})
}

// injectValidationFunctions 注入验证函数
func (e *engineImpl[T]) injectValidationFunctions(dataCtx ast.IDataContext) {
	// 正则表达式匹配
	dataCtx.Add("Matches", func(s, pattern string) bool {
		matched, err := regexp.MatchString(pattern, s)
		if err != nil {
			return false
		}
		return matched
	})
	
	// 邮箱验证
	dataCtx.Add("IsEmail", func(email string) bool {
		emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
		matched, _ := regexp.MatchString(emailRegex, email)
		return matched
	})
	
	// 手机号验证（中国）
	dataCtx.Add("IsPhoneNumber", func(phone string) bool {
		phoneRegex := `^1[3-9]\d{9}$`
		matched, _ := regexp.MatchString(phoneRegex, phone)
		return matched
	})
	
	// 身份证号验证（简化）
	dataCtx.Add("IsIDCard", func(id string) bool {
		idRegex := `^\d{17}[\dXx]$`
		matched, _ := regexp.MatchString(idRegex, id)
		return matched
	})
	
	// 数值范围检查
	dataCtx.Add("Between", func(value, min, max float64) bool {
		return value >= min && value <= max
	})
	
	// 字符串长度检查
	dataCtx.Add("LengthBetween", func(s string, min, max int) bool {
		length := len(s)
		return length >= min && length <= max
	})
}

// 辅助方法
func (e *engineImpl[T]) isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []interface{}:
		return len(val) == 0
	default:
		return false
	}
}
