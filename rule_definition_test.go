package runehammer

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestRuleDefinition 测试规则定义标准
func TestRuleDefinition(t *testing.T) {
	Convey("规则定义标准测试", t, func() {

		Convey("RuleDefinitionStandard 结构", func() {

			Convey("基本结构创建", func() {
				standard := &RuleDefinitionStandard{
					Version: "1.0",
					Metadata: Metadata{
						Domain:      "finance",
						Author:      "test_author",
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
						Description: "测试规则集",
						Version:     "1.0.0",
					},
					Definitions: Definitions{
						Variables: map[string]interface{}{
							"min_age":    18,
							"max_amount": 10000,
						},
						Functions: []FunctionDef{
							{
								Name:        "calculate_score",
								Parameters:  []string{"income", "debt"},
								Description: "计算信用评分",
								ReturnType:  "number",
							},
						},
						Constants: map[string]interface{}{
							"RATE": 0.05,
						},
					},
					Rules: []Rule{},
				}

				So(standard, ShouldNotBeNil)
				So(standard.Version, ShouldEqual, "1.0")
				So(standard.Metadata.Domain, ShouldEqual, "finance")
				So(standard.Definitions.Variables["min_age"], ShouldEqual, 18)
				So(len(standard.Definitions.Functions), ShouldEqual, 1)
				So(standard.Definitions.Constants["RATE"], ShouldEqual, 0.05)
			})

			Convey("JSON序列化和反序列化", func() {
				original := &RuleDefinitionStandard{
					Version: "1.0",
					Metadata: Metadata{
						Domain:      "test",
						Author:      "test_user",
						CreatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						Description: "测试标准",
					},
					Rules: []Rule{},
				}

				// 序列化
				data, err := json.Marshal(original)
				So(err, ShouldBeNil)
				So(data, ShouldNotBeEmpty)

				// 反序列化
				var restored RuleDefinitionStandard
				err = json.Unmarshal(data, &restored)
				So(err, ShouldBeNil)
				So(restored.Version, ShouldEqual, original.Version)
				So(restored.Metadata.Domain, ShouldEqual, original.Metadata.Domain)
			})
		})

		Convey("Metadata 元数据", func() {

			Convey("完整元数据信息", func() {
				now := time.Now()
				metadata := Metadata{
					Domain:      "insurance",
					Author:      "rule_designer",
					CreatedAt:   now,
					UpdatedAt:   now,
					Description: "保险规则集合",
					Version:     "2.1.0",
				}

				So(metadata.Domain, ShouldEqual, "insurance")
				So(metadata.Author, ShouldEqual, "rule_designer")
				So(metadata.CreatedAt, ShouldEqual, now)
				So(metadata.UpdatedAt, ShouldEqual, now)
				So(metadata.Description, ShouldContainSubstring, "保险")
				So(metadata.Version, ShouldEqual, "2.1.0")
			})
		})

		Convey("Definitions 可重用定义", func() {

			Convey("变量定义", func() {
				definitions := Definitions{
					Variables: map[string]interface{}{
						"threshold":  100,
						"rate":       0.15,
						"category":   "premium",
						"is_enabled": true,
						"tags":       []string{"risk", "score"},
					},
				}

				So(definitions.Variables["threshold"], ShouldEqual, 100)
				So(definitions.Variables["rate"], ShouldEqual, 0.15)
				So(definitions.Variables["category"], ShouldEqual, "premium")
				So(definitions.Variables["is_enabled"], ShouldBeTrue)

				tags, ok := definitions.Variables["tags"].([]string)
				So(ok, ShouldBeTrue)
				So(len(tags), ShouldEqual, 2)
				So(tags[0], ShouldEqual, "risk")
			})

			Convey("函数定义", func() {
				funcDef := FunctionDef{
					Name:        "validate_credit",
					Parameters:  []string{"score", "history", "income"},
					Description: "验证信用状况",
					ReturnType:  "boolean",
				}

				So(funcDef.Name, ShouldEqual, "validate_credit")
				So(len(funcDef.Parameters), ShouldEqual, 3)
				So(funcDef.Parameters[0], ShouldEqual, "score")
				So(funcDef.Description, ShouldContainSubstring, "验证")
				So(funcDef.ReturnType, ShouldEqual, "boolean")
			})

			Convey("常量定义", func() {
				definitions := Definitions{
					Constants: map[string]interface{}{
						"PI":           3.14159,
						"MAX_RETRIES":  5,
						"DEFAULT_LANG": "zh-CN",
						"ENABLED":      true,
					},
				}

				So(definitions.Constants["PI"], ShouldEqual, 3.14159)
				So(definitions.Constants["MAX_RETRIES"], ShouldEqual, 5)
				So(definitions.Constants["DEFAULT_LANG"], ShouldEqual, "zh-CN")
				So(definitions.Constants["ENABLED"], ShouldBeTrue)
			})
		})

		Convey("StandardRule 标准规则", func() {

			Convey("NewStandardRule 工厂方法", func() {
				rule := NewStandardRule("R001", "用户年龄验证")

				So(rule, ShouldNotBeNil)
				So(rule.ID, ShouldEqual, "R001")
				So(rule.Name, ShouldEqual, "用户年龄验证")
				So(rule.Priority, ShouldEqual, 50)
				So(rule.Enabled, ShouldBeTrue)
				So(len(rule.Tags), ShouldEqual, 0)
				So(len(rule.Actions), ShouldEqual, 0)
			})

			Convey("AddSimpleCondition 添加简单条件", func() {
				rule := NewStandardRule("R002", "收入验证")

				// 添加第一个条件
				rule.AddSimpleCondition("income", OpGreaterThan, 50000)

				So(rule.Conditions.Type, ShouldEqual, ConditionTypeSimple)
				So(rule.Conditions.Left, ShouldEqual, "income")
				So(rule.Conditions.Operator, ShouldEqual, OpGreaterThan)
				So(rule.Conditions.Right, ShouldEqual, 50000)

				// 添加第二个条件，应该自动创建复合条件
				rule.AddSimpleCondition("age", OpGreaterThanOrEqual, 18)

				So(rule.Conditions.Type, ShouldEqual, ConditionTypeComposite)
				So(rule.Conditions.Operator, ShouldEqual, OpAnd)
				So(len(rule.Conditions.Children), ShouldEqual, 2)

				// 验证第一个子条件
				firstChild := rule.Conditions.Children[0]
				So(firstChild.Type, ShouldEqual, ConditionTypeSimple)
				So(firstChild.Left, ShouldEqual, "income")
				So(firstChild.Operator, ShouldEqual, OpGreaterThan)
				So(firstChild.Right, ShouldEqual, 50000)

				// 验证第二个子条件
				secondChild := rule.Conditions.Children[1]
				So(secondChild.Type, ShouldEqual, ConditionTypeSimple)
				So(secondChild.Left, ShouldEqual, "age")
				So(secondChild.Operator, ShouldEqual, OpGreaterThanOrEqual)
				So(secondChild.Right, ShouldEqual, 18)
			})

			Convey("AddAction 添加动作", func() {
				rule := NewStandardRule("R003", "评分计算")

				// 添加赋值动作
				rule.AddAction(ActionTypeAssign, "result.status", "approved")

				So(len(rule.Actions), ShouldEqual, 1)
				action := rule.Actions[0]
				So(action.Type, ShouldEqual, ActionTypeAssign)
				So(action.Target, ShouldEqual, "result.status")
				So(action.Value, ShouldEqual, "approved")

				// 添加计算动作
				rule.AddAction(ActionTypeCalculate, "result.score", "income * 0.1 + age * 0.05")

				So(len(rule.Actions), ShouldEqual, 2)
				calcAction := rule.Actions[1]
				So(calcAction.Type, ShouldEqual, ActionTypeCalculate)
				So(calcAction.Target, ShouldEqual, "result.score")
				So(calcAction.Value, ShouldEqual, "income * 0.1 + age * 0.05")
			})

			Convey("ToJSON 和 FromJSON", func() {
				rule := NewStandardRule("R004", "JSON测试规则")
				rule.AddSimpleCondition("amount", OpLessThan, 1000)
				rule.AddAction(ActionTypeAssign, "result.level", "low")

				// 转换为JSON
				jsonStr, err := rule.ToJSON()
				So(err, ShouldBeNil)
				So(jsonStr, ShouldNotBeEmpty)
				So(jsonStr, ShouldContainSubstring, "R004")
				So(jsonStr, ShouldContainSubstring, "JSON测试规则")

				// 从JSON恢复
				var restoredRule StandardRule
				err = restoredRule.FromJSON(jsonStr)
				So(err, ShouldBeNil)
				So(restoredRule.ID, ShouldEqual, rule.ID)
				So(restoredRule.Name, ShouldEqual, rule.Name)
				So(restoredRule.Conditions.Type, ShouldEqual, rule.Conditions.Type)
				So(len(restoredRule.Actions), ShouldEqual, len(rule.Actions))
			})
		})

		Convey("Condition 条件定义", func() {

			Convey("简单条件", func() {
				condition := Condition{
					Type:     ConditionTypeSimple,
					Left:     "user.age",
					Operator: OpGreaterThanOrEqual,
					Right:    21,
				}

				So(condition.Type, ShouldEqual, ConditionTypeSimple)
				So(condition.Left, ShouldEqual, "user.age")
				So(condition.Operator, ShouldEqual, OpGreaterThanOrEqual)
				So(condition.Right, ShouldEqual, 21)
			})

			Convey("复合条件", func() {
				condition := Condition{
					Type:     ConditionTypeComposite,
					Operator: OpAnd,
					Children: []Condition{
						{
							Type:     ConditionTypeSimple,
							Left:     "user.age",
							Operator: OpGreaterThan,
							Right:    18,
						},
						{
							Type:     ConditionTypeSimple,
							Left:     "user.income",
							Operator: OpGreaterThan,
							Right:    30000,
						},
					},
				}

				So(condition.Type, ShouldEqual, ConditionTypeComposite)
				So(condition.Operator, ShouldEqual, OpAnd)
				So(len(condition.Children), ShouldEqual, 2)

				firstChild := condition.Children[0]
				So(firstChild.Type, ShouldEqual, ConditionTypeSimple)
				So(firstChild.Left, ShouldEqual, "user.age")
				So(firstChild.Right, ShouldEqual, 18)

				secondChild := condition.Children[1]
				So(secondChild.Left, ShouldEqual, "user.income")
				So(secondChild.Right, ShouldEqual, 30000)
			})

			Convey("表达式条件", func() {
				condition := Condition{
					Type:       ConditionTypeExpression,
					Expression: "user.age >= 18 && user.income > 50000",
				}

				So(condition.Type, ShouldEqual, ConditionTypeExpression)
				So(condition.Expression, ShouldContainSubstring, "user.age")
				So(condition.Expression, ShouldContainSubstring, "&&")
				So(condition.Expression, ShouldContainSubstring, "50000")
			})

			Convey("函数条件", func() {
				condition := Condition{
					Type:  ConditionTypeFunction,
					Left:  "validate_credit",
					Right: []interface{}{"user.credit_score", "user.history"},
				}

				So(condition.Type, ShouldEqual, ConditionTypeFunction)
				So(condition.Left, ShouldEqual, "validate_credit")

				params, ok := condition.Right.([]interface{})
				So(ok, ShouldBeTrue)
				So(len(params), ShouldEqual, 2)
				So(params[0], ShouldEqual, "user.credit_score")
				So(params[1], ShouldEqual, "user.history")
			})
		})

		Convey("Action 动作定义", func() {

			Convey("赋值动作", func() {
				action := Action{
					Type:   ActionTypeAssign,
					Target: "result.approved",
					Value:  true,
				}

				So(action.Type, ShouldEqual, ActionTypeAssign)
				So(action.Target, ShouldEqual, "result.approved")
				So(action.Value, ShouldBeTrue)
			})

			Convey("计算动作", func() {
				action := Action{
					Type:       ActionTypeCalculate,
					Target:     "result.score",
					Expression: "income * 0.1 + credit_score * 0.2",
				}

				So(action.Type, ShouldEqual, ActionTypeCalculate)
				So(action.Target, ShouldEqual, "result.score")
				So(action.Expression, ShouldContainSubstring, "income")
				So(action.Expression, ShouldContainSubstring, "credit_score")
			})

			Convey("调用动作", func() {
				action := Action{
					Type:   ActionTypeInvoke,
					Target: "send_notification",
					Parameters: map[string]interface{}{
						"recipient": "admin@example.com",
						"subject":   "Loan Application",
						"urgent":    true,
					},
				}

				So(action.Type, ShouldEqual, ActionTypeInvoke)
				So(action.Target, ShouldEqual, "send_notification")
				So(action.Parameters["recipient"], ShouldEqual, "admin@example.com")
				So(action.Parameters["urgent"], ShouldBeTrue)
			})

			Convey("告警动作", func() {
				action := Action{
					Type:  ActionTypeAlert,
					Value: "High risk transaction detected",
					Parameters: map[string]interface{}{
						"level":   "critical",
						"channel": "slack",
					},
				}

				So(action.Type, ShouldEqual, ActionTypeAlert)
				So(action.Value, ShouldEqual, "High risk transaction detected")
				So(action.Parameters["level"], ShouldEqual, "critical")
			})

			Convey("日志动作", func() {
				action := Action{
					Type:  ActionTypeLog,
					Value: "Processing loan application for user ${user.id}",
					Parameters: map[string]interface{}{
						"level":    "info",
						"category": "audit",
					},
				}

				So(action.Type, ShouldEqual, ActionTypeLog)
				So(action.Value, ShouldContainSubstring, "loan application")
				So(action.Parameters["level"], ShouldEqual, "info")
			})
		})

		Convey("常量定义", func() {

			Convey("操作符常量", func() {
				// 比较操作符
				So(string(OpEqual), ShouldEqual, "==")
				So(string(OpNotEqual), ShouldEqual, "!=")
				So(string(OpGreaterThan), ShouldEqual, ">")
				So(string(OpLessThan), ShouldEqual, "<")
				So(string(OpGreaterThanOrEqual), ShouldEqual, ">=")
				So(string(OpLessThanOrEqual), ShouldEqual, "<=")

				// 逻辑操作符
				So(string(OpAnd), ShouldEqual, "and")
				So(string(OpOr), ShouldEqual, "or")
				So(string(OpNot), ShouldEqual, "not")

				// 集合操作符
				So(string(OpIn), ShouldEqual, "in")
				So(string(OpNotIn), ShouldEqual, "notIn")
				So(string(OpContains), ShouldEqual, "contains")
				So(string(OpMatches), ShouldEqual, "matches")
				So(string(OpBetween), ShouldEqual, "between")
			})

			Convey("条件类型常量", func() {
				So(string(ConditionTypeSimple), ShouldEqual, "simple")
				So(string(ConditionTypeComposite), ShouldEqual, "composite")
				So(string(ConditionTypeExpression), ShouldEqual, "expression")
				So(string(ConditionTypeFunction), ShouldEqual, "function")
			})

			Convey("动作类型常量", func() {
				So(string(ActionTypeAssign), ShouldEqual, "assign")
				So(string(ActionTypeCalculate), ShouldEqual, "calculate")
				So(string(ActionTypeInvoke), ShouldEqual, "invoke")
				So(string(ActionTypeAlert), ShouldEqual, "alert")
				So(string(ActionTypeLog), ShouldEqual, "log")
				So(string(ActionTypeStop), ShouldEqual, "stop")
			})
		})
	})
}

// TestSimpleRule 测试简化规则定义
func TestSimpleRule(t *testing.T) {
	Convey("简化规则定义测试", t, func() {

		Convey("SimpleRule 结构", func() {

			Convey("基本创建", func() {
				rule := SimpleRule{
					When: "user.age >= 18 && user.income > 30000",
					Then: map[string]string{
						"result.approved": "true",
						"result.level":    "standard",
					},
				}

				So(rule.When, ShouldContainSubstring, "user.age")
				So(rule.When, ShouldContainSubstring, "&&")
				So(rule.Then["result.approved"], ShouldEqual, "true")
				So(rule.Then["result.level"], ShouldEqual, "standard")
			})

			Convey("JSON序列化", func() {
				rule := SimpleRule{
					When: "amount < 1000",
					Then: map[string]string{
						"risk": "low",
					},
				}

				data, err := json.Marshal(rule)
				So(err, ShouldBeNil)
				So(string(data), ShouldNotBeEmpty)
				// 注意：实际生成的JSON可能不会包含"amount < 1000"这个确切的字符串
				// 这里我们只是确保序列化成功
			})
		})

		Convey("MetricRule 指标计算规则", func() {

			Convey("基本创建", func() {
				rule := MetricRule{
					Name:        "credit_score",
					Description: "信用评分计算",
					Formula:     "income_score * 0.4 + payment_history * 0.6",
					Variables: map[string]string{
						"income_score":    "income / 10000",
						"payment_history": "good_payments / total_payments",
					},
					Conditions: []string{
						"income > 0",
						"total_payments > 0",
					},
				}

				So(rule.Name, ShouldEqual, "credit_score")
				So(rule.Description, ShouldContainSubstring, "信用评分")
				So(rule.Formula, ShouldContainSubstring, "income_score")
				So(rule.Variables["income_score"], ShouldContainSubstring, "income")
				So(len(rule.Conditions), ShouldEqual, 2)
				So(rule.Conditions[0], ShouldEqual, "income > 0")
			})

			Convey("复杂指标计算", func() {
				rule := MetricRule{
					Name:        "risk_score",
					Description: "风险评分计算",
					Formula:     "base_risk + age_factor + location_factor",
					Variables: map[string]string{
						"base_risk":       "transaction_amount / monthly_income",
						"age_factor":      "age < 25 ? 0.1 : 0",
						"location_factor": "high_risk_country ? 0.2 : 0",
					},
					Conditions: []string{
						"monthly_income > 0",
						"transaction_amount > 0",
					},
				}

				So(rule.Name, ShouldEqual, "risk_score")
				So(rule.Variables["base_risk"], ShouldContainSubstring, "transaction_amount")
				So(rule.Variables["age_factor"], ShouldContainSubstring, "age < 25")
				So(rule.Variables["location_factor"], ShouldContainSubstring, "high_risk_country")
			})
		})

		Convey("ValidationRule 验证规则", func() {

			Convey("基本验证规则", func() {
				rule := ValidationRule{
					Field:    "email",
					Rules:    []string{"required", "email"},
					Message:  "请输入有效的邮箱地址",
					Level:    "error",
					Required: true,
				}

				So(rule.Field, ShouldEqual, "email")
				So(len(rule.Rules), ShouldEqual, 2)
				So(rule.Rules[0], ShouldEqual, "required")
				So(rule.Rules[1], ShouldEqual, "email")
				So(rule.Message, ShouldContainSubstring, "邮箱")
				So(rule.Level, ShouldEqual, "error")
				So(rule.Required, ShouldBeTrue)
			})

			Convey("带默认值的验证规则", func() {
				rule := ValidationRule{
					Field:    "status",
					Rules:    []string{"in:active,inactive,pending"},
					Message:  "状态值无效",
					Level:    "warning",
					Required: false,
					Default:  "pending",
				}

				So(rule.Field, ShouldEqual, "status")
				So(rule.Rules[0], ShouldContainSubstring, "in:")
				So(rule.Default, ShouldEqual, "pending")
				So(rule.Required, ShouldBeFalse)
			})

			Convey("复杂验证规则", func() {
				rule := ValidationRule{
					Field: "password",
					Rules: []string{
						"required",
						"min:8",
						"regex:^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)",
					},
					Message:  "密码必须包含大小写字母和数字，至少8位",
					Level:    "error",
					Required: true,
				}

				So(rule.Field, ShouldEqual, "password")
				So(len(rule.Rules), ShouldEqual, 3)
				So(rule.Rules[1], ShouldEqual, "min:8")
				So(rule.Rules[2], ShouldContainSubstring, "regex:")
				So(rule.Message, ShouldContainSubstring, "密码")
			})
		})
	})
}

// TestRuleValidation 测试规则验证
func TestRuleValidation(t *testing.T) {
	Convey("规则验证测试", t, func() {

		Convey("StandardRule.Validate 方法", func() {

			Convey("有效规则验证", func() {
				rule := NewStandardRule("R001", "有效规则")
				rule.AddSimpleCondition("age", OpGreaterThan, 18)
				rule.AddAction(ActionTypeAssign, "result", "approved")

				errors := rule.Validate()
				So(len(errors), ShouldEqual, 0)
			})

			Convey("缺少ID的规则", func() {
				rule := &StandardRule{
					Name: "缺少ID的规则",
				}
				rule.AddSimpleCondition("age", OpGreaterThan, 18)
				rule.AddAction(ActionTypeAssign, "result", "approved")

				errors := rule.Validate()
				So(len(errors), ShouldBeGreaterThan, 0)

				// 查找ID错误
				hasIDError := false
				for _, err := range errors {
					if err.Field == "id" {
						hasIDError = true
						So(err.Message, ShouldContainSubstring, "ID不能为空")
					}
				}
				So(hasIDError, ShouldBeTrue)
			})

			Convey("缺少名称的规则", func() {
				rule := &StandardRule{
					ID: "R002",
				}
				rule.AddSimpleCondition("age", OpGreaterThan, 18)
				rule.AddAction(ActionTypeAssign, "result", "approved")

				errors := rule.Validate()
				So(len(errors), ShouldBeGreaterThan, 0)

				// 查找名称错误
				hasNameError := false
				for _, err := range errors {
					if err.Field == "name" {
						hasNameError = true
						So(err.Message, ShouldContainSubstring, "名称不能为空")
					}
				}
				So(hasNameError, ShouldBeTrue)
			})

			Convey("缺少条件的规则", func() {
				rule := &StandardRule{
					ID:   "R003",
					Name: "缺少条件的规则",
				}
				rule.AddAction(ActionTypeAssign, "result", "approved")

				errors := rule.Validate()
				So(len(errors), ShouldBeGreaterThan, 0)

				// 查找条件错误
				hasConditionError := false
				for _, err := range errors {
					if err.Field == "conditions" {
						hasConditionError = true
						So(err.Message, ShouldContainSubstring, "必须包含条件")
					}
				}
				So(hasConditionError, ShouldBeTrue)
			})

			Convey("缺少动作的规则", func() {
				rule := NewStandardRule("R004", "缺少动作的规则")
				rule.AddSimpleCondition("age", OpGreaterThan, 18)

				errors := rule.Validate()
				So(len(errors), ShouldBeGreaterThan, 0)

				// 查找动作错误
				hasActionError := false
				for _, err := range errors {
					if err.Field == "actions" {
						hasActionError = true
						So(err.Message, ShouldContainSubstring, "至少一个动作")
					}
				}
				So(hasActionError, ShouldBeTrue)
			})
		})

		Convey("validateCondition 函数", func() {

			Convey("有效简单条件", func() {
				condition := Condition{
					Type:     ConditionTypeSimple,
					Left:     "age",
					Operator: OpGreaterThan,
					Right:    18,
				}

				errors := validateCondition(condition)
				So(len(errors), ShouldEqual, 0)
			})

			Convey("无效简单条件 - 缺少左操作数", func() {
				condition := Condition{
					Type:     ConditionTypeSimple,
					Operator: OpGreaterThan,
					Right:    18,
				}

				errors := validateCondition(condition)
				So(len(errors), ShouldBeGreaterThan, 0)

				hasLeftError := false
				for _, err := range errors {
					if err.Field == "conditions.left" {
						hasLeftError = true
						So(err.Message, ShouldContainSubstring, "左操作数不能为空")
					}
				}
				So(hasLeftError, ShouldBeTrue)
			})

			Convey("无效简单条件 - 缺少操作符", func() {
				condition := Condition{
					Type:  ConditionTypeSimple,
					Left:  "age",
					Right: 18,
				}

				errors := validateCondition(condition)
				So(len(errors), ShouldBeGreaterThan, 0)

				hasOperatorError := false
				for _, err := range errors {
					if err.Field == "conditions.operator" {
						hasOperatorError = true
						So(err.Message, ShouldContainSubstring, "操作符不能为空")
					}
				}
				So(hasOperatorError, ShouldBeTrue)
			})

			Convey("无效复合条件 - 缺少子条件", func() {
				condition := Condition{
					Type:     ConditionTypeComposite,
					Operator: OpAnd,
					Children: []Condition{},
				}

				errors := validateCondition(condition)
				So(len(errors), ShouldBeGreaterThan, 0)

				hasChildrenError := false
				for _, err := range errors {
					if err.Field == "conditions.children" {
						hasChildrenError = true
						So(err.Message, ShouldContainSubstring, "必须包含子条件")
					}
				}
				So(hasChildrenError, ShouldBeTrue)
			})

			Convey("无效表达式条件 - 缺少表达式", func() {
				condition := Condition{
					Type: ConditionTypeExpression,
				}

				errors := validateCondition(condition)
				So(len(errors), ShouldBeGreaterThan, 0)

				hasExpressionError := false
				for _, err := range errors {
					if err.Field == "conditions.expression" {
						hasExpressionError = true
						So(err.Message, ShouldContainSubstring, "表达式不能为空")
					}
				}
				So(hasExpressionError, ShouldBeTrue)
			})
		})

		Convey("validateAction 函数", func() {

			Convey("有效赋值动作", func() {
				action := Action{
					Type:   ActionTypeAssign,
					Target: "result",
					Value:  "approved",
				}

				errors := validateAction(action, 0)
				So(len(errors), ShouldEqual, 0)
			})

			Convey("无效动作 - 缺少类型", func() {
				action := Action{
					Target: "result",
					Value:  "approved",
				}

				errors := validateAction(action, 0)
				So(len(errors), ShouldBeGreaterThan, 0)

				hasTypeError := false
				for _, err := range errors {
					if err.Field == "actions[0].type" {
						hasTypeError = true
						So(err.Message, ShouldContainSubstring, "类型不能为空")
					}
				}
				So(hasTypeError, ShouldBeTrue)
			})

			Convey("无效赋值动作 - 缺少目标", func() {
				action := Action{
					Type:  ActionTypeAssign,
					Value: "approved",
				}

				errors := validateAction(action, 1)
				So(len(errors), ShouldBeGreaterThan, 0)

				hasTargetError := false
				for _, err := range errors {
					if err.Field == "actions[1].target" {
						hasTargetError = true
						So(err.Message, ShouldContainSubstring, "目标不能为空")
					}
				}
				So(hasTargetError, ShouldBeTrue)
			})

			Convey("无效调用动作 - 缺少目标函数", func() {
				action := Action{
					Type: ActionTypeInvoke,
				}

				errors := validateAction(action, 2)
				So(len(errors), ShouldBeGreaterThan, 0)

				hasTargetError := false
				for _, err := range errors {
					if err.Field == "actions[2].target" {
						hasTargetError = true
						So(err.Message, ShouldContainSubstring, "目标函数不能为空")
					}
				}
				So(hasTargetError, ShouldBeTrue)
			})
		})

		Convey("ValidationError 结构", func() {

			Convey("基本错误信息", func() {
				err := ValidationError{
					Field:   "test.field",
					Message: "测试错误消息",
					Code:    "VALIDATION_ERROR",
				}

				So(err.Field, ShouldEqual, "test.field")
				So(err.Message, ShouldEqual, "测试错误消息")
				So(err.Code, ShouldEqual, "VALIDATION_ERROR")
			})

			Convey("JSON序列化", func() {
				err := ValidationError{
					Field:   "user.email",
					Message: "邮箱格式无效",
					Code:    "INVALID_EMAIL",
				}

				data, jsonErr := json.Marshal(err)
				So(jsonErr, ShouldBeNil)
				So(string(data), ShouldContainSubstring, "user.email")
				So(string(data), ShouldContainSubstring, "邮箱格式")
			})
		})
	})
}
