package runehammer

import (
	"context"
	"fmt"
	"log"
)

// 业务结果类型定义
type UserValidationResult struct {
	Adult    bool   `json:"adult"`
	Eligible bool   `json:"eligible"`
	Level    string `json:"level"`
}

type OrderProcessResult struct {
	Discount float64 `json:"discount"`
	Status   string  `json:"status"`
	Priority int     `json:"priority"`
}

// 输入数据类型
type UserData struct {
	Age    int     `json:"age"`
	Income float64 `json:"income"`
	VIP    bool    `json:"vip"`
}

// ExampleUniversalEngine 展示通用引擎功能的示例函数
func ExampleUniversalEngine() {
	fmt.Println("=== Runehammer 通用引擎示例 ===")
	
	// ============================================================================
	// 启动时创建单个BaseEngine实例
	// ============================================================================
	
	fmt.Println("\n1. 创建通用BaseEngine实例...")
	baseEngine, err := NewBaseEngine(
		WithDSN("file:example.db?mode=memory&cache=shared&_fk=1"),
		WithAutoMigrate(),
		WithLogger(NewNoopLogger()),
	)
	if err != nil {
		log.Fatal("创建BaseEngine失败:", err)
	}
	defer baseEngine.Close()
	
	fmt.Println("✅ BaseEngine创建成功")
	
	// ============================================================================
	// 运行时创建不同类型的TypedEngine包装器
	// ============================================================================
	
	fmt.Println("\n2. 创建不同类型的TypedEngine包装器...")
	
	// 用户验证引擎 - 返回强类型结构体
	userEngine := NewTypedEngine[UserValidationResult](baseEngine)
	
	// 订单处理引擎 - 返回强类型结构体  
	orderEngine := NewTypedEngine[OrderProcessResult](baseEngine)
	
	// 通用map引擎 - 返回灵活的map类型
	mapEngine := NewTypedEngine[map[string]interface{}](baseEngine)
	
	fmt.Println("✅ TypedEngine包装器创建成功")
	
	// ============================================================================
	// 设置测试规则（实际项目中会从数据库加载）
	// ============================================================================
	
	// 注意：这里为演示目的直接设置，实际使用时规则从数据库加载
	fmt.Println("\n3. 设置测试规则...")
	setupTestRules(baseEngine)
	fmt.Println("✅ 测试规则设置完成")
	
	// ============================================================================
	// 测试数据
	// ============================================================================
	
	testUsers := []UserData{
		{Age: 17, Income: 20000, VIP: false}, // 未成年
		{Age: 20, Income: 40000, VIP: false}, // 成年但不符合条件
		{Age: 25, Income: 60000, VIP: true},  // VIP用户
		{Age: 30, Income: 80000, VIP: false}, // 高收入用户
	}
	
	ctx := context.Background()
	
	// ============================================================================
	// 演示：同一个BaseEngine支持多种返回类型
	// ============================================================================
	
	fmt.Println("\n4. 演示多种返回类型...")
	
	for i, user := range testUsers {
		fmt.Printf("\n--- 用户 %d: Age=%d, Income=%.0f, VIP=%v ---\n", 
			i+1, user.Age, user.Income, user.VIP)
		
		// 用户验证 - 强类型结构体结果
		userResult, err := userEngine.Exec(ctx, "USER_VALIDATE", user)
		if err != nil {
			fmt.Printf("❌ 用户验证失败: %v\n", err)
			continue
		}
		fmt.Printf("👤 用户验证结果: Adult=%v, Eligible=%v, Level=%s\n", 
			userResult.Adult, userResult.Eligible, userResult.Level)
		
		// 订单处理 - 强类型结构体结果
		orderResult, err := orderEngine.Exec(ctx, "ORDER_PROCESS", user)
		if err != nil {
			fmt.Printf("❌ 订单处理失败: %v\n", err)
			continue
		}
		fmt.Printf("🛒 订单处理结果: Discount=%.2f, Status=%s, Priority=%d\n", 
			orderResult.Discount, orderResult.Status, orderResult.Priority)
		
		// 通用map - 灵活的map结果
		mapResult, err := mapEngine.Exec(ctx, "USER_VALIDATE", user)
		if err != nil {
			fmt.Printf("❌ 通用执行失败: %v\n", err)
			continue
		}
		fmt.Printf("🗂️  通用map结果: %+v\n", mapResult)
	}
	
	// ============================================================================
	// 演示：性能优势
	// ============================================================================
	
	fmt.Println("\n5. 性能优势演示...")
	fmt.Println("传统方式: 每种返回类型需要一个引擎实例")
	fmt.Println("新方式: 一个BaseEngine + 多个轻量级TypedEngine包装器")
	fmt.Println("优势: 资源共享、配置统一、管理简单")
	
	fmt.Println("\n=== 示例完成 ===")
}

// setupTestRules 设置测试规则（演示用）
func setupTestRules(baseEngine BaseEngine) {
	// 实际项目中，这些规则会存储在数据库中
	// 这里为了演示目的，我们假设已经有了规则
	// 
	// 规则示例：
	// USER_VALIDATE: 用户验证规则
	//   - result["adult"] = age >= 18
	//   - result["eligible"] = age >= 21 AND income > 50000
	//   - result["level"] = VIP用户为"VIP"，否则根据收入确定
	//
	// ORDER_PROCESS: 订单处理规则  
	//   - result["discount"] = VIP用户0.1，高收入用户0.05，其他0
	//   - result["status"] = VIP用户"VIP"，其他"NORMAL"
	//   - result["priority"] = VIP用户1，高收入用户2，其他3
	
	fmt.Println("   📋 用户验证规则: 年龄判断 + 资格验证 + 等级评定")
	fmt.Println("   📋 订单处理规则: 折扣计算 + 状态设置 + 优先级分配")
}