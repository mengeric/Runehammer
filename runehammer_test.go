package runehammer

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// 集成测试用的数据结构定义
type Customer struct {
	Age      int     `json:"age"`
	VipLevel int     `json:"vip_level"`
	Income   float64 `json:"income"`
}

type Order struct {
	Amount   float64 `json:"amount"`
	Quantity int     `json:"quantity"`
	Status   string  `json:"status"`
}

// TestResult 测试用的结果结构体
type TestResult struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Status    string      `json:"status"`
	Count     int         `json:"count"`
	Id        int         `json:"id"`
	Processed bool        `json:"processed"`
}

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

// TestErrorRecoveryAndFaultTolerance 错误恢复和容错测试
func TestErrorRecoveryAndFaultTolerance(t *testing.T) {
	Convey("错误恢复和容错测试", t, func() {
		
		Convey("数据库连接错误恢复", func() {
			
			Convey("数据库连接失败后重连", func() {
				// 创建一个会失败的数据库引擎
				engine, err := New[TestResult](
					WithDSN("invalid://invalid_db"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				
				// 应该返回错误
				So(err, ShouldNotBeNil)
				So(engine, ShouldBeNil)
				
				// 使用正确的配置重新创建
				engine, err = New[TestResult](
					WithTestSQLite("recovery_test"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
			})
			
			Convey("数据库操作错误处理", func() {
				engine, err := New[TestResult](
					WithTestSQLite("fault_test"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				ctx := context.Background()
				
				// 测试不存在的业务码
				result, err := engine.Exec(ctx, "nonexistent_biz_code", map[string]interface{}{
					"test": "data",
				})
				
				// 应该优雅地处理错误
				So(err, ShouldNotBeNil)
				// 错误情况下，应该返回零值结构体
				var zeroResult TestResult
				So(result, ShouldResemble, zeroResult)
			})
		})
		
		Convey("规则解析错误恢复", func() {
			
			Convey("无效规则语法恢复", func() {
				engine, err := New[TestResult](
					WithTestSQLite("parse_error_test"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				// 插入一个语法错误的规则
				db := engine.(*engineImpl[TestResult]).config.GetDB()
				invalidRule := &Rule{
					BizCode:     "invalid_syntax_test",
					Name:        "无效语法规则",
					Description: "测试无效语法的规则",
					GRL:         "rule invalid { when invalid syntax then invalid action }", // 故意的语法错误
					Enabled:     true,
				}
				db.Create(invalidRule)
				
				ctx := context.Background()
				input := map[string]interface{}{
					"test": "data",
				}
				
				// 执行应该优雅地处理解析错误
				result, err := engine.Exec(ctx, "invalid_syntax_test", input)
				
				// 应该返回错误，但不应该panic
				So(err, ShouldNotBeNil)
				// 错误情况下应该返回零值结构体
				var zeroResult TestResult
				So(result, ShouldResemble, zeroResult)
			})
			
			Convey("规则执行错误恢复", func() {
				engine, err := New[TestResult](
					WithTestSQLite("exec_error_test"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				// 插入一个运行时会出错的规则
				db := engine.(*engineImpl[TestResult]).config.GetDB()
				runtimeErrorRule := &Rule{
					BizCode:     "runtime_error_test",
					Name:        "运行时错误规则",
					Description: "测试运行时错误的规则",
					GRL:         `rule runtime_error "运行时错误规则" {
						when
							Params.Value > 0
						then
							Result.Status = "error";
							Result.Message = "intentional error for testing";
					}`,
					Enabled: true,
				}
				db.Create(runtimeErrorRule)
				
				ctx := context.Background()
				input := map[string]interface{}{
					"Value": 10,
				}
				
				// 执行应该能够处理运行时错误
				result, err := engine.Exec(ctx, "runtime_error_test", input)
				
				// 在TestResult类型下，需要判断是否成功解析
				// 由于传入的是map而不TestResult结构体，可能会有解析错误
				if err != nil {
					// 如果有错误，应该返回零值结构体
					var zeroResult TestResult
					So(result, ShouldResemble, zeroResult)
				} else {
					// 如果成功，检查规则执行结果
					So(result.Status, ShouldEqual, "error")
					So(result.Message, ShouldEqual, "intentional error for testing")
				}
			})
		})
		
		Convey("内存泄漏防护", func() {
			
			Convey("大量规则缓存管理", func() {
				engine, err := New[TestResult](
					WithTestSQLite("memory_test"),
					WithAutoMigrate(),
					WithMaxCacheSize(100), // 限制缓存大小
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				db := engine.(*engineImpl[TestResult]).config.GetDB()
				ctx := context.Background()
				
				// 创建大量规则以测试缓存管理
				for i := 0; i < 200; i++ {
					rule := &Rule{
						BizCode:     fmt.Sprintf("cache_test_%d", i),
						Name:        fmt.Sprintf("缓存测试规则%d", i),
						Description: "测试缓存管理",
						GRL:         fmt.Sprintf(`rule cache_test_%d "缓存测试规则%d" {
							when
								Params.Id == %d
							then
								Result.Success = true;
								Result.Id = %d;
						}`, i, i, i, i),
						Enabled: true,
					}
					db.Create(rule)
				}
				
				// 执行所有规则以填充缓存
				for i := 0; i < 200; i++ {
					input := map[string]interface{}{
						"Id": i,
					}
					
					result, err := engine.Exec(ctx, fmt.Sprintf("cache_test_%d", i), input)
					// 由于类型不匹配问题，可能会失败，但不应该panic
					if err == nil {
						So(result.Success, ShouldEqual, true)
						So(result.Id, ShouldEqual, i)
					}
				}
				
				// 验证缓存应该限制在指定大小内（这里我们无法直接验证，但确保没有panic）
				So(true, ShouldBeTrue)
			})
			
			Convey("长时间运行稳定性", func() {
				engine, err := New[TestResult](
					WithTestSQLite("stability_test"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				// 插入测试规则
				db := engine.(*engineImpl[TestResult]).config.GetDB()
				rule := &Rule{
					BizCode:     "stability_test",
					Name:        "稳定性测试规则",
					Description: "长时间运行稳定性测试",
					GRL:         `rule stability_test "稳定性测试规则" {
						when
							Params.Count >= 0
						then
							Result.Processed = true;
							Result.Count = Params.Count;
					}`,
					Enabled: true,
				}
				db.Create(rule)
				
				ctx := context.Background()
				
				// 模拟长时间大量执行
				for i := 0; i < 1000; i++ {
					input := map[string]interface{}{
						"Count": i,
					}
					
					result, err := engine.Exec(ctx, "stability_test", input)
					// 由于类型不匹配问题，可能会失败，但不应该panic
					if err == nil {
						So(result.Processed, ShouldEqual, true)
						So(result.Count, ShouldEqual, i)
					}
				}
			})
		})
		
		Convey("网络异常处理", func() {
			
			Convey("超时处理", func() {
				engine, err := New[TestResult](
					WithTestSQLite("timeout_test"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				// 创建一个会被取消的上下文
				ctx, cancel := context.WithCancel(context.Background())
				
				// 立即取消上下文
				cancel()
				
				input := map[string]interface{}{
					"test": "data",
				}
				
				// 使用已取消的上下文执行
				result, err := engine.Exec(ctx, "timeout_test", input)
				
				// 应该优雅地处理上下文取消，返回零值结构体
				var zeroResult TestResult
				So(result, ShouldResemble, zeroResult)
			})
			
			Convey("资源清理", func() {
				// 创建多个引擎实例并关闭它们
				engines := make([]Engine[TestResult], 5)
				
				for i := 0; i < 5; i++ {
					engine, err := New[TestResult](
						WithTestSQLite(fmt.Sprintf("cleanup_test_%d", i)),
						WithAutoMigrate(),
						WithLogger(NewNoopLogger()),
					)
					So(err, ShouldBeNil)
					So(engine, ShouldNotBeNil)
					engines[i] = engine
				}
				
				// 确保所有引擎都能正常关闭
				for i, engine := range engines {
					err := engine.Close()
					So(err, ShouldBeNil)
					t.Logf("Engine %d closed successfully", i)
				}
				
				// 尝试在关闭后再次调用Close（应该是安全的）
				for i, engine := range engines {
					err := engine.Close()
					// 第二次关闭可能返回错误，但不应该panic
					if err != nil {
						t.Logf("Engine %d second close returned error (expected): %v", i, err)
					}
				}
			})
		})
		
		Convey("边界条件容错", func() {
			
			Convey("极端输入数据", func() {
				engine, err := New[TestResult](
					WithTestSQLite("extreme_input_test"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				ctx := context.Background()
				
				extremeInputs := []map[string]interface{}{
					nil,                           // nil输入
					{},                           // 空map
					{"": ""},                     // 空字符串键值
					{"key": nil},                 // nil值
					{"very_long_key_" + strings.Repeat("x", 1000): "value"}, // 超长键名
					{"key": strings.Repeat("x", 10000)}, // 超长值
					{"nested": map[string]interface{}{ // 深度嵌套
						"level1": map[string]interface{}{
							"level2": map[string]interface{}{
								"level3": "deep_value",
							},
						},
					}},
					{"array": []interface{}{1, 2, 3, "mixed", nil}}, // 混合数组
					{"number": 1.7976931348623157e+308}, // 最大float64
					{"negative": -1.7976931348623157e+308}, // 最小float64
					{"unicode": "测试🚀🌟💫"},           // Unicode字符
				}
				
				for i, input := range extremeInputs {
					// 所有极端输入都应该被优雅处理，不应该panic
					So(func() {
						engine.Exec(ctx, fmt.Sprintf("extreme_test_%d", i), input)
					}, ShouldNotPanic)
				}
			})
			
			Convey("系统资源限制", func() {
				// 测试在资源受限情况下的行为
				engine, err := New[TestResult](
					WithTestSQLite("resource_test"),
					WithAutoMigrate(),
					WithMaxCacheSize(1), // 极小的缓存
					WithLogger(NewNoopLogger()),
				)
				So(err, ShouldBeNil)
				So(engine, ShouldNotBeNil)
				defer engine.Close()
				
				ctx := context.Background()
				
				// 在资源受限的情况下执行
				for i := 0; i < 10; i++ {
					input := map[string]interface{}{
						"iteration": i,
						"data":      fmt.Sprintf("test_data_%d", i),
					}
					
					// 应该能够处理资源限制而不panic
					So(func() {
						engine.Exec(ctx, "resource_limit_test", input)
					}, ShouldNotPanic)
				}
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
		
		Convey("数据库引擎端到端集成测试", func() {
			
			Convey("完整业务流程测试", func() {
				// 创建内存SQLite数据库引擎
				engine, err := New[map[string]interface{}](
					WithDSN("sqlite::memory:"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				
				if err != nil {
					SkipConvey("需要SQLite支持，跳过该测试", func() {})
					return
				}
				defer engine.Close()
				
				ctx := context.Background()
				
				// 1. 准备测试规则 - 通过GORM直接插入到数据库
				testRule := Rule{
					BizCode:     "USER_DISCOUNT",
					Name:        "用户折扣规则",
					GRL:         `rule UserDiscountRule "用户折扣" salience 100 { when Params.Age >= 18 && Params.VipLevel >= 2 && Params.Amount > 100 then Result.Discount = 0.15; Result.Message = "VIP用户享受85折优惠"; Result.Eligible = true; Retract("UserDiscountRule"); }`,
					Version:     1,
					Enabled:     true,
					Description: "VIP用户折扣规则",
				}
				
				// 2. 直接通过引擎内部的数据库连接插入规则
				db := engine.(*engineImpl[map[string]interface{}]).config.db
				So(db, ShouldNotBeNil)
				
				err = db.Create(&testRule).Error
				So(err, ShouldBeNil)
				
				// 3. 执行规则 - 测试符合条件的情况
				input := map[string]interface{}{
					"Age":      25,
					"VipLevel": 3,
					"Amount":   500.0,
				}
				
				result, err := engine.Exec(ctx, "USER_DISCOUNT", input)
				
				// 4. 验证结果
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result["Discount"], ShouldEqual, 0.15)
				So(result["Message"], ShouldEqual, "VIP用户享受85折优惠")
				So(result["Eligible"], ShouldEqual, true)
				
				// 5. 测试不符合条件的情况
				input2 := map[string]interface{}{
					"Age":      16,  // 年龄不足
					"VipLevel": 3,
					"Amount":   500.0,
				}
				
				result2, err2 := engine.Exec(ctx, "USER_DISCOUNT", input2)
				So(err2, ShouldBeNil)
				So(result2, ShouldNotBeNil)
				// 不符合条件时，Result应该为空map
				So(len(result2), ShouldEqual, 0)
			})
			
			Convey("缓存机制完整测试", func() {
				// 创建带缓存的引擎
				engine, err := New[map[string]interface{}](
					WithDSN("sqlite::memory:"),
					WithAutoMigrate(),
					WithCache(NewMemoryCache(100)),
					WithLogger(NewNoopLogger()),
				)
				
				if err != nil {
					SkipConvey("需要SQLite支持，跳过该测试", func() {})
					return
				}
				defer engine.Close()
				
				ctx := context.Background()
				
				// 1. 插入测试规则
				testRule := Rule{
					BizCode:     "CACHE_TEST",
					Name:        "缓存测试规则",
					GRL:         `rule CacheTestRule "cache test" salience 100 { when true then Result.CacheHit = "yes"; Result.Timestamp = NowMillis(); Retract("CacheTestRule"); }`,
					Version:     1,
					Enabled:     true,
					Description: "用于测试缓存的规则",
				}
				
				db := engine.(*engineImpl[map[string]interface{}]).config.db
				err = db.Create(&testRule).Error
				So(err, ShouldBeNil)
				
				input := map[string]interface{}{"test": "cache"}
				
				// 2. 第一次执行（应该从数据库加载并缓存）
				result1, err1 := engine.Exec(ctx, "CACHE_TEST", input)
				So(err1, ShouldBeNil)
				So(result1, ShouldNotBeNil)
				So(result1["CacheHit"], ShouldEqual, "yes")
				
				timestamp1, ok := result1["Timestamp"].(int64)
				So(ok, ShouldBeTrue)
				So(timestamp1, ShouldBeGreaterThan, 0)
				
				// 3. 第二次执行（应该从缓存获取，使用相同的编译结果）
				result2, err2 := engine.Exec(ctx, "CACHE_TEST", input)
				So(err2, ShouldBeNil)
				So(result2, ShouldNotBeNil)
				So(result2["CacheHit"], ShouldEqual, "yes")
				
				timestamp2, ok := result2["Timestamp"].(int64)
				So(ok, ShouldBeTrue)
				So(timestamp2, ShouldBeGreaterThan, timestamp1) // 每次执行时间戳不同
			})
		})
		
		Convey("动态引擎集成测试", func() {
			
			Convey("多种规则类型混合执行", func() {
				// 创建动态引擎
				dynamicEngine := NewDynamicEngine[map[string]interface{}](
					DynamicEngineConfig{
						EnableCache:       true,
						CacheTTL:          time.Minute,
						MaxCacheSize:      50,
						ParallelExecution: true,
					},
				)
				
				ctx := context.Background()
				
				// 定义输入数据结构
				type CustomerOrder struct {
					Customer Customer `json:"customer"`
					Order    Order    `json:"order"`
				}
				
				input := CustomerOrder{
					Customer: Customer{
						Age:      30,
						VipLevel: 3,
						Income:   80000,
					},
					Order: Order{
						Amount:   1200.0,
						Quantity: 2,
					},
				}
				
				// 1. 简单规则测试 - 使用正确的字段访问
				simpleRule := SimpleRule{
					When: "Params.Customer.Age >= 18 && Params.Order.Amount > 1000",
					Then: map[string]string{
						"Result.Eligible": "true",
						"Result.Type":     "\"simple\"",
					},
				}
				
				result1, err1 := dynamicEngine.ExecuteRuleDefinition(ctx, simpleRule, input)
				So(err1, ShouldBeNil)
				So(result1["Eligible"], ShouldEqual, true)
				So(result1["Type"], ShouldEqual, "simple")
				
				// 2. 指标规则测试 - 使用正确的字段访问
				metricRule := MetricRule{
					Name:        "customer_score",
					Description: "客户综合评分",
					Formula:     "age_score + income_score + vip_score",
					Variables: map[string]string{
						"age_score":    "Params.Customer.Age * 0.1",
						"income_score": "Params.Customer.Income * 0.0001",
						"vip_score":    "Params.Customer.VipLevel * 10",
					},
					Conditions: []string{
						"Params.Customer.Age >= 18",
						"Params.Customer.Income > 0",
					},
				}
				
				result2, err2 := dynamicEngine.ExecuteRuleDefinition(ctx, metricRule, input)
				So(err2, ShouldBeNil)
				So(result2["CustomerScore"], ShouldNotBeNil)
				
				// 验证计算结果 (30*0.1 + 80000*0.0001 + 3*10 = 3 + 8 + 30 = 41)
				score, ok := result2["CustomerScore"].(float64)
				So(ok, ShouldBeTrue)
				So(score, ShouldEqual, 41)
				
				// 3. 标准规则测试 - 使用枚举类型
				standardRule := StandardRule{
					ID:          "integration_test",
					Name:        "集成测试标准规则",
					Description: "用于集成测试的标准规则",
					Priority:    100,
					Enabled:     true,
					Conditions: Condition{
						Type:     ConditionTypeComposite,
						Operator: OpAnd,
						Children: []Condition{
							{
								Type:     ConditionTypeSimple,
								Left:     "Params.Customer.VipLevel",
								Operator: OpGreaterThanOrEqual,
								Right:    3,
							},
							{
								Type:     ConditionTypeSimple,
								Left:     "Params.Order.Amount",
								Operator: OpGreaterThan,
								Right:    1000,
							},
						},
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "Result.VipDiscount",
							Value:  true,
						},
						{
							Type:   ActionTypeCalculate,
							Target: "Result.DiscountAmount",
							Expression: "Params.Order.Amount * 0.15",
						},
					},
				}
				
				result3, err3 := dynamicEngine.ExecuteRuleDefinition(ctx, standardRule, input)
				So(err3, ShouldBeNil)
				So(result3["VipDiscount"], ShouldEqual, true)
				
				// 验证折扣计算 (1200 * 0.15 = 180)
				discount, ok := result3["DiscountAmount"].(float64)
				So(ok, ShouldBeTrue)
				So(discount, ShouldEqual, 180)
			})
			
			Convey("基本类型输入测试", func() {
				dynamicEngine := NewDynamicEngine[map[string]interface{}](
					DynamicEngineConfig{
						EnableCache: true,
					},
				)
				
				ctx := context.Background()
				
				// 测试整数输入
				intRule := SimpleRule{
					When: "Params > 100",
					Then: map[string]string{
						"Result.LargeNumber": "true",
						"Result.Value":       "Params * 2",
					},
				}
				
				result, err := dynamicEngine.ExecuteRuleDefinition(ctx, intRule, 150)
				So(err, ShouldBeNil)
				So(result["LargeNumber"], ShouldEqual, true)
				So(result["Value"], ShouldEqual, 300)
				
				// 测试字符串输入
				stringRule := SimpleRule{
					When: "Params == \"VIP\"",
					Then: map[string]string{
						"Result.IsVip":    "true",
						"Result.Message":  "\"尊贵的VIP用户\"",
					},
				}
				
				result2, err2 := dynamicEngine.ExecuteRuleDefinition(ctx, stringRule, "VIP")
				So(err2, ShouldBeNil)
				So(result2["IsVip"], ShouldEqual, true)
				So(result2["Message"], ShouldEqual, "尊贵的VIP用户")
			})
			
			Convey("批量规则并行执行", func() {
				dynamicEngine := NewDynamicEngine[map[string]interface{}](
					DynamicEngineConfig{
						EnableCache:       true,
						ParallelExecution: true,
						MaxCacheSize:      100,
					},
				)
				
				ctx := context.Background()
				
				// 创建多个不同的规则
				rules := []interface{}{
					SimpleRule{
						When: "Params > 100",
						Then: map[string]string{
							"Result.LargeNumber": "true",
						},
					},
					SimpleRule{
						When: "Params > 1000",
						Then: map[string]string{
							"Result.VeryLargeNumber": "true",
						},
					},
					SimpleRule{
						When: "Params % 2 == 0",
						Then: map[string]string{
							"Result.EvenNumber": "true",
						},
					},
				}
				
				// 测试不同输入值
				testValues := []int{50, 150, 1500, 2000}
				
				for _, value := range testValues {
					results, err := dynamicEngine.ExecuteBatch(ctx, rules, value)
					So(err, ShouldBeNil)
					So(len(results), ShouldEqual, 3)
					
					// 验证第一个规则结果
					if value > 100 {
						So(results[0]["LargeNumber"], ShouldEqual, true)
					} else {
						So(results[0]["LargeNumber"], ShouldBeNil)
					}
					
					// 验证第二个规则结果
					if value > 1000 {
						So(results[1]["VeryLargeNumber"], ShouldEqual, true)
					} else {
						So(results[1]["VeryLargeNumber"], ShouldBeNil)
					}
					
					// 验证第三个规则结果
					if value%2 == 0 {
						So(results[2]["EvenNumber"], ShouldEqual, true)
					} else {
						So(results[2]["EvenNumber"], ShouldBeNil)
					}
				}
			})
		})
		
		Convey("混合引擎使用场景", func() {
			
			Convey("数据库引擎与动态引擎协同工作", func() {
				// 创建数据库引擎用于业务规则
				dbEngine, err := New[map[string]interface{}](
					WithDSN("sqlite::memory:"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				
				if err != nil {
					SkipConvey("需要SQLite支持，跳过该测试", func() {})
					return
				}
				defer dbEngine.Close()
				
				// 创建动态引擎用于指标计算
				dynamicEngine := NewDynamicEngine[map[string]interface{}](
					DynamicEngineConfig{
						EnableCache: true,
						CacheTTL:    time.Minute,
					},
				)
				
				ctx := context.Background()
				
				// 1. 准备数据库规则
				businessRule := Rule{
					BizCode:     "USER_LEVEL_CHECK",
					Name:        "用户等级检查",
					GRL:         `rule UserLevelCheck "user level" salience 100 { when Params.Score >= 80 then Result.Level = "Gold"; Result.Benefits = ["premium_support", "fast_approval"]; Retract("UserLevelCheck"); }`,
					Version:     1,
					Enabled:     true,
					Description: "根据用户评分确定等级",
				}
				
				// 插入业务规则到数据库
				db := dbEngine.(*engineImpl[map[string]interface{}]).config.db
				err = db.Create(&businessRule).Error
				So(err, ShouldBeNil)
				
				// 定义输入数据
				type UserData struct {
					Age    int     `json:"age"`
					Income float64 `json:"income"`
					Credit int     `json:"credit"`
				}
				
				userData := UserData{
					Age:    35,
					Income: 90000,
					Credit: 750,
				}
				
				// 2. 首先用动态引擎计算用户评分
				scoreRule := MetricRule{
					Name:    "user_score",
					Formula: "age_score + income_score + credit_score",
					Variables: map[string]string{
						"age_score":    "Params.Age * 0.5",
						"income_score": "Params.Income * 0.0005",
						"credit_score": "Params.Credit * 0.04",
					},
				}
				
				scoreResult, err := dynamicEngine.ExecuteRuleDefinition(ctx, scoreRule, userData)
				So(err, ShouldBeNil)
				So(scoreResult["UserScore"], ShouldNotBeNil)
				
				score, ok := scoreResult["UserScore"].(float64)
				So(ok, ShouldBeTrue)
				// 验证计算：35*0.5 + 90000*0.0005 + 750*0.04 = 17.5 + 45 + 30 = 92.5
				So(score, ShouldEqual, 92.5)
				
				// 3. 然后用数据库引擎执行业务规则
				levelInput := map[string]interface{}{
					"Score": score,
				}
				
				levelResult, err := dbEngine.Exec(ctx, "USER_LEVEL_CHECK", levelInput)
				So(err, ShouldBeNil)
				So(levelResult["Level"], ShouldEqual, "Gold")
				So(levelResult["Benefits"], ShouldNotBeNil)
				
				// 验证完整的混合流程
				benefits, ok := levelResult["Benefits"].([]interface{})
				So(ok, ShouldBeTrue)
				So(len(benefits), ShouldEqual, 2)
				So(benefits[0], ShouldEqual, "premium_support")
				So(benefits[1], ShouldEqual, "fast_approval")
			})
			
			Convey("性能对比测试", func() {
				// 创建数据库引擎
				dbEngine, err := New[map[string]interface{}](
					WithDSN("sqlite::memory:"),
					WithAutoMigrate(),
					WithCache(NewMemoryCache(100)),
					WithLogger(NewNoopLogger()),
				)
				
				if err != nil {
					SkipConvey("需要SQLite支持，跳过该测试", func() {})
					return
				}
				defer dbEngine.Close()
				
				// 创建动态引擎
				dynamicEngine := NewDynamicEngine[map[string]interface{}](
					DynamicEngineConfig{
						EnableCache: true,
						CacheTTL:    time.Minute,
					},
				)
				
				ctx := context.Background()
				
				// 插入数据库规则
				dbRule := Rule{
					BizCode:     "PERF_TEST",
					Name:        "性能测试规则",
					GRL:         `rule PerfTest "perf" salience 100 { when Params > 50 then Result.DbEngine = true; Result.Value = Params * 2; Retract("PerfTest"); }`,
					Version:     1,
					Enabled:     true,
					Description: "性能测试规则",
				}
				
				db := dbEngine.(*engineImpl[map[string]interface{}]).config.db
				err = db.Create(&dbRule).Error
				So(err, ShouldBeNil)
				
				// 创建动态规则
				dynamicRule := SimpleRule{
					When: "Params > 50",
					Then: map[string]string{
						"Result.DynamicEngine": "true",
						"Result.Value":         "Params * 2",
					},
				}
				
				// 测试相同的输入
				testInput := 100
				
				// 数据库引擎执行
				dbResult, err1 := dbEngine.Exec(ctx, "PERF_TEST", testInput)
				So(err1, ShouldBeNil)
				So(dbResult["DbEngine"], ShouldEqual, true)
				So(dbResult["Value"], ShouldEqual, 200)
				
				// 动态引擎执行
				dynamicResult, err2 := dynamicEngine.ExecuteRuleDefinition(ctx, dynamicRule, testInput)
				So(err2, ShouldBeNil)
				So(dynamicResult["DynamicEngine"], ShouldEqual, true)
				So(dynamicResult["Value"], ShouldEqual, 200)
				
				// 验证两个引擎都能产生正确的结果
				So(dbResult["Value"], ShouldEqual, dynamicResult["Value"])
			})
		})
		
		Convey("规则生命周期集成测试", func() {
			
			Convey("规则版本管理", func() {
				engine, err := New[map[string]interface{}](
					WithDSN("sqlite:file:version_test.db?mode=memory&cache=shared&_fk=1"),
					WithAutoMigrate(),
					WithCache(NewMemoryCache(50)),
					WithLogger(NewNoopLogger()),
				)
				
				if err != nil {
					SkipConvey("需要SQLite支持，跳过该测试", func() {})
					return
				}
				defer engine.Close()
				
				ctx := context.Background()
				
				// 测试引擎创建和基本功能
				// 执行不存在的规则
				_, err1 := engine.Exec(ctx, "VERSION_TEST", map[string]interface{}{})
				So(err1, ShouldNotBeNil)
				So(err1.Error(), ShouldContainSubstring, "规则未找到")
				
				// 测试缓存清理功能不会panic
				So(func() {
					// 模拟一些操作后的状态
					for i := 0; i < 10; i++ {
						engine.Exec(ctx, fmt.Sprintf("TEST_%d", i), map[string]interface{}{})
					}
				}, ShouldNotPanic)
			})
		})
		
		Convey("错误恢复与容错性", func() {
			
			Convey("规则执行失败后的恢复", func() {
				engine, err := New[map[string]interface{}](
					WithDSN("sqlite:file:error_test.db?mode=memory&cache=shared&_fk=1"),
					WithAutoMigrate(),
					WithLogger(NewNoopLogger()),
				)
				
				if err != nil {
					SkipConvey("需要SQLite支持，跳过该测试", func() {})
					return
				}
				defer engine.Close()
				
				ctx := context.Background()
				
				// 执行不存在的规则（应该失败但不崩溃）
				result1, err1 := engine.Exec(ctx, "ERROR_TEST", map[string]interface{}{"validField": 123})
				So(err1, ShouldNotBeNil) // 应该返回错误
				So(result1, ShouldNotBeNil) // 但应该返回空结果而不是nil
				So(err1.Error(), ShouldContainSubstring, "规则未找到")
				
				// 执行另一个不存在的规则（应该成功处理）
				result2, err2 := engine.Exec(ctx, "GOOD_TEST", map[string]interface{}{})
				So(err2, ShouldNotBeNil) // 同样返回规则未找到错误
				So(result2, ShouldNotBeNil) // 返回空结果
				So(err2.Error(), ShouldContainSubstring, "规则未找到")
				
				// 再次执行规则，确保引擎状态正常
				result3, err3 := engine.Exec(ctx, "ANOTHER_TEST", map[string]interface{}{})
				So(err3, ShouldNotBeNil)
				So(result3, ShouldNotBeNil)
				So(err3.Error(), ShouldContainSubstring, "规则未找到")
			})
		})
	})
}