package rule

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// 表达式解析器 - 解析各种语法的表达式
// ============================================================================

// ExpressionParser 表达式解析器接口
type ExpressionParser interface {
	// ParseCondition 解析条件表达式
	ParseCondition(expr string) (string, error)

	// ParseExpression 解析普通表达式
	ParseExpression(expr string) (string, error)

	// ParseAction 解析动作表达式
	ParseAction(target, expr string) (string, error)

	// SetSyntax 设置语法类型
	SetSyntax(syntax SyntaxType)
}

// SyntaxType 语法类型
type SyntaxType string

const (
	SyntaxTypeSQL        SyntaxType = "sql"        // SQL-like语法
	SyntaxTypeJavaScript SyntaxType = "javascript" // JavaScript-like语法
)

// DefaultExpressionParser 默认表达式解析器
type DefaultExpressionParser struct {
	syntax    SyntaxType
	operators map[string]string
	functions map[string]string
	keywords  map[string]string
}

// NewExpressionParser 创建表达式解析器
func NewExpressionParser(syntax ...SyntaxType) ExpressionParser {
	syntaxType := SyntaxTypeSQL // 默认SQL语法
	if len(syntax) > 0 {
		syntaxType = syntax[0]
	}

	parser := &DefaultExpressionParser{
		syntax:    syntaxType,
		operators: map[string]string{},
		functions: map[string]string{},
		keywords:  map[string]string{},
	}

	return parser
}

// ParseCondition 解析条件表达式
func (p *DefaultExpressionParser) ParseCondition(expr string) (string, error) {
	if expr == "" {
		return "", fmt.Errorf("条件表达式不能为空")
	}

	// 根据语法类型选择解析策略
	switch p.syntax {
	case SyntaxTypeSQL:
		return p.parseSQLCondition(expr)
	case SyntaxTypeJavaScript:
		return p.parseJSCondition(expr)
	default:
		return "", fmt.Errorf("不支持的语法类型: %s", p.syntax)
	}
}

// ParseExpression 解析普通表达式
func (p *DefaultExpressionParser) ParseExpression(expr string) (string, error) {
	if expr == "" {
		return "", fmt.Errorf("表达式不能为空")
	}

	// 通用表达式解析
	result := expr

	// 替换函数
	for chinese, english := range p.functions {
		result = strings.ReplaceAll(result, chinese, english)
	}

	// 替换操作符
	for chinese, english := range p.operators {
		result = strings.ReplaceAll(result, chinese, english)
	}

	// 处理三元运算符 condition ? value1 : value2
	if matched := p.parseTernaryOperator(result); matched != "" {
		result = matched
	}

	return result, nil
}

// ParseAction 解析动作表达式
func (p *DefaultExpressionParser) ParseAction(target, expr string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("动作目标不能为空")
	}

	// 解析表达式
	parsedExpr, err := p.ParseExpression(expr)
	if err != nil {
		return "", err
	}

	// 生成赋值语句
	resolvedTarget := p.resolveTarget(target)
	return fmt.Sprintf("%s = %s", resolvedTarget, parsedExpr), nil
}

// SetSyntax 设置语法类型
func (p *DefaultExpressionParser) SetSyntax(syntax SyntaxType) {
	p.syntax = syntax
}

// ============================================================================
// 各种语法的解析实现
// ============================================================================

// parseSQLCondition 解析SQL-like条件
func (p *DefaultExpressionParser) parseSQLCondition(expr string) (string, error) {
	result := expr

	// 替换SQL关键词
	replacements := map[string]string{
		" AND ":         " && ",
		" and ":         " && ",
		" OR ":          " || ",
		" or ":          " || ",
		" NOT ":         " !",
		" not ":         " !",
		" BETWEEN ":     " >= ",
		" between ":     " >= ",
		" IN ":          " Contains ",
		" in ":          " Contains ",
		" LIKE ":        " Matches ",
		" like ":        " Matches ",
		" IS NULL ":     " == null",
		" is null ":     " == null",
		" IS NOT NULL ": " != null",
		" is not null ": " != null",
	}

	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 处理BETWEEN操作符: field BETWEEN min AND max -> field >= min && field <= max
	betweenRegex := regexp.MustCompile(`(\w+(?:\.\w+)*)\s+>=\s+(\d+(?:\.\d+)?)\s+AND\s+(\d+(?:\.\d+)?)`)
	result = betweenRegex.ReplaceAllString(result, "$1 >= $2 && $1 <= $3")

	// 处理IN操作符: field Contains (val1, val2, val3) -> (Contains([val1, val2, val3], field))
	inRegex := regexp.MustCompile(`(\w+(?:\.\w+)*)\s+Contains\s+\(([^)]+)\)`)
	result = inRegex.ReplaceAllStringFunc(result, func(match string) string {
		parts := inRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		field := parts[1]
		values := strings.Split(parts[2], ",")
		var cleanValues []string
		for _, v := range values {
			cleanValues = append(cleanValues, strings.TrimSpace(v))
		}
		return fmt.Sprintf("Contains([%s], %s)", strings.Join(cleanValues, ", "), field)
	})

	// 基本语法验证
	if err := p.validateSQLSyntax(result); err != nil {
		return "", err
	}

	return result, nil
}

// parseJSCondition 解析JavaScript-like条件
func (p *DefaultExpressionParser) parseJSCondition(expr string) (string, error) {
	result := expr

	// JavaScript操作符映射
	jsReplacements := map[string]string{
		"===": "==",
		"!==": "!=",
		"&&":  "&&",
		"||":  "||",
		"!":   "!",
	}

	for old, new := range jsReplacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// 处理数组方法调用
	// orders.filter(o => o.amount > 100).length
	// 转换为: Count(Filter(orders, "amount > 100"))
	filterRegex := regexp.MustCompile(`(\w+)\.filter\((\w+)\s*=>\s*([^)]+)\)\.length`)
	result = filterRegex.ReplaceAllString(result, "Count(Filter($1, \"$3\"))")

	// 处理map方法
	mapRegex := regexp.MustCompile(`(\w+)\.map\((\w+)\s*=>\s*([^)]+)\)`)
	result = mapRegex.ReplaceAllString(result, "Map($1, \"$3\")")

	return result, nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// parseTernaryOperator 解析三元运算符
func (p *DefaultExpressionParser) parseTernaryOperator(expr string) string {
	// 匹配 condition ? value1 : value2 格式
	ternaryRegex := regexp.MustCompile(`([^?]+)\?([^:]+):(.+)`)
	matches := ternaryRegex.FindStringSubmatch(expr)

	if len(matches) == 4 {
		condition := strings.TrimSpace(matches[1])
		trueValue := strings.TrimSpace(matches[2])
		falseValue := strings.TrimSpace(matches[3])

		// 在GRL中使用条件表达式
		return fmt.Sprintf("(%s) ? %s : %s", condition, trueValue, falseValue)
	}

	return ""
}

// resolveTarget 解析目标字段
func (p *DefaultExpressionParser) resolveTarget(target string) string {
    // 处理结果字段
    if strings.HasPrefix(target, "Result.") || strings.HasPrefix(target, "result.") {
        field := strings.TrimPrefix(strings.TrimPrefix(target, "Result."), "result.")
        return fmt.Sprintf("Result[\"%s\"]", field)
    }

    return target
}

// parseNumber 解析数字
func (p *DefaultExpressionParser) parseNumber(s string) (float64, error) {
	// 移除千分位分隔符
	s = strings.ReplaceAll(s, ",", "")
	return strconv.ParseFloat(s, 64)
}

// parseDate 解析日期
func (p *DefaultExpressionParser) parseDate(s string) (time.Time, error) {
	// 支持多种日期格式
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"01/02/2006",
		"02-01-2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期: %s", s)
}

// isStringLiteral 检查是否是字符串字面量
func (p *DefaultExpressionParser) isStringLiteral(s string) bool {
	return (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'"))
}

// isNumberLiteral 检查是否是数字字面量
func (p *DefaultExpressionParser) isNumberLiteral(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// isBooleanLiteral 检查是否是布尔字面量
func (p *DefaultExpressionParser) isBooleanLiteral(s string) bool {
	return s == "true" || s == "false"
}

// normalizeBooleanLiteral 标准化布尔字面量
func (p *DefaultExpressionParser) normalizeBooleanLiteral(s string) string {
	return s
}

// validateSQLSyntax 验证SQL表达式语法
func (p *DefaultExpressionParser) validateSQLSyntax(expr string) error {
	expr = strings.TrimSpace(expr)
	
	// 检查空表达式
	if expr == "" {
		return fmt.Errorf("表达式不能为空")
	}
	
	// 检查是否以操作符开始
	invalidStarts := []string{"&&", "||", ">=", "<=", "==", "!=", ">", "<", "AND", "OR"}
	for _, start := range invalidStarts {
		if strings.HasPrefix(expr, start) {
			return fmt.Errorf("表达式不能以操作符开始: %s", start)
		}
	}
	
	// 检查是否以操作符结束
	invalidEnds := []string{"&&", "||", ">=", "<=", "==", "!=", ">", "<", "AND", "OR"}
	for _, end := range invalidEnds {
		if strings.HasSuffix(expr, " "+end) || strings.HasSuffix(expr, end+" ") {
			return fmt.Errorf("表达式不能以操作符结束: %s", end)
		}
	}
	
	return nil
}
