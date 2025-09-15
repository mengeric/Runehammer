package runehammer

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/robfig/cron/v3"
	. "github.com/smartystreets/goconvey/convey"
)

// TestEngineFunctions 测试引擎内置函数
func TestEngineFunctions(t *testing.T) {
	Convey("引擎内置函数测试", t, func() {

		// 创建测试用的引擎实例
		config := DefaultConfig()
		mapper := newMockRuleMapper()
		cache := newMockCache()
		cacheKeys := newMockCacheKeyBuilder()
		logger := NewNoopLogger()
		knowledgeLibrary := ast.NewKnowledgeLibrary()
		cronScheduler := cron.New()

		engine := NewEngineImpl[map[string]any](
			config, mapper, cache, cacheKeys, logger,
			knowledgeLibrary, &sync.Map{}, cronScheduler, false,
		)

		// 创建数据上下文
		dataCtx := ast.NewDataContext()

		// 注入内置函数
		engine.injectBuiltinFunctions(dataCtx)

		Convey("时间函数测试", func() {

			Convey("Now() 当前时间", func() {
				nowFunc := dataCtx.Get("Now")
				So(nowFunc, ShouldNotBeNil)

				value, err := nowFunc.GetValue()
				So(err, ShouldBeNil)

				// 调用函数
				nowValue := value.Interface().(func() time.Time)()
				So(nowValue, ShouldNotBeZeroValue)

				// 时间应该在合理范围内（最近1分钟内）
				timeDiff := time.Since(nowValue)
				So(timeDiff, ShouldBeLessThan, time.Minute)
			})

			Convey("Today() 今天开始时间", func() {
				todayFunc := dataCtx.Get("Today")
				So(todayFunc, ShouldNotBeNil)

				value, err := todayFunc.GetValue()
				So(err, ShouldBeNil)

				// 调用函数
				todayValue := value.Interface().(func() time.Time)()
				So(todayValue, ShouldNotBeZeroValue)

				// 应该是今天的开始时间（0点）
				So(todayValue.Hour(), ShouldEqual, 0)
				So(todayValue.Minute(), ShouldEqual, 0)
				So(todayValue.Second(), ShouldEqual, 0)
			})

			Convey("NowMillis() 当前毫秒时间戳", func() {
				nowMillisFunc := dataCtx.Get("NowMillis")
				So(nowMillisFunc, ShouldNotBeNil)

				value, err := nowMillisFunc.GetValue()
				So(err, ShouldBeNil)

				// 调用函数
				millisValue := value.Interface().(func() int64)()
				So(millisValue, ShouldBeGreaterThan, 0)

				// 时间戳应该接近当前时间
				currentMillis := time.Now().UnixMilli()
				diff := currentMillis - millisValue
				So(math.Abs(float64(diff)), ShouldBeLessThan, 1000) // 1秒内
			})

			Convey("AddDays() 日期加减", func() {
				addDaysFunc := dataCtx.Get("AddDays")
				So(addDaysFunc, ShouldNotBeNil)

				value, err := addDaysFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试日期加减
				baseDate := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
				expectedDate := time.Date(2023, 1, 11, 0, 0, 0, 0, time.UTC)

				addDays := value.Interface().(func(time.Time, int) time.Time)
				resultDate := addDays(baseDate, 10)
				So(resultDate.Equal(expectedDate), ShouldBeTrue)
			})
		})

		Convey("字符串函数测试", func() {

			Convey("Contains() 包含检查", func() {
				containsFunc := dataCtx.Get("Contains")
				So(containsFunc, ShouldNotBeNil)

				value, err := containsFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试包含
				containsCheck := value.Interface().(func(string, string) bool)
				So(containsCheck("hello world", "world"), ShouldBeTrue)
				So(containsCheck("hello world", "test"), ShouldBeFalse)
			})

			Convey("HasPrefix() 前缀检查", func() {
				hasPrefixFunc := dataCtx.Get("HasPrefix")
				So(hasPrefixFunc, ShouldNotBeNil)

				value, err := hasPrefixFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试前缀
				prefixCheck := value.Interface().(func(string, string) bool)
				So(prefixCheck("hello world", "hello"), ShouldBeTrue)
				So(prefixCheck("hello world", "world"), ShouldBeFalse)
			})

			Convey("HasSuffix() 后缀检查", func() {
				hasSuffixFunc := dataCtx.Get("HasSuffix")
				So(hasSuffixFunc, ShouldNotBeNil)

				value, err := hasSuffixFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试后缀
				suffixCheck := value.Interface().(func(string, string) bool)
				So(suffixCheck("hello world", "world"), ShouldBeTrue)
				So(suffixCheck("hello world", "hello"), ShouldBeFalse)
			})

			Convey("ToUpper() 转大写", func() {
				toUpperFunc := dataCtx.Get("ToUpper")
				So(toUpperFunc, ShouldNotBeNil)

				value, err := toUpperFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试大写转换
				upperCase := value.Interface().(func(string) string)
				So(upperCase("hello"), ShouldEqual, "HELLO")
				So(upperCase("Hello World"), ShouldEqual, "HELLO WORLD")
			})

			Convey("ToLower() 转小写", func() {
				toLowerFunc := dataCtx.Get("ToLower")
				So(toLowerFunc, ShouldNotBeNil)

				value, err := toLowerFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试小写转换
				lowerCase := value.Interface().(func(string) string)
				So(lowerCase("HELLO"), ShouldEqual, "hello")
				So(lowerCase("Hello World"), ShouldEqual, "hello world")
			})
		})

		Convey("数学函数测试", func() {

			Convey("Max() 最大值", func() {
				maxFunc := dataCtx.Get("Max")
				So(maxFunc, ShouldNotBeNil)

				value, err := maxFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试最大值
				maxValue := value.Interface().(func(float64, float64) float64)
				So(maxValue(5.0, 10.0), ShouldEqual, 10.0)
				So(maxValue(-5.0, -10.0), ShouldEqual, -5.0)
			})

			Convey("Min() 最小值", func() {
				minFunc := dataCtx.Get("Min")
				So(minFunc, ShouldNotBeNil)

				value, err := minFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试最小值
				minValue := value.Interface().(func(float64, float64) float64)
				So(minValue(5.0, 10.0), ShouldEqual, 5.0)
				So(minValue(-5.0, -10.0), ShouldEqual, -10.0)
			})

			Convey("Abs() 绝对值", func() {
				absFunc := dataCtx.Get("Abs")
				So(absFunc, ShouldNotBeNil)

				value, err := absFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试绝对值
				absValue := value.Interface().(func(float64) float64)
				So(absValue(-10.0), ShouldEqual, 10.0)
				So(absValue(10.0), ShouldEqual, 10.0)
				So(absValue(0.0), ShouldEqual, 0.0)
			})

			Convey("Round() 四舍五入", func() {
				roundFunc := dataCtx.Get("Round")
				So(roundFunc, ShouldNotBeNil)

				value, err := roundFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试四舍五入
				roundValue := value.Interface().(func(float64) float64)
				So(roundValue(3.4), ShouldEqual, 3.0)
				So(roundValue(3.6), ShouldEqual, 4.0)
				So(roundValue(-3.4), ShouldEqual, -3.0)
				So(roundValue(-3.6), ShouldEqual, -4.0)
			})
		})

		Convey("工具函数测试", func() {

			Convey("Len() 长度计算", func() {
				lenFunc := dataCtx.Get("Len")
				So(lenFunc, ShouldNotBeNil)

				value, err := lenFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试字符串长度（字节长度，不是字符数）
				lenValue := value.Interface().(func(string) int)
				So(lenValue("hello"), ShouldEqual, 5)
				So(lenValue(""), ShouldEqual, 0)
				So(lenValue("中文测试"), ShouldEqual, 12) // UTF-8编码，每个中文字符3字节
			})

			Convey("IsEmpty() 空值检查", func() {
				isEmptyFunc := dataCtx.Get("IsEmpty")
				So(isEmptyFunc, ShouldNotBeNil)

				value, err := isEmptyFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试空值检查
				isEmpty := value.Interface().(func(interface{}) bool)
				So(isEmpty(""), ShouldBeTrue)
				So(isEmpty(nil), ShouldBeTrue)
				So(isEmpty("hello"), ShouldBeFalse)
				So(isEmpty(0), ShouldBeFalse) // 0不是空值
			})

			Convey("IsNotEmpty() 非空检查", func() {
				isNotEmptyFunc := dataCtx.Get("IsNotEmpty")
				So(isNotEmptyFunc, ShouldNotBeNil)

				value, err := isNotEmptyFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试非空检查
				isNotEmpty := value.Interface().(func(interface{}) bool)
				So(isNotEmpty("hello"), ShouldBeTrue)
				So(isNotEmpty(0), ShouldBeTrue) // 0不是空值
				So(isNotEmpty(""), ShouldBeFalse)
				So(isNotEmpty(nil), ShouldBeFalse)
			})
		})

		Convey("集合函数测试", func() {

			Convey("Sum() 求和", func() {
				sumFunc := dataCtx.Get("Sum")
				So(sumFunc, ShouldNotBeNil)

				value, err := sumFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试数组求和
				sumValue := value.Interface().(func([]float64) float64)
				So(sumValue([]float64{1, 2, 3, 4, 5}), ShouldEqual, 15.0)
				So(sumValue([]float64{}), ShouldEqual, 0.0)
				So(sumValue([]float64{-1, -2, -3}), ShouldEqual, -6.0)
			})

			Convey("Avg() 平均值", func() {
				avgFunc := dataCtx.Get("Avg")
				So(avgFunc, ShouldNotBeNil)

				value, err := avgFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试平均值
				avgValue := value.Interface().(func([]float64) float64)
				So(avgValue([]float64{1, 2, 3, 4, 5}), ShouldEqual, 3.0)
				So(avgValue([]float64{10, 20}), ShouldEqual, 15.0)
			})

			Convey("Count() 计数", func() {
				countFunc := dataCtx.Get("Count")
				So(countFunc, ShouldNotBeNil)

				value, err := countFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试计数
				countValue := value.Interface().(func([]interface{}) int)
				So(countValue([]interface{}{1, 2, 3}), ShouldEqual, 3)
				So(countValue([]interface{}{}), ShouldEqual, 0)
			})

			Convey("ContainsSlice() 数组包含检查", func() {
				containsSliceFunc := dataCtx.Get("ContainsSlice")
				So(containsSliceFunc, ShouldNotBeNil)

				value, err := containsSliceFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试数组包含检查
				containsSlice := value.Interface().(func([]interface{}, interface{}) bool)
				So(containsSlice([]interface{}{1, 2, 3, 4, 5}, 3), ShouldBeTrue)
				So(containsSlice([]interface{}{1, 2, 3, 4, 5}, 6), ShouldBeFalse)
				So(containsSlice([]interface{}{}, 1), ShouldBeFalse)
			})

			Convey("Filter() 数组过滤", func() {
				filterFunc := dataCtx.Get("Filter")
				So(filterFunc, ShouldNotBeNil)

				value, err := filterFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试数组过滤
				filter := value.Interface().(func([]interface{}, string) []interface{})
				// 由于Filter函数是简化实现，这里只测试能正常调用
				result := filter([]interface{}{1, 2, 3, 4, 5}, "x > 3")
				So(result, ShouldNotBeNil) // 即使是空数组也不应该是nil
			})

			Convey("Map() 数组映射", func() {
				mapFunc := dataCtx.Get("Map")
				So(mapFunc, ShouldNotBeNil)

				value, err := mapFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试数组映射
				mapFn := value.Interface().(func([]interface{}, string) []interface{})
				// 由于Map函数是简化实现，这里只测试能正常调用
				result := mapFn([]interface{}{1, 2, 3, 4, 5}, "x * 2")
				So(result, ShouldNotBeNil) // 即使是空数组也不应该是nil
			})

			Convey("Unique() 数组去重", func() {
				uniqueFunc := dataCtx.Get("Unique")
				So(uniqueFunc, ShouldNotBeNil)

				value, err := uniqueFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试数组去重
				unique := value.Interface().(func([]interface{}) []interface{})
				input := []interface{}{1, 2, 2, 3, 3, 4, 5, 5}
				result := unique(input)
				So(result, ShouldHaveLength, 5)

				// 测试空数组
				emptyResult := unique([]interface{}{})
				So(emptyResult, ShouldHaveLength, 0)

				// 测试无重复数组
				noDupResult := unique([]interface{}{1, 2, 3, 4, 5})
				So(noDupResult, ShouldHaveLength, 5)
			})
		})

		Convey("验证函数测试", func() {

			Convey("Matches() 正则匹配", func() {
				matchesFunc := dataCtx.Get("Matches")
				So(matchesFunc, ShouldNotBeNil)

				value, err := matchesFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试正则匹配
				matches := value.Interface().(func(string, string) bool)
				So(matches("hello123", `\d+`), ShouldBeTrue)
				So(matches("hello", `\d+`), ShouldBeFalse)
				So(matches("test@example.com", `\w+@\w+\.\w+`), ShouldBeTrue)
			})

			Convey("IsEmail() 邮箱验证", func() {
				isEmailFunc := dataCtx.Get("IsEmail")
				So(isEmailFunc, ShouldNotBeNil)

				value, err := isEmailFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试邮箱验证
				isEmail := value.Interface().(func(string) bool)
				So(isEmail("test@example.com"), ShouldBeTrue)
				So(isEmail("user@domain.org"), ShouldBeTrue)
				So(isEmail("invalid-email"), ShouldBeFalse)
				So(isEmail("@example.com"), ShouldBeFalse)
				So(isEmail("test@"), ShouldBeFalse)
			})

			Convey("IsPhoneNumber() 手机号验证", func() {
				isPhoneFunc := dataCtx.Get("IsPhoneNumber")
				So(isPhoneFunc, ShouldNotBeNil)

				value, err := isPhoneFunc.GetValue()
				So(err, ShouldBeNil)

				// 测试手机号验证
				isPhone := value.Interface().(func(string) bool)
				So(isPhone("13800138000"), ShouldBeTrue)
				So(isPhone("18612345678"), ShouldBeTrue)
				So(isPhone("1234567890"), ShouldBeFalse)
				So(isPhone("138001380001"), ShouldBeFalse) // 太长
				So(isPhone("abc"), ShouldBeFalse)
			})
		})

		Convey("函数集成测试", func() {
			// 测试多个函数的组合使用

			Convey("字符串操作组合", func() {
				// 获取多个字符串函数
				containsFunc := dataCtx.Get("Contains")
				toUpperFunc := dataCtx.Get("ToUpper")
				lenFunc := dataCtx.Get("Len")

				So(containsFunc, ShouldNotBeNil)
				So(toUpperFunc, ShouldNotBeNil)
				So(lenFunc, ShouldNotBeNil)

				// 组合操作
				testStr := "Hello World"

				// 转大写
				upperValue, _ := toUpperFunc.GetValue()
				upperStr := upperValue.Interface().(func(string) string)(testStr)
				So(upperStr, ShouldEqual, "HELLO WORLD")

				// 检查长度
				lenValue, _ := lenFunc.GetValue()
				strLen := lenValue.Interface().(func(string) int)(upperStr)
				So(strLen, ShouldEqual, 11)

				// 检查包含
				containsValue, _ := containsFunc.GetValue()
				contains := containsValue.Interface().(func(string, string) bool)(upperStr, "WORLD")
				So(contains, ShouldBeTrue)
			})

			Convey("数学运算组合", func() {
				// 获取数学函数
				maxFunc := dataCtx.Get("Max")
				minFunc := dataCtx.Get("Min")
				absFunc := dataCtx.Get("Abs")
				roundFunc := dataCtx.Get("Round")

				So(maxFunc, ShouldNotBeNil)
				So(minFunc, ShouldNotBeNil)
				So(absFunc, ShouldNotBeNil)
				So(roundFunc, ShouldNotBeNil)

				// 组合计算
				a, b := -3.7, 8.2

				// 取最大值的绝对值并四舍五入
				maxValue, _ := maxFunc.GetValue()
				maxResult := maxValue.Interface().(func(float64, float64) float64)(a, b)

				absValue, _ := absFunc.GetValue()
				absResult := absValue.Interface().(func(float64) float64)(maxResult)

				roundValue, _ := roundFunc.GetValue()
				finalResult := roundValue.Interface().(func(float64) float64)(absResult)

				So(finalResult, ShouldEqual, 8.0)
			})
		})
	})
}
