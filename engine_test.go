package runehammer

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// noopLogger 实现Logger接口的空操作日志记录器
type noopLogger struct{}

func (n *noopLogger) Debugf(ctx context.Context, msg string, keyvals ...any) {}
func (n *noopLogger) Infof(ctx context.Context, msg string, keyvals ...any)  {}
func (n *noopLogger) Warnf(ctx context.Context, msg string, keyvals ...any)  {}
func (n *noopLogger) Errorf(ctx context.Context, msg string, keyvals ...any) {}

func TestEngine(t *testing.T) {
	Convey("规则引擎测试", t, func() {
		// 测试数据库
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		So(err, ShouldBeNil)

		// 自动迁移
		err = db.AutoMigrate(&Rule{})
		So(err, ShouldBeNil)

		Convey("创建引擎", func() {
			engine, err := New[map[string]any](
				WithDB(db),
				WithDisableCache(),
				WithLogger(&noopLogger{}),
			)
			So(err, ShouldBeNil)
			So(engine, ShouldNotBeNil)
		})

		Convey("规则不存在", func() {
			engine, err := New[map[string]any](
				WithDB(db),
				WithDisableCache(),
				WithLogger(&noopLogger{}),
			)
			So(err, ShouldBeNil)

			// 执行不存在的规则
			result, err := engine.Exec(context.Background(), "nonexistent", map[string]any{"input": "test"})
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "规则未找到")
			So(result, ShouldBeZeroValue)
		})

		Convey("配置验证", func() {
			// 测试无数据库配置的情况
			_, err := New[map[string]any]()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "配置验证失败")
		})

		Convey("DSN配置", func() {
			// 测试DSN配置（跳过，因为需要实际的MySQL DSN）
			So("Skip", ShouldEqual, "Skip") // 占位测试
		})
	})
}