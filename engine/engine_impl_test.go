package engine

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"gitee.com/damengde/runehammer/cache"
	"gitee.com/damengde/runehammer/config"
	logger "gitee.com/damengde/runehammer/logger"
	"gitee.com/damengde/runehammer/rule"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/robfig/cron/v3"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestEngineImpl 测试引擎实现
func TestEngineImpl(t *testing.T) {
	Convey("引擎实现测试", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		Convey("引擎创建", func() {
			cfg := config.DefaultConfig()
			mapper := rule.NewMockRuleMapper(ctrl)
			cacheImpl := cache.NewMockCache(ctrl)
			cacheKeys := cache.CacheKeyBuilder{}
			lgr := logger.NewNoopLogger()
			knowledgeLibrary := ast.NewKnowledgeLibrary()
			knowledgeBases := &sync.Map{}
			cronScheduler := cron.New()

			engine := NewEngineImpl[map[string]any](
				cfg, mapper, cacheImpl, cacheKeys, lgr,
				knowledgeLibrary, knowledgeBases, cronScheduler, false,
			)

			So(engine, ShouldNotBeNil)
		})

		Convey("执行规则", func() {
			cfg := config.DefaultConfig()
			mapper := rule.NewMockRuleMapper(ctrl)
			cacheImpl := cache.NewMockCache(ctrl)
			cacheKeys := cache.CacheKeyBuilder{}
			lgr := logger.NewNoopLogger()
			knowledgeLibrary := ast.NewKnowledgeLibrary()
			knowledgeBases := &sync.Map{}
			cronScheduler := cron.New()

			engine := NewEngineImpl[map[string]any](
				cfg, mapper, cacheImpl, cacheKeys, lgr,
				knowledgeLibrary, knowledgeBases, cronScheduler, false,
			)

			Convey("正常执行", func() {
				// 设置测试规则
				rules := []*rule.Rule{
					{
						ID:      1,
						BizCode: "test_biz",
						Name:    "测试规则",
						GRL:     `rule TestRule "测试规则" { when Params["age"] >= 18 then Result["adult"] = true; Retract("TestRule"); }`,
						Enabled: true,
					},
				}

				// 设置mock期望 - 先从缓存获取（返回错误表示缓存未命中）
				cacheImpl.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("cache miss"))
				
				// 然后从数据库获取
				mapper.EXPECT().FindByBizCode(gomock.Any(), "test_biz").Return(rules, nil)
				
				// 设置缓存
				cacheImpl.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				// 执行规则
				input := map[string]any{"age": 25}
				result, err := engine.Exec(context.Background(), "test_biz", input)

				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result["adult"], ShouldEqual, true)
			})

			Convey("空业务码", func() {
				input := map[string]any{"age": 25}
				result, err := engine.Exec(context.Background(), "", input)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "无效的业务码")
				So(result, ShouldBeZeroValue)
			})

			Convey("空输入", func() {
				result, err := engine.Exec(context.Background(), "test_biz", nil)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "输入参数为空")
				So(result, ShouldBeZeroValue)
			})

			Convey("规则不存在", func() {
				// 设置mock期望：缓存未命中和返回空规则列表
				cacheImpl.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("cache miss"))
				mapper.EXPECT().FindByBizCode(gomock.Any(), "nonexistent").Return([]*rule.Rule{}, nil)

				input := map[string]any{"age": 25}
				result, err := engine.Exec(context.Background(), "nonexistent", input)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "规则未找到")
				So(result, ShouldNotBeNil) // 引擎返回空的map而不是nil
			})
		})

		Convey("引擎关闭", func() {
			cfg := config.DefaultConfig()
			mapper := rule.NewMockRuleMapper(ctrl)
			cacheImpl := cache.NewMockCache(ctrl)
			cacheKeys := cache.CacheKeyBuilder{}
			lgr := logger.NewNoopLogger()
			knowledgeLibrary := ast.NewKnowledgeLibrary()
			knowledgeBases := &sync.Map{}
			cronScheduler := cron.New()

			engine := NewEngineImpl[map[string]any](
				cfg, mapper, cacheImpl, cacheKeys, lgr,
				knowledgeLibrary, knowledgeBases, cronScheduler, false,
			)

			Convey("正常关闭", func() {
				// 设置cache close期望
				cacheImpl.EXPECT().Close().Return(nil)

				err := engine.Close()
				So(err, ShouldBeNil)

				// 关闭后执行规则应该失败
				input := map[string]any{"test": "value"}
				result, err := engine.Exec(context.Background(), "test_biz", input)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeZeroValue)
			})

			Convey("重复关闭", func() {
				// 设置cache close期望 - 可能被调用多次
				cacheImpl.EXPECT().Close().Return(nil).AnyTimes()

				err1 := engine.Close()
				So(err1, ShouldBeNil)

				err2 := engine.Close()
				So(err2, ShouldBeNil) // 重复关闭不应该报错
			})
		})

		Convey("数据库集成测试", func() {
			// 创建内存数据库
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			So(err, ShouldBeNil)

			// 自动迁移
			err = db.AutoMigrate(&rule.Rule{})
			So(err, ShouldBeNil)

			// 插入测试规则
			testRule := &rule.Rule{
				BizCode: "db_test",
				Name:    "数据库测试规则",
				GRL:     `rule DBTestRule "数据库测试" { when Params["amount"] >= 100 then Result["discount"] = 0.1; Retract("DBTestRule"); }`,
				Enabled: true,
			}
			err = db.Create(testRule).Error
			So(err, ShouldBeNil)

			Convey("使用真实数据库映射器", func() {
				cfg := config.DefaultConfig()
				mapper := rule.NewRuleMapper(db)
				cacheImpl := cache.NewMemoryCache(1000) // 使用真实cache而不是mock
				cacheKeys := cache.CacheKeyBuilder{}
				lgr := logger.NewNoopLogger()
				knowledgeLibrary := ast.NewKnowledgeLibrary()
				knowledgeBases := &sync.Map{}
				cronScheduler := cron.New()

				engine := NewEngineImpl[map[string]any](
					cfg, mapper, cacheImpl, cacheKeys, lgr,
					knowledgeLibrary, knowledgeBases, cronScheduler, false,
				)

				input := map[string]any{"amount": 150}
				result, err := engine.Exec(context.Background(), "db_test", input)

				So(err, ShouldBeNil)
				So(result, ShouldNotBeNil)
				So(result["discount"], ShouldEqual, 0.1)
			})
		})
	})
}
