package runehammer

import (
	"context"
	"testing"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
)

// TestDynamicEngine 测试动态引擎
func TestDynamicEngine(t *testing.T) {
	// 创建动态引擎
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

	t.Run("执行简单规则", func(t *testing.T) {
		// 定义简单规则
		simpleRule := SimpleRule{
			When: "customer.age >= 18",
			Then: map[string]string{
				"result.eligible": "true",
				"result.message":  "\"符合条件\"",
			},
		}

		// 输入数据
		input := map[string]interface{}{
			"customer": map[string]interface{}{
				"age": 25,
			},
		}

		// 执行规则
		result, err := engine.ExecuteRuleDefinition(context.Background(), simpleRule, input)
		if err != nil {
			t.Fatalf("执行简单规则失败: %v", err)
		}

		// 验证结果
		if result == nil {
			t.Fatal("结果为空")
		}

		if eligible, ok := result["eligible"]; !ok || eligible != true {
			t.Errorf("期望 eligible=true, 实际: %v", eligible)
		}

		if message, ok := result["message"]; !ok || message != "符合条件" {
			t.Errorf("期望 message='符合条件', 实际: %v", message)
		}
	})

	t.Run("执行指标规则", func(t *testing.T) {
		// 定义指标规则
		metricRule := MetricRule{
			Name:        "customer_score",
			Description: "计算客户评分",
			Formula:     "age * 0.1 + income * 0.0001",
			Variables: map[string]string{
				"age":    "customer.age",
				"income": "customer.income",
			},
			Conditions: []string{
				"customer.age >= 18",
				"customer.income > 0",
			},
		}

		// 输入数据
		input := map[string]interface{}{
			"customer": map[string]interface{}{
				"age":    30,
				"income": 50000,
			},
		}

		// 执行规则
		result, err := engine.ExecuteRuleDefinition(context.Background(), metricRule, input)
		if err != nil {
			t.Fatalf("执行指标规则失败: %v", err)
		}

		// 验证结果
		if result == nil {
			t.Fatal("结果为空")
		}

		score, ok := result["customer_score"]
		if !ok {
			t.Error("未找到 customer_score")
		} else if scoreFloat, ok := score.(float64); !ok || scoreFloat != 8.0 {
			t.Errorf("期望 customer_score=8.0, 实际: %v", score)
		}
	})

	t.Run("执行标准规则", func(t *testing.T) {
		// 定义标准规则
		standardRule := StandardRule{
			ID:          "test_rule_1",
			Name:        "年龄验证规则",
			Description: "验证客户年龄是否符合要求",
			Priority:    100,
			Enabled:     true,
			Tags:        []string{"validation", "age"},
			Conditions: Condition{
				Type:     ConditionTypeSimple,
				Left:     "customer.age",
				Operator: ">=",
				Right:    18,
			},
			Actions: []Action{
				{
					Type:   ActionTypeAssign,
					Target: "result.valid",
					Value:  true,
				},
				{
					Type:       ActionTypeCalculate,
					Target:     "result.score",
					Expression: "customer.age * 2",
				},
			},
		}

		// 输入数据
		input := map[string]interface{}{
			"customer": map[string]interface{}{
				"age": 25,
			},
		}

		// 执行规则
		result, err := engine.ExecuteRuleDefinition(context.Background(), standardRule, input)
		if err != nil {
			t.Fatalf("执行标准规则失败: %v", err)
		}

		// 验证结果
		if result == nil {
			t.Fatal("结果为空")
		}

		if valid, ok := result["valid"]; !ok || valid != true {
			t.Errorf("期望 valid=true, 实际: %v", valid)
		}

		if score, ok := result["score"]; !ok || score != 50.0 {
			t.Errorf("期望 score=50.0, 实际: %v", score)
		}
	})

	t.Run("批量执行规则", func(t *testing.T) {
		// 定义多个规则
		rules := []interface{}{
			SimpleRule{
				When: "order.amount > 100",
				Then: map[string]string{
					"result.discount": "0.1",
				},
			},
			SimpleRule{
				When: "customer.vip == true",
				Then: map[string]string{
					"result.priority": "\"high\"",
				},
			},
			SimpleRule{
				When: "order.amount > 500",
				Then: map[string]string{
					"result.free_shipping": "true",
				},
			},
		}

		// 输入数据
		input := map[string]interface{}{
			"order": map[string]interface{}{
				"amount": 600,
			},
			"customer": map[string]interface{}{
				"vip": true,
			},
		}

		// 执行规则
		results, err := engine.ExecuteBatch(context.Background(), rules, input)
		if err != nil {
			t.Fatalf("批量执行规则失败: %v", err)
		}

		// 验证结果
		if len(results) != len(rules) {
			t.Errorf("期望 %d 个结果, 实际: %d", len(rules), len(results))
		}

		// 验证每个结果
		for i, result := range results {
			if result == nil {
				t.Errorf("第 %d 个结果为空", i)
				continue
			}

			t.Logf("第 %d 个结果: %v", i, result)
		}
	})

	t.Run("缓存测试", func(t *testing.T) {
		// 创建带缓存的引擎
		cacheEngine := NewDynamicEngine[map[string]interface{}](
			DynamicEngineConfig{
				EnableCache:  true,
				CacheTTL:     1 * time.Minute,
				MaxCacheSize: 10,
			},
		)

		rule := SimpleRule{
			When: "data.value > 10",
			Then: map[string]string{
				"result.status": "\"pass\"",
			},
		}

		input := map[string]interface{}{
			"data": map[string]interface{}{
				"value": 15,
			},
		}

		// 第一次执行（应该缓存）
		start1 := time.Now()
		result1, err := cacheEngine.ExecuteRuleDefinition(context.Background(), rule, input)
		duration1 := time.Since(start1)
		if err != nil {
			t.Fatalf("第一次执行失败: %v", err)
		}

		// 第二次执行（应该从缓存获取）
		start2 := time.Now()
		result2, err := cacheEngine.ExecuteRuleDefinition(context.Background(), rule, input)
		duration2 := time.Since(start2)
		if err != nil {
			t.Fatalf("第二次执行失败: %v", err)
		}

		// 验证结果一致
		if result1["status"] != result2["status"] {
			t.Error("缓存前后结果不一致")
		}

		// 获取缓存统计
		stats := cacheEngine.GetCacheStats()
		t.Logf("缓存统计: TotalHits=%d, Size=%d", stats.TotalHits, stats.Size)

		// 第二次执行应该更快（从缓存获取）
		if duration2 >= duration1 {
			t.Logf("警告: 缓存可能未生效，第一次: %v, 第二次: %v", duration1, duration2)
		}
	})

	t.Run("验证测试", func(t *testing.T) {
		// 创建启用验证的引擎
		validationEngine := NewDynamicEngine[map[string]interface{}](
			DynamicEngineConfig{
				StrictValidation: true,
			},
		)

		// 无效规则（缺少when条件）
		invalidRule := SimpleRule{
			When: "", // 空条件
			Then: map[string]string{
				"result.test": "\"value\"",
			},
		}

		input := map[string]interface{}{
			"data": map[string]interface{}{
				"value": 1,
			},
		}

		// 执行应该失败
		_, err := validationEngine.ExecuteRuleDefinition(context.Background(), invalidRule, input)
		if err == nil {
			t.Error("期望验证失败，但执行成功")
		}
	})

	t.Run("自定义函数测试", func(t *testing.T) {
		// 注册自定义函数
		engine.RegisterCustomFunction("CustomDoubleValue", func(x float64) float64 {
			return x * 2
		})

		engine.RegisterCustomFunctions(map[string]interface{}{
			"CustomAddTen": func(x float64) float64 {
				return x + 10
			},
		})

		// 使用自定义函数的规则
		rule := SimpleRule{
			When: "data.value > 5",
			Then: map[string]string{
				"result.doubled": "CustomDoubleValue(data.value)",
				"result.added":   "CustomAddTen(data.value)",
			},
		}

		input := map[string]interface{}{
			"data": map[string]interface{}{
				"value": 15.0,
			},
		}

		result, err := engine.ExecuteRuleDefinition(context.Background(), rule, input)
		if err != nil {
			t.Fatalf("执行自定义函数规则失败: %v", err)
		}

		if doubled, ok := result["doubled"]; !ok || doubled != 30.0 {
			t.Errorf("期望 doubled=30.0, 实际: %v", doubled)
		}

		if added, ok := result["added"]; !ok || added != 25.0 {
			t.Errorf("期望 added=25.0, 实际: %v", added)
		}
	})
}

// TestRuleConverter 测试规则转换器
func TestRuleConverter(t *testing.T) {
	converter := NewGRLConverter()

	t.Run("转换简单规则", func(t *testing.T) {
		rule := SimpleRule{
			When: "customer.age >= 18 and customer.income > 30000",
			Then: map[string]string{
				"result.eligible": "true",
				"result.level":    "\"gold\"",
			},
		}

		grl, err := converter.ConvertSimpleRule(rule)
		if err != nil {
			t.Fatalf("转换简单规则失败: %v", err)
		}

		if grl == "" {
			t.Error("生成的GRL为空")
		}

		t.Logf("生成的GRL:\n%s", grl)

		// 验证GRL包含关键元素
		if !contains(grl, "rule") {
			t.Error("GRL应包含rule关键字")
		}
		if !contains(grl, "when") {
			t.Error("GRL应包含when关键字")
		}
		if !contains(grl, "then") {
			t.Error("GRL应包含then关键字")
		}
	})

	t.Run("转换指标规则", func(t *testing.T) {
		rule := MetricRule{
			Name:        "risk_score",
			Description: "风险评分计算",
			Formula:     "age * 0.1 + debt_ratio * 0.5",
			Variables: map[string]string{
				"age":        "customer.age",
				"debt_ratio": "customer.debt / customer.income",
			},
			Conditions: []string{
				"customer.age >= 18",
				"customer.income > 0",
			},
		}

		grl, err := converter.ConvertMetricRule(rule)
		if err != nil {
			t.Fatalf("转换指标规则失败: %v", err)
		}

		if grl == "" {
			t.Error("生成的GRL为空")
		}

		t.Logf("生成的GRL:\n%s", grl)

		// 验证GRL包含关键元素
		if !contains(grl, "risk_score") {
			t.Error("GRL应包含指标名称")
		}
	})

	t.Run("转换标准规则", func(t *testing.T) {
		rule := StandardRule{
			ID:          "std_rule_1",
			Name:        "标准规则测试",
			Description: "测试标准规则转换",
			Priority:    80,
			Enabled:     true,
			Conditions: Condition{
				Type:     ConditionTypeComposite,
				Operator: "and",
				Children: []Condition{
					{
						Type:     ConditionTypeSimple,
						Left:     "user.age",
						Operator: ">=",
						Right:    21,
					},
					{
						Type:     ConditionTypeSimple,
						Left:     "user.country",
						Operator: "==",
						Right:    "CN",
					},
				},
			},
			Actions: []Action{
				{
					Type:   ActionTypeAssign,
					Target: "result.approved",
					Value:  true,
				},
			},
		}

		grl, err := converter.ConvertRule(rule, Definitions{})
		if err != nil {
			t.Fatalf("转换标准规则失败: %v", err)
		}

		if grl == "" {
			t.Error("生成的GRL为空")
		}

		t.Logf("生成的GRL:\n%s", grl)

		// 验证GRL包含关键元素
		if !contains(grl, "std_rule_1") {
			t.Error("GRL应包含规则ID")
		}
		if !contains(grl, "salience 80") {
			t.Error("GRL应包含优先级")
		}
	})
}

// TestExpressionParser 测试表达式解析器
func TestExpressionParser(t *testing.T) {
	parser := NewExpressionParser()

	t.Run("SQL语法解析", func(t *testing.T) {
		parser.SetSyntax(SyntaxTypeSQL)

		testCases := []struct {
			input    string
			expected string
		}{
			{
				"age >= 18 AND income > 30000",
				"age >= 18 && income > 30000",
			},
			{
				"name LIKE '%张%' OR phone IS NOT NULL",
				"name  Matches  '%张%' || phone != null",
			},
			{
				"score BETWEEN 60 AND 100",
				"score >= 60 && score <= 100",
			},
		}

		for _, tc := range testCases {
			result, err := parser.ParseCondition(tc.input)
			if err != nil {
				t.Errorf("解析失败 '%s': %v", tc.input, err)
				continue
			}

			if !contains(result, "&&") && contains(tc.expected, "&&") {
				t.Errorf("解析结果不包含期望的&&操作符: %s", result)
			}

			t.Logf("输入: %s\n输出: %s\n期望: %s\n", tc.input, result, tc.expected)
		}
	})
}

// TestBuiltinFunctions 测试内置函数
func TestBuiltinFunctions(t *testing.T) {
	// 创建模拟的数据上下文
	dataCtx := ast.NewDataContext()
	
	// 创建引擎实例用于注入函数
	engine := &engineImpl[map[string]interface{}]{}
	
	// 注入内置函数
	engine.injectBuiltinFunctions(dataCtx)

	t.Run("数学函数测试", func(t *testing.T) {
		// 测试数学函数是否正确注入
		mathFunctions := []string{
			"Abs", "Max", "Min", "Round", "Floor", "Ceil",
			"Pow", "Sqrt", "Sin", "Cos", "Tan", "Log", "Log10",
			"Sum", "Avg", "MaxSlice", "MinSlice",
		}

		for _, funcName := range mathFunctions {
			if dataCtx.Get(funcName) == nil {
				t.Errorf("数学函数 %s 未正确注入", funcName)
			}
		}
	})

	t.Run("字符串函数测试", func(t *testing.T) {
		stringFunctions := []string{
			"Contains", "HasPrefix", "HasSuffix", "Len",
			"ToUpper", "ToLower", "Split", "Join", "Replace", "TrimSpace",
		}

		for _, funcName := range stringFunctions {
			if dataCtx.Get(funcName) == nil {
				t.Errorf("字符串函数 %s 未正确注入", funcName)
			}
		}
	})

	t.Run("验证函数测试", func(t *testing.T) {
		validationFunctions := []string{
			"Matches", "IsEmail", "IsPhoneNumber", "IsIDCard",
			"Between", "LengthBetween",
		}

		for _, funcName := range validationFunctions {
			if dataCtx.Get(funcName) == nil {
				t.Errorf("验证函数 %s 未正确注入", funcName)
			}
		}
	})
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		   (s == substr || (len(s) >= len(substr) && 
		    indexInString(s, substr) >= 0))
}

func indexInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}