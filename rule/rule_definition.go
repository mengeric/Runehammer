package rule

import (
	"encoding/json"
	"fmt"
	"time"
)

// ============================================================================
// 规则定义标准 (Rule Definition Standard v1.0)
// ============================================================================

// RuleDefinitionStandard 规则定义标准 - 通用规则描述格式
type RuleDefinitionStandard struct {
	Version     string      `json:"version" yaml:"version"`         // 标准版本号
	Metadata    Metadata    `json:"metadata" yaml:"metadata"`       // 元数据信息
	Definitions Definitions `json:"definitions" yaml:"definitions"` // 可重用定义
	Rules       []Rule      `json:"rules" yaml:"rules"`             // 规则列表
}

// Metadata 规则元数据
type Metadata struct {
	Domain      string    `json:"domain" yaml:"domain"`           // 业务领域
	Author      string    `json:"author" yaml:"author"`           // 创建者
	CreatedAt   time.Time `json:"createdAt" yaml:"createdAt"`     // 创建时间
	UpdatedAt   time.Time `json:"updatedAt" yaml:"updatedAt"`     // 更新时间
	Description string    `json:"description" yaml:"description"` // 描述信息
	Version     string    `json:"version" yaml:"version"`         // 业务版本
}

// Definitions 可重用定义 - 变量、函数、常量定义
type Definitions struct {
	Variables map[string]interface{} `json:"variables" yaml:"variables"` // 变量定义
	Functions []FunctionDef          `json:"functions" yaml:"functions"` // 函数定义
	Constants map[string]interface{} `json:"constants" yaml:"constants"` // 常量定义
}

// FunctionDef 函数定义
type FunctionDef struct {
	Name        string   `json:"name" yaml:"name"`               // 函数名
	Parameters  []string `json:"parameters" yaml:"parameters"`   // 参数列表
	Description string   `json:"description" yaml:"description"` // 函数描述
	ReturnType  string   `json:"returnType" yaml:"returnType"`   // 返回类型
}

// StandardRule 标准规则定义
type StandardRule struct {
	ID          string      `json:"id" yaml:"id"`                   // 规则唯一标识
	Name        string      `json:"name" yaml:"name"`               // 规则名称
	Description string      `json:"description" yaml:"description"` // 规则描述
	Priority    int         `json:"priority" yaml:"priority"`       // 优先级 (salience)
	Enabled     bool        `json:"enabled" yaml:"enabled"`         // 是否启用
	Tags        []string    `json:"tags" yaml:"tags"`               // 标签
	Conditions  Condition   `json:"conditions" yaml:"conditions"`   // 条件定义
	Actions     []Action    `json:"actions" yaml:"actions"`         // 动作定义
}

// ============================================================================
// 枚举类型定义
// ============================================================================

// ConditionType 条件类型
type ConditionType string

// ConditionType 条件类型枚举
const (
	ConditionTypeSimple     ConditionType = "simple"     // 简单条件: field op value
	ConditionTypeComposite  ConditionType = "composite"  // 复合条件: 包含子条件
	ConditionTypeExpression ConditionType = "expression" // 表达式条件: 自由表达式
	ConditionTypeFunction   ConditionType = "function"   // 函数条件: 调用函数
	ConditionTypeAnd        ConditionType = "and"        // 逻辑与条件
	ConditionTypeOr         ConditionType = "or"         // 逻辑或条件
	ConditionTypeNot        ConditionType = "not"        // 逻辑非条件
)

// Operator 操作符类型
type Operator string

// Operator 操作符枚举
const (
	// 比较操作符
	OpEqual              Operator = "=="        // 等于
	OpNotEqual           Operator = "!="        // 不等于
	OpGreaterThan        Operator = ">"         // 大于
	OpLessThan           Operator = "<"         // 小于
	OpGreaterThanOrEqual Operator = ">="        // 大于等于
	OpLessThanOrEqual    Operator = "<="        // 小于等于
	
	// 逻辑操作符
	OpAnd Operator = "and"    // 与
	OpOr  Operator = "or"     // 或
	OpNot Operator = "not"    // 非
	
	// 集合操作符
	OpIn       Operator = "in"       // 包含于
	OpNotIn    Operator = "notIn"    // 不包含于
	OpContains Operator = "contains" // 包含
	OpMatches  Operator = "matches"  // 正则匹配
	OpBetween  Operator = "between"  // 范围
)

// ActionType 动作类型
type ActionType string

// ActionType 动作类型枚举
const (
	ActionTypeAssign    ActionType = "assign"    // 赋值: target = value
	ActionTypeCalculate ActionType = "calculate" // 计算: target = expression
	ActionTypeInvoke    ActionType = "invoke"    // 调用: 调用函数或方法
	ActionTypeAlert     ActionType = "alert"     // 告警: 发送告警
	ActionTypeLog       ActionType = "log"       // 日志: 记录日志
	ActionTypeStop      ActionType = "stop"      // 停止: 停止规则执行
)

// Condition 条件定义 - 支持嵌套和复合条件
type Condition struct {
	Type       ConditionType `json:"type" yaml:"type"`             // 条件类型
	Operator   Operator      `json:"operator" yaml:"operator"`     // 操作符
	Left       interface{}   `json:"left" yaml:"left"`             // 左操作数
	Right      interface{}   `json:"right" yaml:"right"`           // 右操作数
	Children   []Condition   `json:"children" yaml:"children"`     // 子条件（用于复合条件）
	Expression string        `json:"expression" yaml:"expression"` // 表达式字符串（用于复杂表达式）
}

// Action 动作定义
type Action struct {
	Type       ActionType             `json:"type" yaml:"type"`             // 动作类型
	Target     string                 `json:"target" yaml:"target"`         // 目标字段或函数
	Value      interface{}            `json:"value" yaml:"value"`           // 设置的值
	Expression string                 `json:"expression" yaml:"expression"` // 表达式
	Parameters map[string]interface{} `json:"parameters" yaml:"parameters"` // 参数
}

// ============================================================================
// 简化的规则定义格式
// ============================================================================

// SimpleRule 简化规则定义 - 用于快速定义简单规则
type SimpleRule struct {
	When string            `json:"when" yaml:"when"` // 条件表达式
	Then map[string]string `json:"then" yaml:"then"` // 结果表达式
}

// MetricRule 指标计算规则 - 专门用于指标计算
type MetricRule struct {
	Name        string            `json:"name" yaml:"name"`               // 指标名称
	Description string            `json:"description" yaml:"description"` // 描述
	Formula     string            `json:"formula" yaml:"formula"`         // 计算公式
	Variables   map[string]string `json:"variables" yaml:"variables"`     // 变量定义
	Conditions  []string          `json:"conditions" yaml:"conditions"`   // 计算条件
}

// ValidationRule 验证规则 - 专门用于数据验证
type ValidationRule struct {
	Field    string      `json:"field" yaml:"field"`       // 验证字段
	Rules    []string    `json:"rules" yaml:"rules"`       // 验证规则
	Message  string      `json:"message" yaml:"message"`   // 错误消息
	Level    string      `json:"level" yaml:"level"`       // 级别: error, warning
	Required bool        `json:"required" yaml:"required"` // 是否必填
	Default  interface{} `json:"default" yaml:"default"`   // 默认值
}

// ============================================================================
// 工厂方法和转换函数
// ============================================================================

// NewStandardRule 创建标准规则
func NewStandardRule(id, name string) *StandardRule {
	return &StandardRule{
		ID:       id,
		Name:     name,
		Priority: 50, // 默认优先级
		Enabled:  true,
		Tags:     []string{},
		Actions:  []Action{},
	}
}

// AddSimpleCondition 添加简单条件
func (r *StandardRule) AddSimpleCondition(field string, operator Operator, value interface{}) *StandardRule {
	condition := Condition{
		Type:     ConditionTypeSimple,
		Left:     field,
		Operator: operator,
		Right:    value,
	}
	
	if r.Conditions.Type == "" {
		// 第一个条件
		r.Conditions = condition
	} else {
		// 追加条件，创建复合条件
		if r.Conditions.Type != ConditionTypeComposite {
			// 将现有条件包装为复合条件
			existing := r.Conditions
			r.Conditions = Condition{
				Type:     ConditionTypeComposite,
				Operator: OpAnd,
				Children: []Condition{existing},
			}
		}
		r.Conditions.Children = append(r.Conditions.Children, condition)
	}
	
	return r
}

// AddAction 添加动作
func (r *StandardRule) AddAction(actionType ActionType, target string, value interface{}) *StandardRule {
	action := Action{
		Type:   actionType,
		Target: target,
		Value:  value,
	}
	r.Actions = append(r.Actions, action)
	return r
}

// ToJSON 转换为JSON字符串
func (r *StandardRule) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON 从JSON字符串解析
func (r *StandardRule) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), r)
}

// ============================================================================
// 验证函数
// ============================================================================

// Validate 验证规则定义的有效性
func (r *StandardRule) Validate() []ValidationError {
	var errors []ValidationError
	
	// 检查必填字段
	if r.ID == "" {
		errors = append(errors, ValidationError{
			Field:   "id",
			Message: "规则ID不能为空",
		})
	}
	
	if r.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "规则名称不能为空",
		})
	}
	
	// 验证条件
	if r.Conditions.Type == "" {
		errors = append(errors, ValidationError{
			Field:   "conditions",
			Message: "规则必须包含条件",
		})
	} else {
		errors = append(errors, validateCondition(r.Conditions)...)
	}
	
	// 验证动作
	if len(r.Actions) == 0 {
		errors = append(errors, ValidationError{
			Field:   "actions",
			Message: "规则必须包含至少一个动作",
		})
	} else {
		for i, action := range r.Actions {
			errors = append(errors, validateAction(action, i)...)
		}
	}
	
	return errors
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// validateCondition 验证条件
func validateCondition(cond Condition) []ValidationError {
	var errors []ValidationError
	
	switch cond.Type {
	case ConditionTypeSimple:
		if cond.Left == nil {
			errors = append(errors, ValidationError{
				Field:   "conditions.left",
				Message: "简单条件的左操作数不能为空",
			})
		}
		if cond.Operator == "" {
			errors = append(errors, ValidationError{
				Field:   "conditions.operator",
				Message: "简单条件的操作符不能为空",
			})
		}
		
	case ConditionTypeComposite:
		if len(cond.Children) == 0 {
			errors = append(errors, ValidationError{
				Field:   "conditions.children",
				Message: "复合条件必须包含子条件",
			})
		}
		// 递归验证子条件
		for _, child := range cond.Children {
			errors = append(errors, validateCondition(child)...)
		}
		
	case ConditionTypeExpression:
		if cond.Expression == "" {
			errors = append(errors, ValidationError{
				Field:   "conditions.expression",
				Message: "表达式条件的表达式不能为空",
			})
		}
	}
	
	return errors
}

// validateAction 验证动作
func validateAction(action Action, index int) []ValidationError {
	var errors []ValidationError
	fieldPrefix := fmt.Sprintf("actions[%d]", index)
	
	if action.Type == "" {
		errors = append(errors, ValidationError{
			Field:   fieldPrefix + ".type",
			Message: "动作类型不能为空",
		})
	}
	
	switch action.Type {
	case ActionTypeAssign, ActionTypeCalculate:
		if action.Target == "" {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".target",
				Message: "赋值和计算动作的目标不能为空",
			})
		}
		
	case ActionTypeInvoke:
		if action.Target == "" {
			errors = append(errors, ValidationError{
				Field:   fieldPrefix + ".target",
				Message: "调用动作的目标函数不能为空",
			})
		}
	}
	
	return errors
}