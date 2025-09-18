package cache

import (
	"context"
	"fmt"
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
				retrievedValue, err := cache.Get(ctx, key)
				So(err, ShouldBeNil)
				So(retrievedValue, ShouldResemble, value)
			})
			
			Convey("Del操作", func() {
				key := "delete_test_key"
				value := []byte("delete_test_value")
				ttl := 1 * time.Hour
				
				// 设置缓存
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 验证存在
				_, err = cache.Get(ctx, key)
				So(err, ShouldBeNil)
				
				// 删除缓存
				err = cache.Del(ctx, key)
				So(err, ShouldBeNil)
				
				// 验证已删除
				_, err = cache.Get(ctx, key)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "not found")
			})
			
			Convey("不存在的键", func() {
				_, err := cache.Get(ctx, "non_existent_key")
				So(err, ShouldNotBeNil)
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
				ttl := 1 * time.Nanosecond // 极短时间
				
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 稍微等待确保过期
				time.Sleep(10 * time.Millisecond)
				
				_, err = cache.Get(ctx, key)
				So(err, ShouldNotBeNil)
			})
			
			Convey("正常TTL", func() {
				key := "expire_normal"
				value := []byte("expire_value")
				ttl := 100 * time.Millisecond
				
				// 设置缓存
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 立即获取应该成功
				retrievedValue, err := cache.Get(ctx, key)
				So(err, ShouldBeNil)
				So(retrievedValue, ShouldResemble, value)
				
				// 等待过期后获取应该失败
				time.Sleep(150 * time.Millisecond)
				_, err = cache.Get(ctx, key)
				So(err, ShouldNotBeNil)
			})
			
			Convey("长TTL", func() {
				key := "expire_long"
				value := []byte("long_value")
				ttl := 10 * time.Second
				
				err := cache.Set(ctx, key, value, ttl)
				So(err, ShouldBeNil)
				
				// 短时间内多次获取都应该成功
				for i := 0; i < 3; i++ {
					retrievedValue, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
					So(retrievedValue, ShouldResemble, value)
					time.Sleep(50 * time.Millisecond)
				}
			})
		})
		
		Convey("容量限制测试", func() {
			maxSize := 5
			cache := NewMemoryCache(maxSize)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("达到容量限制", func() {
				// 填满缓存
				for i := 0; i < maxSize; i++ {
					key := fmt.Sprintf("key_%d", i)
					value := []byte(fmt.Sprintf("value_%d", i))
					err := cache.Set(ctx, key, value, 1*time.Hour)
					So(err, ShouldBeNil)
				}
				
				// 验证所有键都存在
				for i := 0; i < maxSize; i++ {
					key := fmt.Sprintf("key_%d", i)
					_, err := cache.Get(ctx, key)
					So(err, ShouldBeNil)
				}
				
				// 添加新键，应该触发清理
				newKey := "new_key"
				newValue := []byte("new_value")
				err := cache.Set(ctx, newKey, newValue, 1*time.Hour)
				So(err, ShouldBeNil)
				
				// 新键应该存在
				_, err = cache.Get(ctx, newKey)
				So(err, ShouldBeNil)
			})
			
			Convey("过期项优先清理", func() {
				// 创建容量为10的缓存避免容量限制影响
				cache10 := NewMemoryCache(10)
				defer cache10.Close()
				
				// 添加一些即将过期的项
				for i := 0; i < 3; i++ {
					key := fmt.Sprintf("expire_key_%d", i)
					value := []byte(fmt.Sprintf("expire_value_%d", i))
					err := cache10.Set(ctx, key, value, 50*time.Millisecond)
					So(err, ShouldBeNil)
				}
				
				// 添加一些长期有效的项
				for i := 0; i < 2; i++ {
					key := fmt.Sprintf("long_key_%d", i)
					value := []byte(fmt.Sprintf("long_value_%d", i))
					err := cache10.Set(ctx, key, value, 10*time.Hour)
					So(err, ShouldBeNil)
				}
				
				// 等待短期项过期
				time.Sleep(100 * time.Millisecond)
				
				// 现在创建容量受限的缓存来测试清理逻辑
				cache5 := NewMemoryCache(5)
				defer cache5.Close()
				
				// 先添加4个长期项到新缓存
				for i := 0; i < 4; i++ {
					key := fmt.Sprintf("long_key_%d", i)
					value := []byte(fmt.Sprintf("long_value_%d", i))
					err := cache5.Set(ctx, key, value, 10*time.Hour)
					So(err, ShouldBeNil)
				}
				
				// 添加1个即将过期的项
				err := cache5.Set(ctx, "expire_key", []byte("expire_value"), 50*time.Millisecond)
				So(err, ShouldBeNil)
				
				// 等待短期项过期
				time.Sleep(100 * time.Millisecond)
				
				// 添加新项应该优先清理过期项
				err = cache5.Set(ctx, "final_key", []byte("final_value"), 1*time.Hour)
				So(err, ShouldBeNil)
				
				// 长期项应该仍然存在
				for i := 0; i < 4; i++ {
					key := fmt.Sprintf("long_key_%d", i)
					_, err := cache5.Get(ctx, key)
					So(err, ShouldBeNil)
				}
				
				// 过期项应该已被清理
				_, err = cache5.Get(ctx, "expire_key")
				So(err, ShouldNotBeNil)
			})
		})
		
		Convey("并发安全测试", func() {
			cache := NewMemoryCache(100)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("并发读写", func() {
				var wg sync.WaitGroup
				numGoroutines := 10
				numOperations := 20
				
				// 并发写入
				wg.Add(numGoroutines)
				for i := 0; i < numGoroutines; i++ {
					go func(id int) {
						defer wg.Done()
						for j := 0; j < numOperations; j++ {
							key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
							value := []byte(fmt.Sprintf("value_%d_%d", id, j))
							cache.Set(ctx, key, value, 1*time.Hour)
						}
					}(i)
				}
				wg.Wait()
				
				// 并发读取
				wg.Add(numGoroutines)
				for i := 0; i < numGoroutines; i++ {
					go func(id int) {
						defer wg.Done()
						for j := 0; j < numOperations; j++ {
							key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
							cache.Get(ctx, key)
						}
					}(i)
				}
				wg.Wait()
			})
			
			Convey("并发删除", func() {
				// 先添加一些数据
				for i := 0; i < 50; i++ {
					key := fmt.Sprintf("delete_key_%d", i)
					value := []byte(fmt.Sprintf("delete_value_%d", i))
					cache.Set(ctx, key, value, 1*time.Hour)
				}
				
				var wg sync.WaitGroup
				numGoroutines := 5
				
				// 并发删除
				wg.Add(numGoroutines)
				for i := 0; i < numGoroutines; i++ {
					go func(id int) {
						defer wg.Done()
						for j := 0; j < 10; j++ {
							key := fmt.Sprintf("delete_key_%d", id*10+j)
							cache.Del(ctx, key)
						}
					}(i)
				}
				wg.Wait()
			})
		})
		
		Convey("边界条件测试", func() {
			ctx := context.Background()
			
			Convey("零容量缓存", func() {
				cache := NewMemoryCache(0)
				defer cache.Close()
				
				err := cache.Set(ctx, "key", []byte("value"), 1*time.Hour)
				So(err, ShouldBeNil) // 应该可以设置，但可能立即被清理
			})
			
			Convey("负容量缓存", func() {
				cache := NewMemoryCache(-1)
				defer cache.Close()
				So(cache, ShouldNotBeNil)
			})
			
			Convey("空键名", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				
				err := cache.Set(ctx, "", []byte("value"), 1*time.Hour)
				So(err, ShouldBeNil)
				
				value, err := cache.Get(ctx, "")
				So(err, ShouldBeNil)
				So(string(value), ShouldEqual, "value")
			})
			
			Convey("空值", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				
				err := cache.Set(ctx, "empty", []byte(""), 1*time.Hour)
				So(err, ShouldBeNil)
				
				value, err := cache.Get(ctx, "empty")
				So(err, ShouldBeNil)
				So(len(value), ShouldEqual, 0)
			})
			
			Convey("nil值", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				
				err := cache.Set(ctx, "nil", nil, 1*time.Hour)
				So(err, ShouldBeNil)
				
				value, err := cache.Get(ctx, "nil")
				So(err, ShouldBeNil)
				So(value, ShouldBeNil)
			})
			
			Convey("零TTL", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				
				err := cache.Set(ctx, "zero_ttl", []byte("value"), 0)
				So(err, ShouldBeNil)
				
				// 零TTL意味着立即过期
				time.Sleep(10 * time.Millisecond)
				_, err = cache.Get(ctx, "zero_ttl")
				So(err, ShouldNotBeNil)
			})
			
			Convey("负TTL", func() {
				cache := NewMemoryCache(10)
				defer cache.Close()
				
				err := cache.Set(ctx, "negative_ttl", []byte("value"), -1*time.Hour)
				So(err, ShouldBeNil)
				
				// 负TTL意味着已经过期
				_, err = cache.Get(ctx, "negative_ttl")
				So(err, ShouldNotBeNil)
			})
		})
		
		Convey("关闭和资源清理测试", func() {
			
			Convey("正常关闭", func() {
				cache := NewMemoryCache(10)
				
				// 添加一些数据
				ctx := context.Background()
				cache.Set(ctx, "key", []byte("value"), 1*time.Hour)
				
				// 关闭缓存
				err := cache.Close()
				So(err, ShouldBeNil)
			})
			
			Convey("重复关闭", func() {
				cache := NewMemoryCache(10)
				
				// 第一次关闭
				err := cache.Close()
				So(err, ShouldBeNil)
				
				// 第二次关闭应该也成功
				err = cache.Close()
				So(err, ShouldBeNil)
			})
		})
		
		Convey("数据类型兼容性测试", func() {
			cache := NewMemoryCache(10)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("各种字节数据", func() {
				testCases := []struct {
					name  string
					data  []byte
				}{
					{"普通字符串", []byte("hello world")},
					{"UTF-8中文", []byte("你好世界")},
					{"JSON数据", []byte(`{"name":"test","value":123}`)},
					{"二进制数据", []byte{0x00, 0x01, 0x02, 0xFF}},
					{"大数据", make([]byte, 1024)},
				}
				
				for _, tc := range testCases {
					Convey("数据类型: "+tc.name, func() {
						key := "data_type_" + tc.name
						
						err := cache.Set(ctx, key, tc.data, 1*time.Hour)
						So(err, ShouldBeNil)
						
						value, err := cache.Get(ctx, key)
						So(err, ShouldBeNil)
						So(value, ShouldResemble, tc.data)
					})
				}
			})
		})
		
		Convey("性能测试", func() {
			cache := NewMemoryCache(1000)
			defer cache.Close()
			ctx := context.Background()
			
			Convey("大量数据操作", func() {
				numItems := 1000
				
				// 批量设置
				start := time.Now()
				for i := 0; i < numItems; i++ {
					key := fmt.Sprintf("perf_key_%d", i)
					value := []byte(fmt.Sprintf("perf_value_%d", i))
					cache.Set(ctx, key, value, 1*time.Hour)
				}
				setDuration := time.Since(start)
				
				// 批量获取
				start = time.Now()
				for i := 0; i < numItems; i++ {
					key := fmt.Sprintf("perf_key_%d", i)
					cache.Get(ctx, key)
				}
				getDuration := time.Since(start)
				
				// 基本性能断言（这些数值可能需要根据实际环境调整）
				So(setDuration, ShouldBeLessThan, 1*time.Second)
				So(getDuration, ShouldBeLessThan, 1*time.Second)
				
				// 批量删除
				start = time.Now()
				for i := 0; i < numItems; i++ {
					key := fmt.Sprintf("perf_key_%d", i)
					cache.Del(ctx, key)
				}
				delDuration := time.Since(start)
				So(delDuration, ShouldBeLessThan, 1*time.Second)
			})
		})
	})
}