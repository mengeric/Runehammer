package runehammer

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

func TestCacheWrapper(t *testing.T) {
	Convey("缓存包装器测试", t, func() {
		primary := NewMemoryCache(10)
		secondary := NewMemoryCache(10)
		
		wrapper := NewCacheWrapper(primary, secondary, nil)
		defer wrapper.Close()
		
		Convey("正常读写", func() {
			ctx := context.Background()
			err := wrapper.Set(ctx, "key", []byte("value"), time.Minute)
			So(err, ShouldBeNil)
			
			value, err := wrapper.Get(ctx, "key")
			So(err, ShouldBeNil)
			So(string(value), ShouldEqual, "value")
		})
		
		Convey("主缓存失败降级", func() {
			ctx := context.Background()
			// 先设置数据
			wrapper.Set(ctx, "key", []byte("value"), time.Minute)
			
			// 关闭主缓存模拟失败
			primary.Close()
			
			// 应该能从备用缓存读取
			value, err := wrapper.Get(ctx, "key")
			So(err, ShouldBeNil)
			So(string(value), ShouldEqual, "value")
		})
	})
}

func TestRuleCacheItem(t *testing.T) {
	Convey("规则缓存项测试", t, func() {
		rules := []*Rule{
			{
				ID:          1,
				BizCode:     "test",
				Name:        "TestRule",
				Description: "测试规则",
				GRL:         "rule Test {}",
				Version:     1,
				Enabled:     true,
			},
		}
		
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
			So(newItem.Rules[0].BizCode, ShouldEqual, "test")
			So(newItem.Version, ShouldEqual, 1)
		})
	})
}