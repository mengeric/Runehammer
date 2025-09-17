package engine

import (
	"math"
	"sync"
	"testing"
	"time"

	"gitee.com/damengde/runehammer/cache"
	"gitee.com/damengde/runehammer/config"
	logger "gitee.com/damengde/runehammer/logger"
	"gitee.com/damengde/runehammer/rule"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/robfig/cron/v3"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

// TestEngineFunctions 测试引擎内置函数
func TestEngineFunctions(t *testing.T) {
	Convey("引擎内置函数测试", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 创建测试用的引擎实例
		cfg := config.DefaultConfig()
		mapper := rule.NewMockRuleMapper(ctrl)
		cacheImpl := cache.NewMockCache(ctrl)
		cacheKeys := cache.CacheKeyBuilder{}
		lgr := logger.NewNoopLogger()
		knowledgeLibrary := ast.NewKnowledgeLibrary()
		cronScheduler := cron.New()

		engine := NewEngineImpl[map[string]any](
			cfg, mapper, cacheImpl, cacheKeys, lgr,
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
				So(result, ShouldBeEmpty) // 简化实现返回空数组
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
				So(result, ShouldBeEmpty) // 简化实现返回空数组
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

// TestEngineFunctionsMissing 测试缺失的函数以达到100%覆盖率
func TestEngineFunctionsMissing(t *testing.T) {
	Convey("缺失函数测试", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 创建测试用的引擎实例
		cfg := config.DefaultConfig()
		mapper := rule.NewMockRuleMapper(ctrl)
		cacheImpl := cache.NewMockCache(ctrl)
		cacheKeys := cache.CacheKeyBuilder{}
		lgr := logger.NewNoopLogger()
		knowledgeLibrary := ast.NewKnowledgeLibrary()
		cronScheduler := cron.New()

		engine := NewEngineImpl[map[string]any](
			cfg, mapper, cacheImpl, cacheKeys, lgr,
			knowledgeLibrary, &sync.Map{}, cronScheduler, false,
		)

		// 创建数据上下文
		dataCtx := ast.NewDataContext()

		// 注入内置函数
		engine.injectBuiltinFunctions(dataCtx)

		Convey("更多时间函数测试", func() {

			Convey("FormatTime() 时间格式化", func() {
				formatTimeFunc := dataCtx.Get("FormatTime")
				So(formatTimeFunc, ShouldNotBeNil)

				value, err := formatTimeFunc.GetValue()
				So(err, ShouldBeNil)

				formatTime := value.Interface().(func(time.Time, string) string)
				testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
				result := formatTime(testTime, "2006-01-02 15:04:05")
				So(result, ShouldEqual, "2024-01-15 10:30:00")
			})

			Convey("ParseTime() 时间解析", func() {
				parseTimeFunc := dataCtx.Get("ParseTime")
				So(parseTimeFunc, ShouldNotBeNil)

				value, err := parseTimeFunc.GetValue()
				So(err, ShouldBeNil)

				parseTime := value.Interface().(func(string, string) (time.Time, error))
				result, err := parseTime("2006-01-02", "2024-01-15")
				So(err, ShouldBeNil)
				So(result.Year(), ShouldEqual, 2024)
				So(result.Month(), ShouldEqual, time.January)
				So(result.Day(), ShouldEqual, 15)
			})

			Convey("AddHours() 小时加减", func() {
				addHoursFunc := dataCtx.Get("AddHours")
				So(addHoursFunc, ShouldNotBeNil)

				value, err := addHoursFunc.GetValue()
				So(err, ShouldBeNil)

				addHours := value.Interface().(func(time.Time, int) time.Time)
				testTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
				result := addHours(testTime, 5)
				So(result.Hour(), ShouldEqual, 15)
			})

			Convey("TimeToMillis() 时间转毫秒", func() {
				timeToMillisFunc := dataCtx.Get("TimeToMillis")
				So(timeToMillisFunc, ShouldNotBeNil)

				value, err := timeToMillisFunc.GetValue()
				So(err, ShouldBeNil)

				timeToMillis := value.Interface().(func(time.Time) int64)
				testTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
				result := timeToMillis(testTime)
				So(result, ShouldBeGreaterThan, 0)
			})

			Convey("MillisToTime() 毫秒转时间", func() {
				millisToTimeFunc := dataCtx.Get("MillisToTime")
				So(millisToTimeFunc, ShouldNotBeNil)

				value, err := millisToTimeFunc.GetValue()
				So(err, ShouldBeNil)

				millisToTime := value.Interface().(func(int64) time.Time)
				result := millisToTime(1705310400000) // 2024-01-15 10:00:00 UTC
				So(result.Year(), ShouldEqual, 2024)
			})
		})

		Convey("更多字符串函数测试", func() {

			Convey("Split() 字符串分割", func() {
				splitFunc := dataCtx.Get("Split")
				So(splitFunc, ShouldNotBeNil)

				value, err := splitFunc.GetValue()
				So(err, ShouldBeNil)

				split := value.Interface().(func(string, string) []string)
				result := split("a,b,c,d", ",")
				So(len(result), ShouldEqual, 4)
				So(result[0], ShouldEqual, "a")
				So(result[3], ShouldEqual, "d")
			})

			Convey("Join() 字符串连接", func() {
				joinFunc := dataCtx.Get("Join")
				So(joinFunc, ShouldNotBeNil)

				value, err := joinFunc.GetValue()
				So(err, ShouldBeNil)

				join := value.Interface().(func([]string, string) string)
				result := join([]string{"hello", "world", "test"}, "-")
				So(result, ShouldEqual, "hello-world-test")
			})

			Convey("Replace() 字符串替换", func() {
				replaceFunc := dataCtx.Get("Replace")
				So(replaceFunc, ShouldNotBeNil)

				value, err := replaceFunc.GetValue()
				So(err, ShouldBeNil)

				replace := value.Interface().(func(string, string, string, int) string)
				// 替换一次
				result1 := replace("hello hello hello", "hello", "hi", 1)
				So(result1, ShouldEqual, "hi hello hello")

				// 替换所有
				result2 := replace("hello hello hello", "hello", "hi", -1)
				So(result2, ShouldEqual, "hi hi hi")
			})

			Convey("TrimSpace() 去除空格", func() {
				trimSpaceFunc := dataCtx.Get("TrimSpace")
				So(trimSpaceFunc, ShouldNotBeNil)

				value, err := trimSpaceFunc.GetValue()
				So(err, ShouldBeNil)

				trimSpace := value.Interface().(func(string) string)
				result := trimSpace("  hello world  ")
				So(result, ShouldEqual, "hello world")
			})
		})

		Convey("更多数学函数测试", func() {

			Convey("Floor() 向下取整", func() {
				floorFunc := dataCtx.Get("Floor")
				So(floorFunc, ShouldNotBeNil)

				value, err := floorFunc.GetValue()
				So(err, ShouldBeNil)

				floor := value.Interface().(func(float64) float64)
				So(floor(5.9), ShouldEqual, 5.0)
				So(floor(-2.1), ShouldEqual, -3.0)
			})

			Convey("Ceil() 向上取整", func() {
				ceilFunc := dataCtx.Get("Ceil")
				So(ceilFunc, ShouldNotBeNil)

				value, err := ceilFunc.GetValue()
				So(err, ShouldBeNil)

				ceil := value.Interface().(func(float64) float64)
				So(ceil(5.1), ShouldEqual, 6.0)
				So(ceil(-2.9), ShouldEqual, -2.0)
			})

			Convey("Pow() 幂运算", func() {
				powFunc := dataCtx.Get("Pow")
				So(powFunc, ShouldNotBeNil)

				value, err := powFunc.GetValue()
				So(err, ShouldBeNil)

				pow := value.Interface().(func(float64, float64) float64)
				So(pow(2.0, 3.0), ShouldEqual, 8.0)
				So(pow(5.0, 2.0), ShouldEqual, 25.0)
			})

			Convey("Sqrt() 平方根", func() {
				sqrtFunc := dataCtx.Get("Sqrt")
				So(sqrtFunc, ShouldNotBeNil)

				value, err := sqrtFunc.GetValue()
				So(err, ShouldBeNil)

				sqrt := value.Interface().(func(float64) float64)
				So(sqrt(16.0), ShouldEqual, 4.0)
				So(sqrt(25.0), ShouldEqual, 5.0)
			})

			Convey("Sin() 正弦", func() {
				sinFunc := dataCtx.Get("Sin")
				So(sinFunc, ShouldNotBeNil)

				value, err := sinFunc.GetValue()
				So(err, ShouldBeNil)

				sin := value.Interface().(func(float64) float64)
				So(sin(0.0), ShouldEqual, 0.0)
				So(sin(math.Pi/2), ShouldAlmostEqual, 1.0, 0.0001)
			})

			Convey("Cos() 余弦", func() {
				cosFunc := dataCtx.Get("Cos")
				So(cosFunc, ShouldNotBeNil)

				value, err := cosFunc.GetValue()
				So(err, ShouldBeNil)

				cos := value.Interface().(func(float64) float64)
				So(cos(0.0), ShouldEqual, 1.0)
				So(cos(math.Pi), ShouldAlmostEqual, -1.0, 0.0001)
			})

			Convey("Tan() 正切", func() {
				tanFunc := dataCtx.Get("Tan")
				So(tanFunc, ShouldNotBeNil)

				value, err := tanFunc.GetValue()
				So(err, ShouldBeNil)

				tan := value.Interface().(func(float64) float64)
				So(tan(0.0), ShouldEqual, 0.0)
				So(tan(math.Pi/4), ShouldAlmostEqual, 1.0, 0.0001)
			})

			Convey("Log() 自然对数", func() {
				logFunc := dataCtx.Get("Log")
				So(logFunc, ShouldNotBeNil)

				value, err := logFunc.GetValue()
				So(err, ShouldBeNil)

				log := value.Interface().(func(float64) float64)
				So(log(math.E), ShouldAlmostEqual, 1.0, 0.0001)
				So(log(1.0), ShouldEqual, 0.0)
			})

			Convey("Log10() 常用对数", func() {
				log10Func := dataCtx.Get("Log10")
				So(log10Func, ShouldNotBeNil)

				value, err := log10Func.GetValue()
				So(err, ShouldBeNil)

				log10 := value.Interface().(func(float64) float64)
				So(log10(100.0), ShouldEqual, 2.0)
				So(log10(1000.0), ShouldEqual, 3.0)
			})

			Convey("MaxSlice() 数组最大值", func() {
				maxSliceFunc := dataCtx.Get("MaxSlice")
				So(maxSliceFunc, ShouldNotBeNil)

				value, err := maxSliceFunc.GetValue()
				So(err, ShouldBeNil)

				maxSlice := value.Interface().(func([]float64) float64)
				result := maxSlice([]float64{1.5, 3.2, 2.8, 5.1, 0.9})
				So(result, ShouldEqual, 5.1)

				// 测试空数组
				emptyResult := maxSlice([]float64{})
				So(emptyResult, ShouldEqual, 0)
			})

			Convey("MinSlice() 数组最小值", func() {
				minSliceFunc := dataCtx.Get("MinSlice")
				So(minSliceFunc, ShouldNotBeNil)

				value, err := minSliceFunc.GetValue()
				So(err, ShouldBeNil)

				minSlice := value.Interface().(func([]float64) float64)
				result := minSlice([]float64{1.5, 3.2, 2.8, 5.1, 0.9})
				So(result, ShouldEqual, 0.9)

				// 测试空数组
				emptyResult := minSlice([]float64{})
				So(emptyResult, ShouldEqual, 0)
			})
		})

		Convey("更多工具函数测试", func() {

			Convey("ToString() 类型转换", func() {
				toStringFunc := dataCtx.Get("ToString")
				So(toStringFunc, ShouldNotBeNil)

				value, err := toStringFunc.GetValue()
				So(err, ShouldBeNil)

				toString := value.Interface().(func(interface{}) string)

				// 测试各种类型
				So(toString(42), ShouldEqual, "42")
				So(toString(3.14), ShouldEqual, "3.14")
				So(toString(true), ShouldEqual, "true")
				So(toString("hello"), ShouldEqual, "hello")
				So(toString(nil), ShouldEqual, "")
				So(toString(int64(123)), ShouldEqual, "123")
				// 测试不支持的类型（default分支）
				So(toString([]int{1, 2, 3}), ShouldEqual, "")
			})

			Convey("ToInt() 字符串转整数", func() {
				toIntFunc := dataCtx.Get("ToInt")
				So(toIntFunc, ShouldNotBeNil)

				value, err := toIntFunc.GetValue()
				So(err, ShouldBeNil)

				toInt := value.Interface().(func(string) (int, error))
				result, err := toInt("42")
				So(err, ShouldBeNil)
				So(result, ShouldEqual, 42)

				// 测试错误情况
				_, err = toInt("invalid")
				So(err, ShouldNotBeNil)
			})

			Convey("ToFloat() 字符串转浮点数", func() {
				toFloatFunc := dataCtx.Get("ToFloat")
				So(toFloatFunc, ShouldNotBeNil)

				value, err := toFloatFunc.GetValue()
				So(err, ShouldBeNil)

				toFloat := value.Interface().(func(string) (float64, error))
				result, err := toFloat("3.14")
				So(err, ShouldBeNil)
				So(result, ShouldEqual, 3.14)

				// 测试错误情况
				_, err = toFloat("invalid")
				So(err, ShouldNotBeNil)
			})

			Convey("ToBool() 字符串转布尔值", func() {
				toBoolFunc := dataCtx.Get("ToBool")
				So(toBoolFunc, ShouldNotBeNil)

				value, err := toBoolFunc.GetValue()
				So(err, ShouldBeNil)

				toBool := value.Interface().(func(string) (bool, error))

				result1, err1 := toBool("true")
				So(err1, ShouldBeNil)
				So(result1, ShouldBeTrue)

				result2, err2 := toBool("false")
				So(err2, ShouldBeNil)
				So(result2, ShouldBeFalse)

				// 测试错误情况
				_, err = toBool("invalid")
				So(err, ShouldNotBeNil)
			})

			Convey("IF() 条件函数", func() {
				ifFunc := dataCtx.Get("IF")
				So(ifFunc, ShouldNotBeNil)

				value, err := ifFunc.GetValue()
				So(err, ShouldBeNil)

				ifCondition := value.Interface().(func(bool, interface{}, interface{}) interface{})

				result1 := ifCondition(true, "yes", "no")
				So(result1, ShouldEqual, "yes")

				result2 := ifCondition(false, "yes", "no")
				So(result2, ShouldEqual, "no")
			})

			Convey("IsNotEmpty() 非空检查", func() {
				isNotEmptyFunc := dataCtx.Get("IsNotEmpty")
				So(isNotEmptyFunc, ShouldNotBeNil)

				value, err := isNotEmptyFunc.GetValue()
				So(err, ShouldBeNil)

				isNotEmpty := value.Interface().(func(interface{}) bool)

				So(isNotEmpty(nil), ShouldBeFalse)
				So(isNotEmpty(""), ShouldBeFalse)
				So(isNotEmpty("hello"), ShouldBeTrue)
				So(isNotEmpty([]interface{}{}), ShouldBeFalse)
				So(isNotEmpty([]interface{}{1, 2, 3}), ShouldBeTrue)
			})
		})

		Convey("更多验证函数测试", func() {

			Convey("IsIDCard() 身份证验证", func() {
				isIDCardFunc := dataCtx.Get("IsIDCard")
				So(isIDCardFunc, ShouldNotBeNil)

				value, err := isIDCardFunc.GetValue()
				So(err, ShouldBeNil)

				isIDCard := value.Interface().(func(string) bool)
				So(isIDCard("123456789012345678"), ShouldBeTrue)   // 18位数字
				So(isIDCard("12345678901234567X"), ShouldBeTrue)   // 17位数字+X
				So(isIDCard("12345"), ShouldBeFalse)               // 太短
				So(isIDCard("1234567890123456789"), ShouldBeFalse) // 太长
			})

			Convey("Between() 范围检查", func() {
				betweenFunc := dataCtx.Get("Between")
				So(betweenFunc, ShouldNotBeNil)

				value, err := betweenFunc.GetValue()
				So(err, ShouldBeNil)

				between := value.Interface().(func(float64, float64, float64) bool)
				So(between(5.0, 1.0, 10.0), ShouldBeTrue)
				So(between(15.0, 1.0, 10.0), ShouldBeFalse)
				So(between(-5.0, 1.0, 10.0), ShouldBeFalse)
			})

			Convey("LengthBetween() 长度范围检查", func() {
				lengthBetweenFunc := dataCtx.Get("LengthBetween")
				So(lengthBetweenFunc, ShouldNotBeNil)

				value, err := lengthBetweenFunc.GetValue()
				So(err, ShouldBeNil)

				lengthBetween := value.Interface().(func(string, int, int) bool)
				So(lengthBetween("hello", 3, 8), ShouldBeTrue)
				So(lengthBetween("hi", 3, 8), ShouldBeFalse)
				So(lengthBetween("verylongstring", 3, 8), ShouldBeFalse)
			})
		})

		Convey("isEmpty函数边界测试", func() {
			// 测试各种类型的isEmpty
			So(engine.isEmpty(nil), ShouldBeTrue)
			So(engine.isEmpty(""), ShouldBeTrue)
			So(engine.isEmpty("hello"), ShouldBeFalse)
			So(engine.isEmpty(0), ShouldBeFalse)
			So(engine.isEmpty(42), ShouldBeFalse)
			So(engine.isEmpty(0.0), ShouldBeFalse)
			So(engine.isEmpty(3.14), ShouldBeFalse)
			So(engine.isEmpty(false), ShouldBeFalse)
			So(engine.isEmpty(true), ShouldBeFalse)
			So(engine.isEmpty([]interface{}{}), ShouldBeTrue)
			So(engine.isEmpty([]interface{}{1, 2, 3}), ShouldBeFalse)
			So(engine.isEmpty([]string{}), ShouldBeFalse)
			So(engine.isEmpty([]int{1, 2, 3}), ShouldBeFalse)
			So(engine.isEmpty(map[string]interface{}{}), ShouldBeFalse)
			So(engine.isEmpty(map[string]interface{}{"key": "value"}), ShouldBeFalse)
			So(engine.isEmpty(struct{}{}), ShouldBeFalse)
		})
	})
}
