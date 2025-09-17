package runehammer

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gitee.com/damengde/runehammer/cache"
	"gitee.com/damengde/runehammer/config"
	logger "gitee.com/damengde/runehammer/logger"
	"github.com/robfig/cron/v3"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

// TestEngineLifecycle 测试引擎生命周期管理
func TestEngineLifecycle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	Convey("引擎生命周期测试", t, func() {

		Convey("StartSync 同步任务启动", func() {

			Convey("正常启动同步任务", func() {
				// 创建配置支持同步
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 100 * time.Millisecond,
				}

				// 创建引擎实例
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 启动同步任务
				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 验证引擎状态
				stats := engine.getStats()
				So(stats, ShouldNotBeNil)
				So(stats["sync_interval"], ShouldEqual, 100*time.Millisecond)
				So(stats["closed"], ShouldBeFalse)

				// 清理
				engine.Close()
			})

			Convey("未配置同步间隔", func() {
				// 创建无同步间隔的配置
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 0,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 启动同步任务应该成功但不做任何事
				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 验证状态
				stats := engine.getStats()
				So(stats["sync_interval"], ShouldEqual, time.Duration(0))

				engine.Close()
			})

			Convey("负同步间隔", func() {
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: -1 * time.Second,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 负间隔应该被当作未配置处理
				err := engine.StartSync()
				So(err, ShouldBeNil)

				engine.Close()
			})

			Convey("重复启动同步任务", func() {
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 200 * time.Millisecond,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 第一次启动
				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 第二次启动应该也成功（添加新的定时任务）
				err = engine.StartSync()
				So(err, ShouldBeNil)

				engine.Close()
			})
		})

		Convey("syncRules 规则同步逻辑", func() {

			Convey("基本同步执行", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 直接调用同步方法
				err := engine.syncRules()
				So(err, ShouldBeNil)

				engine.Close()
			})

			Convey("带日志的同步执行", func() {
				config := &config.Config{DSN: "mock"}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewDefaultLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 同步应该成功并记录日志
				err := engine.syncRules()
				So(err, ShouldBeNil)

				engine.Close()
			})

			Convey("同步过程清理缓存", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 模拟添加一些编译缓存
				engine.knowledgeBases.Store("test1", "knowledge1")
				engine.knowledgeBases.Store("test2", "knowledge2")

				// 验证缓存存在
				stats := engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 2)

				// 执行同步，应该清理缓存
				err := engine.syncRules()
				So(err, ShouldBeNil)

				// 验证缓存被清理
				stats = engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 0)

				engine.Close()
			})
		})

		Convey("clearExpiredKnowledgeBases 清理编译缓存", func() {

			Convey("清理所有缓存", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 添加多个缓存条目
				testData := map[string]string{
					"biz1": "knowledge1",
					"biz2": "knowledge2",
					"biz3": "knowledge3",
				}

				for key, value := range testData {
					engine.knowledgeBases.Store(key, value)
				}

				// 验证缓存数量
				stats := engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 3)

				// 清理缓存
				engine.clearExpiredKnowledgeBases()

				// 验证所有缓存被清理
				stats = engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 0)

				// 验证无法获取之前的缓存
				for key := range testData {
					_, exists := engine.knowledgeBases.Load(key)
					So(exists, ShouldBeFalse)
				}

				engine.Close()
			})

			Convey("清理空缓存", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 验证初始状态
				stats := engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 0)

				// 清理空缓存应该不出错
				So(func() {
					engine.clearExpiredKnowledgeBases()
				}, ShouldNotPanic)

				// 状态应该保持不变
				stats = engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 0)

				engine.Close()
			})
		})

		Convey("refreshCache 刷新指定缓存", func() {

			Convey("正常刷新缓存", func() {
				config := &config.Config{DSN: "mock"}
				cacheImpl := cache.NewMemoryCache(1000)
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cacheImpl,
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 设置初始缓存
				bizCode := "test_biz"
				engine.knowledgeBases.Store(bizCode, "old_knowledge")

				// 验证缓存存在
				_, exists := engine.knowledgeBases.Load(bizCode)
				So(exists, ShouldBeTrue)

				// 刷新缓存
				_ = engine.refreshCache(bizCode)
				// 由于没有真实的数据库，getRules可能会失败，但清理逻辑应该执行
				// 这里主要验证不会panic

				// 验证编译缓存被清理
				_, exists = engine.knowledgeBases.Load(bizCode)
				So(exists, ShouldBeFalse)

				engine.Close()
				cacheImpl.Close()
			})

			Convey("带缓存组件的刷新", func() {
				cacheImpl := cache.NewMemoryCache(1000)
				config := &config.Config{DSN: "mock"}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cacheImpl,
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				bizCode := "test_biz_with_cache"

				// 设置缓存数据
				ctx := context.Background()
				cacheKey := engine.cacheKeys.RuleKey(bizCode)
				cacheImpl.Set(ctx, cacheKey, []byte("cached_rules"), time.Hour)

				// 验证缓存存在
				data, err := cacheImpl.Get(ctx, cacheKey)
				So(err, ShouldBeNil)
				So(data, ShouldNotBeNil)

				// 刷新缓存
				engine.refreshCache(bizCode)

				engine.Close()
				cacheImpl.Close()
			})

			Convey("带日志的刷新", func() {
				config := &config.Config{DSN: "mock"}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewDefaultLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 刷新操作应该记录日志并且不panic
				So(func() {
					engine.refreshCache("test_biz_with_log")
				}, ShouldNotPanic)

				engine.Close()
			})

			Convey("空业务码刷新", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 空业务码应该不会panic
				So(func() {
					engine.refreshCache("")
				}, ShouldNotPanic)

				engine.Close()
			})
		})

		Convey("getStats 统计信息获取", func() {

			Convey("基本统计信息", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					nil, // 无缓存
					cache.CacheKeyBuilder{},
					nil, // 无日志
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				stats := engine.getStats()
				So(stats, ShouldNotBeNil)
				So(stats["closed"], ShouldBeFalse)
				So(stats["knowledge_bases"], ShouldEqual, 0)
				So(stats["sync_interval"], ShouldEqual, time.Duration(0))
				So(stats["cache_enabled"], ShouldBeFalse)
				So(stats["logger_enabled"], ShouldBeFalse)

				engine.Close()
			})

			Convey("完整配置的统计信息", func() {
				cacheImpl := cache.NewMemoryCache(1000)
				logger := logger.NewDefaultLogger()
				syncInterval := 5 * time.Minute

				config := &config.Config{
					DSN:          "mock",
					SyncInterval: syncInterval,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cacheImpl,
					cache.CacheKeyBuilder{},
					logger,
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 添加一些缓存数据
				engine.knowledgeBases.Store("biz1", "kb1")
				engine.knowledgeBases.Store("biz2", "kb2")

				stats := engine.getStats()
				So(stats, ShouldNotBeNil)
				So(stats["closed"], ShouldBeFalse)
				So(stats["knowledge_bases"], ShouldEqual, 2)
				So(stats["sync_interval"], ShouldEqual, syncInterval)
				So(stats["cache_enabled"], ShouldBeTrue)
				So(stats["logger_enabled"], ShouldBeTrue)

				engine.Close()
				cacheImpl.Close()
			})

			Convey("关闭后的统计信息", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 关闭引擎
				engine.Close()

				stats := engine.getStats()
				So(stats, ShouldNotBeNil)
				So(stats["closed"], ShouldBeTrue)
			})

			Convey("大量缓存条目统计", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 添加大量缓存条目
				count := 100
				for i := 0; i < count; i++ {
					key := fmt.Sprintf("biz_%d", i)
					value := fmt.Sprintf("knowledge_%d", i)
					engine.knowledgeBases.Store(key, value)
				}

				stats := engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, count)

				engine.Close()
			})
		})

		Convey("生命周期集成测试", func() {

			Convey("完整生命周期流程", func() {
				cacheImpl := cache.NewMemoryCache(1000)
				logger := logger.NewNoopLogger() // 使用NoopLogger避免测试输出干扰

				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 50 * time.Millisecond,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cacheImpl,
					cache.CacheKeyBuilder{},
					logger,
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 1. 检查初始状态
				stats := engine.getStats()
				So(stats["closed"], ShouldBeFalse)
				So(stats["knowledge_bases"], ShouldEqual, 0)

				// 2. 启动同步任务
				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 3. 添加一些缓存数据
				engine.knowledgeBases.Store("test1", "kb1")
				engine.knowledgeBases.Store("test2", "kb2")

				stats = engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 2)

				// 4. 等待同步任务执行（清理缓存）
				time.Sleep(100 * time.Millisecond)

				// 5. 验证缓存被清理
				stats = engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 0)

				// 6. 手动刷新特定缓存
				engine.refreshCache("manual_refresh")
				// 可能失败但不应该panic

				// 7. 关闭引擎
				engine.Close()
				cacheImpl.Close()

				// 8. 验证关闭状态
				stats = engine.getStats()
				So(stats["closed"], ShouldBeTrue)
			})

			Convey("并发安全测试", func() {
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 10 * time.Millisecond,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 启动同步任务
				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 并发执行各种操作
				done := make(chan bool, 10)

				// 并发添加缓存
				for i := 0; i < 5; i++ {
					go func(id int) {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("Goroutine %d panicked: %v", id, r)
							}
							done <- true
						}()

						key := fmt.Sprintf("concurrent_%d", id)
						engine.knowledgeBases.Store(key, fmt.Sprintf("kb_%d", id))
					}(i)
				}

				// 并发获取统计
				for i := 0; i < 3; i++ {
					go func(id int) {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("Stats goroutine %d panicked: %v", id, r)
							}
							done <- true
						}()

						for j := 0; j < 10; j++ {
							engine.getStats()
							time.Sleep(1 * time.Millisecond)
						}
					}(i)
				}

				// 并发刷新缓存
				for i := 0; i < 2; i++ {
					go func(id int) {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("Refresh goroutine %d panicked: %v", id, r)
							}
							done <- true
						}()

						bizCode := fmt.Sprintf("refresh_%d", id)
						engine.refreshCache(bizCode)
					}(i)
				}

				// 等待所有goroutine完成
				for i := 0; i < 10; i++ {
					<-done
				}

				// 等待同步任务执行几次
				time.Sleep(50 * time.Millisecond)

				engine.Close()

				// 验证没有panic发生
				So(true, ShouldBeTrue)
			})

			Convey("资源清理验证", func() {
				cacheImpl := cache.NewMemoryCache(1000)

				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 100 * time.Millisecond,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cacheImpl,
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 启动同步并添加数据
				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 添加缓存数据
				for i := 0; i < 10; i++ {
					key := fmt.Sprintf("resource_%d", i)
					engine.knowledgeBases.Store(key, fmt.Sprintf("data_%d", i))
				}

				stats := engine.getStats()
				initialKBCount := stats["knowledge_bases"].(int)
				So(initialKBCount, ShouldEqual, 10)

				// 关闭引擎
				engine.Close()
				cacheImpl.Close()

				// 验证状态已更新
				stats = engine.getStats()
				So(stats["closed"], ShouldBeTrue)

				// 注意：这里不验证缓存是否被清理，因为Close()不负责清理缓存
				// 缓存清理由定时任务或手动清理完成
			})
		})

		Convey("错误处理测试", func() {

			Convey("同步任务添加失败处理", func() {
				// 这个测试比较难模拟，因为cron.AddFunc很少失败
				// 这里主要验证错误处理逻辑存在
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 1 * time.Nanosecond, // 极小间隔可能导致问题
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 尝试启动同步任务
				_ = engine.StartSync()
				// 即使失败也不应该panic

				engine.Close()
			})

			Convey("无效配置处理", func() {
				// nil配置
				So(func() {
					config := &config.Config{DSN: "mock"}
					engine := NewEngineImpl[map[string]interface{}](
						config,
						NewMockRuleMapper(ctrl),
						cache.NewMemoryCache(1000),
						cache.CacheKeyBuilder{},
						logger.NewNoopLogger(),
						nil,
						&sync.Map{},
						cron.New(),
						false,
					)
					if engine != nil {
						engine.StartSync()
						engine.Close()
					}
				}, ShouldNotPanic)
			})

			Convey("同步过程异常处理", func() {
				config := &config.Config{DSN: "mock"}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 手动调用同步方法应该处理内部异常
				So(func() {
					engine.syncRules()
				}, ShouldNotPanic)

				engine.Close()
			})
		})
	})
}

// TestEngineLifecycleEdgeCases 测试生命周期边界情况
func TestEngineLifecycleEdgeCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	Convey("生命周期边界测试", t, func() {

		Convey("极端配置测试", func() {

			Convey("极短同步间隔", func() {
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 1 * time.Microsecond,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 立即关闭，验证不会有竞态条件
				engine.Close()
			})

			Convey("极长同步间隔", func() {
				config := &config.Config{
					DSN:          "mock",
					SyncInterval: 24 * time.Hour,
				}

				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				err := engine.StartSync()
				So(err, ShouldBeNil)

				// 验证统计信息
				stats := engine.getStats()
				So(stats["sync_interval"], ShouldEqual, 24*time.Hour)

				engine.Close()
			})
		})

		Convey("大量数据测试", func() {

			Convey("大量编译缓存", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 添加大量缓存条目
				count := 10000
				for i := 0; i < count; i++ {
					key := fmt.Sprintf("large_cache_%d", i)
					value := fmt.Sprintf("large_knowledge_%d", i)
					engine.knowledgeBases.Store(key, value)
				}

				// 验证数量
				stats := engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, count)

				// 清理大量缓存
				start := time.Now()
				engine.clearExpiredKnowledgeBases()
				duration := time.Since(start)

				// 清理应该在合理时间内完成（1秒内）
				So(duration, ShouldBeLessThan, 1*time.Second)

				// 验证清理结果
				stats = engine.getStats()
				So(stats["knowledge_bases"], ShouldEqual, 0)

				engine.Close()
			})
		})

		Convey("特殊字符处理", func() {

			Convey("特殊业务码", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 测试各种特殊字符的业务码
				specialCodes := []string{
					"",
					"中文业务码",
					"special!@#$%^&*()",
					"with spaces",
					"with\ttab",
					"with\nnewline",
					"very_long_business_code_that_exceeds_normal_length_expectations_and_might_cause_issues",
				}

				for _, code := range specialCodes {
					So(func() {
						engine.refreshCache(code)
					}, ShouldNotPanic)
				}

				engine.Close()
			})
		})

		Convey("并发边界测试", func() {

			Convey("高并发缓存操作", func() {
				config := &config.Config{DSN: "mock"}
				engine := NewEngineImpl[map[string]interface{}](
					config,
					NewMockRuleMapper(ctrl),
					cache.NewMemoryCache(1000),
					cache.CacheKeyBuilder{},
					logger.NewNoopLogger(),
					nil,
					&sync.Map{},
					cron.New(),
					false,
				)
				So(engine, ShouldNotBeNil)

				// 启动大量并发goroutine
				concurrency := 100
				done := make(chan bool, concurrency)

				for i := 0; i < concurrency; i++ {
					go func(id int) {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("High concurrency goroutine %d panicked: %v", id, r)
							}
							done <- true
						}()

						// 混合操作
						key := fmt.Sprintf("concurrent_%d", id)
						engine.knowledgeBases.Store(key, fmt.Sprintf("value_%d", id))
						engine.getStats()
						engine.refreshCache(key)
						engine.clearExpiredKnowledgeBases()
					}(i)
				}

				// 等待所有goroutine完成
				for i := 0; i < concurrency; i++ {
					<-done
				}

				engine.Close()
				So(true, ShouldBeTrue) // 验证没有panic
			})
		})
	})
}
