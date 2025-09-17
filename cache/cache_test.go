package cache

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMemoryCache(t *testing.T) {
	Convey("内存缓存测试", t, func() {
		cache := NewMemoryCache(10)
		defer cache.Close()
		
		Convey("基本操作", func() {
			ctx := context.Background()
			// 设置值
			err := cache.Set(ctx, "key1", []byte("value1"), time.Minute)
			So(err, ShouldBeNil)
			
			// 获取值
			value, err := cache.Get(ctx, "key1")
			So(err, ShouldBeNil)
			So(string(value), ShouldEqual, "value1")
			
			// 删除值
			err = cache.Del(ctx, "key1")
			So(err, ShouldBeNil)
			
			// 获取已删除的值
			_, err = cache.Get(ctx, "key1")
			So(err.Error(), ShouldEqual, "cache key not found")
		})
		
		Convey("过期测试", func() {
			ctx := context.Background()
			// 设置短过期时间
			err := cache.Set(ctx, "key2", []byte("value2"), time.Millisecond*100)
			So(err, ShouldBeNil)
			
			// 立即获取应该成功
			value, err := cache.Get(ctx, "key2")
			So(err, ShouldBeNil)
			So(string(value), ShouldEqual, "value2")
			
			// 等待过期
			time.Sleep(time.Millisecond * 200)
			
			// 获取过期值应该失败
			_, err = cache.Get(ctx, "key2")
			So(err.Error(), ShouldEqual, "cache key not found")
		})
		
		Convey("容量限制测试", func() {
			ctx := context.Background()
			smallCache := NewMemoryCache(2)
			defer smallCache.Close()
			
			// 添加多个值
			smallCache.Set(ctx, "k1", []byte("v1"), time.Hour)
			smallCache.Set(ctx, "k2", []byte("v2"), time.Hour)
			smallCache.Set(ctx, "k3", []byte("v3"), time.Hour) // 这个应该触发清理
			
			// 由于容量限制，某些值可能被清理
			// 具体行为取决于清理策略
		})
	})
}

func TestRuleCacheItem(t *testing.T) {
	Convey("规则缓存项测试", t, func() {
		// Create a simple rule-like object for testing
		rule := map[string]interface{}{
			"ID":          1,
			"BizCode":     "test",
			"Name":        "TestRule",
			"Description": "测试规则",
			"GRL":         "rule Test {}",
			"Version":     1,
			"Enabled":     true,
		}
		
		rules := []Rule{rule}
		
		item := &RuleCacheItem{
			Rules:     rules,
			UpdatedAt: time.Now(),
			Version:   1,
		}
		
		Convey("序列化和反序列化", func() {
			// 序列化
			data, err := item.ToBytes()
			So(err, ShouldBeNil)
			So(data, ShouldNotBeNil)
			
			// 反序列化
			newItem := &RuleCacheItem{}
			err = newItem.FromBytes(data)
			So(err, ShouldBeNil)
			So(len(newItem.Rules), ShouldEqual, 1)
			So(newItem.Version, ShouldEqual, 1)
		})
	})
}