package runehammer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 定义缓存错误
var ErrCacheNotFound = errors.New("cache key not found")

// mockRuleMapper Mock规则映射器
type mockRuleMapper struct {
	rules map[string][]*Rule
}

func newMockRuleMapper() *mockRuleMapper {
	return &mockRuleMapper{
		rules: make(map[string][]*Rule),
	}
}

func (m *mockRuleMapper) FindByBizCode(ctx context.Context, bizCode string) ([]*Rule, error) {
	rules, ok := m.rules[bizCode]
	if !ok {
		return nil, nil
	}
	return rules, nil
}

func (m *mockRuleMapper) SetRules(bizCode string, rules []*Rule) {
	m.rules[bizCode] = rules
}

// mockCache Mock缓存
type mockCache struct {
	data map[string][]byte
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string][]byte),
	}
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return nil, ErrCacheNotFound
}

func (m *mockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func (m *mockCache) Del(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCache) Close() error {
	m.data = make(map[string][]byte)
	return nil
}

// mockCacheKeyBuilder Mock缓存键构建器 - 使用真实的CacheKeyBuilder
func newMockCacheKeyBuilder() CacheKeyBuilder {
	return CacheKeyBuilder{}
}

// TestEngineImpl 测试引擎实现
func TestEngineImpl(t *testing.T) {
	Convey("引擎实现测试", t, func() {
		
		Convey("引擎创建", func() {
			config := DefaultConfig()
			mapper := newMockRuleMapper()
			cache := newMockCache()
			cacheKeys := newMockCacheKeyBuilder()
			logger := NewNoopLogger()
			knowledgeLibrary := ast.NewKnowledgeLibrary()
			knowledgeBases := &sync.Map{}
			cronScheduler := cron.New()
			
			engine := NewEngineImpl[map[string]any](
				config, mapper, cache, cacheKeys, logger,
				knowledgeLibrary, knowledgeBases, cronScheduler, false,
			)
			
			So(engine, ShouldNotBeNil)
			So(engine.config, ShouldEqual, config)
			So(engine.mapper, ShouldEqual, mapper)
			So(engine.cache, ShouldEqual, cache)
			So(engine.logger, ShouldEqual, logger)
			So(engine.closed, ShouldBeFalse)
		})

		Convey("执行规则", func() {
			config := DefaultConfig()
			mapper := newMockRuleMapper()
			cache := newMockCache()
			cacheKeys := newMockCacheKeyBuilder()
			logger := NewNoopLogger()
			knowledgeLibrary := ast.NewKnowledgeLibrary()
			knowledgeBases := &sync.Map{}
			cronScheduler := cron.New()
			
			engine := NewEngineImpl[map[string]any](
				config, mapper, cache, cacheKeys, logger,
				knowledgeLibrary, knowledgeBases, cronScheduler, false,
			)

			Convey("正常执行", func() {
				// 设置测试规则
				rules := []*Rule{
					{
						ID:      1,
						BizCode: "test_biz",
						Name:    "测试规则",
						GRL:     `rule TestRule "测试规则" { when Params.age >= 18 then result["adult"] = true; }`,
						Enabled: true,
					},
				}
				mapper.SetRules("test_biz", rules)

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
				input := map[string]any{"age": 25}
				result, err := engine.Exec(context.Background(), "nonexistent", input)
				
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "规则未找到")
				So(result, ShouldBeZeroValue)
			})

			Convey("引擎已关闭", func() {
				engine.closed = true
				
				input := map[string]any{"age": 25}
				result, err := engine.Exec(context.Background(), "test_biz", input)
				
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "引擎已关闭")
				So(result, ShouldBeZeroValue)
			})
		})

		Convey("引擎关闭", func() {
			config := DefaultConfig()
			mapper := newMockRuleMapper()
			cache := newMockCache()
			cacheKeys := newMockCacheKeyBuilder()
			logger := NewNoopLogger()
			knowledgeLibrary := ast.NewKnowledgeLibrary()
			knowledgeBases := &sync.Map{}
			cronScheduler := cron.New()
			
			engine := NewEngineImpl[map[string]any](
				config, mapper, cache, cacheKeys, logger,
				knowledgeLibrary, knowledgeBases, cronScheduler, false,
			)

			Convey("正常关闭", func() {
				err := engine.Close()
				So(err, ShouldBeNil)
				So(engine.closed, ShouldBeTrue)

				// 关闭后执行规则应该失败
				input := map[string]any{"test": "value"}
				result, err := engine.Exec(context.Background(), "test_biz", input)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeZeroValue)
			})

			Convey("重复关闭", func() {
				err1 := engine.Close()
				So(err1, ShouldBeNil)

				err2 := engine.Close()
				So(err2, ShouldBeNil) // 重复关闭不应该报错
			})
		})

		Convey("并发安全性", func() {
			config := DefaultConfig()
			mapper := newMockRuleMapper()
			cache := newMockCache()
			cacheKeys := newMockCacheKeyBuilder()
			logger := NewNoopLogger()
			knowledgeLibrary := ast.NewKnowledgeLibrary()
			knowledgeBases := &sync.Map{}
			cronScheduler := cron.New()
			
			engine := NewEngineImpl[map[string]any](
				config, mapper, cache, cacheKeys, logger,
				knowledgeLibrary, knowledgeBases, cronScheduler, false,
			)

			rules := []*Rule{
				{
					ID:      1,
					BizCode: "concurrent_test",
					Name:    "并发测试规则",
					GRL:     `rule ConcurrentRule "并发测试" { when Params.id >= 0 then result["processed"] = Params.id; }`,
					Enabled: true,
				},
			}
			mapper.SetRules("concurrent_test", rules)

			Convey("并发执行规则", func() {
				var wg sync.WaitGroup
				errors := make([]error, 10)
				results := make([]map[string]any, 10)

				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						input := map[string]any{"id": id}
						result, err := engine.Exec(context.Background(), "concurrent_test", input)
						results[id] = result
						errors[id] = err
					}(i)
				}

				wg.Wait()

				// 验证所有执行都成功
				for i := 0; i < 10; i++ {
					So(errors[i], ShouldBeNil)
					So(results[i], ShouldNotBeNil)
					So(results[i]["processed"], ShouldEqual, i)
				}
			})
		})

		Convey("数据库集成测试", func() {
			// 创建内存数据库
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			So(err, ShouldBeNil)

			// 自动迁移
			err = db.AutoMigrate(&Rule{})
			So(err, ShouldBeNil)

			// 插入测试规则
			rule := &Rule{
				BizCode: "db_test",
				Name:    "数据库测试规则",
				GRL:     `rule DBTestRule "数据库测试" { when Params.amount > 100 then result["discount"] = 0.1; }`,
				Enabled: true,
			}
			err = db.Create(rule).Error
			So(err, ShouldBeNil)

			Convey("使用真实数据库映射器", func() {
				config := DefaultConfig()
				mapper := NewRuleMapper(db)
				cache := newMockCache()
				cacheKeys := newMockCacheKeyBuilder()
				logger := NewNoopLogger()
				knowledgeLibrary := ast.NewKnowledgeLibrary()
				knowledgeBases := &sync.Map{}
				cronScheduler := cron.New()
				
				engine := NewEngineImpl[map[string]any](
					config, mapper, cache, cacheKeys, logger,
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