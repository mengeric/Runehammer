package runehammer

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestConfig 配置测试
func TestConfig(t *testing.T) {
	Convey("配置管理测试", t, func() {
		
		Convey("默认配置", func() {
			config := DefaultConfig()
			
			So(config, ShouldNotBeNil)
			So(config.tableName, ShouldEqual, "runehammer_rules")
			So(config.cacheTTL, ShouldEqual, 10*time.Minute)
			So(config.syncInterval, ShouldEqual, 5*time.Minute)
			So(config.autoMigrate, ShouldBeFalse)
			So(config.maxCacheSize, ShouldEqual, 1000)
			So(config.enableCache, ShouldBeTrue)
			So(config.redisDB, ShouldEqual, 0)
		})

		Convey("配置选项应用", func() {
			config := DefaultConfig()
			
			Convey("WithDB选项", func() {
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				WithDB(db)(config)
				
				So(config.GetDB(), ShouldEqual, db)
			})
			
			Convey("WithDSN选项", func() {
				dsn := "user:password@tcp(localhost:3306)/database?charset=utf8mb4"
				WithDSN(dsn)(config)
				
				So(config.dsn, ShouldEqual, dsn)
			})
			
			Convey("WithGormConfig选项", func() {
				gormConfig := &gorm.Config{SkipDefaultTransaction: true}
				WithGormConfig(gormConfig)(config)
				
				So(config.gormConfig, ShouldEqual, gormConfig)
			})
			
			Convey("WithAutoMigrate选项", func() {
				WithAutoMigrate()(config)
				
				So(config.GetAutoMigrate(), ShouldBeTrue)
			})
			
			Convey("WithTableName选项", func() {
				tableName := "custom_rules"
				WithTableName(tableName)(config)
				
				So(config.tableName, ShouldEqual, tableName)
			})
			
			Convey("WithCache选项", func() {
				cache := NewMemoryCache(100)
				WithCache(cache)(config)
				
				So(config.GetCache(), ShouldEqual, cache)
				So(config.enableCache, ShouldBeTrue)
				
				// 测试nil缓存
				WithCache(nil)(config)
				So(config.enableCache, ShouldBeFalse)
			})
			
			Convey("WithRedis选项", func() {
				addr := "localhost:6379"
				password := "password"
				db := 1
				WithRedis(addr, password, db)(config)
				
				So(config.redisAddr, ShouldEqual, addr)
				So(config.redisPass, ShouldEqual, password)
				So(config.redisDB, ShouldEqual, db)
			})
			
			Convey("WithCacheTTL选项", func() {
				ttl := 30 * time.Minute
				WithCacheTTL(ttl)(config)
				
				So(config.cacheTTL, ShouldEqual, ttl)
			})
			
			Convey("WithSyncInterval选项", func() {
				interval := 2 * time.Minute
				WithSyncInterval(interval)(config)
				
				So(config.GetSyncInterval(), ShouldEqual, interval)
			})
			
			Convey("WithLogger选项", func() {
				logger := &noopLogger{}
				WithLogger(logger)(config)
				
				So(config.GetLogger(), ShouldEqual, logger)
			})
			
			Convey("WithMaxCacheSize选项", func() {
				size := 500
				WithMaxCacheSize(size)(config)
				
				So(config.maxCacheSize, ShouldEqual, size)
			})
			
			Convey("WithDisableCache选项", func() {
				WithDisableCache()(config)
				
				So(config.enableCache, ShouldBeFalse)
			})
		})

		Convey("配置验证", func() {
			Convey("有效配置", func() {
				config := DefaultConfig()
				config.dsn = "test_dsn"
				
				err := config.Validate()
				So(err, ShouldBeNil)
			})
			
			Convey("无效配置 - 缺少数据库配置", func() {
				config := DefaultConfig()
				// 既没有DB实例也没有DSN
				
				err := config.Validate()
				So(err, ShouldEqual, ErrNoDatabaseConfig)
			})
			
			Convey("有DB实例时配置有效", func() {
				config := DefaultConfig()
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				config.db = db
				
				err := config.Validate()
				So(err, ShouldBeNil)
			})
		})

		Convey("数据库设置", func() {
			Convey("已有DB实例", func() {
				config := DefaultConfig()
				db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
				config.db = db
				
				err := config.SetupDB()
				So(err, ShouldBeNil)
				So(config.GetDB(), ShouldEqual, db)
			})
			
			Convey("无效DSN", func() {
				config := DefaultConfig()
				config.dsn = "invalid_dsn"
				
				err := config.SetupDB()
				So(err, ShouldNotBeNil)
			})
		})

		Convey("缓存设置", func() {
			Convey("禁用缓存", func() {
				config := DefaultConfig()
				config.enableCache = false
				
				err := config.SetupCache()
				So(err, ShouldBeNil)
				So(config.GetCache(), ShouldBeNil)
			})
			
			Convey("已有缓存实例", func() {
				config := DefaultConfig()
				cache := NewMemoryCache(100)
				config.cache = cache
				
				err := config.SetupCache()
				So(err, ShouldBeNil)
				So(config.GetCache(), ShouldEqual, cache)
			})
			
			Convey("降级到内存缓存", func() {
				config := DefaultConfig()
				// 没有Redis地址，应该降级到内存缓存
				
				err := config.SetupCache()
				So(err, ShouldBeNil)
				So(config.GetCache(), ShouldNotBeNil)
			})
		})

		Convey("动态配置", func() {
			Convey("默认动态配置", func() {
				dynamicConfig := DefaultDynamicConfig()
				
				So(dynamicConfig, ShouldNotBeNil)
				So(dynamicConfig.ParserConfig.DefaultSyntax, ShouldEqual, SyntaxTypeSQL)
				So(len(dynamicConfig.ParserConfig.SupportedSyntax), ShouldEqual, 2)
				So(dynamicConfig.ParserConfig.SupportedSyntax, ShouldContain, SyntaxTypeSQL)
				So(dynamicConfig.ParserConfig.SupportedSyntax, ShouldContain, SyntaxTypeJavaScript)
				
				So(dynamicConfig.CacheConfig.Enabled, ShouldBeTrue)
				So(dynamicConfig.CacheConfig.MaxSize, ShouldEqual, 500)
				So(dynamicConfig.CacheConfig.TTL, ShouldEqual, 5*time.Minute)
				
				So(dynamicConfig.ValidatorConfig.Enabled, ShouldBeTrue)
				So(dynamicConfig.ValidatorConfig.StrictMode, ShouldBeFalse)
				
				So(dynamicConfig.ExecutionConfig.EnableParallel, ShouldBeTrue)
				So(dynamicConfig.ExecutionConfig.MaxConcurrency, ShouldEqual, 10)
			})
			
			Convey("获取动态配置", func() {
				config := DefaultConfig()
				
				dynamicConfig := config.GetDynamicConfig()
				So(dynamicConfig, ShouldNotBeNil)
				So(dynamicConfig.ParserConfig.DefaultSyntax, ShouldEqual, SyntaxTypeSQL)
			})
		})

		Convey("动态配置选项", func() {
			config := DefaultConfig()
			
			Convey("WithDynamicConfig选项", func() {
				dynamicConfig := &DynamicConfig{
					ParserConfig: ParserConfig{
						DefaultSyntax: SyntaxTypeJavaScript,
					},
				}
				WithDynamicConfig(dynamicConfig)(config)
				
				So(config.dynamicConfig, ShouldEqual, dynamicConfig)
			})
			
			Convey("WithConverterConfig选项", func() {
				converterConfig := ConverterConfig{
					DefaultPriority: 100,
				}
				WithConverterConfig(converterConfig)(config)
				
				So(config.GetDynamicConfig().ConverterConfig.DefaultPriority, ShouldEqual, 100)
			})
			
			Convey("WithParserConfig选项", func() {
				parserConfig := ParserConfig{
					DefaultSyntax: SyntaxTypeJavaScript,
				}
				WithParserConfig(parserConfig)(config)
				
				So(config.GetDynamicConfig().ParserConfig.DefaultSyntax, ShouldEqual, SyntaxTypeJavaScript)
			})
			
			Convey("WithDynamicCacheConfig选项", func() {
				cacheConfig := DynamicCacheConfig{
					Enabled: false,
					MaxSize: 200,
				}
				WithDynamicCacheConfig(cacheConfig)(config)
				
				So(config.GetDynamicConfig().CacheConfig.Enabled, ShouldBeFalse)
				So(config.GetDynamicConfig().CacheConfig.MaxSize, ShouldEqual, 200)
			})
			
			Convey("WithValidatorConfig选项", func() {
				validatorConfig := ValidatorConfig{
					Enabled:    false,
					StrictMode: true,
				}
				WithValidatorConfig(validatorConfig)(config)
				
				So(config.GetDynamicConfig().ValidatorConfig.Enabled, ShouldBeFalse)
				So(config.GetDynamicConfig().ValidatorConfig.StrictMode, ShouldBeTrue)
			})
			
			Convey("WithExecutionConfig选项", func() {
				executionConfig := ExecutionConfig{
					EnableParallel: false,
					MaxConcurrency: 5,
				}
				WithExecutionConfig(executionConfig)(config)
				
				So(config.GetDynamicConfig().ExecutionConfig.EnableParallel, ShouldBeFalse)
				So(config.GetDynamicConfig().ExecutionConfig.MaxConcurrency, ShouldEqual, 5)
			})
			
			Convey("WithCustomFunctions选项", func() {
				functions := map[string]interface{}{
					"customFunc": func() string { return "test" },
				}
				WithCustomFunctions(functions)(config)
				
				So(config.GetDynamicConfig().CustomFunctions, ShouldNotBeEmpty)
				So(config.GetDynamicConfig().CustomFunctions["customFunc"], ShouldNotBeNil)
			})
			
			Convey("WithDefaultSyntax选项", func() {
				WithDefaultSyntax(SyntaxTypeJavaScript)(config)
				
				So(config.GetDynamicConfig().ParserConfig.DefaultSyntax, ShouldEqual, SyntaxTypeJavaScript)
			})
			
			Convey("WithSupportedSyntax选项", func() {
				WithSupportedSyntax(SyntaxTypeSQL)(config)
				
				So(len(config.GetDynamicConfig().ParserConfig.SupportedSyntax), ShouldEqual, 1)
				So(config.GetDynamicConfig().ParserConfig.SupportedSyntax[0], ShouldEqual, SyntaxTypeSQL)
			})
			
			Convey("WithCustomOperators选项", func() {
				operators := map[string]string{
					"customOp": "CustomOperator",
				}
				WithCustomOperators(operators)(config)
				
				So(config.GetDynamicConfig().ParserConfig.CustomOperators, ShouldNotBeEmpty)
				So(config.GetDynamicConfig().ParserConfig.CustomOperators["customOp"], ShouldEqual, "CustomOperator")
			})
			
			Convey("WithExecutionTimeout选项", func() {
				timeout := 60 * time.Second
				WithExecutionTimeout(timeout)(config)
				
				So(config.GetDynamicConfig().ExecutionConfig.ExecutionTimeout, ShouldEqual, timeout)
			})
			
			Convey("WithMaxConcurrency选项", func() {
				maxConcurrency := 20
				WithMaxConcurrency(maxConcurrency)(config)
				
				So(config.GetDynamicConfig().ExecutionConfig.MaxConcurrency, ShouldEqual, maxConcurrency)
			})
		})

		Convey("配置选项组合", func() {
			config := DefaultConfig()
			
			// 应用多个选项
			opts := []Option{
				WithTableName("test_rules"),
				WithCacheTTL(15 * time.Minute),
				WithMaxCacheSize(2000),
				WithAutoMigrate(),
				WithDefaultSyntax(SyntaxTypeJavaScript),
			}
			
			for _, opt := range opts {
				config.ApplyOption(opt)
			}
			
			So(config.tableName, ShouldEqual, "test_rules")
			So(config.cacheTTL, ShouldEqual, 15*time.Minute)
			So(config.maxCacheSize, ShouldEqual, 2000)
			So(config.autoMigrate, ShouldBeTrue)
			So(config.GetDynamicConfig().ParserConfig.DefaultSyntax, ShouldEqual, SyntaxTypeJavaScript)
		})
	})
}