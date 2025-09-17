package runehammer

import (
	"fmt"
	"time"
	
	"gitee.com/damengde/runehammer/config"
)

// ============================================================================
// SQLite 测试配置 - 专为测试环境提供优化的SQLite配置
// ============================================================================
//
// 注意：本文件专为测试使用，核心业务代码不依赖此文件
// 提供各种测试场景的SQLite配置预设，简化测试代码的数据库配置

// SQLiteTestConfig SQLite测试数据库配置选项
type SQLiteTestConfig struct {
	// 基础配置
	FilePath string // 数据库文件路径，空字符串表示内存数据库
	
	// 性能优化配置
	JournalMode    string        // 日志模式: DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF
	SynchronousMode string       // 同步模式: OFF, NORMAL, FULL, EXTRA
	CacheSize      int           // 缓存大小（页数，负数表示KB）
	TempStore      string        // 临时存储: DEFAULT, FILE, MEMORY
	LockingMode    string        // 锁定模式: NORMAL, EXCLUSIVE
	
	// 连接池配置
	MaxOpenConns   int           // 最大打开连接数
	MaxIdleConns   int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	
	// 事务配置
	BusyTimeout    time.Duration // 忙碌超时时间
	
	// 功能开关
	ForeignKeys    bool          // 是否启用外键约束
	CaseSensitiveLike bool       // LIKE操作是否区分大小写
	AutoVacuum     string        // 自动清理模式: NONE, FULL, INCREMENTAL
}

// SQLiteConfig 向后兼容的类型别名
// Deprecated: 使用 SQLiteTestConfig 替代，命名更清晰
type SQLiteConfig = SQLiteTestConfig

// ProductionSQLiteConfig 返回生产环境的默认SQLite配置
// 注意：此函数仅为测试提供生产环境模拟配置，实际生产环境请直接使用WithDSN()
func ProductionSQLiteConfig() *SQLiteTestConfig {
	return &SQLiteTestConfig{
		// 性能优化
		JournalMode:     "WAL",          // 写前日志，提高并发性能
		SynchronousMode: "NORMAL",       // 平衡性能和安全性
		CacheSize:       -64000,         // 64MB 缓存
		TempStore:       "MEMORY",       // 临时数据存储在内存中
		LockingMode:     "NORMAL",       // 标准锁定模式
		
		// 连接池配置
		MaxOpenConns:    25,                    // 最大25个连接
		MaxIdleConns:    10,                    // 最大10个空闲连接
		ConnMaxLifetime: time.Hour,             // 连接最大生命周期1小时
		
		// 事务配置
		BusyTimeout:     30 * time.Second,      // 30秒忙碌超时
		
		// 功能开关
		ForeignKeys:       true,               // 启用外键约束
		CaseSensitiveLike: false,              // LIKE不区分大小写
		AutoVacuum:        "INCREMENTAL",      // 增量清理
	}
}

// TestMemoryOnlyConfig 返回测试专用的纯内存SQLite配置
func TestMemoryOnlyConfig() *SQLiteTestConfig {
	config := ProductionSQLiteConfig()
	config.FilePath = ":memory:"
	config.JournalMode = "MEMORY"     // 内存模式下使用内存日志
	config.SynchronousMode = "OFF"    // 内存模式下不需要同步
	config.TempStore = "MEMORY"       // 临时数据在内存中
	config.CacheSize = -32000         // 32MB 缓存（内存模式下较小）
	config.BusyTimeout = 5 * time.Second // 较短的超时时间
	return config
}

// TestSharedMemoryConfig 返回测试专用的共享内存SQLite配置
func TestSharedMemoryConfig(testName string) *SQLiteTestConfig {
	config := TestMemoryOnlyConfig()
	// 使用共享缓存的内存数据库，允许多个连接访问同一数据库
	config.FilePath = fmt.Sprintf("file:%s_test.db?mode=memory&cache=shared&_fk=1", testName)
	return config
}

// BuildDSN 根据配置构建SQLite DSN字符串
func (c *SQLiteTestConfig) BuildDSN() string {
	dsn := c.FilePath
	
	// 如果是文件路径且不包含参数，添加基本参数
	if c.FilePath != ":memory:" && !contains(c.FilePath, "?") {
		dsn += "?"
	} else if c.FilePath != ":memory:" && contains(c.FilePath, "?") {
		dsn += "&"
	} else {
		// 内存数据库的情况
		return dsn
	}
	
	// 构建参数字符串
	params := []string{}
	
	if c.JournalMode != "" {
		params = append(params, fmt.Sprintf("_journal_mode=%s", c.JournalMode))
	}
	
	if c.SynchronousMode != "" {
		params = append(params, fmt.Sprintf("_synchronous=%s", c.SynchronousMode))
	}
	
	if c.CacheSize != 0 {
		params = append(params, fmt.Sprintf("_cache_size=%d", c.CacheSize))
	}
	
	if c.TempStore != "" {
		params = append(params, fmt.Sprintf("_temp_store=%s", c.TempStore))
	}
	
	if c.LockingMode != "" {
		params = append(params, fmt.Sprintf("_locking_mode=%s", c.LockingMode))
	}
	
	if c.BusyTimeout > 0 {
		params = append(params, fmt.Sprintf("_busy_timeout=%d", int(c.BusyTimeout.Milliseconds())))
	}
	
	if c.ForeignKeys {
		params = append(params, "_fk=1")
	}
	
	if c.CaseSensitiveLike {
		params = append(params, "_case_sensitive_like=1")
	}
	
	if c.AutoVacuum != "" && c.AutoVacuum != "NONE" {
		params = append(params, fmt.Sprintf("_auto_vacuum=%s", c.AutoVacuum))
	}
	
	// 连接参数
	for i, param := range params {
		if i == 0 {
			dsn += param
		} else {
			dsn += "&" + param
		}
	}
	
	return dsn
}

// WithOptimizedSQLite 返回使用优化SQLite配置的选项（测试专用）
func WithOptimizedSQLite(sqliteConfig *SQLiteTestConfig) Option {
	return func(c *config.Config) {
		if sqliteConfig == nil {
			sqliteConfig = ProductionSQLiteConfig()
		}
		c.DSN = "sqlite:" + sqliteConfig.BuildDSN()
	}
}

// WithTestSQLite 返回测试用的SQLite配置选项
func WithTestSQLite(testName string) Option {
	sqliteConfig := TestSharedMemoryConfig(testName)
	return WithOptimizedSQLite(sqliteConfig)
}

// contains 辅助函数
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}