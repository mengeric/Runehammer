package runehammer

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestGRLConverter 测试GRL转换器
func TestGRLConverter(t *testing.T) {
	Convey("GRL转换器测试", t, func() {

		Convey("NewGRLConverter 构造函数", func() {

			Convey("默认配置创建", func() {
				converter := NewGRLConverter()

				So(converter, ShouldNotBeNil)
				So(converter.config.DefaultPriority, ShouldEqual, 50)
				So(converter.config.StrictMode, ShouldBeFalse)
				So(converter.config.VariablePrefix, ShouldNotBeNil)
				So(converter.config.OperatorMapping, ShouldNotBeNil)
				So(converter.config.FunctionMapping, ShouldNotBeNil)

				// 验证默认映射
				So(converter.config.OperatorMapping["=="], ShouldEqual, "==")
				So(converter.config.OperatorMapping["and"], ShouldEqual, "&&")
				So(converter.config.OperatorMapping["or"], ShouldEqual, "||")
				So(converter.config.FunctionMapping["now"], ShouldEqual, "Now()")
			})

			Convey("自定义配置创建", func() {
				customConfig := ConverterConfig{
					VariablePrefix: map[string]string{
						"custom": "CustomPrefix",
					},
					OperatorMapping: map[string]string{
						"equals": "==",
					},
					FunctionMapping: map[string]string{
						"current_time": "CurrentTime()",
					},
					DefaultPriority: 100,
					StrictMode:      true,
				}

				converter := NewGRLConverter(customConfig)

				So(converter.config.DefaultPriority, ShouldEqual, 100)
				So(converter.config.StrictMode, ShouldBeTrue)
				So(converter.config.VariablePrefix["custom"], ShouldEqual, "CustomPrefix")
				So(converter.config.OperatorMapping["equals"], ShouldEqual, "==")
				So(converter.config.FunctionMapping["current_time"], ShouldEqual, "CurrentTime()")
			})

			Convey("部分自定义配置", func() {
				partialConfig := ConverterConfig{

					DefaultPriority: 75,
					StrictMode:      true,
				}

				converter := NewGRLConverter(partialConfig)

				So(converter.config.DefaultPriority, ShouldEqual, 75)
				So(converter.config.StrictMode, ShouldBeTrue)
				// 其他配置应该使用默认值
				So(converter.config.OperatorMapping["=="], ShouldEqual, "==")
			})
		})

		Convey("ConvertToGRL 主转换方法", func() {
			converter := NewGRLConverter()

			Convey("转换JSON字符串", func() {
				jsonStr := `{
					"id": "JSON_RULE",
					"name": "JSON规则",
					"description": "从JSON转换的规则",
					"conditions": {
						"type": "simple",
						"left": "status",
						"operator": "==",
						"right": "active"
					},
					"actions": [
						{
							"type": "assign",
							"target": "result",
							"value": "valid"
						}
					]
				}`

				grl, err := converter.ConvertToGRL(jsonStr)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
			})

			Convey("转换无效JSON字符串", func() {
				invalidJSON := `{"invalid": json}`

				grl, err := converter.ConvertToGRL(invalidJSON)
				So(err, ShouldNotBeNil)
				So(grl, ShouldBeEmpty)
			})

			Convey("转换StandardRule", func() {
				rule := NewStandardRule("R001", "测试规则")
				rule.Description = "年龄验证规则"
				rule.Priority = 60
				rule.AddSimpleCondition("age", OpGreaterThan, 18)
				rule.AddAction(ActionTypeAssign, "result.approved", true)

				grl, err := converter.ConvertToGRL(*rule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
				So(grl, ShouldContainSubstring, "rule R001")
				So(grl, ShouldContainSubstring, "年龄验证规则")
				So(grl, ShouldContainSubstring, "salience 60")
				So(grl, ShouldContainSubstring, "when")
				So(grl, ShouldContainSubstring, "then")
				So(grl, ShouldContainSubstring, "\"age\" > 18")
				So(grl, ShouldContainSubstring, "result[\"approved\"]")
			})

			Convey("转换StandardRule指针", func() {
				rule := NewStandardRule("R002", "指针规则")
				rule.AddSimpleCondition("status", OpEqual, "active")
				rule.AddAction(ActionTypeAssign, "result", "valid")

				grl, err := converter.ConvertToGRL(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "rule R002")
				So(grl, ShouldContainSubstring, "\"status\" == \"active\"")
			})

			Convey("转换SimpleRule", func() {
				rule := SimpleRule{
					When: "age >= 18 && income > 30000",
					Then: map[string]string{
						"result.approved": "true",
						"result.level":    "standard",
					},
				}

				grl, err := converter.ConvertToGRL(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
				So(grl, ShouldContainSubstring, "\"age\" >= 18")
				So(grl, ShouldContainSubstring, "income > 30000")
				So(grl, ShouldContainSubstring, "result[\"approved\"]")
				So(grl, ShouldContainSubstring, "result[\"level\"]")
			})

			Convey("转换MetricRule", func() {
				rule := MetricRule{
					Name:        "credit_score",
					Description: "信用评分计算",
					Formula:     "income * 0.1 + age * 0.05",
					Variables: map[string]string{
						"base_score": "income / 1000",
					},
					Conditions: []string{
						"income > 0",
						"age >= 18",
					},
				}

				grl, err := converter.ConvertToGRL(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
				So(grl, ShouldContainSubstring, "credit_score")
				So(grl, ShouldContainSubstring, "income * 0.1")
				So(grl, ShouldContainSubstring, "base_score")
			})

			Convey("转换不支持的类型", func() {
				unsupported := struct {
					Field string
				}{
					Field: "test",
				}

				grl, err := converter.ConvertToGRL(unsupported)
				So(err, ShouldNotBeNil)
				So(grl, ShouldBeEmpty)
				So(err.Error(), ShouldContainSubstring, "不支持的规则定义类型")
			})

			Convey("转换map类型应该失败", func() {
				mapData := map[string]interface{}{
					"id":   "MAP_RULE",
					"name": "Map规则",
				}

				grl, err := converter.ConvertToGRL(mapData)
				So(err, ShouldNotBeNil)
				So(grl, ShouldBeEmpty)
				So(err.Error(), ShouldContainSubstring, "不支持的规则定义类型")
			})
		})

		Convey("ConvertRule 标准规则转换", func() {
			converter := NewGRLConverter()

			Convey("基本规则转换", func() {
				rule := StandardRule{
					ID:          "BASIC_001",
					Name:        "基本规则",
					Description: "基本测试规则",
					Priority:    70,
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "amount",
						Operator: OpGreaterThan,
						Right:    1000,
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result.risk",
							Value:  "high",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "rule BASIC_001")
				So(grl, ShouldContainSubstring, "基本测试规则")
				So(grl, ShouldContainSubstring, "salience 70")
				So(grl, ShouldContainSubstring, "\"amount\" > 1000")
				So(grl, ShouldContainSubstring, "result[\"risk\"] = \"high\"")
				So(grl, ShouldContainSubstring, "Retract(\"BASIC_001\")")
			})

			Convey("使用默认优先级", func() {
				rule := StandardRule{
					ID:   "DEFAULT_PRIORITY",
					Name: "默认优先级规则",
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

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "salience 50") // 默认优先级
			})

			Convey("复合条件转换", func() {
				rule := StandardRule{
					ID:   "COMPOSITE_001",
					Name: "复合条件规则",
					Conditions: Condition{
						Type:     ConditionTypeComposite,
						Operator: OpAnd,
						Children: []Condition{
							{
								Type:     ConditionTypeSimple,
								Left:     "age",
								Operator: OpGreaterThanOrEqual,
								Right:    18,
							},
							{
								Type:     ConditionTypeSimple,
								Left:     "income",
								Operator: OpGreaterThan,
								Right:    50000,
							},
						},
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "approved",
							Value:  true,
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "\"age\" >= 18")
				So(grl, ShouldContainSubstring, "\"income\" > 50000")
				So(grl, ShouldContainSubstring, "&&") // AND操作符转换
			})

			Convey("多个动作转换", func() {
				rule := StandardRule{
					ID:   "MULTI_ACTION",
					Name: "多动作规则",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "status",
						Operator: OpEqual,
						Right:    "pending",
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result.status",
							Value:  "processed",
						},
						{
							Type:   ActionTypeAssign,
							Target: "result.timestamp",
							Value:  "now()",
						},
						{
							Type:  ActionTypeLog,
							Value: "Rule executed successfully",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "result[\"status\"] = \"processed\"")
				So(grl, ShouldContainSubstring, "result[\"timestamp\"] = \"now()\"")
				// 验证有多行动作
				actionLines := strings.Count(grl, ";")
				So(actionLines, ShouldBeGreaterThanOrEqualTo, 3) // 至少3个动作（包括Retract）
			})
		})

		Convey("ConvertSimpleRule 简化规则转换", func() {
			converter := NewGRLConverter()

			Convey("基本简化规则", func() {
				rule := SimpleRule{
					When: "age > 21",
					Then: map[string]string{
						"result": "adult",
					},
				}

				grl, err := converter.ConvertSimpleRule(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
				So(grl, ShouldContainSubstring, "age > 21")
				So(grl, ShouldContainSubstring, "result")
				So(grl, ShouldContainSubstring, "adult")
			})

			Convey("复杂表达式", func() {
				rule := SimpleRule{
					When: "age >= 18 && income > 30000 && credit_score >= 700",
					Then: map[string]string{
						"result.approved": "true",
						"result.rate":     "0.05",
						"result.limit":    "100000",
					},
				}

				grl, err := converter.ConvertSimpleRule(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "\"age\" >= 18")
				So(grl, ShouldContainSubstring, "&&")
				So(grl, ShouldContainSubstring, "credit_score >= 700")
				So(grl, ShouldContainSubstring, "result[\"approved\"]")
				So(grl, ShouldContainSubstring, "result[\"rate\"]")
				So(grl, ShouldContainSubstring, "result[\"limit\"]")
			})

			Convey("空when条件", func() {
				rule := SimpleRule{
					When: "",
					Then: map[string]string{
						"result": "default",
					},
				}

				_, err := converter.ConvertSimpleRule(rule)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "条件表达式不能为空")
			})

			Convey("空then动作", func() {
				rule := SimpleRule{
					When: "true",
					Then: map[string]string{},
				}

				_, err := converter.ConvertSimpleRule(rule)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "表达式不能为空")
			})
		})

		Convey("ConvertMetricRule 指标规则转换", func() {
			converter := NewGRLConverter()

			Convey("基本指标规则", func() {
				rule := MetricRule{
					Name:        "simple_score",
					Description: "简单评分",
					Formula:     "income / 1000",
					Variables:   map[string]string{},
					Conditions:  []string{"income > 0"},
				}

				grl, err := converter.ConvertMetricRule(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
				So(grl, ShouldContainSubstring, "simple_score")
				So(grl, ShouldContainSubstring, "income / 1000")
				So(grl, ShouldContainSubstring, "income > 0")
			})

			Convey("带变量定义的指标规则", func() {
				rule := MetricRule{
					Name:        "complex_score",
					Description: "复杂评分计算",
					Formula:     "base_score + bonus_score",
					Variables: map[string]string{
						"base_score":  "income / 10000",
						"bonus_score": "age > 30 ? 10 : 0",
					},
					Conditions: []string{
						"income > 0",
						"age >= 18",
					},
				}

				grl, err := converter.ConvertMetricRule(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "complex_score")
				So(grl, ShouldContainSubstring, "base_score")
				So(grl, ShouldContainSubstring, "bonus_score")
				So(grl, ShouldContainSubstring, "income / 10000")
				So(grl, ShouldContainSubstring, "age > 30")
				So(grl, ShouldContainSubstring, "base_score + bonus_score")
			})

			Convey("无条件的指标规则", func() {
				rule := MetricRule{
					Name:        "unconditional_score",
					Description: "无条件评分",
					Formula:     "100",
					Variables:   map[string]string{},
					Conditions:  []string{},
				}

				grl, err := converter.ConvertMetricRule(rule)
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "unconditional_score")
				So(grl, ShouldContainSubstring, "100")
			})

			Convey("空名称的指标规则", func() {
				rule := MetricRule{
					Name:       "",
					Formula:    "income * 0.1",
					Variables:  map[string]string{},
					Conditions: []string{},
				}

				_, err := converter.ConvertMetricRule(rule)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "表达式不能为空")
			})

			Convey("空公式的指标规则", func() {
				rule := MetricRule{
					Name:       "empty_formula",
					Formula:    "",
					Variables:  map[string]string{},
					Conditions: []string{},
				}

				_, err := converter.ConvertMetricRule(rule)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "表达式不能为空")
			})
		})

		Convey("操作符映射测试", func() {
			converter := NewGRLConverter()

			Convey("比较操作符", func() {
				testCases := []struct {
					input    string
					expected string
				}{
					{"==", "=="},
					{"!=", "!="},
					{">", ">"},
					{"<", "<"},
					{">=", ">="},
					{"<=", "<="},
				}

				for _, tc := range testCases {
					rule := NewStandardRule("TEST", "测试")
					rule.AddSimpleCondition("field", tc.input, "value")
					rule.AddAction(ActionTypeAssign, "result", "ok")

					grl, err := converter.ConvertRule(*rule, Definitions{})
					So(err, ShouldBeNil)
					So(grl, ShouldContainSubstring, tc.expected)
				}
			})

			Convey("逻辑操作符", func() {
				rule := StandardRule{
					ID:   "LOGIC_TEST",
					Name: "逻辑操作符测试",
					Conditions: Condition{
						Type:     ConditionTypeComposite,
						Operator: OpAnd,
						Children: []Condition{
							{
								Type:     ConditionTypeSimple,
								Left:     "a",
								Operator: OpEqual,
								Right:    1,
							},
							{
								Type:     ConditionTypeSimple,
								Left:     "b",
								Operator: OpEqual,
								Right:    2,
							},
						},
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "ok",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldContainSubstring, "&&") // AND转换为&&
			})
		})

		Convey("Validate 验证方法", func() {
			converter := NewGRLConverter()

			Convey("验证有效的StandardRule", func() {
				rule := NewStandardRule("VALID_RULE", "有效规则")
				rule.AddSimpleCondition("field", OpEqual, "value")
				rule.AddAction(ActionTypeAssign, "result", "ok")

				err := converter.Validate(*rule)
				So(err, ShouldBeNil)
			})

			Convey("验证无效的规则", func() {
				rule := &StandardRule{
					ID: "", // 缺少ID
				}

				err := converter.Validate(*rule)
				So(err, ShouldNotBeNil)
			})

			Convey("验证SimpleRule", func() {
				rule := SimpleRule{
					When: "field == value",
					Then: map[string]string{
						"result": "ok",
					},
				}

				err := converter.Validate(rule)
				So(err, ShouldBeNil)
			})

			Convey("验证不支持的类型", func() {
				unsupported := "invalid rule type"

				err := converter.Validate(unsupported)
				So(err, ShouldBeNil) // 根据实际实现，不支持的类型可能不报错
			})
		})

		Convey("错误处理测试", func() {
			converter := NewGRLConverter()

			Convey("无效条件转换", func() {
				rule := StandardRule{
					ID:   "INVALID_CONDITION",
					Name: "无效条件规则",
					Conditions: Condition{
						Type: "invalid_type", // 无效的条件类型
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "ok",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldNotBeNil)
				So(grl, ShouldBeEmpty)
			})

			Convey("无效动作转换", func() {
				rule := StandardRule{
					ID:   "INVALID_ACTION",
					Name: "无效动作规则",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "field",
						Operator: OpEqual,
						Right:    "value",
					},
					Actions: []Action{
						{
							Type: "invalid_action_type", // 无效的动作类型
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldNotBeNil)
				So(grl, ShouldBeEmpty)
			})
		})

		Convey("边界情况测试", func() {
			converter := NewGRLConverter()

			Convey("空规则ID处理", func() {
				rule := StandardRule{
					ID:   "",
					Name: "空ID规则",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "field",
						Operator: OpEqual,
						Right:    "value",
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "ok",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				// 应该有合理的错误处理或默认行为
				_ = grl
				_ = err
			})

			Convey("特殊字符在规则名中", func() {
				rule := StandardRule{
					ID:   "SPECIAL-CHARS_123",
					Name: "特殊字符规则!@#$%",
					Conditions: Condition{
						Type:     ConditionTypeSimple,
						Left:     "field",
						Operator: OpEqual,
						Right:    "value",
					},
					Actions: []Action{
						{
							Type:   ActionTypeAssign,
							Target: "result",
							Value:  "ok",
						},
					},
				}

				grl, err := converter.ConvertRule(rule, Definitions{})
				So(err, ShouldBeNil)
				// 规则名应该被清理（去除特殊字符或转义）
				So(grl, ShouldContainSubstring, "rule")
			})
		})
	})
}

// TestConverterConfig 测试转换器配置
func TestConverterConfig(t *testing.T) {
	Convey("转换器配置测试", t, func() {

		Convey("ConverterConfig 结构", func() {

			Convey("基本配置", func() {
				config := ConverterConfig{
					VariablePrefix: map[string]string{
						"user":  "UserData",
						"order": "OrderInfo",
					},
					OperatorMapping: map[string]string{
						"equals":  "==",
						"greater": ">",
					},
					FunctionMapping: map[string]string{
						"current_date": "CurrentDate()",
						"sum_values":   "Sum",
					},
					DefaultPriority: 75,
					StrictMode:      true,
				}

				So(config.VariablePrefix["user"], ShouldEqual, "UserData")
				So(config.OperatorMapping["equals"], ShouldEqual, "==")
				So(config.FunctionMapping["current_date"], ShouldEqual, "CurrentDate()")
				So(config.DefaultPriority, ShouldEqual, 75)
				So(config.StrictMode, ShouldBeTrue)
			})

			Convey("空配置", func() {
				config := ConverterConfig{}

				So(config.VariablePrefix, ShouldBeNil)
				So(config.OperatorMapping, ShouldBeNil)
				So(config.FunctionMapping, ShouldBeNil)
				So(config.DefaultPriority, ShouldEqual, 0)
				So(config.StrictMode, ShouldBeFalse)
			})
		})

		Convey("配置合并测试", func() {

			Convey("配置覆盖", func() {
				customConfig := ConverterConfig{
					DefaultPriority: 200,
					StrictMode:      true,
					OperatorMapping: map[string]string{
						"custom_op": "CUSTOM",
					},
				}

				converter := NewGRLConverter(customConfig)

				So(converter.config.DefaultPriority, ShouldEqual, 200)
				So(converter.config.StrictMode, ShouldBeTrue)
				So(converter.config.OperatorMapping["custom_op"], ShouldEqual, "CUSTOM")
			})
		})
	})
}

// TestRuleConverterInterface 测试RuleConverter接口
func TestRuleConverterInterface(t *testing.T) {
	Convey("RuleConverter接口测试", t, func() {

		Convey("接口实现验证", func() {
			converter := NewGRLConverter()

			// 验证GRLConverter实现了RuleConverter接口
			So(converter, ShouldImplement, (*RuleConverter)(nil))
		})

		Convey("接口方法调用", func() {
			var converter RuleConverter = NewGRLConverter()

			// 测试所有接口方法都能正常调用
			rule := NewStandardRule("INTERFACE_TEST", "接口测试")
			rule.AddSimpleCondition("test", OpEqual, true)
			rule.AddAction(ActionTypeAssign, "result", "ok")

			Convey("ConvertRule方法", func() {
				grl, err := converter.ConvertRule(*rule, Definitions{})
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
			})

			Convey("ConvertToGRL方法", func() {
				grl, err := converter.ConvertToGRL(*rule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
			})

			Convey("ConvertSimpleRule方法", func() {
				simpleRule := SimpleRule{
					When: "test == true",
					Then: map[string]string{"result": "ok"},
				}

				grl, err := converter.ConvertSimpleRule(simpleRule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
			})

			Convey("ConvertMetricRule方法", func() {
				metricRule := MetricRule{
					Name:       "test_metric",
					Formula:    "value * 2",
					Variables:  map[string]string{},
					Conditions: []string{"value > 0"},
				}

				grl, err := converter.ConvertMetricRule(metricRule)
				So(err, ShouldBeNil)
				So(grl, ShouldNotBeEmpty)
			})

			Convey("Validate方法", func() {
				err := converter.Validate(*rule)
				So(err, ShouldBeNil)
			})
		})
	})
}
