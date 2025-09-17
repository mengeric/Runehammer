package engine

import (
	"reflect"
	"sync"
	"testing"

	"gitee.com/damengde/runehammer/cache"
	"gitee.com/damengde/runehammer/config"
	logger "gitee.com/damengde/runehammer/logger"
	"gitee.com/damengde/runehammer/rule"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/robfig/cron/v3"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

// TestEngineContext 测试引擎上下文数据管理
func TestEngineContext(t *testing.T) {
	Convey("引擎数据上下文测试", t, func() {
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

		Convey("输入数据注入测试", func() {

			Convey("Map类型注入", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 测试Map类型输入
				input := map[string]interface{}{
					"customer": map[string]interface{}{
						"age":  25,
						"vip":  true,
						"name": "张三",
					},
					"order": map[string]interface{}{
						"amount": 1000.0,
						"status": "paid",
					},
				}

				// 注入输入数据
				err := engine.injectInputData(dataCtx, input)
				So(err, ShouldBeNil)

				// 验证注入结果 - Map类型应该作为Params整体注入
				paramsValue := dataCtx.Get("Params")
				So(paramsValue, ShouldNotBeNil)

				actualValue, err := paramsValue.GetValue()
				So(err, ShouldBeNil)

				actualData := actualValue.Interface()
				actualMap, ok := actualData.(map[string]interface{})
				So(ok, ShouldBeTrue)
				So(actualMap["customer"], ShouldNotBeNil)
				So(actualMap["order"], ShouldNotBeNil)

				customerData := actualMap["customer"].(map[string]interface{})
				So(customerData["age"], ShouldEqual, 25)
				So(customerData["vip"], ShouldEqual, true)
				So(customerData["name"], ShouldEqual, "张三")

				orderData := actualMap["order"].(map[string]interface{})
				So(orderData["amount"], ShouldEqual, 1000.0)
				So(orderData["status"], ShouldEqual, "paid")
			})

			Convey("结构体类型注入", func() {
				// 定义测试结构体
				type Customer struct {
					Age  int    `json:"age"`
					VIP  bool   `json:"vip"`
					Name string `json:"name"`
				}

				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 测试结构体类型输入
				customer := Customer{
					Age:  30,
					VIP:  true,
					Name: "李四",
				}

				// 注入输入数据
				err := engine.injectInputData(dataCtx, customer)
				So(err, ShouldBeNil)

				// 验证注入结果 - 结构体应该使用类型名（小写）
				customerValue := dataCtx.Get("customer")
				So(customerValue, ShouldNotBeNil)

				actualValue, err := customerValue.GetValue()
				So(err, ShouldBeNil)

				actualData := actualValue.Interface()
				actualCustomer, ok := actualData.(Customer)
				So(ok, ShouldBeTrue)
				So(actualCustomer.Age, ShouldEqual, 30)
				So(actualCustomer.VIP, ShouldEqual, true)
				So(actualCustomer.Name, ShouldEqual, "李四")
			})

			Convey("匿名结构体注入", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 测试匿名结构体输入
				input := struct {
					Value int
					Flag  bool
				}{
					Value: 100,
					Flag:  true,
				}

				// 注入输入数据
				err := engine.injectInputData(dataCtx, input)
				So(err, ShouldBeNil)

				// 验证注入结果 - 匿名结构体应该使用"Params"
				paramsValue := dataCtx.Get("Params")
				So(paramsValue, ShouldNotBeNil)

				actualValue, err := paramsValue.GetValue()
				So(err, ShouldBeNil)

				actualData := actualValue.Interface()
				// 由于是匿名结构体，类型比较需要使用反射
				actualType := reflect.TypeOf(actualData)
				inputType := reflect.TypeOf(input)
				So(actualType, ShouldEqual, inputType)
			})

			Convey("基本类型注入", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 测试基本类型输入
				input := 42

				// 注入输入数据
				err := engine.injectInputData(dataCtx, input)
				So(err, ShouldBeNil)

				// 验证注入结果 - 基本类型应该使用"Params"
				paramsValue := dataCtx.Get("Params")
				So(paramsValue, ShouldNotBeNil)

				actualValue, err := paramsValue.GetValue()
				So(err, ShouldBeNil)

				actualData := actualValue.Interface()
				So(actualData, ShouldEqual, 42)
			})

			Convey("指针类型注入", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 测试指针类型输入
				value := "test string"
				input := &value

				// 注入输入数据
				err := engine.injectInputData(dataCtx, input)
				So(err, ShouldBeNil)

				// 验证注入结果
				paramsValue := dataCtx.Get("Params")
				So(paramsValue, ShouldNotBeNil)

				actualValue, err := paramsValue.GetValue()
				So(err, ShouldBeNil)

				actualData := actualValue.Interface()
				// 指针被解引用，应该是原始值
				actualPtr, ok := actualData.(*string)
				So(ok, ShouldBeTrue)
				So(*actualPtr, ShouldEqual, "test string")
			})
		})

		Convey("结果提取测试", func() {

			Convey("interface{}类型结果", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 设置result变量
				resultData := map[string]interface{}{
					"success": true,
					"message": "操作成功",
					"data":    42,
				}
				err := dataCtx.Add("result", resultData)
				So(err, ShouldBeNil)

				// 提取结果
				result, err := engine.extractResult(dataCtx)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result["success"], ShouldEqual, true)
				So(result["message"], ShouldEqual, "操作成功")
				So(result["data"], ShouldEqual, 42)
			})

			Convey("Map类型结果", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 设置result变量
				resultData := map[string]any{
					"key1": "value1",
					"key2": "value2",
				}
				err := dataCtx.Add("result", resultData)
				So(err, ShouldBeNil)

				// 提取结果 - 使用正确的interface{}类型
				result, err := engine.extractResult(dataCtx)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
			})

			Convey("空结果处理", func() {
				// 创建数据上下文，不设置result变量
				dataCtx := ast.NewDataContext()

				// 提取结果
				result, err := engine.extractResult(dataCtx)
				So(err, ShouldBeNil)
				So(result, ShouldBeZeroValue) // 应该返回零值
			})

			Convey("结构体结果", func() {
				// 定义结果结构体
				type ResultStruct struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
					Success bool   `json:"success"`
				}

				// 创建专门用于结构体类型的引擎
				structEngine := NewEngineImpl[ResultStruct](
					cfg, mapper, cacheImpl, cacheKeys, lgr,
					knowledgeLibrary, &sync.Map{}, cronScheduler, false,
				)

				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 设置result变量 - 使用map模拟结构体数据
				resultData := map[string]interface{}{
					"code":    200,
					"message": "成功",
					"success": true,
				}
				err := dataCtx.Add("result", resultData)
				So(err, ShouldBeNil)

				// 提取结果
				result, err := structEngine.extractResult(dataCtx)
				So(err, ShouldBeNil)
				So(result.Code, ShouldEqual, 200)
				So(result.Message, ShouldEqual, "成功")
				So(result.Success, ShouldEqual, true)
			})
		})

		Convey("错误处理测试", func() {

			Convey("无效的结果值", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 设置一个无效的result值（例如channel，无法序列化）
				invalidResult := make(chan int)
				err := dataCtx.Add("result", invalidResult)
				So(err, ShouldBeNil)

				// 尝试提取结果
				result, err := engine.extractResult(dataCtx)
				So(err, ShouldNotBeNil) // 应该出错
				So(result, ShouldBeZeroValue)
			})
		})

		Convey("数据类型兼容性测试", func() {

			Convey("各种数值类型", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				// 测试不同数值类型
				testCases := []struct {
					name  string
					value interface{}
				}{
					{"int", int(42)},
					{"int32", int32(42)},
					{"int64", int64(42)},
					{"float32", float32(3.14)},
					{"float64", float64(3.14159)},
					{"uint", uint(42)},
				}

				for _, tc := range testCases {
					Convey("类型: "+tc.name, func() {
						err := engine.injectInputData(dataCtx, tc.value)
						So(err, ShouldBeNil)

						paramsValue := dataCtx.Get("Params")
						So(paramsValue, ShouldNotBeNil)

						actualValue, err := paramsValue.GetValue()
						So(err, ShouldBeNil)
						So(actualValue.Interface(), ShouldEqual, tc.value)
					})
				}
			})

			Convey("字符串类型", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				testStrings := []string{
					"普通字符串",
					"",
					"包含特殊字符: !@#$%^&*()",
					"中文字符串测试",
					"Multi\nLine\nString",
				}

				for _, str := range testStrings {
					Convey("字符串: "+str, func() {
						err := engine.injectInputData(dataCtx, str)
						So(err, ShouldBeNil)

						paramsValue := dataCtx.Get("Params")
						So(paramsValue, ShouldNotBeNil)

						actualValue, err := paramsValue.GetValue()
						So(err, ShouldBeNil)
						So(actualValue.Interface(), ShouldEqual, str)
					})
				}
			})

			Convey("布尔类型", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()

				boolTests := []struct {
					name  string
					value bool
				}{
					{"布尔值 true", true},
					{"布尔值 false", false},
				}

				for _, test := range boolTests {
					Convey(test.name, func() {
						err := engine.injectInputData(dataCtx, test.value)
						So(err, ShouldBeNil)

						paramsValue := dataCtx.Get("Params")
						So(paramsValue, ShouldNotBeNil)

						actualValue, err := paramsValue.GetValue()
						So(err, ShouldBeNil)
						So(actualValue.Interface(), ShouldEqual, test.value)
					})
				}
			})
		})
	})
}

// TestExtractResultFunctions 专门测试结果提取相关函数
func TestExtractResultFunctions(t *testing.T) {
	Convey("结果提取函数详细测试", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 创建测试用的引擎实例
		cfg2 := config.DefaultConfig()
		mapper := rule.NewMockRuleMapper(ctrl)
		cache2 := cache.NewMockCache(ctrl)
		cacheKeys := cache.CacheKeyBuilder{}
		lgr := logger.NewNoopLogger()
		knowledgeLibrary := ast.NewKnowledgeLibrary()
		cronScheduler := cron.New()

		Convey("extractInterfaceResult 函数", func() {
			// 创建 interface{} 类型引擎
			interfaceEngine := NewEngineImpl[any](
				cfg2, mapper, cache2, cacheKeys, lgr,
				knowledgeLibrary, &sync.Map{}, cronScheduler, false,
			)

			Convey("正常值提取", func() {
				testValues := []interface{}{
					"string value",
					42,
					true,
					map[string]interface{}{"key": "value"},
					[]int{1, 2, 3},
				}

				for _, testValue := range testValues {
					result, err := interfaceEngine.extractInterfaceResult(testValue)
					So(err, ShouldBeNil)
					So(result, ShouldEqual, testValue)
				}
			})

			Convey("nil值处理", func() {
				result, err := interfaceEngine.extractInterfaceResult(nil)
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})

		Convey("extractMapResult 函数", func() {
			// 创建 map 类型引擎
			mapEngine := NewEngineImpl[map[string]any](
				cfg2, mapper, cache2, cacheKeys, lgr,
				knowledgeLibrary, &sync.Map{}, cronScheduler, false,
			)

			Convey("有效map提取", func() {
				testMap := map[string]any{
					"key1":   "value1",
					"key2":   42,
					"key3":   true,
					"nested": map[string]any{"inner": "value"},
				}

				result, err := mapEngine.extractMapResult(testMap)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result["key1"], ShouldEqual, "value1")
				So(result["key2"], ShouldEqual, 42)
				So(result["key3"], ShouldEqual, true)

				nested, ok := result["nested"].(map[string]any)
				So(ok, ShouldBeTrue)
				So(nested["inner"], ShouldEqual, "value")
			})

			Convey("空map提取", func() {
				emptyMap := map[string]any{}
				result, err := mapEngine.extractMapResult(emptyMap)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(len(result), ShouldEqual, 0)
			})

			Convey("非map类型错误", func() {
				invalidTypes := []interface{}{
					"not a map",
					42,
					[]string{"array"},
					struct{ field string }{"value"},
				}

				for _, invalidType := range invalidTypes {
					_, err := mapEngine.extractMapResult(invalidType)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "不是有效的map类型")
				}
			})
		})

		Convey("extractPointerResult 函数", func() {
			type TestStruct struct {
				Name  string
				Value int
			}

			// 创建指针类型引擎
			ptrEngine := NewEngineImpl[*TestStruct](
				cfg2, mapper, cache2, cacheKeys, lgr,
				knowledgeLibrary, &sync.Map{}, cronScheduler, false,
			)

			Convey("有效指针提取", func() {
				testData := &TestStruct{
					Name:  "test name",
					Value: 123,
				}

				result, err := ptrEngine.extractPointerResult(testData)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result.Name, ShouldEqual, "test name")
				So(result.Value, ShouldEqual, 123)
			})

			Convey("nil指针提取", func() {
				var nilPtr *TestStruct = nil
				result, err := ptrEngine.extractPointerResult(nilPtr)
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
		})

		Convey("extractGenericResult 函数", func() {
			type GenericStruct struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
				Active  bool   `json:"active"`
			}

			// 创建泛型结构体引擎
			genericEngine := NewEngineImpl[GenericStruct](
				cfg2, mapper, cache2, cacheKeys, lgr,
				knowledgeLibrary, &sync.Map{}, cronScheduler, false,
			)

			Convey("有效数据转换", func() {
				inputData := map[string]interface{}{
					"message": "success",
					"code":    200,
					"active":  true,
				}

				result, err := genericEngine.extractGenericResult(inputData)
				So(err, ShouldBeNil)
				So(result.Message, ShouldEqual, "success")
				So(result.Code, ShouldEqual, 200)
				So(result.Active, ShouldEqual, true)
			})

			Convey("部分字段数据", func() {
				inputData := map[string]interface{}{
					"message": "partial data",
					// 缺少 code 和 active 字段
				}

				result, err := genericEngine.extractGenericResult(inputData)
				So(err, ShouldBeNil)
				So(result.Message, ShouldEqual, "partial data")
				So(result.Code, ShouldEqual, 0)       // 默认值
				So(result.Active, ShouldEqual, false) // 默认值
			})

			Convey("类型不匹配处理", func() {
				inputData := map[string]interface{}{
					"message": 123,    // 应该是string
					"code":    "text", // 应该是int
					"active":  "yes",  // 应该是bool
				}

				// JSON序列化/反序列化会尽力转换类型，某些可能失败
				_, err := genericEngine.extractGenericResult(inputData)
				So(err, ShouldNotBeNil) // 类型不匹配应该出错
			})

			Convey("无法序列化的数据", func() {
				// channel 类型无法序列化为JSON
				invalidData := make(chan int)

				_, err := genericEngine.extractGenericResult(invalidData)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "序列化结果失败")
			})

			Convey("循环引用数据", func() {
				// 创建循环引用的map
				cyclicMap := make(map[string]interface{})
				cyclicMap["self"] = cyclicMap

				_, err := genericEngine.extractGenericResult(cyclicMap)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "序列化结果失败")
			})
		})

		Convey("完整结果提取流程", func() {
			// 测试 extractResult 主函数
			engine := NewEngineImpl[map[string]any](
				cfg2, mapper, cache2, cacheKeys, lgr,
				knowledgeLibrary, &sync.Map{}, cronScheduler, false,
			)

			Convey("result变量为nil", func() {
				dataCtx := ast.NewDataContext()
				// 不添加result变量

				result, err := engine.extractResult(dataCtx)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil) // 应该返回零值，而不是nil
			})

			Convey("result变量获取失败", func() {
				// 这个测试比较难模拟，因为需要mock Grule的内部行为
				// 在实际情况下，GetValue()失败的情况比较少见
			})
		})
	})
}
