package runehammer

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// 测试用结构体定义
type UserResult struct {
	Adult    bool    `json:"adult"`
	Eligible bool    `json:"eligible"`
	Score    int     `json:"score"`
	Level    string  `json:"level"`
}

type OrderResult struct {
	Discount float64 `json:"discount"`
	Status   string  `json:"status"`
	Priority int     `json:"priority"`
}

type UserInput struct {
	Age    int     `json:"age"`
	Income float64 `json:"income"`
	VIP    bool    `json:"vip"`
}

// TestUniversalEngine 测试通用引擎功能
func TestUniversalEngine(t *testing.T) {
	Convey("通用引擎测试", t, func() {
		
		Convey("创建BaseEngine实例", func() {
			// 使用Mock数据进行测试
			mapper := newMockRuleMapper()
			
			// 添加测试规则
			rules := []*Rule{
				{
					ID:      1,
					BizCode: "USER_VALIDATE",
					Name:    "用户验证规则",
					GRL:     `rule UserValidation "用户验证" { when userinput.Age >= 18 then Result["Adult"] = true; Result["Eligible"] = userinput.Age >= 21; }`,
					Enabled: true,
				},
				{
					ID:      2,
					BizCode: "ORDER_PROCESS",
					Name:    "订单处理规则",
					GRL:     `rule OrderProcess "订单处理" { when userinput.VIP == true then Result["Discount"] = 0.1; Result["Status"] = "VIP"; }`,
					Enabled: true,
				},
			}
			mapper.SetRules("USER_VALIDATE", rules[:1])
			mapper.SetRules("ORDER_PROCESS", rules[1:])
			
			// 创建BaseEngine - 启动时只需要一个实例
			baseEngine, err := NewBaseEngine(
				WithDSN("sqlite:file:test.db?mode=memory&cache=shared&_fk=1"),
				WithAutoMigrate(),
				WithLogger(NewNoopLogger()),
			)
			So(err, ShouldBeNil)
			So(baseEngine, ShouldNotBeNil)
			defer baseEngine.Close()
			
			// 手动设置mapper到wrapper内部引擎
			if wrapper, ok := baseEngine.(*baseEngineWrapper); ok {
				if engineImpl, ok := wrapper.engine.(*engineImpl[map[string]interface{}]); ok {
					engineImpl.mapper = mapper
				}
			}
			
			Convey("测试原始执行功能", func() {
				input := UserInput{Age: 25, Income: 50000, VIP: true}
				
				// 执行用户验证规则
				result, err := baseEngine.ExecRaw(context.Background(), "USER_VALIDATE", input)
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result["Adult"], ShouldEqual, true)
				So(result["Eligible"], ShouldEqual, true)
			})
			
			Convey("测试TypedEngine - UserResult", func() {
				// 创建用户结果类型的引擎
				userEngine := NewTypedEngine[UserResult](baseEngine)
				
				input := UserInput{Age: 19, Income: 30000, VIP: false}
				result, err := userEngine.Exec(context.Background(), "USER_VALIDATE", input)
				
				So(err, ShouldBeNil)
				So(result.Adult, ShouldEqual, true)
				So(result.Eligible, ShouldEqual, false) // 19 < 21
			})
			
			Convey("测试TypedEngine - OrderResult", func() {
				// 创建订单结果类型的引擎
				orderEngine := NewTypedEngine[OrderResult](baseEngine)
				
				input := UserInput{Age: 30, Income: 80000, VIP: true}
				result, err := orderEngine.Exec(context.Background(), "ORDER_PROCESS", input)
				
				So(err, ShouldBeNil)
				So(result.Discount, ShouldEqual, 0.1)
				So(result.Status, ShouldEqual, "VIP")
			})
			
			Convey("测试同一BaseEngine支持多种类型", func() {
				// 同一个BaseEngine实例可以用于多种类型
				userEngine := NewTypedEngine[UserResult](baseEngine)
				orderEngine := NewTypedEngine[OrderResult](baseEngine)
				mapEngine := NewTypedEngine[map[string]interface{}](baseEngine)
				
				input := UserInput{Age: 22, Income: 60000, VIP: true}
				
				// 用户验证
				userResult, err := userEngine.Exec(context.Background(), "USER_VALIDATE", input)
				So(err, ShouldBeNil)
				So(userResult.Adult, ShouldEqual, true)
				So(userResult.Eligible, ShouldEqual, true)
				
				// 订单处理  
				orderResult, err := orderEngine.Exec(context.Background(), "ORDER_PROCESS", input)
				So(err, ShouldBeNil)
				So(orderResult.Discount, ShouldEqual, 0.1)
				So(orderResult.Status, ShouldEqual, "VIP")
				
				// 通用map类型
				mapResult, err := mapEngine.Exec(context.Background(), "USER_VALIDATE", input)
				So(err, ShouldBeNil)
				So(mapResult["adult"], ShouldEqual, true)
				So(mapResult["eligible"], ShouldEqual, true)
			})
			
			Convey("测试类型转换错误处理", func() {
				// 定义一个无法从map转换的复杂结构体
				type ComplexStruct struct {
					Channel chan int `json:"-"` // 无法JSON序列化的字段
				}
				
				complexEngine := NewTypedEngine[ComplexStruct](baseEngine)
				input := UserInput{Age: 25, Income: 50000, VIP: true}
				
				result, err := complexEngine.Exec(context.Background(), "USER_VALIDATE", input)
				// 应该能转换成功，因为Channel字段会被忽略
				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
			})
		})
		
		Convey("性能对比测试", func() {
			mapper1 := newMockRuleMapper()
			mapper2 := newMockRuleMapper()
			mapper3 := newMockRuleMapper()
			
			rules := []*Rule{
				{
					ID:      1,
					BizCode: "PERF_TEST",
					Name:    "性能测试规则",
					GRL:     `rule PerfTest "性能测试" { when userinput.Age >= 18 then Result["Adult"] = true; Result["Score"] = userinput.Age * 2; }`,
					Enabled: true,
				},
			}
			mapper1.SetRules("PERF_TEST", rules)
			mapper2.SetRules("PERF_TEST", rules)
			mapper3.SetRules("PERF_TEST", rules)
			
			// 传统方式 - 每种类型一个引擎实例
			userEngine, err := New[UserResult](
				WithDSN("sqlite:file:test_user.db?mode=memory&cache=shared&_fk=1"),
				WithAutoMigrate(),
				WithLogger(NewNoopLogger()),
			)
			So(err, ShouldBeNil)
			defer userEngine.Close()
			
			mapEngine, err := New[map[string]interface{}](
				WithDSN("sqlite:file:test_map.db?mode=memory&cache=shared&_fk=1"),
				WithAutoMigrate(),
				WithLogger(NewNoopLogger()),
			)
			So(err, ShouldBeNil)
			defer mapEngine.Close()
			
			// 手动设置mapper
			if engineImpl, ok := userEngine.(*engineImpl[UserResult]); ok {
				engineImpl.mapper = mapper1
			}
			if engineImpl, ok := mapEngine.(*engineImpl[map[string]interface{}]); ok {
				engineImpl.mapper = mapper2
			}
			
			// 新方式 - 一个BaseEngine + 多个TypedEngine包装器
			baseEngine, err := NewBaseEngine(
				WithDSN("sqlite:file:test_base.db?mode=memory&cache=shared&_fk=1"),
				WithAutoMigrate(),
				WithLogger(NewNoopLogger()),
			)
			So(err, ShouldBeNil)
			defer baseEngine.Close()
			
			// 手动设置mapper
			if wrapper, ok := baseEngine.(*baseEngineWrapper); ok {
				if engineImpl, ok := wrapper.engine.(*engineImpl[map[string]interface{}]); ok {
					engineImpl.mapper = mapper3
				}
			}
			
			universalUserEngine := NewTypedEngine[UserResult](baseEngine)
			universalMapEngine := NewTypedEngine[map[string]interface{}](baseEngine)
			
			input := UserInput{Age: 25, Income: 50000, VIP: true}
			ctx := context.Background()
			
			Convey("验证结果一致性", func() {
				// 传统方式结果
				traditionalUserResult, err1 := userEngine.Exec(ctx, "PERF_TEST", input)
				traditionalMapResult, err2 := mapEngine.Exec(ctx, "PERF_TEST", input)
				
				// 新方式结果
				universalUserResult, err3 := universalUserEngine.Exec(ctx, "PERF_TEST", input)
				universalMapResult, err4 := universalMapEngine.Exec(ctx, "PERF_TEST", input)
				
				// 验证所有执行都成功
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(err3, ShouldBeNil)
				So(err4, ShouldBeNil)
				
				// 验证结果一致
				So(traditionalUserResult.Adult, ShouldEqual, universalUserResult.Adult)
				So(traditionalUserResult.Score, ShouldEqual, universalUserResult.Score)
				So(traditionalMapResult["adult"], ShouldEqual, universalMapResult["adult"])
				So(traditionalMapResult["score"], ShouldEqual, universalMapResult["score"])
			})
		})
	})
}

// TestTypeConversion 测试类型转换功能
func TestTypeConversion(t *testing.T) {
	Convey("类型转换测试", t, func() {
		
		Convey("map到结构体转换", func() {
			rawResult := map[string]interface{}{
				"adult":    true,
				"eligible": false,
				"score":    85,
				"level":    "gold",
			}
			
			result, err := convertToType[UserResult](rawResult)
			So(err, ShouldBeNil)
			So(result.Adult, ShouldEqual, true)
			So(result.Eligible, ShouldEqual, false)
			So(result.Score, ShouldEqual, 85)
			So(result.Level, ShouldEqual, "gold")
		})
		
		Convey("map到map转换", func() {
			rawResult := map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			}
			
			// 转换为map[string]interface{}
			result1, err := convertToType[map[string]interface{}](rawResult)
			So(err, ShouldBeNil)
			So(result1["key1"], ShouldEqual, "value1")
			So(result1["key2"], ShouldEqual, 42)
			So(result1["key3"], ShouldEqual, true)
			
			// 转换为map[string]any
			result2, err := convertToType[map[string]any](rawResult)
			So(err, ShouldBeNil)
			So(result2["key1"], ShouldEqual, "value1")
			So(result2["key2"], ShouldEqual, 42)
			So(result2["key3"], ShouldEqual, true)
		})
		
		Convey("接口类型转换", func() {
			rawResult := map[string]interface{}{
				"data": "test",
			}
			
			result, err := convertToType[interface{}](rawResult)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
		})
		
		Convey("转换错误处理", func() {
			// 包含无法JSON序列化的数据
			rawResult := map[string]interface{}{
				"valid":   "data",
				"invalid": make(chan int), // channel类型无法JSON序列化
			}
			
			_, err := convertToType[UserResult](rawResult)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "JSON序列化失败")
		})
	})
}