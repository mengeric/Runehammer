package runehammer

import (
	"testing"
	"reflect"
	"sync"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/robfig/cron/v3"
)

// TestEngineContext 测试引擎上下文数据管理
func TestEngineContext(t *testing.T) {
	Convey("引擎数据上下文测试", t, func() {
		
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
		
		Convey("输入数据注入测试", func() {
			
			Convey("Map类型注入", func() {
				// 创建数据上下文
				dataCtx := ast.NewDataContext()
				
				// 测试Map类型输入
				input := map[string]interface{}{
					"customer": map[string]interface{}{
						"age":    25,
						"vip":    true,
						"name":   "张三",
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
					config, mapper, cache, cacheKeys, logger,
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
				
				boolTests := []struct{
					name string
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