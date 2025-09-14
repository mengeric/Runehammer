package runehammer

import (
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
	
	// 可以扩展更多内置函数类别
	// e.injectMathFunctions(dataCtx)
	// e.injectUtilFunctions(dataCtx)
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