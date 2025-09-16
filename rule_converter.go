package runehammer

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// ============================================================================
// 规则转换器 - 将标准格式转换为GRL
// ============================================================================

// RuleConverter 规则转换器接口
type RuleConverter interface {
	// ConvertToGRL 将标准规则定义转换为GRL
	ConvertToGRL(definition interface{}) (string, error)

	// ConvertRule 转换单个标准规则
	ConvertRule(rule StandardRule, defs Definitions) (string, error)

	// ConvertSimpleRule 转换简化规则
	ConvertSimpleRule(rule SimpleRule) (string, error)

	// ConvertMetricRule 转换指标规则
	ConvertMetricRule(rule MetricRule) (string, error)

	// Validate 验证规则定义
	Validate(definition interface{}) error
}

// GRLConverter GRL转换器实现
type GRLConverter struct {
	config           ConverterConfig
	expressionParser ExpressionParser
}

// ConverterConfig 转换器配置
type ConverterConfig struct {
	// 变量前缀映射
	VariablePrefix map[string]string

	// 函数映射
	FunctionMapping map[string]string

	// 操作符映射
	OperatorMapping map[string]string

	// 是否严格模式
	StrictMode bool

	// 默认优先级
	DefaultPriority int
}

// NewGRLConverter 创建GRL转换器
func NewGRLConverter(config ...ConverterConfig) *GRLConverter {
	defaultConfig := ConverterConfig{
		VariablePrefix: map[string]string{
			"customer": "customer",
			"order":    "order",
			"user":     "user",
			"data":     "data",
			"Result":   "Result",
		},
		OperatorMapping: map[string]string{
			"==":       "==",
			"!=":       "!=",
			">":        ">",
			"<":        "<",
			">=":       ">=",
			"<=":       "<=",
			"and":      "&&",
			"or":       "||",
			"not":      "!",
			"in":       "Contains",
			"contains": "Contains",
			"matches":  "Matches",
			"between":  "BETWEEN", // 特殊处理
		},
		FunctionMapping: map[string]string{
			"now":         "Now()",
			"today":       "Today()",
			"daysBetween": "DaysBetween",
			"sum":         "Sum",
			"avg":         "Avg",
			"max":         "Max",
			"min":         "Min",
			"count":       "Count",
		},
		DefaultPriority: 50,
		StrictMode:      false,
	}

	if len(config) > 0 {
		// 合并配置
		cfg := config[0]
		if cfg.VariablePrefix != nil {
			defaultConfig.VariablePrefix = cfg.VariablePrefix
		}
		if cfg.OperatorMapping != nil {
			defaultConfig.OperatorMapping = cfg.OperatorMapping
		}
		if cfg.FunctionMapping != nil {
			defaultConfig.FunctionMapping = cfg.FunctionMapping
		}
		defaultConfig.StrictMode = cfg.StrictMode
		if cfg.DefaultPriority > 0 {
			defaultConfig.DefaultPriority = cfg.DefaultPriority
		}
	}

	return &GRLConverter{
		config:           defaultConfig,
		expressionParser: NewExpressionParser(),
	}
}

// ConvertToGRL 转换标准格式到GRL
func (c *GRLConverter) ConvertToGRL(definition interface{}) (string, error) {
	switch def := definition.(type) {
	case StandardRule:
		return c.ConvertRule(def, Definitions{})

	case *StandardRule:
		return c.ConvertRule(*def, Definitions{})

	case SimpleRule:
		return c.ConvertSimpleRule(def)

	case *SimpleRule:
		return c.ConvertSimpleRule(*def)

	case MetricRule:
		return c.ConvertMetricRule(def)

	case *MetricRule:
		return c.ConvertMetricRule(*def)

	case RuleDefinitionStandard:
		// 转换完整的规则定义标准
		return c.convertStandard(def)

	default:
		return "", fmt.Errorf("不支持的规则定义类型: %T", definition)
	}
}

// ConvertRule 转换标准规则
func (c *GRLConverter) ConvertRule(rule StandardRule, defs Definitions) (string, error) {
	var grl strings.Builder

	// 规则头
	priority := rule.Priority
	if priority == 0 {
		priority = c.config.DefaultPriority
	}

	grl.WriteString(fmt.Sprintf("rule %s \"%s\" salience %d {\n",
		c.sanitizeRuleName(rule.ID),
		rule.Description,
		priority))

	// when子句
	grl.WriteString("    when\n        ")
	condition, err := c.convertCondition(rule.Conditions, defs)
	if err != nil {
		return "", fmt.Errorf("转换条件失败: %w", err)
	}
	grl.WriteString(condition)
	grl.WriteString("\n")

	// then子句
	grl.WriteString("    then\n")
	for _, action := range rule.Actions {
		actionGRL, err := c.convertAction(action, defs)
		if err != nil {
			return "", fmt.Errorf("转换动作失败: %w", err)
		}
		grl.WriteString(fmt.Sprintf("        %s;\n", actionGRL))
	}

	// 添加Retract
	grl.WriteString(fmt.Sprintf("        Retract(\"%s\");\n", c.sanitizeRuleName(rule.ID)))
	grl.WriteString("}")

	return grl.String(), nil
}

// ConvertSimpleRule 转换简化规则
func (c *GRLConverter) ConvertSimpleRule(rule SimpleRule) (string, error) {
	var grl strings.Builder

	// 生成规则名
	ruleName := "SimpleRule_" + c.generateRuleID()

	grl.WriteString(fmt.Sprintf("rule %s \"动态生成的简化规则\" salience %d {\n",
		ruleName, c.config.DefaultPriority))

	// when子句 - 解析条件表达式
	grl.WriteString("    when\n        ")
	condition, err := c.expressionParser.ParseCondition(rule.When)
	if err != nil {
		return "", fmt.Errorf("解析when条件失败: %w", err)
	}
	grl.WriteString(condition)
	grl.WriteString("\n")

	// then子句 - 解析结果表达式
	grl.WriteString("    then\n")
	for key, expr := range rule.Then {
		action, err := c.expressionParser.ParseAction(key, expr)
		if err != nil {
			return "", fmt.Errorf("解析then动作失败 (%s): %w", key, err)
		}
		grl.WriteString(fmt.Sprintf("        %s;\n", action))
	}

	// 添加Retract
	grl.WriteString(fmt.Sprintf("        Retract(\"%s\");\n", ruleName))
	grl.WriteString("}")

	return grl.String(), nil
}

// ConvertMetricRule 转换指标规则
func (c *GRLConverter) ConvertMetricRule(rule MetricRule) (string, error) {
	var grl strings.Builder

	// 生成规则名
	ruleName := c.sanitizeRuleName("Metric_" + rule.Name)

	grl.WriteString(fmt.Sprintf("rule %s \"%s\" salience %d {\n",
		ruleName, rule.Description, c.config.DefaultPriority))

	// when子句 - 组合所有条件
	grl.WriteString("    when\n        ")
	if len(rule.Conditions) > 0 {
		var conditions []string
		for _, cond := range rule.Conditions {
			parsed, err := c.expressionParser.ParseCondition(cond)
			if err != nil {
				return "", fmt.Errorf("解析指标条件失败: %w", err)
			}
			conditions = append(conditions, parsed)
		}
		grl.WriteString(strings.Join(conditions, " && "))
	} else {
		grl.WriteString("true") // 无条件
	}
	grl.WriteString("\n")

	// then子句 - 变量定义和指标计算
	grl.WriteString("    then\n")

	// 定义变量
	for varName, expr := range rule.Variables {
		varDef, err := c.expressionParser.ParseAction(varName, expr)
		if err != nil {
			return "", fmt.Errorf("解析变量定义失败 (%s): %w", varName, err)
		}
		grl.WriteString(fmt.Sprintf("        %s;\n", varDef))
	}

	// 计算指标
	formula, err := c.expressionParser.ParseExpression(rule.Formula)
	if err != nil {
		return "", fmt.Errorf("解析指标公式失败: %w", err)
	}

	grl.WriteString(fmt.Sprintf("        Result[\"%s\"] = %s;\n", rule.Name, formula))

	// 添加Retract
	grl.WriteString(fmt.Sprintf("        Retract(\"%s\");\n", ruleName))
	grl.WriteString("}")

	return grl.String(), nil
}

// convertCondition 转换条件
func (c *GRLConverter) convertCondition(cond Condition, defs Definitions) (string, error) {
	switch cond.Type {
	case ConditionTypeSimple:
		return c.convertSimpleCondition(cond, defs)

	case ConditionTypeComposite:
		return c.convertCompositeCondition(cond, defs)

	case ConditionTypeExpression:
		return c.expressionParser.ParseCondition(cond.Expression)

	case ConditionTypeFunction:
		return c.convertFunctionCondition(cond, defs)

	default:
		return "", fmt.Errorf("不支持的条件类型: %s", cond.Type)
	}
}

// convertSimpleCondition 转换简单条件
func (c *GRLConverter) convertSimpleCondition(cond Condition, defs Definitions) (string, error) {
	// 左操作数
	left, err := c.convertOperand(cond.Left, defs)
	if err != nil {
		return "", fmt.Errorf("转换左操作数失败: %w", err)
	}

	// 操作符
	operator, err := c.convertOperator(string(cond.Operator), cond.Right)
	if err != nil {
		return "", fmt.Errorf("转换操作符失败: %w", err)
	}

	// 右操作数
	right, err := c.convertOperand(cond.Right, defs)
	if err != nil {
		return "", fmt.Errorf("转换右操作数失败: %w", err)
	}

	// 特殊操作符处理
	switch cond.Operator {
	case OpBetween:
		// 处理BETWEEN操作符
		if reflect.TypeOf(cond.Right).Kind() == reflect.Slice {
			values := reflect.ValueOf(cond.Right)
			if values.Len() == 2 {
				return fmt.Sprintf("%s >= %v && %s <= %v",
					left, values.Index(0).Interface(),
					left, values.Index(1).Interface()), nil
			}
		}
		return "", fmt.Errorf("between操作符需要两个值的数组")

	case "in":
		// 处理IN操作符 - 转换为Contains函数调用
		return fmt.Sprintf("Contains(%s, %s)", right, left), nil

	case "contains":
		// 处理CONTAINS操作符
		return fmt.Sprintf("Contains(%s, %s)", left, right), nil

	case "matches":
		// 处理正则匹配
		return fmt.Sprintf("Matches(%s, %s)", left, right), nil

	default:
		return fmt.Sprintf("%s %s %s", left, operator, right), nil
	}
}

// convertCompositeCondition 转换复合条件
func (c *GRLConverter) convertCompositeCondition(cond Condition, defs Definitions) (string, error) {
	if len(cond.Children) == 0 {
		return "", fmt.Errorf("复合条件必须包含子条件")
	}

	var conditions []string
	for _, child := range cond.Children {
		childCond, err := c.convertCondition(child, defs)
		if err != nil {
			return "", err
		}
		conditions = append(conditions, fmt.Sprintf("(%s)", childCond))
	}

	// 操作符
	operator := c.config.OperatorMapping[string(cond.Operator)]
	if operator == "" {
		operator = string(cond.Operator)
	}

	return strings.Join(conditions, " "+operator+" "), nil
}

// convertFunctionCondition 转换函数条件
func (c *GRLConverter) convertFunctionCondition(cond Condition, defs Definitions) (string, error) {
	// 解析函数调用
	return c.expressionParser.ParseCondition(cond.Expression)
}

// convertAction 转换动作
func (c *GRLConverter) convertAction(action Action, defs Definitions) (string, error) {
	switch action.Type {
	case ActionTypeAssign:
		// 赋值动作: target = value
		target := c.resolveTarget(action.Target)
		value := c.convertValue(action.Value)
		return fmt.Sprintf("%s = %s", target, value), nil

	case ActionTypeCalculate:
		// 计算动作: target = expression
		target := c.resolveTarget(action.Target)
		expr, err := c.expressionParser.ParseExpression(action.Expression)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s = %s", target, expr), nil

	case ActionTypeInvoke:
		// 调用动作: function(params)
		var params []string
		for key, val := range action.Parameters {
			params = append(params, fmt.Sprintf("%s=%s", key, c.convertValue(val)))
		}
		if len(params) > 0 {
			return fmt.Sprintf("%s(%s)", action.Target, strings.Join(params, ", ")), nil
		}
		return fmt.Sprintf("%s()", action.Target), nil

	case ActionTypeLog:
		// 日志动作
		return fmt.Sprintf("Log(\"%s\")", action.Value), nil

	case ActionTypeAlert:
		// 告警动作
		return fmt.Sprintf("Alert(\"%s\")", action.Value), nil

	default:
		return "", fmt.Errorf("不支持的动作类型: %s", action.Type)
	}
}

// 辅助函数

// convertOperand 转换操作数
func (c *GRLConverter) convertOperand(operand interface{}, defs Definitions) (string, error) {
	switch v := operand.(type) {
	case string:
		// 检查是否是字段引用
		if strings.Contains(v, ".") || c.isVariable(v) {
			return v, nil
		}
		// 字符串字面量
		return fmt.Sprintf("\"%s\"", v), nil

	case int, int64, float32, float64:
		return fmt.Sprintf("%v", v), nil

	case bool:
		return fmt.Sprintf("%v", v), nil

	case nil:
		return "null", nil

	default:
		// 尝试JSON序列化
		return fmt.Sprintf("%v", v), nil
	}
}

// convertOperator 转换操作符
func (c *GRLConverter) convertOperator(op string, rightOperand interface{}) (string, error) {
	if mapped, ok := c.config.OperatorMapping[op]; ok {
		return mapped, nil
	}
	return op, nil
}

// convertValue 转换值
func (c *GRLConverter) convertValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case int, int64, float32, float64, bool:
		return fmt.Sprintf("%v", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("\"%v\"", v)
	}
}

// resolveTarget 解析目标
func (c *GRLConverter) resolveTarget(target string) string {
	// 检查是否是结果字段
	if strings.HasPrefix(target, "Result.") {
		field := strings.TrimPrefix(target, "Result.")
		return fmt.Sprintf("Result[\"%s\"]", field)
	}
	return target
}

// isVariable 检查是否是变量
func (c *GRLConverter) isVariable(name string) bool {
	for prefix := range c.config.VariablePrefix {
		if strings.HasPrefix(name, prefix+".") {
			return true
		}
	}
	return false
}

// sanitizeRuleName 清理规则名称
func (c *GRLConverter) sanitizeRuleName(name string) string {
	// 移除特殊字符，只保留字母、数字和下划线
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	return reg.ReplaceAllString(name, "_")
}

// generateRuleID 生成规则ID
func (c *GRLConverter) generateRuleID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// convertStandard 转换完整标准
func (c *GRLConverter) convertStandard(standard RuleDefinitionStandard) (string, error) {
	var allRules []string

	for _, rule := range standard.Rules {
		if !rule.Enabled {
			continue
		}

		// 如果数据库Rule已包含完整的GRL，直接返回
		if rule.GRL != "" {
			allRules = append(allRules, rule.GRL)
			continue
		}

		// 否则需要从Rule转换为StandardRule再生成GRL
		// 但Rule结构不包含Conditions和Actions，这里需要特殊处理
		// 对于动态生成的场景，Rule应该包含完整信息或者使用StandardRule
		return "", fmt.Errorf("数据库Rule模型不包含足够信息进行GRL转换，请使用StandardRule或确保Rule.GRL不为空")
	}

	return strings.Join(allRules, "\n\n"), nil
}

// Validate 验证规则定义
func (c *GRLConverter) Validate(definition interface{}) error {
	switch def := definition.(type) {
	case StandardRule:
		errors := def.Validate()
		if len(errors) > 0 {
			return fmt.Errorf("规则验证失败: %s", errors[0].Message)
		}

	case *StandardRule:
		errors := def.Validate()
		if len(errors) > 0 {
			return fmt.Errorf("规则验证失败: %s", errors[0].Message)
		}

	case SimpleRule:
		if def.When == "" {
			return fmt.Errorf("简化规则的when条件不能为空")
		}
		if len(def.Then) == 0 {
			return fmt.Errorf("简化规则的then动作不能为空")
		}

	case MetricRule:
		if def.Name == "" {
			return fmt.Errorf("指标规则的名称不能为空")
		}
		if def.Formula == "" {
			return fmt.Errorf("指标规则的公式不能为空")
		}
	}

	return nil
}
