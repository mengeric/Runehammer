package runehammer

import (
	"context"
	"fmt"
	"time"
)

// ExampleRulesUsage 演示自定义规则的基本使用方式
func ExampleRulesUsage() {
	fmt.Println("=== Runehammer 自定义规则使用示例 ===")
	
	// ============================================================================
	// 示例1: 动态引擎 - 简单规则
	// ============================================================================
	
	fmt.Println("\n--- 动态引擎简单规则示例 ---")
	
	// 创建动态引擎
	engine := NewDynamicEngine[map[string]interface{}](
		DynamicEngineConfig{
			EnableCache:       true,
			CacheTTL:          5 * time.Minute,
			MaxCacheSize:      100,
			StrictValidation:  false, // 放宽验证以便示例运行
			ParallelExecution: false, // 简化示例
			DefaultTimeout:    10 * time.Second,
		},
	)
	
	// 定义简单的年龄验证规则
	ageRule := SimpleRule{
		When: "Params >= 18", // 对于基本类型输入，使用Params访问
		Then: map[string]string{
			"Result.Adult":   "true",
			"Result.Message": "\"符合年龄要求\"",
		},
	}
	
	// 执行规则
	result, err := engine.ExecuteRuleDefinition(context.Background(), ageRule, 25)
	if err != nil {
		fmt.Printf("❌ 执行年龄规则失败: %v\n", err)
	} else {
		fmt.Printf("✅ 年龄验证结果: %+v\n", result)
	}
	
	// ============================================================================
	// 示例2: 自定义函数注册
	// ============================================================================
	
	fmt.Println("\n--- 自定义函数示例 ---")
	
	// 注册自定义函数
	engine.RegisterCustomFunction("IsAdult", func(age int) bool {
		return age >= 18
	})
	
	engine.RegisterCustomFunction("CalculateDiscount", func(amount float64, rate float64) float64 {
		return amount * rate
	})
	
	// 使用自定义函数的规则
	customFuncRule := SimpleRule{
		When: "IsAdult(Params)",
		Then: map[string]string{
			"Result.Adult":    "true",
			"Result.Discount": "CalculateDiscount(100.0, 0.1)",
		},
	}
	
	result, err = engine.ExecuteRuleDefinition(context.Background(), customFuncRule, 25)
	if err != nil {
		fmt.Printf("❌ 执行自定义函数规则失败: %v\n", err)
	} else {
		fmt.Printf("✅ 自定义函数结果: %+v\n", result)
	}
	
	// ============================================================================
	// 示例3: 字符串输入示例
	// ============================================================================
	
	fmt.Println("\n--- 字符串输入示例 ---")
	
	stringRule := SimpleRule{
		When: "Params == \"VIP\"",
		Then: map[string]string{
			"Result.IsVip":    "true",
			"Result.Privilege": "\"高级权限\"",
		},
	}
	
	result, err = engine.ExecuteRuleDefinition(context.Background(), stringRule, "VIP")
	if err != nil {
		fmt.Printf("❌ 执行字符串规则失败: %v\n", err)
	} else {
		fmt.Printf("✅ 字符串规则结果: %+v\n", result)
	}
	
	// ============================================================================
	// 示例4: 批量规则执行
	// ============================================================================
	
	fmt.Println("\n--- 批量规则执行示例 ---")
	
	batchRules := []interface{}{
		SimpleRule{
			When: "Params > 100",
			Then: map[string]string{
				"Result.LargeAmount": "true",
			},
		},
		SimpleRule{
			When: "Params <= 100",
			Then: map[string]string{
				"Result.SmallAmount": "true",
			},
		},
	}
	
	results, err := engine.ExecuteBatch(context.Background(), batchRules, 150)
	if err != nil {
		fmt.Printf("❌ 批量执行失败: %v\n", err)
	} else {
		fmt.Println("✅ 批量执行结果:")
		for i, result := range results {
			fmt.Printf("   规则%d: %+v\n", i+1, result)
		}
	}
	
	fmt.Println("\n=== 示例完成 ===")
}

// ExampleUniversalEngineUsage 演示通用引擎的使用方式
func ExampleUniversalEngineUsage() {
	fmt.Println("=== 通用引擎使用示例 ===")
	
	// 创建BaseEngine实例（仅需一个）
	baseEngine, err := NewBaseEngine(
		WithDSN("sqlite:file:example.db?mode=memory&cache=shared&_fk=1"),
		WithAutoMigrate(),
		WithLogger(NewNoopLogger()),
	)
	if err != nil {
		fmt.Printf("❌ 创建BaseEngine失败: %v\n", err)
		return
	}
	defer baseEngine.Close()
	
	// 创建不同类型的TypedEngine包装器
	mapEngine := NewTypedEngine[map[string]interface{}](baseEngine)
	
	// 测试数据
	testData := map[string]interface{}{
		"age":    25,
		"income": 80000.0,
		"vip":    true,
	}
	
	// 执行规则（注意：实际使用时规则需要在数据库中存在）
	result, err := mapEngine.Exec(context.Background(), "TEST_RULE", testData)
	if err != nil {
		fmt.Printf("❌ 执行规则失败（这是预期的，因为数据库中没有规则）: %v\n", err)
	} else {
		fmt.Printf("✅ 执行结果: %+v\n", result)
	}
	
	fmt.Println("✅ 通用引擎创建成功，可以通过TypedEngine包装器支持多种返回类型")
	fmt.Println("=== 通用引擎示例完成 ===")
}