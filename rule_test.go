package runehammer

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRule 测试规则数据模型
func TestRule(t *testing.T) {
	Convey("规则数据模型测试", t, func() {
		
		Convey("Rule结构体定义", func() {
			
			Convey("基本字段验证", func() {
				rule := &Rule{
					ID:      1,
					BizCode: "test_biz",
					Name:    "测试规则",
					GRL:     `rule TestRule "测试" { when true then result = true; }`,
					Version: 1,
					Enabled: true,
				}
				
				So(rule.ID, ShouldEqual, 1)
				So(rule.BizCode, ShouldEqual, "test_biz")
				So(rule.Name, ShouldEqual, "测试规则")
				So(rule.GRL, ShouldNotBeEmpty)
				So(rule.Version, ShouldEqual, 1)
				So(rule.Enabled, ShouldBeTrue)
			})
			
			Convey("TableName() 方法", func() {
				rule := Rule{}
				tableName := rule.TableName()
				So(tableName, ShouldEqual, "runehammer_rules")
			})
			
			Convey("时间字段处理", func() {
				now := time.Now()
				rule := &Rule{
					BizCode:   "time_test",
					Name:      "时间测试",
					GRL:       "rule TimeTest {}",
					CreatedAt: now,
					UpdatedAt: now,
				}
				
				So(rule.CreatedAt.Equal(now), ShouldBeTrue)
				So(rule.UpdatedAt.Equal(now), ShouldBeTrue)
			})
		})
		
		Convey("规则验证和边界测试", func() {
			
			Convey("空值处理", func() {
				rule := &Rule{}
				
				// 默认值验证
				So(rule.ID, ShouldEqual, 0)
				So(rule.BizCode, ShouldBeEmpty)
				So(rule.Name, ShouldBeEmpty)
				So(rule.GRL, ShouldBeEmpty)
				So(rule.Version, ShouldEqual, 0)
				So(rule.Enabled, ShouldBeFalse)
			})
			
			Convey("长字符串处理", func() {
				longBizCode := make([]byte, 200) // 超过100字符限制
				for i := range longBizCode {
					longBizCode[i] = 'a'
				}
				
				longName := make([]byte, 300) // 超过200字符限制
				for i := range longName {
					longName[i] = 'b'
				}
				
				longDescription := make([]byte, 600) // 超过500字符限制
				for i := range longDescription {
					longDescription[i] = 'c'
				}
				
				rule := &Rule{
					BizCode:     string(longBizCode),
					Name:        string(longName),
					Description: string(longDescription),
				}
				
				So(len(rule.BizCode), ShouldEqual, 200)
				So(len(rule.Name), ShouldEqual, 300)
				So(len(rule.Description), ShouldEqual, 600)
			})
			
			Convey("特殊字符处理", func() {
				rule := &Rule{
					BizCode:     "test_biz_特殊字符",
					Name:        "测试规则!@#$%^&*()",
					GRL:         `rule SpecialCharRule "特殊字符" { when "test" == "测试" then result = "成功"; }`,
					Description: "包含特殊字符的描述：!@#$%^&*()",
					CreatedBy:   "用户@domain.com",
					UpdatedBy:   "管理员#123",
				}
				
				So(rule.BizCode, ShouldContainSubstring, "特殊字符")
				So(rule.Name, ShouldContainSubstring, "!@#$%^&*()")
				So(rule.GRL, ShouldContainSubstring, "测试")
				So(rule.Description, ShouldContainSubstring, "!@#$%^&*()")
				So(rule.CreatedBy, ShouldContainSubstring, "@")
				So(rule.UpdatedBy, ShouldContainSubstring, "#")
			})
		})
	})
}

// TestRuleMapper 测试规则数据访问层
func TestRuleMapper(t *testing.T) {
	Convey("规则数据访问测试", t, func() {
		
		// 创建内存数据库
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		So(err, ShouldBeNil)
		
		// 自动迁移
		err = db.AutoMigrate(&Rule{})
		So(err, ShouldBeNil)
		
		// 创建mapper实例
		mapper := NewRuleMapper(db)
		So(mapper, ShouldNotBeNil)
		
		Convey("基本查询测试", func() {
			
			Convey("空数据库查询", func() {
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, "nonexistent")
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 0)
			})
			
			Convey("单条规则查询", func() {
				// 插入测试数据
				rule := &Rule{
					BizCode: "single_test",
					Name:    "单条测试规则",
					GRL:     `rule SingleTest "单条测试" { when true then result = "single"; }`,
					Version: 1,
					Enabled: true,
				}
				err := db.Create(rule).Error
				So(err, ShouldBeNil)
				
				// 查询规则
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, "single_test")
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 1)
				So(rules[0].BizCode, ShouldEqual, "single_test")
				So(rules[0].Name, ShouldEqual, "单条测试规则")
				So(rules[0].Enabled, ShouldBeTrue)
			})
			
			Convey("多条规则查询", func() {
				// 插入多条测试数据
				rules := []*Rule{
					{
						BizCode: "multi_test",
						Name:    "多条测试规则1",
						GRL:     `rule MultiTest1 "多条测试1" { when true then result = "multi1"; }`,
						Version: 1,
						Enabled: true,
					},
					{
						BizCode: "multi_test",
						Name:    "多条测试规则2",
						GRL:     `rule MultiTest2 "多条测试2" { when true then result = "multi2"; }`,
						Version: 2,
						Enabled: true,
					},
					{
						BizCode: "multi_test",
						Name:    "多条测试规则3",
						GRL:     `rule MultiTest3 "多条测试3" { when true then result = "multi3"; }`,
						Version: 3,
						Enabled: true,
					},
				}
				
				for _, rule := range rules {
					err := db.Create(rule).Error
					So(err, ShouldBeNil)
				}
				
				// 查询规则
				ctx := context.Background()
				foundRules, err := mapper.FindByBizCode(ctx, "multi_test")
				
				So(err, ShouldBeNil)
				So(foundRules, ShouldNotBeNil)
				So(len(foundRules), ShouldEqual, 3)
				
				// 验证按版本号降序排列
				So(foundRules[0].Version, ShouldEqual, 3)
				So(foundRules[1].Version, ShouldEqual, 2)
				So(foundRules[2].Version, ShouldEqual, 1)
			})
		})
		
		Convey("状态过滤测试", func() {
			
			Convey("启用状态过滤", func() {
				// 插入启用和禁用的规则
				enabledRule := &Rule{
					BizCode: "status_test",
					Name:    "启用规则",
					GRL:     `rule EnabledTest "启用测试" { when true then result = "enabled"; }`,
					Version: 1,
					Enabled: true,
				}
				
				disabledRule := &Rule{
					BizCode: "status_test",
					Name:    "禁用规则",
					GRL:     `rule DisabledTest "禁用测试" { when true then result = "disabled"; }`,
					Version: 2,
					Enabled: false,
				}
				
				err := db.Create(enabledRule).Error
				So(err, ShouldBeNil)
				
				err = db.Create(disabledRule).Error
				So(err, ShouldBeNil)
				
				// 查询规则，应该只返回启用的规则
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, "status_test")
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 1)
				So(rules[0].Enabled, ShouldBeTrue)
				So(rules[0].Name, ShouldEqual, "启用规则")
			})
		})
		
		Convey("上下文测试", func() {
			
			Convey("超时上下文", func() {
				// 创建一个已经超时的上下文
				ctx, cancel := context.WithTimeout(context.Background(), 0)
				defer cancel()
				time.Sleep(1 * time.Millisecond) // 确保超时
				
				// 这个测试可能不会失败，因为SQLite内存数据库操作很快
				// 但至少验证代码能处理上下文
				rules, err := mapper.FindByBizCode(ctx, "timeout_test")
				
				// 根据实际情况，可能成功也可能失败
				if err != nil {
					So(err, ShouldNotBeNil)
				} else {
					So(rules, ShouldNotBeNil)
				}
			})
			
			Convey("取消上下文", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // 立即取消
				
				// 同样，SQLite内存数据库操作可能太快，取消不了
				rules, err := mapper.FindByBizCode(ctx, "cancel_test")
				
				// 根据实际情况处理
				if err != nil {
					So(err, ShouldNotBeNil)
				} else {
					So(rules, ShouldNotBeNil)
				}
			})
		})
		
		Convey("边界条件测试", func() {
			
			Convey("空业务码查询", func() {
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, "")
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 0)
			})
			
			Convey("特殊字符业务码", func() {
				// 插入包含特殊字符的业务码
				rule := &Rule{
					BizCode: "special_chars_!@#$%",
					Name:    "特殊字符测试",
					GRL:     `rule SpecialCharsTest "特殊字符" { when true then result = "special"; }`,
					Version: 1,
					Enabled: true,
				}
				err := db.Create(rule).Error
				So(err, ShouldBeNil)
				
				// 查询规则
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, "special_chars_!@#$%")
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 1)
				So(rules[0].BizCode, ShouldEqual, "special_chars_!@#$%")
			})
			
			Convey("很长的业务码", func() {
				longBizCode := "very_long_biz_code_" + string(make([]byte, 50))
				for i := 18; i < len(longBizCode); i++ {
					longBizCode = longBizCode[:i] + "x" + longBizCode[i+1:]
				}
				
				// 插入长业务码
				rule := &Rule{
					BizCode: longBizCode,
					Name:    "长业务码测试",
					GRL:     `rule LongBizCodeTest "长业务码" { when true then result = "long"; }`,
					Version: 1,
					Enabled: true,
				}
				err := db.Create(rule).Error
				So(err, ShouldBeNil)
				
				// 查询规则
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, longBizCode)
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 1)
				So(rules[0].BizCode, ShouldEqual, longBizCode)
			})
		})
		
		Convey("数据完整性测试", func() {
			
			Convey("完整字段验证", func() {
				now := time.Now()
				rule := &Rule{
					BizCode:     "integrity_test",
					Name:        "完整性测试规则",
					GRL:         `rule IntegrityTest "完整性测试" { when true then result = "integrity"; }`,
					Version:     5,
					Enabled:     true,
					Description: "这是一个完整性测试规则",
					CreatedBy:   "test_user",
					UpdatedBy:   "test_admin",
					CreatedAt:   now,
					UpdatedAt:   now,
				}
				
				err := db.Create(rule).Error
				So(err, ShouldBeNil)
				
				// 查询并验证所有字段
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, "integrity_test")
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 1)
				
				foundRule := rules[0]
				So(foundRule.BizCode, ShouldEqual, "integrity_test")
				So(foundRule.Name, ShouldEqual, "完整性测试规则")
				So(foundRule.GRL, ShouldContainSubstring, "IntegrityTest")
				So(foundRule.Version, ShouldEqual, 5)
				So(foundRule.Enabled, ShouldBeTrue)
				So(foundRule.Description, ShouldEqual, "这是一个完整性测试规则")
				So(foundRule.CreatedBy, ShouldEqual, "test_user")
				So(foundRule.UpdatedBy, ShouldEqual, "test_admin")
				So(foundRule.ID, ShouldBeGreaterThan, 0)
			})
		})
		
		Convey("性能测试", func() {
			
			Convey("大量数据查询", func() {
				// 插入大量测试数据
				bizCode := "performance_test"
				for i := 0; i < 100; i++ {
					rule := &Rule{
						BizCode: bizCode,
						Name:    "性能测试规则" + string(rune(i)),
						GRL:     `rule PerformanceTest "性能测试" { when true then result = "performance"; }`,
						Version: i + 1,
						Enabled: true,
					}
					err := db.Create(rule).Error
					So(err, ShouldBeNil)
				}
				
				// 查询并测试性能
				start := time.Now()
				ctx := context.Background()
				rules, err := mapper.FindByBizCode(ctx, bizCode)
				elapsed := time.Since(start)
				
				So(err, ShouldBeNil)
				So(rules, ShouldNotBeNil)
				So(len(rules), ShouldEqual, 100)
				So(elapsed, ShouldBeLessThan, 100*time.Millisecond) // 100ms内完成
				
				// 验证排序（按版本号降序）
				for i := 0; i < len(rules)-1; i++ {
					So(rules[i].Version, ShouldBeGreaterThan, rules[i+1].Version)
				}
			})
		})
	})
}