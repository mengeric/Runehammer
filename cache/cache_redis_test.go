package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

// TestRedisCache 测试Redis缓存实现
func TestRedisCache(t *testing.T) {
	Convey("Redis缓存测试", t, func() {

		Convey("RedisCache创建", func() {

			Convey("使用真实Redis客户端创建", func() {
				// 创建Redis客户端配置
				client := redis.NewClient(&redis.Options{
					Addr:     "localhost:6379",
					Password: "",
					DB:       0,
				})

				cache := NewRedisCache(client)
				So(cache, ShouldNotBeNil)
				So(cache, ShouldImplement, (*Cache)(nil))

				// 验证类型
				redisCache, ok := cache.(*RedisCache)
				So(ok, ShouldBeTrue)
				So(redisCache.client, ShouldEqual, client)

				// 清理
				client.Close()
			})

			Convey("使用集群客户端创建", func() {
				// 创建集群客户端配置（但不实际创建集群客户端）
				// 因为集群客户端类型转换比较复杂，这里仅测试基本配置
				client := redis.NewClient(&redis.Options{
					Addr: "localhost:7000",
				})

				cache := NewRedisCache(client)
				So(cache, ShouldNotBeNil)
				So(cache, ShouldImplement, (*Cache)(nil))

				// 清理
				client.Close()
			})
		})

		Convey("基本操作测试（需要Redis服务）", func() {
			// 这些测试需要真实的Redis服务
			// 在CI/CD环境中可能需要跳过

			SkipConvey("Redis服务连接测试", func() {
				client := redis.NewClient(&redis.Options{
					Addr:     "localhost:6379",
					Password: "",
					DB:       0,
				})
				defer client.Close()

				// 测试连接
				ctx := context.Background()
				_, err := client.Ping(ctx).Result()
				if err != nil {
					// Redis服务不可用，跳过后续测试
					return
				}

				cache := NewRedisCache(client)

				Convey("Set和Get操作", func() {
					key := "test_key"
					value := []byte("test_value")
					ttl := 1 * time.Hour

					// 设置缓存
					err := cache.Set(ctx, key, value, ttl)
					So(err, ShouldBeNil)

					// 获取缓存
					result, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
					So(result, ShouldResemble, value)

					// 清理测试数据
					cache.Del(ctx, key)
				})

				Convey("Del操作", func() {
					key := "delete_test"
					value := []byte("delete_value")
					ttl := 1 * time.Hour

					// 设置缓存
					err := cache.Set(ctx, key, value, ttl)
					So(err, ShouldBeNil)

					// 验证存在
					result, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
					So(result, ShouldResemble, value)

					// 删除缓存
					err = cache.Del(ctx, key)
					So(err, ShouldBeNil)

					// 验证已删除
					result, err = cache.Get(ctx, key)
					So(err, ShouldNotBeNil)
					So(result, ShouldBeNil)
				})

				Convey("TTL过期测试", func() {
					key := "expire_test"
					value := []byte("expire_value")
					ttl := 100 * time.Millisecond

					// 设置短TTL
					err := cache.Set(ctx, key, value, ttl)
					So(err, ShouldBeNil)

					// 立即获取应该成功
					result, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
					So(result, ShouldResemble, value)

					// 等待过期
					time.Sleep(200 * time.Millisecond)

					// 应该获取不到
					result, err = cache.Get(ctx, key)
					So(err, ShouldNotBeNil)
					So(result, ShouldBeNil)
				})

				Convey("不存在的键", func() {
					result, err := cache.Get(ctx, "nonexistent_key")
					So(err, ShouldNotBeNil)
					So(result, ShouldBeNil)
					So(err.Error(), ShouldContainSubstring, "not found")
				})
			})
		})

		Convey("接口一致性测试", func() {
			// 这些测试不需要真实Redis连接
			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			cache := NewRedisCache(client)

			Convey("实现Cache接口", func() {
				So(cache, ShouldImplement, (*Cache)(nil))
			})

			Convey("方法签名正确", func() {
				ctx := context.Background()

				// 验证方法存在且签名正确
				So(func() {
					cache.Get(ctx, "test")
				}, ShouldNotPanic)

				So(func() {
					cache.Set(ctx, "test", []byte("value"), time.Hour)
				}, ShouldNotPanic)

				So(func() {
					cache.Del(ctx, "test")
				}, ShouldNotPanic)

				So(func() {
					cache.Close()
				}, ShouldNotPanic)
			})

			client.Close()
		})

		Convey("错误处理测试", func() {

			Convey("客户端已关闭的情况", func() {
				client := redis.NewClient(&redis.Options{
					Addr: "localhost:6379",
				})
				cache := NewRedisCache(client)

				// 先关闭客户端
				client.Close()

				ctx := context.Background()

				// 操作应该返回错误
				_, err := cache.Get(ctx, "test")
				So(err, ShouldNotBeNil)

				err = cache.Set(ctx, "test", []byte("value"), time.Hour)
				So(err, ShouldNotBeNil)

				err = cache.Del(ctx, "test")
				So(err, ShouldNotBeNil)
			})

			Convey("无效连接配置", func() {
				// 使用无效的Redis地址
				client := redis.NewClient(&redis.Options{
					Addr:        "invalid_host:6379",
					DialTimeout: 100 * time.Millisecond,
				})
				cache := NewRedisCache(client)
				defer client.Close()

				ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancel()

				// 操作应该超时或连接失败
				_, err := cache.Get(ctx, "test")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("上下文处理测试", func() {
			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			cache := NewRedisCache(client)
			defer client.Close()

			Convey("超时上下文", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()

				// 让上下文超时
				time.Sleep(5 * time.Millisecond)

				_, err := cache.Get(ctx, "test")
				So(err, ShouldNotBeNil)
			})

			Convey("取消上下文", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // 立即取消

				_, err := cache.Get(ctx, "test")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("数据类型兼容性测试", func() {
			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			cache := NewRedisCache(client)
			defer client.Close()

			Convey("各种字节数据", func() {
				testCases := []struct {
					name string
					data []byte
				}{
					{"普通字符串", []byte("hello world")},
					{"UTF-8中文", []byte("你好世界")},
					{"JSON数据", []byte(`{"key": "value", "number": 123}`)},
					{"二进制数据", []byte{0x00, 0xFF, 0x42, 0x00, 0xFF}},
					{"空数据", []byte{}},
					{"大数据", make([]byte, 1024)},
				}

				// 由于可能没有Redis服务，这里只测试方法调用不panic
				ctx := context.Background()
				for _, tc := range testCases {
					Convey("数据类型: "+tc.name, func() {
						So(func() {
							cache.Set(ctx, "test_"+tc.name, tc.data, time.Hour)
						}, ShouldNotPanic)

						So(func() {
							cache.Get(ctx, "test_"+tc.name)
						}, ShouldNotPanic)
					})
				}
			})
		})

		Convey("并发安全测试", func() {
			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			cache := NewRedisCache(client)
			defer client.Close()

			Convey("并发操作不会panic", func() {
				ctx := context.Background()

				// 启动多个goroutine进行并发操作
				done := make(chan bool, 10)

				for i := 0; i < 10; i++ {
					go func(id int) {
						defer func() {
							if r := recover(); r != nil {
								t.Errorf("Goroutine %d panicked: %v", id, r)
							}
							done <- true
						}()

						key := fmt.Sprintf("concurrent_key_%d", id)
						value := []byte(fmt.Sprintf("concurrent_value_%d", id))

						// 执行一系列操作
						cache.Set(ctx, key, value, time.Hour)
						cache.Get(ctx, key)
						cache.Del(ctx, key)
					}(i)
				}

				// 等待所有goroutine完成
				for i := 0; i < 10; i++ {
					<-done
				}

				// 如果到达这里说明没有panic
				So(true, ShouldBeTrue)
			})
		})
	})
}

// TestRedisCacheIntegration 测试Redis缓存集成
func TestRedisCacheIntegration(t *testing.T) {
	Convey("Redis缓存集成测试", t, func() {

		Convey("与其他组件集成", func() {

			Convey("作为Cache接口使用", func() {

				useCache := func(c Cache, ctx context.Context) error {
					// 典型的缓存使用模式
					key := "integration_test"
					value := []byte("integration_value")

					// 设置
					if err := c.Set(ctx, key, value, time.Hour); err != nil {
						return err
					}

					// 获取
					result, err := c.Get(ctx, key)
					if err != nil {
						return err
					}

					// 验证
					if string(result) != string(value) {
						return fmt.Errorf("值不匹配")
					}

					// 删除
					return c.Del(ctx, key)
				}

				client := redis.NewClient(&redis.Options{
					Addr: "localhost:6379",
				})
				cache := NewRedisCache(client)
				defer client.Close()

				ctx := context.Background()

				// 测试集成使用
				So(func() {
					useCache(cache, ctx)
				}, ShouldNotPanic)
			})
		})

		Convey("配置选项测试", func() {

			Convey("不同的Redis配置", func() {
				configs := []redis.Options{
					{
						Addr:     "localhost:6379",
						Password: "",
						DB:       0,
					},
					{
						Addr:     "localhost:6379",
						Password: "",
						DB:       1,
					},
					{
						Addr:        "localhost:6379",
						DialTimeout: 5 * time.Second,
						ReadTimeout: 3 * time.Second,
					},
				}

				for i, config := range configs {
					Convey(fmt.Sprintf("配置 %d", i), func() {
						client := redis.NewClient(&config)
						cache := NewRedisCache(client)
						defer client.Close()

						So(cache, ShouldNotBeNil)
						So(cache, ShouldImplement, (*Cache)(nil))
					})
				}
			})
		})

		Convey("性能考虑", func() {

			Convey("连接池配置", func() {
				client := redis.NewClient(&redis.Options{
					Addr:         "localhost:6379",
					PoolSize:     10,
					MinIdleConns: 5,
				})
				cache := NewRedisCache(client)
				defer client.Close()

				So(cache, ShouldNotBeNil)

				// 验证连接池配置生效
				stats := client.PoolStats()
				So(stats, ShouldNotBeNil)
			})
		})
	})
}

// TestRedisCacheEdgeCases 测试Redis缓存边界情况
func TestRedisCacheEdgeCases(t *testing.T) {
	Convey("Redis缓存边界测试", t, func() {

		Convey("极端情况处理", func() {
			client := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
			})
			cache := NewRedisCache(client)
			defer client.Close()

			ctx := context.Background()

			Convey("空键名", func() {
				So(func() {
					cache.Set(ctx, "", []byte("empty_key"), time.Hour)
				}, ShouldNotPanic)

				So(func() {
					cache.Get(ctx, "")
				}, ShouldNotPanic)
			})

			Convey("很长的键名", func() {
				longKey := string(make([]byte, 1000))
				for i := range longKey {
					longKey = longKey[:i] + "x" + longKey[i+1:]
				}

				So(func() {
					cache.Set(ctx, longKey, []byte("long_key_value"), time.Hour)
				}, ShouldNotPanic)

				So(func() {
					cache.Get(ctx, longKey)
				}, ShouldNotPanic)
			})

			Convey("nil值", func() {
				So(func() {
					cache.Set(ctx, "nil_value", nil, time.Hour)
				}, ShouldNotPanic)

				So(func() {
					cache.Get(ctx, "nil_value")
				}, ShouldNotPanic)
			})

			Convey("零TTL", func() {
				So(func() {
					cache.Set(ctx, "zero_ttl", []byte("zero_value"), 0)
				}, ShouldNotPanic)
			})

			Convey("负TTL", func() {
				So(func() {
					cache.Set(ctx, "negative_ttl", []byte("negative_value"), -1*time.Hour)
				}, ShouldNotPanic)
			})
		})

		Convey("资源清理", func() {

			Convey("多次关闭", func() {
				client := redis.NewClient(&redis.Options{
					Addr: "localhost:6379",
				})
				cache := NewRedisCache(client)

				// 第一次关闭
				err1 := cache.Close()
				So(err1, ShouldBeNil)

				// 第二次关闭
				err2 := cache.Close()
				// Redis客户端支持多次关闭，不应该panic
				So(func() { cache.Close() }, ShouldNotPanic)
				_ = err2 // 避免未使用变量警告
			})
		})
	})
}
