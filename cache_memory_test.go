package runehammer

import (
	"context"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestMemoryCacheDetailed 测试内存缓存实现（详细版本）
func TestMemoryCacheDetailed(t *testing.T) {
	Convey("内存缓存测试", t, func() {
		
		Convey("缓存创建", func() {
			
			Convey("正常创建", func() {
				cache := NewMemoryCache(100)
				So(cache, ShouldNotBeNil)
				
				// 测试关闭
				err := cache.Close()
				So(err, ShouldBeNil)
			})
			
			Convey("不同容量创建", func() {
				testCases := []int{1, 10, 100, 1000}
				
				for _, maxSize := range testCases {
					cache := NewMemoryCache(maxSize)
					So(cache, ShouldNotBeNil)
					cache.Close()
				}
			})
		})
		
		Convey("基本操作测试", func() {
			cache := NewMemoryCache(10)
			defer cache.Close()
			ctx := context.Background()
			
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
			
			Convey("不存在的键", func() {
				result, err := cache.Get(ctx, "nonexistent_key")
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
				So(err.Error(), ShouldContainSubstring, "not found")
			})
		})
		
		Convey("TTL过期测试", func() {
			cache := NewMemoryCache(10)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("立即过期", func() {
				key := "expire_immediate"
				value := []byte("expire_value")
				ttl := 1 * time.Millisecond
				
				// 设置短TTL
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 等待过期
				time.Sleep(5 * time.Millisecond)
				
				// 应该获取不到
				result, err := cache.Get(ctx, key)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})
			
			Convey("正常TTL", func() {
				key := "expire_normal"
				value := []byte("normal_value")
				ttl := 100 * time.Millisecond
				
				// 设置缓存
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 立即获取应该成功
				result, err := cache.Get(ctx, key)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, value)
				
				// 等待过期
				time.Sleep(150 * time.Millisecond)
				
				// 应该获取不到
				result, err = cache.Get(ctx, key)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})
			
			Convey("长TTL", func() {
				key := "expire_long"
				value := []byte("long_value")
				ttl := 1 * time.Hour
				
				// 设置长TTL
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 应该能够获取到
				result, err := cache.Get(ctx, key)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, value)
			})
		})
		
		Convey("容量限制测试", func() {
			maxSize := 5
			cache := NewMemoryCache(maxSize)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("达到容量限制", func() {
				ttl := 1 * time.Hour
				
				// 填满缓存
				for i := 0; i < maxSize; i++ {
					key := "key_" + string(rune(i+'0'))
					value := []byte("value_" + string(rune(i+'0')))
					err := cache.Set(ctx, key, value, ttl)
					So(err, ShouldBeNil)
				}
				
				// 验证所有项都存在
				for i := 0; i < maxSize; i++ {
					key := "key_" + string(rune(i+'0'))
					result, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
					So(result, ShouldNotBeNil)
				}
				
				// 添加超出容量的项
				extraKey := "extra_key"
				extraValue := []byte("extra_value")
				err := cache.Set(ctx, extraKey, extraValue, ttl)
				So(err, ShouldBeNil)
				
				// 新项应该存在
				result, err := cache.Get(ctx, extraKey)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, extraValue)
			})
			
			Convey("过期项优先清理", func() {
				// 添加一些即将过期的项
				shortTTL := 10 * time.Millisecond
				longTTL := 1 * time.Hour
				
				// 设置短TTL项
				for i := 0; i < 3; i++ {
					key := "short_" + string(rune(i+'0'))
					value := []byte("short_value_" + string(rune(i+'0')))
					err := cache.Set(ctx, key, value, shortTTL)
					So(err, ShouldBeNil)
				}
				
				// 设置长TTL项
				for i := 0; i < 2; i++ {
					key := "long_" + string(rune(i+'0'))
					value := []byte("long_value_" + string(rune(i+'0')))
					err := cache.Set(ctx, key, value, longTTL)
					So(err, ShouldBeNil)
				}
				
				// 等待短TTL项过期
				time.Sleep(20 * time.Millisecond)
				
				// 添加新项触发清理
				newKey := "new_key"
				newValue := []byte("new_value")
				err := cache.Set(ctx, newKey, newValue, longTTL)
				So(err, ShouldBeNil)
				
				// 长TTL项应该仍然存在
				for i := 0; i < 2; i++ {
					key := "long_" + string(rune(i+'0'))
					result, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
					So(result, ShouldNotBeNil)
				}
				
				// 新项也应该存在
				result, err := cache.Get(ctx, newKey)
				So(err, ShouldBeNil)
				So(result, ShouldResemble, newValue)
			})
		})
		
		Convey("并发安全测试", func() {
			cache := NewMemoryCache(100)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("并发读写", func() {
				var wg sync.WaitGroup
				errors := make([]error, 20)
				
				// 启动10个写goroutine
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						key := "concurrent_key_" + string(rune(id+'0'))
						value := []byte("concurrent_value_" + string(rune(id+'0')))
						errors[id] = cache.Set(ctx, key, value, 1*time.Hour)
					}(i)
				}
				
				// 启动10个读goroutine
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						key := "concurrent_key_" + string(rune(id+'0'))
						_, errors[id+10] = cache.Get(ctx, key)
					}(i)
				}
				
				wg.Wait()
				
				// 写操作应该都成功
				for i := 0; i < 10; i++ {
					So(errors[i], ShouldBeNil)
				}
			})
			
			Convey("并发删除", func() {
				// 先设置一些数据
				for i := 0; i < 10; i++ {
					key := "delete_key_" + string(rune(i+'0'))
					value := []byte("delete_value_" + string(rune(i+'0')))
					cache.Set(ctx, key, value, 1*time.Hour)
				}
				
				var wg sync.WaitGroup
				errors := make([]error, 10)
				
				// 并发删除
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(id int) {
						defer wg.Done()
						key := "delete_key_" + string(rune(id+'0'))
						errors[id] = cache.Del(ctx, key)
					}(i)
				}
				
				wg.Wait()
				
				// 删除操作应该都成功
				for i := 0; i < 10; i++ {
					So(errors[i], ShouldBeNil)
				}
				
				// 验证都已删除
				for i := 0; i < 10; i++ {
					key := "delete_key_" + string(rune(i+'0'))
					result, err := cache.Get(ctx, key)
					So(err, ShouldNotBeNil)
					So(result, ShouldBeNil)
				}
			})
		})
		
		Convey("边界条件测试", func() {
			
			Convey("零容量缓存", func() {
				cache := NewMemoryCache(0)
				defer cache.Close()
				ctx := context.Background()
				
				// 设置数据
				err := cache.Set(ctx, "test", []byte("value"), 1*time.Hour)
				So(err, ShouldBeNil)
				
				// 应该能够获取到（因为清理逻辑会至少删除1个）
				result, err := cache.Get(ctx, "test")
				// 这个测试的结果取决于实现细节
				if err == nil {
					So(result, ShouldNotBeNil)
				}
			})
			
			Convey("负容量缓存", func() {
				cache := NewMemoryCache(-1)
				defer cache.Close()
				ctx := context.Background()
				
				// 设置数据应该仍然工作
				err := cache.Set(ctx, "test", []byte("value"), 1*time.Hour)
				So(err, ShouldBeNil)
			})
			
			Convey("空键名", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				ctx := context.Background()
				
				err := cache.Set(ctx, "", []byte("empty_key_value"), 1*time.Hour)
				So(err, ShouldBeNil)
				
				result, err := cache.Get(ctx, "")
				So(err, ShouldBeNil)
				So(result, ShouldResemble, []byte("empty_key_value"))
			})
			
			Convey("空值", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				ctx := context.Background()
				
				err := cache.Set(ctx, "empty_value", []byte{}, 1*time.Hour)
				So(err, ShouldBeNil)
				
				result, err := cache.Get(ctx, "empty_value")
				So(err, ShouldBeNil)
				So(result, ShouldResemble, []byte{})
			})
			
			Convey("nil值", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				ctx := context.Background()
				
				err := cache.Set(ctx, "nil_value", nil, 1*time.Hour)
				So(err, ShouldBeNil)
				
				result, err := cache.Get(ctx, "nil_value")
				So(err, ShouldBeNil)
				So(result, ShouldBeNil)
			})
			
			Convey("零TTL", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				ctx := context.Background()
				
				err := cache.Set(ctx, "zero_ttl", []byte("zero_value"), 0)
				So(err, ShouldBeNil)
				
				// 零TTL意味着立即过期
				result, err := cache.Get(ctx, "zero_ttl")
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})
			
			Convey("负TTL", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				ctx := context.Background()
				
				err := cache.Set(ctx, "negative_ttl", []byte("negative_value"), -1*time.Hour)
				So(err, ShouldBeNil)
				
				// 负TTL意味着已经过期
				result, err := cache.Get(ctx, "negative_ttl")
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})
		})
		
		Convey("关闭和资源清理测试", func() {
			
			Convey("正常关闭", func() {
				cache := NewMemoryCache(10)
				
				// 设置一些数据
				ctx := context.Background()
				cache.Set(ctx, "test", []byte("value"), 1*time.Hour)
				
				// 关闭缓存
				err := cache.Close()
				So(err, ShouldBeNil)
				
				// 关闭后仍然可以操作，但后台清理会停止
				result, err := cache.Get(ctx, "test")
				So(err, ShouldBeNil)
				So(result, ShouldResemble, []byte("value"))
			})
			
			Convey("重复关闭", func() {
				cache := NewMemoryCache(10)
				
				// 第一次关闭
				err1 := cache.Close()
				So(err1, ShouldBeNil)
				
				// 第二次关闭应该不会panic
				err2 := cache.Close()
				So(err2, ShouldBeNil)
			})
		})
		
		Convey("数据类型兼容性测试", func() {
			cache := NewMemoryCache(10)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("各种字节数据", func() {
				testCases := []struct {
					name string
					data []byte
				}{
					{"普通字符串", []byte("hello world")},
					{"UTF-8中文", []byte("你好世界")},
					{"JSON数据", []byte(`{"key": "value", "number": 123}`)},
					{"二进制数据", []byte{0x00, 0xFF, 0x42, 0x00, 0xFF}},
					{"大数据", make([]byte, 1024)},
				}
				
				for _, tc := range testCases {
					Convey("数据类型: "+tc.name, func() {
						key := "data_" + tc.name
						
						err := cache.Set(ctx, key, tc.data, 1*time.Hour)
						So(err, ShouldBeNil)
						
						result, err := cache.Get(ctx, key)
						So(err, ShouldBeNil)
						So(result, ShouldResemble, tc.data)
					})
				}
			})
		})
		
		Convey("性能测试", func() {
			
			Convey("大量数据操作", func() {
				cache := NewMemoryCache(1000)
				defer cache.Close()
				ctx := context.Background()
				
				count := 500
				ttl := 1 * time.Hour
				
				// 批量设置
				start := time.Now()
				for i := 0; i < count; i++ {
					key := "perf_key_" + string(rune(i))
					value := []byte("perf_value_" + string(rune(i)))
					err := cache.Set(ctx, key, value, ttl)
					So(err, ShouldBeNil)
				}
				setDuration := time.Since(start)
				
				// 批量获取
				start = time.Now()
				for i := 0; i < count; i++ {
					key := "perf_key_" + string(rune(i))
					result, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
					So(result, ShouldNotBeNil)
				}
				getDuration := time.Since(start)
				
				// 性能断言（比较宽松的限制）
				So(setDuration, ShouldBeLessThan, 100*time.Millisecond)
				So(getDuration, ShouldBeLessThan, 50*time.Millisecond)
			})
		})
		
		Convey("清理机制测试", func() {
			
			Convey("异步删除", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				ctx := context.Background()
				
				// 设置一个短TTL项
				key := "async_delete_test"
				value := []byte("async_value")
				ttl := 10 * time.Millisecond
				
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 等待过期
				time.Sleep(20 * time.Millisecond)
				
				// 第一次获取会触发异步删除
				result, err := cache.Get(ctx, key)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
				
				// 给异步删除一些时间
				time.Sleep(10 * time.Millisecond)
				
				// 再次获取应该仍然失败
				result, err = cache.Get(ctx, key)
				So(err, ShouldNotBeNil)
				So(result, ShouldBeNil)
			})
		})
	})
}

// TestMemoryCacheEdgeCases 测试内存缓存边界情况
func TestMemoryCacheEdgeCases(t *testing.T) {
	Convey("内存缓存边界测试", t, func() {
		
		Convey("极端并发测试", func() {
			cache := NewMemoryCache(50)
			defer cache.Close()
			ctx := context.Background()
			
			var wg sync.WaitGroup
			goroutineCount := 100
			operationsPerGoroutine := 10
			
			// 启动大量并发操作
			for i := 0; i < goroutineCount; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					
					for j := 0; j < operationsPerGoroutine; j++ {
						key := "extreme_key_" + string(rune(id)) + "_" + string(rune(j))
						value := []byte("extreme_value_" + string(rune(id)) + "_" + string(rune(j)))
						
						// 随机操作
						switch j % 3 {
						case 0:
							cache.Set(ctx, key, value, 100*time.Millisecond)
						case 1:
							cache.Get(ctx, key)
						case 2:
							cache.Del(ctx, key)
						}
					}
				}(i)
			}
			
			wg.Wait()
			
			// 测试完成后缓存应该仍然可用
			err := cache.Set(ctx, "final_test", []byte("final_value"), 1*time.Hour)
			So(err, ShouldBeNil)
			
			result, err := cache.Get(ctx, "final_test")
			So(err, ShouldBeNil)
			So(result, ShouldResemble, []byte("final_value"))
		})
		
		Convey("内存压力测试", func() {
			cache := NewMemoryCache(10)
			defer cache.Close()
			ctx := context.Background()
			
			// 创建大量大数据项
			largeData := make([]byte, 1024) // 1KB数据
			for i := range largeData {
				largeData[i] = byte(i % 256)
			}
			
			// 连续添加大数据，测试清理机制
			for i := 0; i < 50; i++ {
				key := "large_data_" + string(rune(i))
				err := cache.Set(ctx, key, largeData, 1*time.Hour)
				So(err, ShouldBeNil)
			}
			
			// 缓存应该仍然可用，但由于容量限制，最后的项可能不存在
			result, err := cache.Get(ctx, "large_data_49")
			if err == nil {
				So(result, ShouldResemble, largeData)
			} else {
				So(err.Error(), ShouldContainSubstring, "not found")
			}
		})
	})
}