package runehammer

import (
	"context"
	"testing"
	"time"

	logger "gitee.com/damengde/runehammer/logger"
	. "github.com/smartystreets/goconvey/convey"
)

// 测试用结构体定义
type TestCustomer struct {
	Age          int     `json:"age"`
	Name         string  `json:"name"`
	VipLevel     int     `json:"vip_level"`
	Income       float64 `json:"income"`
	PurchaseCount int    `json:"purchase_count"`
}

type TestOrder struct {
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
}

type TestInput struct {
	Customer TestCustomer `json:"customer"`
	Order    TestOrder    `json:"order"`
}

// testValidator 实现 RuleValidator 接口的测试验证器
type testValidator struct{}

func (tv *testValidator) Validate(definition interface{}) []ValidationError {
	if sr, ok := definition.(SimpleRule); ok && sr.When == "" {
		return []ValidationError{
			{
				Message: "规则条件不能为空",
				Field:   "When",
			},
		}
	}
	return nil
}

// TestDynamicEngine 测试动态引擎 - 使用结构体类型
func TestDynamicEngine(t *testing.T) {
	Convey("动态规则引擎测试", t, func() {
		// 创建动态引擎 - 使用结构体类型
		engine := NewDynamicEngine[map[string]interface{}](
			DynamicEngineConfig{
				EnableCache:       true,
				CacheTTL:          5 * time.Minute,
				MaxCacheSize:      100,
				StrictValidation:  true,
				ParallelExecution: true,
				DefaultTimeout:    10 * time.Second,
			},
		)

		Convey("执行简单规则", func() {
			// 定义简单规则 - 使用结构体字段访问
			simpleRule := SimpleRule{
				When: "testinput.Customer.Age >= 18",
				Then: map[string]string{
					"Result.Eligible": "true",
					"Result.Message":  "\"符合条件\"",
				},
			}

			// 输入数据 - 使用结构体
			input := TestInput{
				Customer: TestCustomer{
					Age:  25,
					Name: "张三",
				},
			}

			// 执行规则
			result, err := engine.ExecuteRuleDefinition(context.Background(), simpleRule, input)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)

			// 验证结果
			So(result["Eligible"], ShouldEqual, true)
			So(result["Message"], ShouldEqual, "符合条件")
		})

		Convey("执行指标规则", func() {
			metricRule := MetricRule{
				Name:        "customer_score",
				Description: "客户评分计算",
				Formula:     "age_score + income_score",
				Variables: map[string]string{
					"age_score":    "testinput.Customer.Age * 0.1",
					"income_score": "testinput.Customer.Income * 0.0001",
				},
				Conditions: []string{
					"testinput.Customer.Age >= 18",
					"testinput.Customer.Income > 0",
				},
			}

			input := TestInput{
				Customer: TestCustomer{
					Age:    30,
					Income: 50000,
				},
			}

			result, err := engine.ExecuteRuleDefinition(context.Background(), metricRule, input)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result["CustomerScore"], ShouldNotBeNil)
		})

		Convey("执行标准规则", func() {
			standardRule := StandardRule{
				ID:          "vip_check",
				Name:        "VIP客户检查",
				Description: "检查客户是否为VIP",
				Priority:    100,
				Enabled:     true,
				Tags:        []string{"vip", "customer"},
				Conditions: Condition{
					Type:     "simple",
					Left:     "testinput.Customer.VipLevel",
					Operator: ">=",
					Right:    3,
				},
				Actions: []Action{
					{
						Type:   "assign",
						Target: "Result.IsVip",
						Value:  true,
					},
					{
						Type:   "assign",
						Target: "Result.VipBenefits",
						Value:  []string{"专属客服", "优先放款"},
					},
				},
			}

			input := TestInput{
				Customer: TestCustomer{
					VipLevel: 4,
				},
			}

			result, err := engine.ExecuteRuleDefinition(context.Background(), standardRule, input)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result["IsVip"], ShouldEqual, true)
		})

		Convey("批量执行规则", func() {
			rules := []interface{}{
				SimpleRule{
					When: "testinput.Order.Amount > 500",
					Then: map[string]string{
						"Result.FreeShipping": "true",
					},
				},
				MetricRule{
					Name:    "loyalty_score",
					Formula: "purchase_count * 10",
					Variables: map[string]string{
						"purchase_count": "testinput.Customer.PurchaseCount",
					},
				},
			}

			input := TestInput{
				Customer: TestCustomer{
					PurchaseCount: 5,
				},
				Order: TestOrder{
					Amount: 600.0,
				},
			}

			results, err := engine.ExecuteBatch(context.Background(), rules, input)
			So(err, ShouldBeNil)
			So(len(results), ShouldEqual, 2)
			So(results[0]["FreeShipping"], ShouldEqual, true)
			So(results[1]["LoyaltyScore"], ShouldNotBeNil)
		})

		Convey("自定义函数注册", func() {
			// 注册自定义函数
			engine.RegisterCustomFunction("CalculateDiscount", func(amount float64, rate float64) float64 {
				return amount * rate
			})

			engine.RegisterCustomFunctions(map[string]interface{}{
				"ValidateAge": func(age int) bool {
					return age >= 18 && age <= 120
				},
			})

			// 使用自定义函数的规则
			rule := SimpleRule{
				When: "ValidateAge(testinput.Customer.Age)",
				Then: map[string]string{
					"Result.Discount": "CalculateDiscount(testinput.Order.Amount, 0.1)",
				},
			}

			input := TestInput{
				Customer: TestCustomer{
					Age: 25,
				},
				Order: TestOrder{
					Amount: 100.0,
				},
			}

			result, err := engine.ExecuteRuleDefinition(context.Background(), rule, input)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result["Discount"], ShouldEqual, 10.0)
		})

		Convey("错误处理", func() {
			Convey("无效规则", func() {
				invalidRule := SimpleRule{
					When: "invalid syntax here",
					Then: map[string]string{
						"Result.Test": "true",
					},
				}

				input := TestInput{
					Customer: TestCustomer{Age: 25},
				}

				_, err := engine.ExecuteRuleDefinition(context.Background(), invalidRule, input)
				So(err, ShouldNotBeNil)
			})

			Convey("空规则", func() {
				emptyRule := SimpleRule{}
				input := TestInput{Customer: TestCustomer{Age: 25}}

				_, err := engine.ExecuteRuleDefinition(context.Background(), emptyRule, input)
				So(err, ShouldNotBeNil)
			})

			Convey("Map类型输入应该失败", func() {
				rule := SimpleRule{
					When: "true",
					Then: map[string]string{
						"Result.Test": "true",
					},
				}

				// 使用 map 类型输入，应该失败
				mapInput := map[string]interface{}{
					"test": "data",
				}

				_, err := engine.ExecuteRuleDefinition(context.Background(), rule, mapInput)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "不支持 map 类型")
			})
		})

		Convey("引擎管理功能", func() {
			Convey("超时执行", func() {
				// 创建一个会超时的规则
				timeoutRule := SimpleRule{
					When: "testinput.Customer.Age >= 0", // 简单条件
					Then: map[string]string{
						"Result.Processed": "true",
					},
				}

				input := TestInput{
					Customer: TestCustomer{Age: 25},
				}

				// 使用很短的超时时间
				_, err := engine.ExecuteWithTimeout(context.Background(), timeoutRule, input, 1*time.Nanosecond)
				So(err, ShouldNotBeNil)
			})

			Convey("注册验证器", func() {
				// 创建一个实现 RuleValidator 接口的验证器
				validator := &testValidator{}
				engine.RegisterValidator(validator)

				// 测试验证器生效
				invalidRule := SimpleRule{
					When: "", // 空条件应该触发验证错误
					Then: map[string]string{"Result.Test": "true"},
				}

				input := TestInput{Customer: TestCustomer{Age: 25}}
				_, err := engine.ExecuteRuleDefinition(context.Background(), invalidRule, input)
				So(err, ShouldNotBeNil)
			})

			Convey("设置日志器", func() {
				logger := logger.NewNoopLogger()
				engine.SetLogger(logger)
				// 验证日志器设置成功（通过后续操作不出错来验证）
				So(func() { engine.SetLogger(logger) }, ShouldNotPanic)
			})

			Convey("缓存管理", func() {
				// 先执行一个规则来填充缓存
				rule := SimpleRule{
					When: "testinput.Customer.Age >= 18",
					Then: map[string]string{"Result.Cached": "true"},
				}
				input := TestInput{Customer: TestCustomer{Age: 25}}
				
				_, err := engine.ExecuteRuleDefinition(context.Background(), rule, input)
				So(err, ShouldBeNil)

				// 获取缓存统计
				stats := engine.GetCacheStats()
				So(stats, ShouldNotBeNil)
				
				// 清理缓存
				engine.ClearCache()
				
				// 再次获取统计，应该显示缓存已清空
				statsAfterClear := engine.GetCacheStats()
				So(statsAfterClear, ShouldNotBeNil)
			})
		})

		Convey("批量执行模式", func() {
			Convey("顺序执行批量规则", func() {
				// 创建不支持并行的引擎
				seqEngine := NewDynamicEngine[map[string]interface{}](
					DynamicEngineConfig{
						EnableCache:       true,
						ParallelExecution: false, // 关闭并行执行
						DefaultTimeout:    10 * time.Second,
					},
				)

				rules := []interface{}{
					SimpleRule{
						When: "testinput.Customer.Age >= 18",
						Then: map[string]string{"Result.Adult": "true"},
					},
					SimpleRule{
						When: "testinput.Order.Amount > 100",
						Then: map[string]string{"Result.HighValue": "true"},
					},
				}

				input := TestInput{
					Customer: TestCustomer{Age: 25},
					Order:    TestOrder{Amount: 150.0},
				}

				results, err := seqEngine.ExecuteBatch(context.Background(), rules, input)
				So(err, ShouldBeNil)
				So(len(results), ShouldEqual, 2)
				So(results[0]["Adult"], ShouldEqual, true)
				So(results[1]["HighValue"], ShouldEqual, true)
			})
		})

		Convey("数据注入测试", func() {
			Convey("非结构体类型数据注入", func() {
				rule := SimpleRule{
					When: "Params > 100", // 对于非结构体类型，使用 Params 访问
					Then: map[string]string{"Result.Large": "true"},
				}

				// 使用基本类型
				result, err := engine.ExecuteRuleDefinition(context.Background(), rule, 150)
				So(err, ShouldBeNil)
				So(result["Large"], ShouldEqual, true)
			})

			Convey("字符串类型数据注入", func() {
				rule := SimpleRule{
					When: "Params == \"test\"",
					Then: map[string]string{"Result.Matched": "true"},
				}

				result, err := engine.ExecuteRuleDefinition(context.Background(), rule, "test")
				So(err, ShouldBeNil)
				So(result["Matched"], ShouldEqual, true)
			})
		})
	})
}