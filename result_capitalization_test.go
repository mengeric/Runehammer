package runehammer

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestResultCapitalization 测试Result首字母大写的功能
func TestResultCapitalization(t *testing.T) {
	Convey("Result首字母大写测试", t, func() {
		
		Convey("动态引擎使用Result.字段名", func() {
			// 创建动态引擎
			engine := NewDynamicEngine[map[string]interface{}](
				DynamicEngineConfig{
					EnableCache:       false,
					StrictValidation:  false,
					ParallelExecution: false,
					DefaultTimeout:    5 * time.Second,
				},
			)
			
			// 定义使用Result.字段名的规则
			rule := SimpleRule{
				When: "Params >= 18",
				Then: map[string]string{
					"Result.Adult":    "true",
					"Result.Message":  "\"用户已成年\"",
					"Result.Category": "\"成年人\"",
				},
			}
			
			// 执行规则
			result, err := engine.ExecuteRuleDefinition(context.Background(), rule, 25)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			
			// 验证结果字段都是大驼峰命名
			So(result["Adult"], ShouldEqual, true)
			So(result["Message"], ShouldEqual, "用户已成年")
			So(result["Category"], ShouldEqual, "成年人")
		})
		
		Convey("批量规则使用Result.字段名", func() {
			// 创建动态引擎
			engine := NewDynamicEngine[map[string]interface{}](
				DynamicEngineConfig{
					EnableCache:       false,
					StrictValidation:  false,
					ParallelExecution: false,
					DefaultTimeout:    5 * time.Second,
				},
			)
			
			// 定义多个使用Result.字段名的规则
			rules := []interface{}{
				SimpleRule{
					When: "Params > 100",
					Then: map[string]string{
						"Result.IsLarge": "true",
						"Result.Level":    "\"high\"",
					},
				},
				SimpleRule{
					When: "Params <= 100",
					Then: map[string]string{
						"Result.IsSmall": "true",
						"Result.Level":    "\"low\"",
					},
				},
			}
			
			// 执行批量规则
			results, err := engine.ExecuteBatch(context.Background(), rules, 150)
			So(err, ShouldBeNil)
			So(len(results), ShouldEqual, 2)
			
			// 验证第一个规则结果
			So(results[0]["IsLarge"], ShouldEqual, true)
			So(results[0]["Level"], ShouldEqual, "high")
			
			// 验证第二个规则结果为空（条件不满足）
			So(len(results[1]), ShouldEqual, 0)
		})
		
		Convey("字符串类型输入使用Result.字段名", func() {
			// 创建动态引擎
			engine := NewDynamicEngine[map[string]interface{}](
				DynamicEngineConfig{
					EnableCache:       false,
					StrictValidation:  false,
					ParallelExecution: false,
					DefaultTimeout:    5 * time.Second,
				},
			)
			
			// 定义字符串规则
			rule := SimpleRule{
				When: "Params == \"VIP\"",
				Then: map[string]string{
					"Result.IsVip":     "true",
					"Result.Privilege":  "\"高级权限\"",
					"Result.Discount":   "0.2",
				},
			}
			
			// 执行规则
			result, err := engine.ExecuteRuleDefinition(context.Background(), rule, "VIP")
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			
			// 验证结果
			So(result["IsVip"], ShouldEqual, true)
			So(result["Privilege"], ShouldEqual, "高级权限")
			So(result["Discount"], ShouldEqual, 0.2)
		})
	})
}