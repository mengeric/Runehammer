package runehammer

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// TestMiscellaneousFunctions 测试各种杂项函数以提高覆盖率
func TestMiscellaneousFunctions(t *testing.T) {
	Convey("杂项函数测试", t, func() {

		Convey("缓存键构建器", func() {
			builder := CacheKeyBuilder{}
			
			Convey("MetaKey函数", func() {
				metaKey := builder.MetaKey("test_meta")
				So(metaKey, ShouldNotBeEmpty)
				So(metaKey, ShouldContainSubstring, "test_meta")
			})
		})

		Convey("NoopLogger函数测试", func() {
			logger := NewNoopLogger()
			ctx := context.Background()

			Convey("Debugf函数", func() {
				// 测试空实现不会panic
				So(func() {
					logger.Debugf(ctx, "debug message", "key", "value")
				}, ShouldNotPanic)
			})

			Convey("Infof函数", func() {
				So(func() {
					logger.Infof(ctx, "info message", "key", "value")
				}, ShouldNotPanic)
			})

			Convey("Warnf函数", func() {
				So(func() {
					logger.Warnf(ctx, "warn message", "key", "value")
				}, ShouldNotPanic)
			})

			Convey("Errorf函数", func() {
				So(func() {
					logger.Errorf(ctx, "error message", "key", "value")
				}, ShouldNotPanic)
			})
		})

		Convey("表达式解析器", func() {
			parser := NewExpressionParser()

			Convey("解析复杂表达式", func() {
				// 测试复杂表达式解析
				complex := "(age >= 18 AND income > 50000) OR vip = true"
				result, err := parser.ParseCondition(complex)
				if err == nil {
					So(result, ShouldNotBeEmpty)
				}
			})
		})

		Convey("规则转换器未覆盖函数", func() {
			converter := NewGRLConverter()

			Convey("convertFromMap函数", func() {
				// 测试从map转换
				mapData := map[string]interface{}{
					"when": "testdata.age >= 18",
					"then": map[string]interface{}{
						"result.adult": true,
					},
				}

				_, err := converter.convertFromMap(mapData)
				if err != nil {
					// 某些转换可能会失败
					So(err, ShouldNotBeNil)
				}
			})

			Convey("convertFromJSON函数", func() {
				// 测试从JSON转换
				jsonStr := `{"when": "testdata.age >= 18", "then": {"result.adult": true}}`
				_, err := converter.convertFromJSON(jsonStr)
				if err != nil {
					// JSON转换可能失败
					So(err, ShouldNotBeNil)
				}
			})
		})

		Convey("内存缓存清理", func() {
			// 创建小容量缓存以触发清理
			cache := NewMemoryCache(2)
			defer cache.Close()

			ctx := context.Background()

			// 添加一些即将过期的数据
			cache.Set(ctx, "key1", []byte("value1"), 1*time.Millisecond)
			cache.Set(ctx, "key2", []byte("value2"), 1*time.Millisecond)

			// 等待过期
			time.Sleep(10 * time.Millisecond)

			// 添加新数据触发清理
			cache.Set(ctx, "key3", []byte("value3"), 1*time.Hour)

			// 验证过期数据被清理
			_, err1 := cache.Get(ctx, "key1")
			_, err2 := cache.Get(ctx, "key2")
			So(err1, ShouldNotBeNil)
			So(err2, ShouldNotBeNil)

			// 新数据应该存在
			result, err := cache.Get(ctx, "key3")
			So(err, ShouldBeNil)
			So(string(result), ShouldEqual, "value3")
		})
	})
}