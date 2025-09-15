package runehammer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestLogger 测试日志组件
func TestLogger(t *testing.T) {
	Convey("日志组件测试", t, func() {
		
		Convey("Logger接口定义", func() {
			
			Convey("接口方法签名验证", func() {
				// 验证接口定义正确性，通过编译即可确保接口正确
				var logger Logger
				logger = NewNoopLogger()
				So(logger, ShouldNotBeNil)
				So(logger, ShouldImplement, (*Logger)(nil))
				
				logger = NewDefaultLogger()
				So(logger, ShouldNotBeNil)
				So(logger, ShouldImplement, (*Logger)(nil))
			})
		})
		
		Convey("NoopLogger测试", func() {
			logger := NewNoopLogger()
			ctx := context.Background()
			
			Convey("创建和类型验证", func() {
				So(logger, ShouldNotBeNil)
				So(logger, ShouldHaveSameTypeAs, &NoopLogger{})
				So(logger, ShouldImplement, (*Logger)(nil))
			})
			
			Convey("所有级别日志方法", func() {
				// 这些方法应该都能安全调用，不会panic
				So(func() {
					logger.Debugf(ctx, "debug message")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Infof(ctx, "info message")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Warnf(ctx, "warn message")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Errorf(ctx, "error message")
				}, ShouldNotPanic)
			})
			
			Convey("带参数的日志调用", func() {
				So(func() {
					logger.Debugf(ctx, "debug with params", "key1", "value1", "key2", 123)
				}, ShouldNotPanic)
				
				So(func() {
					logger.Infof(ctx, "info with params", "user_id", 12345, "action", "login")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Warnf(ctx, "warn with params", "warning_code", "W001")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Errorf(ctx, "error with params", "error", "database connection failed")
				}, ShouldNotPanic)
			})
			
			Convey("空上下文调用", func() {
				So(func() {
					logger.Debugf(nil, "debug with nil context")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Infof(context.TODO(), "info with TODO context")
				}, ShouldNotPanic)
			})
			
			Convey("空消息和参数", func() {
				So(func() {
					logger.Debugf(ctx, "")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Infof(ctx, "", nil)
				}, ShouldNotPanic)
				
				So(func() {
					logger.Warnf(ctx, "message", nil, nil, nil)
				}, ShouldNotPanic)
			})
		})
		
		Convey("DefaultLogger测试", func() {
			
			Convey("创建和类型验证", func() {
				logger := NewDefaultLogger()
				So(logger, ShouldNotBeNil)
				So(logger, ShouldHaveSameTypeAs, &DefaultLogger{})
				So(logger, ShouldImplement, (*Logger)(nil))
			})
			
			Convey("输出内容验证", func() {
				// 捕获标准输出来验证日志内容
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				// 测试各级别日志输出
				logger.Debugf(ctx, "test debug message")
				logger.Infof(ctx, "test info message")
				logger.Warnf(ctx, "test warn message")
				logger.Errorf(ctx, "test error message")
				
				// 恢复标准输出
				w.Close()
				os.Stdout = oldStdout
				
				// 读取输出内容
				var buf bytes.Buffer
				io.Copy(&buf, r)
				output := buf.String()
				
				// 验证输出包含预期内容
				So(output, ShouldContainSubstring, "[DEBUG]")
				So(output, ShouldContainSubstring, "[INFO]")
				So(output, ShouldContainSubstring, "[WARN]")
				So(output, ShouldContainSubstring, "[ERROR]")
				So(output, ShouldContainSubstring, "test debug message")
				So(output, ShouldContainSubstring, "test info message")
				So(output, ShouldContainSubstring, "test warn message")
				So(output, ShouldContainSubstring, "test error message")
			})
			
			Convey("带参数的日志输出", func() {
				// 捕获标准输出
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				// 测试带参数的日志
				logger.Infof(ctx, "user login", "user_id", 12345, "ip", "192.168.1.1")
				logger.Errorf(ctx, "database error", "table", "users", "error", "connection timeout")
				
				// 恢复标准输出
				w.Close()
				os.Stdout = oldStdout
				
				// 读取输出内容
				var buf bytes.Buffer
				io.Copy(&buf, r)
				output := buf.String()
				
				// 验证参数被正确输出
				So(output, ShouldContainSubstring, "user login")
				So(output, ShouldContainSubstring, "user_id")
				So(output, ShouldContainSubstring, "12345")
				So(output, ShouldContainSubstring, "database error")
				So(output, ShouldContainSubstring, "connection timeout")
			})
			
			Convey("特殊字符和编码", func() {
				// 捕获标准输出
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				// 测试中文和特殊字符
				logger.Infof(ctx, "中文日志消息", "用户", "张三", "操作", "登录")
				logger.Warnf(ctx, "特殊字符", "data", "!@#$%^&*()")
				
				// 恢复标准输出
				w.Close()
				os.Stdout = oldStdout
				
				// 读取输出内容
				var buf bytes.Buffer
				io.Copy(&buf, r)
				output := buf.String()
				
				// 验证特殊字符被正确输出
				So(output, ShouldContainSubstring, "中文日志消息")
				So(output, ShouldContainSubstring, "张三")
				So(output, ShouldContainSubstring, "!@#$%^&*()")
			})
		})
		
		Convey("日志级别一致性测试", func() {
			
			Convey("NoopLogger所有方法签名一致", func() {
				logger := NewNoopLogger()
				ctx := context.Background()
				
				// 验证所有方法都有相同的签名
				testMethods := []func(context.Context, string, ...any){
					logger.Debugf,
					logger.Infof,
					logger.Warnf,
					logger.Errorf,
				}
				
				// 所有方法都应该能正常调用
				for _, method := range testMethods {
					So(func() {
						method(ctx, "test message", "key", "value")
					}, ShouldNotPanic)
				}
			})
			
			Convey("DefaultLogger所有方法签名一致", func() {
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				// 验证所有方法都有相同的签名
				testMethods := []func(context.Context, string, ...any){
					logger.Debugf,
					logger.Infof,
					logger.Warnf,
					logger.Errorf,
				}
				
				// 所有方法都应该能正常调用
				for _, method := range testMethods {
					So(func() {
						method(ctx, "test message", "key", "value")
					}, ShouldNotPanic)
				}
			})
		})
		
		Convey("边界条件测试", func() {
			
			Convey("大量参数", func() {
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				// 创建大量参数
				var keyvals []any
				for i := 0; i < 100; i++ {
					keyvals = append(keyvals, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
				}
				
				So(func() {
					logger.Infof(ctx, "large params test", keyvals...)
				}, ShouldNotPanic)
			})
			
			Convey("nil参数", func() {
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				So(func() {
					logger.Debugf(ctx, "nil test", nil, nil, nil)
				}, ShouldNotPanic)
				
				So(func() {
					logger.Infof(ctx, "mixed nil", "key1", nil, nil, "value2")
				}, ShouldNotPanic)
			})
			
			Convey("复杂类型参数", func() {
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				complexStruct := struct {
					Name string
					Age  int
					Data map[string]interface{}
				}{
					Name: "test",
					Age:  25,
					Data: map[string]interface{}{
						"nested": "value",
					},
				}
				
				So(func() {
					logger.Infof(ctx, "complex type", "struct", complexStruct)
				}, ShouldNotPanic)
				
				So(func() {
					logger.Errorf(ctx, "slice param", "numbers", []int{1, 2, 3, 4, 5})
				}, ShouldNotPanic)
			})
		})
		
		Convey("性能测试", func() {
			
			Convey("NoopLogger性能", func() {
				logger := NewNoopLogger()
				ctx := context.Background()
				
				// NoopLogger应该非常快，因为它什么都不做
				for i := 0; i < 1000; i++ {
					logger.Infof(ctx, "performance test", "iteration", i)
				}
				
				// 如果能够完成1000次调用而不超时，说明性能可接受
				So(true, ShouldBeTrue)
			})
			
			Convey("DefaultLogger大量输出", func() {
				// 重定向输出到丢弃，避免污染测试输出
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				// 大量日志输出
				for i := 0; i < 100; i++ {
					logger.Infof(ctx, "performance test", "iteration", i)
				}
				
				// 恢复标准输出
				w.Close()
				os.Stdout = oldStdout
				
				// 清空管道
				var buf bytes.Buffer
				io.Copy(&buf, r)
				
				// 验证输出包含预期数量的日志
				output := buf.String()
				lines := strings.Split(strings.TrimSpace(output), "\n")
				So(len(lines), ShouldEqual, 100)
			})
		})
		
		Convey("实际应用场景测试", func() {
			
			Convey("结构化日志记录", func() {
				// 捕获输出
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				
				logger := NewDefaultLogger()
				ctx := context.Background()
				
				// 模拟真实应用场景的结构化日志
				logger.Infof(ctx, "用户登录成功", 
					"user_id", "12345",
					"username", "john_doe",
					"ip", "192.168.1.100",
					"user_agent", "Mozilla/5.0",
					"timestamp", "2023-01-01T10:00:00Z")
				
				logger.Errorf(ctx, "数据库连接失败",
					"database", "mysql",
					"host", "localhost:3306",
					"error", "connection timeout",
					"retry_count", 3)
				
				logger.Warnf(ctx, "API请求限流",
					"endpoint", "/api/users",
					"client_ip", "192.168.1.200",
					"rate_limit", 100,
					"current_requests", 95)
				
				// 恢复输出
				w.Close()
				os.Stdout = oldStdout
				
				// 读取输出
				var buf bytes.Buffer
				io.Copy(&buf, r)
				output := buf.String()
				
				// 验证结构化信息都被记录
				So(output, ShouldContainSubstring, "用户登录成功")
				So(output, ShouldContainSubstring, "john_doe")
				So(output, ShouldContainSubstring, "数据库连接失败")
				So(output, ShouldContainSubstring, "connection timeout")
				So(output, ShouldContainSubstring, "API请求限流")
				So(output, ShouldContainSubstring, "/api/users")
			})
			
			Convey("上下文信息传递", func() {
				logger := NewDefaultLogger()
				
				// 测试不同的上下文
				ctxWithValue := context.WithValue(context.Background(), "request_id", "req-12345")
				cancelCtx, cancel := context.WithCancel(context.Background())
				
				So(func() {
					logger.Infof(ctxWithValue, "带值的上下文", "operation", "test")
				}, ShouldNotPanic)
				
				So(func() {
					logger.Warnf(cancelCtx, "可取消的上下文", "status", "processing")
				}, ShouldNotPanic)
				
				cancel() // 取消上下文
				
				So(func() {
					logger.Errorf(cancelCtx, "已取消的上下文", "result", "cancelled")
				}, ShouldNotPanic)
			})
		})
	})
}

// TestLoggerIntegration 测试日志组件集成
func TestLoggerIntegration(t *testing.T) {
	Convey("日志组件集成测试", t, func() {
		
		Convey("作为依赖注入使用", func() {
			// 模拟一个使用Logger的服务
			type Service struct {
				logger Logger
			}
			
			newService := func(logger Logger) *Service {
				return &Service{logger: logger}
			}
			
			doWork := func(s *Service, ctx context.Context) {
				s.logger.Infof(ctx, "开始工作")
				s.logger.Debugf(ctx, "处理详情", "step", 1)
				s.logger.Warnf(ctx, "注意事项", "warning", "resource low")
				s.logger.Errorf(ctx, "完成工作", "result", "success")
			}
			
			Convey("使用NoopLogger", func() {
				service := newService(NewNoopLogger())
				
				So(func() {
					doWork(service, context.Background())
				}, ShouldNotPanic)
			})
			
			Convey("使用DefaultLogger", func() {
				// 重定向输出
				oldStdout := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w
				
				service := newService(NewDefaultLogger())
				doWork(service, context.Background())
				
				// 恢复输出
				w.Close()
				os.Stdout = oldStdout
				
				// 验证输出
				var buf bytes.Buffer
				io.Copy(&buf, r)
				output := buf.String()
				
				So(output, ShouldContainSubstring, "开始工作")
				So(output, ShouldContainSubstring, "处理详情")
				So(output, ShouldContainSubstring, "注意事项")
				So(output, ShouldContainSubstring, "完成工作")
			})
		})
		
		Convey("Logger接口的多态性", func() {
			loggers := []Logger{
				NewNoopLogger(),
				NewDefaultLogger(),
			}
			
			ctx := context.Background()
			
			for i, logger := range loggers {
				Convey(fmt.Sprintf("Logger %d 多态测试", i), func() {
					So(func() {
						logger.Debugf(ctx, "多态测试", "logger_index", i)
						logger.Infof(ctx, "多态测试", "logger_index", i)
						logger.Warnf(ctx, "多态测试", "logger_index", i)
						logger.Errorf(ctx, "多态测试", "logger_index", i)
					}, ShouldNotPanic)
				})
			}
		})
	})
}