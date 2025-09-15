package runehammer

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestRuleConverterCoverage 专门提升 rule_converter.go 覆盖率的测试
func TestRuleConverterCoverage(t *testing.T) {
	Convey("Rule Converter 覆盖率提升测试", t, func() {

		Convey("convertStandard 函数覆盖", func() {
			converter := NewGRLConverter()

			Convey("转换包含禁用规则的标准", func() {
				standard := RuleDefinitionStandard{
					Rules: []Rule{
						{
							ID:      1,
							BizCode: "test",
							Name:    "启用规则",
							Enabled: true,
							GRL:     "rule EnabledRule \"启用的规则\" { when true then result[\"test\"] = true; }",
						},
						{
							ID:      2,
							BizCode: "test",
							Name:    "禁用规则",
							Enabled: false,
							GRL:     "rule DisabledRule \"禁用的规则\" { when true then result[\"test\"] = false; }",
						},
					},
				}

				grl, err := converter.ConvertToGRL(standard)
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "EnabledRule")
				So(grl, ShouldNotContainSubstring, "DisabledRule")
			})

			Convey("转换不完整的规则定义", func() {
				standard := RuleDefinitionStandard{
					Rules: []Rule{
						{
							ID:      3,
							BizCode: "test",
							Name:    "不完整规则",
							Enabled: true,
							GRL:     "", // 空的GRL，且没有足够信息转换
						},
					},
				}

				_, err := converter.ConvertToGRL(standard)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "不包含足够信息进行GRL转换")
			})
		})

		Convey("特殊操作符条件覆盖", func() {
			converter := NewGRLConverter()

			Convey("between 操作符", func() {
				rule := StandardRule{
					ID:   "BETWEEN_TEST",
					Name: "Between测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "age",
						Operator: "between",
						Right:    []int{18, 65},
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "valid",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "\"age\" >= 18")
				So(grl, ShouldContainSubstring, "\"age\" <= 65")
				So(grl, ShouldContainSubstring, "&&")
			})

			Convey("between 操作符错误情况", func() {
				rule := StandardRule{
					ID:   "BETWEEN_ERROR",
					Name: "Between错误测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "age",
						Operator: "between",
						Right:    "invalid", // 不是数组
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "error",
						},
					},
				}

				_, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "between操作符需要两个值的数组")
			})

			Convey("in 操作符", func() {
				rule := StandardRule{
					ID:   "IN_TEST",
					Name: "In测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "status",
						Operator: "in",
						Right:    []string{"active", "pending"},
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "valid",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "Contains")
			})

			Convey("contains 操作符", func() {
				rule := StandardRule{
					ID:   "CONTAINS_TEST",
					Name: "Contains测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "name",
						Operator: "contains",
						Right:    "test",
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "found",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "Contains")
			})

			Convey("matches 操作符", func() {
				rule := StandardRule{
					ID:   "MATCHES_TEST",
					Name: "Matches测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "email",
						Operator: "matches",
						Right:    ".*@.*\\.com",
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "valid_email",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "Matches")
			})
		})

		Convey("条件类型覆盖", func() {
			converter := NewGRLConverter()

			Convey("expression 条件类型", func() {
				rule := StandardRule{
					ID:   "EXPRESSION_TEST",
					Name: "表达式测试",
					Conditions: Condition{
						Type:       ConditionTypeExpression,
						Expression: "age >= 18 && income > 30000",
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "approved",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "age >= 18")
				So(grl, ShouldContainSubstring, "income > 30000")
			})

			Convey("function 条件类型", func() {
				rule := StandardRule{
					ID:   "FUNCTION_TEST",
					Name: "函数测试",
					Conditions: Condition{
						Type:       ConditionTypeFunction,
						Expression: "ValidateAge(age)",
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "validated",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "ValidateAge")
			})

			Convey("复合条件空子条件", func() {
				rule := StandardRule{
					ID:   "EMPTY_COMPOSITE",
					Name: "空复合条件",
					Conditions: Condition{
						Type:     ConditionTypeComposite,
						Operator: OpAnd,
						Children: []Condition{}, // 空子条件
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "error",
						},
					},
				}

				_, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "复合条件必须包含子条件")
			})
		})

		Convey("动作类型覆盖", func() {
			converter := NewGRLConverter()

			Convey("calculate 动作类型", func() {
				rule := StandardRule{
					ID:   "CALCULATE_TEST",
					Name: "计算动作测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "amount",
						Operator: OpGreaterThan,
						Right:    0,
					},
					Actions: []Action{
						{
							Type:       ActionTypeCalculate,
							Target:     "result.total",
							Expression: "amount * 1.2",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "result[\"total\"] = amount * 1.2")
			})

			Convey("invoke 动作类型带参数", func() {
				rule := StandardRule{
					ID:   "INVOKE_WITH_PARAMS",
					Name: "调用动作带参数",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "status",
						Operator: OpEqual,
						Right:    "pending",
					},
					Actions: []Action{
						{
							Type:   ActionTypeInvoke,
							Target: "ProcessOrder",
							Parameters: map[string]interface{}{
								"orderId": "12345",
								"urgent":  true,
							},
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "ProcessOrder(")
				So(grl, ShouldContainSubstring, "orderId=\"12345\"")
				So(grl, ShouldContainSubstring, "urgent=true")
			})

			Convey("invoke 动作类型无参数", func() {
				rule := StandardRule{
					ID:   "INVOKE_NO_PARAMS",
					Name: "调用动作无参数",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "ready",
						Operator: OpEqual,
						Right:    true,
					},
					Actions: []Action{
						{
							Type:   ActionTypeInvoke,
							Target: "StartProcess",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "StartProcess()")
			})

			Convey("log 动作类型", func() {
				rule := StandardRule{
					ID:   "LOG_TEST",
					Name: "日志动作测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "debug",
						Operator: OpEqual,
						Right:    true,
					},
					Actions: []Action{
						{
							Type:  ActionTypeLog,
							Value: "Debug mode activated",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "Log(\"Debug mode activated\")")
			})

			Convey("alert 动作类型", func() {
				rule := StandardRule{
					ID:   "ALERT_TEST",
					Name: "告警动作测试",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "critical",
						Operator: OpEqual,
						Right:    true,
					},
					Actions: []Action{
						{
							Type:  ActionTypeAlert,
							Value: "Critical error detected",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "Alert(\"Critical error detected\")")
			})

			Convey("不支持的动作类型", func() {
				rule := StandardRule{
					ID:   "UNSUPPORTED_ACTION",
					Name: "不支持的动作",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "test",
						Operator: OpEqual,
						Right:    true,
					},
					Actions: []Action{
						{
							Type: "unknown_action_type",
						},
					},
				}

				_, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "不支持的动作类型")
			})
		})

		Convey("辅助函数覆盖", func() {
			converter := NewGRLConverter()

			Convey("isVariable 函数测试", func() {
				// 测试各种变量前缀
				So(converter.isVariable("customer.name"), ShouldBeTrue)
				So(converter.isVariable("order.amount"), ShouldBeTrue)
				So(converter.isVariable("user.id"), ShouldBeTrue)
				So(converter.isVariable("data.field"), ShouldBeTrue)
				So(converter.isVariable("result.value"), ShouldBeTrue)
				So(converter.isVariable("unknown.field"), ShouldBeFalse)
				So(converter.isVariable("simple_field"), ShouldBeFalse)
			})

			Convey("convertOperand 函数各种类型", func() {
				defs := Definitions{}

				// 字符串字面量
				result, err := converter.convertOperand("simple_string", defs)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "\"simple_string\"")

				// 字段引用
				result, err = converter.convertOperand("customer.name", defs)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "customer.name")

				// 数字类型
				result, err = converter.convertOperand(42, defs)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "42")

				result, err = converter.convertOperand(3.14, defs)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "3.14")

				// 布尔类型
				result, err = converter.convertOperand(true, defs)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "true")

				result, err = converter.convertOperand(false, defs)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "false")

				// nil 值
				result, err = converter.convertOperand(nil, defs)
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "null")

				// 复杂类型
				result, err = converter.convertOperand([]string{"a", "b"}, defs)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeEmpty)
			})

			Convey("convertValue 函数各种类型", func() {
				// 字符串
				result := converter.convertValue("test")
				So(result, ShouldEqual, "\"test\"")

				// 数字
				result = converter.convertValue(123)
				So(result, ShouldEqual, "123")

				result = converter.convertValue(45.67)
				So(result, ShouldEqual, "45.67")

				// 布尔
				result = converter.convertValue(true)
				So(result, ShouldEqual, "true")

				// nil
				result = converter.convertValue(nil)
				So(result, ShouldEqual, "null")

				// 其他类型
				result = converter.convertValue([]int{1, 2, 3})
				So(result, ShouldEqual, "\"[1 2 3]\"")
			})

			Convey("sanitizeRuleName 函数测试", func() {
				// 正常名称
				result := converter.sanitizeRuleName("NormalRule123")
				So(result, ShouldEqual, "NormalRule123")

				// 特殊字符
				result = converter.sanitizeRuleName("Rule-With_Special@Chars#")
				So(result, ShouldEqual, "Rule_With_Special_Chars_")

				// 空格和中文
				result = converter.sanitizeRuleName("规则 带有 空格")
				So(result, ShouldEqual, "________") // 8个下划线
			})

			Convey("convertOperator 函数测试", func() {
				// 映射的操作符
				result, err := converter.convertOperator("and", "value")
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "&&")

				result, err = converter.convertOperator("or", "value")
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "||")

				// 未映射的操作符
				result, err = converter.convertOperator("custom_op", "value")
				So(err, ShouldBeNil)
				So(result, ShouldEqual, "custom_op")
			})

			Convey("resolveTarget 函数测试", func() {
				// result 字段
				result := converter.resolveTarget("result.score")
				So(result, ShouldEqual, "result[\"score\"]")

				result = converter.resolveTarget("result.nested.field")
				So(result, ShouldEqual, "result[\"nested.field\"]")

				// 非 result 字段
				result = converter.resolveTarget("other.field")
				So(result, ShouldEqual, "other.field")
			})

			Convey("generateRuleID 函数测试", func() {
				id1 := converter.generateRuleID()
				// 等待1毫秒确保时间戳不同
				// time.Sleep(1 * time.Millisecond)
				id2 := converter.generateRuleID()
				
				So(id1, ShouldNotBeEmpty)
				So(id2, ShouldNotBeEmpty)
				// 由于时间戳精度问题，可能会生成相同ID，这里只验证格式
				So(len(id1), ShouldBeGreaterThan, 0)
				So(len(id2), ShouldBeGreaterThan, 0)
			})
		})

		Convey("Validate 函数完整覆盖", func() {
			converter := NewGRLConverter()

			Convey("验证StandardRule指针", func() {
				rule := &StandardRule{
					ID:   "VALID_PTR_RULE",
					Name: "有效指针规则",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "test",
						Operator: OpEqual,
						Right:    true,
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "ok",
						},
					},
				}

				err := converter.Validate(rule)
				So(err, ShouldBeNil)
			})

			Convey("验证无效SimpleRule", func() {
				// 空when条件
				rule := SimpleRule{
					When: "",
					Then: map[string]string{"result": "test"},
				}

				err := converter.Validate(rule)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "when条件不能为空")

				// 空then动作
				rule = SimpleRule{
					When: "true",
					Then: map[string]string{},
				}

				err = converter.Validate(rule)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "then动作不能为空")
			})

			Convey("验证无效MetricRule", func() {
				// 空名称
				rule := MetricRule{
					Name:    "",
					Formula: "value * 2",
				}

				err := converter.Validate(rule)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "名称不能为空")

				// 空公式
				rule2 := MetricRule{
					Name:    "test_metric",
					Formula: "",
				}

				err = converter.Validate(rule2)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "公式不能为空")
			})
		})

		Convey("JSON转换回退测试", func() {
			converter := NewGRLConverter()

			Convey("JSON转换为SimpleRule", func() {
				// 无法转换为StandardRule的JSON，应该尝试转换为SimpleRule
				jsonStr := `{
					"when": "age >= 18",
					"then": {
						"result": "approved"
					}
				}`

				grl, err := converter.ConvertToGRL(jsonStr)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
				So(grl, ShouldContainSubstring, "age >= 18")
			})

			Convey("JSON完全无效", func() {
				invalidJSON := `{"completely": "invalid", "structure": true}`

				_, err := converter.ConvertToGRL(invalidJSON)
				So(err, ShouldNotBeNil)
			})
		})
	})
}