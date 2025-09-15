package runehammer

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestRunehammer 测试主接口和工厂方法
func TestRunehammer(t *testing.T) {
	Convey("Runehammer主接口测试", t, func() {
		
		Convey("Engine接口定义", func() {
			
			Convey("接口方法签名验证", func() {
				// 验证接口定义正确性，通过编译即可确保接口正确
				var engine Engine[map[string]interface{}]
				
				// 模拟引擎实现
				engine = NewEngineImpl[map[string]interface{}](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				
				So(engine, ShouldNotBeNil)
				So(engine, ShouldImplement, (*Engine[map[string]interface{}])(nil))
				
				engine.Close()
			})
			
			Convey("泛型支持验证", func() {
				// 测试不同的泛型类型
				
				// string类型
				var stringEngine Engine[string]
				stringEngine = NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				So(stringEngine, ShouldNotBeNil)
				stringEngine.Close()
				
				// 结构体类型
				type TestResult struct {
					Score int    `json:"score"`
					Grade string `json:"grade"`
				}
				
				var structEngine Engine[TestResult]
				structEngine = NewEngineImpl[TestResult](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				So(structEngine, ShouldNotBeNil)
				structEngine.Close()
				
				// 切片类型
				var sliceEngine Engine[[]string]
				sliceEngine = NewEngineImpl[[]string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				So(sliceEngine, ShouldNotBeNil)
				sliceEngine.Close()
			})
		})
		
		Convey("New工厂方法", func() {
			
			Convey("基本配置创建引擎", func() {
				SkipConvey("需要真实数据库连接", func() {
					// 这个测试需要真实的数据库配置
					// 在实际环境中取消Skip并提供正确的数据库配置
					engine, err := New[map[string]interface{}](
						WithDSN("mysql://user:pass@localhost/test"),
					)
					So(err, ShouldBeNil)
					So(engine, ShouldNotBeNil)
					So(engine, ShouldImplement, (*Engine[map[string]interface{}])(nil))
					
					engine.Close()
				})
			})
			
			Convey("完整配置创建引擎", func() {
				SkipConvey("需要真实数据库和Redis连接", func() {
					// 这个测试需要真实的外部依赖
					engine, err := New[map[string]interface{}](
						WithDSN("mysql://user:pass@localhost/test"),
						WithCache(NewMemoryCache(1000)),
						WithLogger(NewDefaultLogger()),
						WithAutoMigrate(),
					)
					So(err, ShouldBeNil)
					So(engine, ShouldNotBeNil)
					
					engine.Close()
				})
			})
			
			Convey("配置验证失败", func() {
				// 测试无效配置
				engine, err := New[map[string]interface{}]()
				So(err, ShouldNotBeNil)
				So(engine, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "配置验证失败")
			})
			
			Convey("数据库连接失败", func() {
				// 测试无效数据库DSN
				engine, err := New[map[string]interface{}](
					WithDSN("invalid_dsn"),
				)
				So(err, ShouldNotBeNil)
				So(engine, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "数据库初始化失败")
			})
		})
		
		Convey("引擎执行接口", func() {
			
			Convey("Exec方法签名", func() {
				engine := NewEngineImpl[map[string]interface{}](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				// 验证方法存在且签名正确
				So(func() {
					engine.Exec(ctx, "test_biz", map[string]interface{}{"test": "data"})
				}, ShouldNotPanic)
			})
			
			Convey("不同输入类型支持", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				// map类型输入
				So(func() {
					engine.Exec(ctx, "test_biz", map[string]interface{}{
						"user": "john",
						"age":  25,
					})
				}, ShouldNotPanic)
				
				// 结构体类型输入
				type User struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}
				
				So(func() {
					engine.Exec(ctx, "test_biz", User{Name: "alice", Age: 30})
				}, ShouldNotPanic)
				
				// 基础类型输入
				So(func() {
					engine.Exec(ctx, "test_biz", "simple_string")
				}, ShouldNotPanic)
			})
			
			Convey("不同返回类型支持", func() {
				// string返回类型
				stringEngine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer stringEngine.Close()
				
				// int返回类型
				intEngine := NewEngineImpl[int](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer intEngine.Close()
				
				// 结构体返回类型
				type Result struct {
					Status  string `json:"status"`
					Message string `json:"message"`
				}
				
				structEngine := NewEngineImpl[Result](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer structEngine.Close()
				
				ctx := context.Background()
				input := map[string]interface{}{"test": true}
				
				// 所有引擎都应该能正常调用
				So(func() {
					stringEngine.Exec(ctx, "test", input)
				}, ShouldNotPanic)
				
				So(func() {
					intEngine.Exec(ctx, "test", input)
				}, ShouldNotPanic)
				
				So(func() {
					structEngine.Exec(ctx, "test", input)
				}, ShouldNotPanic)
			})
		})
		
		Convey("资源管理", func() {
			
			Convey("Close方法", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				
				// 验证Close方法存在
				err := engine.Close()
				So(err, ShouldBeNil)
			})
			
			Convey("重复关闭", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				
				// 第一次关闭
				err1 := engine.Close()
				So(err1, ShouldBeNil)
				
				// 第二次关闭应该也成功
				err2 := engine.Close()
				So(err2, ShouldBeNil)
			})
			
			Convey("关闭后调用方法", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				
				// 关闭引擎
				err := engine.Close()
				So(err, ShouldBeNil)
				
				// 关闭后调用Exec应该返回错误
				ctx := context.Background()
				result, err := engine.Exec(ctx, "test", map[string]interface{}{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "引擎已关闭")
				
				// 验证返回值为零值
				var zeroValue string
				So(result, ShouldEqual, zeroValue)
			})
		})
		
		Convey("上下文处理", func() {
			
			Convey("正常上下文", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				So(func() {
					engine.Exec(ctx, "test", map[string]interface{}{})
				}, ShouldNotPanic)
			})
			
			Convey("带值的上下文", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.WithValue(context.Background(), "request_id", "req-123")
				
				So(func() {
					engine.Exec(ctx, "test", map[string]interface{}{})
				}, ShouldNotPanic)
			})
			
			Convey("取消的上下文", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // 立即取消
				
				// 取消的上下文应该导致快速失败
				result, err := engine.Exec(ctx, "test", map[string]interface{}{})
				So(err, ShouldNotBeNil)
				
				var zeroValue string
				So(result, ShouldEqual, zeroValue)
			})
			
			Convey("nil上下文", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				// nil上下文应该不会panic，但可能返回错误
				So(func() {
					engine.Exec(nil, "test", map[string]interface{}{})
				}, ShouldNotPanic)
			})
		})
		
		Convey("业务码处理", func() {
			
			Convey("正常业务码", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				normalCodes := []string{
					"USER_VALIDATE",
					"LOAN_APPROVAL",
					"RISK_ASSESSMENT",
					"business_code_123",
				}
				
				for _, code := range normalCodes {
					So(func() {
						engine.Exec(ctx, code, map[string]interface{}{})
					}, ShouldNotPanic)
				}
			})
			
			Convey("空业务码", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				// 空业务码应该返回错误
				result, err := engine.Exec(ctx, "", map[string]interface{}{})
				So(err, ShouldNotBeNil)
				
				var zeroValue string
				So(result, ShouldEqual, zeroValue)
			})
			
			Convey("特殊字符业务码", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				specialCodes := []string{
					"中文业务码",
					"code with spaces",
					"code-with-dashes",
					"code_with_underscores",
					"code.with.dots",
					"code!@#$%^&*()",
				}
				
				for _, code := range specialCodes {
					So(func() {
						engine.Exec(ctx, code, map[string]interface{}{})
					}, ShouldNotPanic)
				}
			})
		})
		
		Convey("并发安全性", func() {
			
			Convey("并发执行", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				concurrency := 50
				done := make(chan bool, concurrency)
				
				for i := 0; i < concurrency; i++ {
					go func(id int) {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("Concurrent goroutine %d panicked: %v", id, r)
							}
							done <- true
						}()
						
						ctx := context.Background()
						bizCode := "concurrent_test"
						input := map[string]interface{}{
							"id":   id,
							"data": "test_data",
						}
						
						// 执行多次调用
						for j := 0; j < 10; j++ {
							engine.Exec(ctx, bizCode, input)
						}
					}(i)
				}
				
				// 等待所有goroutine完成
				for i := 0; i < concurrency; i++ {
					<-done
				}
				
				// 验证没有panic发生
				So(true, ShouldBeTrue)
			})
			
			Convey("并发关闭", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				
				concurrency := 10
				done := make(chan bool, concurrency)
				
				// 启动并发执行
				for i := 0; i < concurrency-1; i++ {
					go func(id int) {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("Concurrent exec goroutine %d panicked: %v", id, r)
							}
							done <- true
						}()
						
						ctx := context.Background()
						for j := 0; j < 100; j++ {
							engine.Exec(ctx, "test", map[string]interface{}{})
						}
					}(i)
				}
				
				// 同时执行关闭
				go func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("Close goroutine panicked: %v", r)
						}
						done <- true
					}()
					
					engine.Close()
				}()
				
				// 等待所有goroutine完成
				for i := 0; i < concurrency; i++ {
					<-done
				}
				
				// 验证没有panic发生
				So(true, ShouldBeTrue)
			})
		})
	})
}

// TestRunehammerIntegration 测试集成场景
func TestRunehammerIntegration(t *testing.T) {
	Convey("Runehammer集成测试", t, func() {
		
		Convey("接口多态性", func() {
			
			Convey("不同实现的兼容性", func() {
				// 创建不同配置的引擎实例
				engines := []Engine[string]{
					NewEngineImpl[string](
						&Config{dsn: "mock"},
						&mockRuleMapper{},
						NewMemoryCache(1000),
						CacheKeyBuilder{},
						NewNoopLogger(),
						nil,
						nil,
						nil,
						false,
					),
					NewEngineImpl[string](
						&Config{dsn: "mock"},
						&mockRuleMapper{},
						nil, // 无缓存
						CacheKeyBuilder{},
						NewDefaultLogger(),
						nil,
						nil,
						nil,
						false,
					),
				}
				
				ctx := context.Background()
				input := map[string]interface{}{"test": "data"}
				
				for i, engine := range engines {
					Convey(fmt.Sprintf("引擎 %d 接口兼容性", i), func() {
						// 验证所有引擎都实现了Engine接口
						So(engine, ShouldImplement, (*Engine[string])(nil))
						
						// 验证所有方法都能正常调用
						So(func() {
							engine.Exec(ctx, "test_biz", input)
							engine.Close()
						}, ShouldNotPanic)
					})
				}
			})
		})
		
		Convey("实际使用场景模拟", func() {
			
			Convey("典型业务使用模式", func() {
				engine := NewEngineImpl[map[string]interface{}](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				// 模拟典型的业务使用场景
				testCases := []struct {
					bizCode string
					input   map[string]interface{}
					desc    string
				}{
					{
						bizCode: "USER_REGISTRATION",
						input: map[string]interface{}{
							"email":    "user@example.com",
							"age":      25,
							"country":  "US",
							"verified": true,
						},
						desc: "用户注册验证",
					},
					{
						bizCode: "LOAN_APPROVAL",
						input: map[string]interface{}{
							"amount":      50000,
							"credit_score": 750,
							"income":      80000,
							"debt_ratio":  0.3,
						},
						desc: "贷款审批",
					},
					{
						bizCode: "RISK_ASSESSMENT",
						input: map[string]interface{}{
							"transaction_amount": 1000,
							"user_level":         "premium",
							"location":           "domestic",
							"time_of_day":        "business_hours",
						},
						desc: "风险评估",
					},
				}
				
				for _, tc := range testCases {
					Convey("场景: "+tc.desc, func() {
						So(func() {
							result, err := engine.Exec(ctx, tc.bizCode, tc.input)
							// 即使返回错误（因为没有真实规则），也不应该panic
							_ = result
							_ = err
						}, ShouldNotPanic)
					})
				}
			})
		})
		
		Convey("错误恢复能力", func() {
			
			Convey("处理无效输入", func() {
				engine := NewEngineImpl[string](
					&Config{dsn: "mock"},
					&mockRuleMapper{},
					NewMemoryCache(1000),
					CacheKeyBuilder{},
					NewNoopLogger(),
					nil,
					nil,
					nil,
					false,
				)
				defer engine.Close()
				
				ctx := context.Background()
				
				// 测试各种无效输入
				invalidInputs := []interface{}{
					nil,
					make(chan int), // 无法序列化的类型
					func() {},      // 函数类型
					complex(1, 2),  // 复数类型
				}
				
				for _, input := range invalidInputs {
					So(func() {
						engine.Exec(ctx, "test", input)
					}, ShouldNotPanic)
				}
			})
		})
	})
}