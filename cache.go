package runehammer

import (
	"context"
	"encoding/json"
	"time"
)

// ============================================================================
// 缓存接口定义 - 统一的缓存抽象层
// ============================================================================

// Cache 缓存接口 - 支持Redis或其他实现的统一抽象
//
// 设计原则:
//   - 支持多种实现（Redis、内存等）
//   - 统一的错误处理机制
//   - 支持上下文传递和超时控制
//   - 简单易用的API设计
type Cache interface {
	// Get 获取缓存值
	//
	// 参数:
	//   ctx - 上下文，用于超时控制和取消操作
	//   key - 缓存键
	//
	// 返回值:
	//   []byte - 缓存的字节数据
	//   error  - 操作错误，键不存在时返回ErrCacheNotFound
	Get(ctx context.Context, key string) ([]byte, error)

	// Set 设置缓存值
	//
	// 参数:
	//   ctx   - 上下文，用于超时控制和取消操作
	//   key   - 缓存键
	//   value - 缓存值（字节数据）
	//   ttl   - 生存时间，过期后自动删除
	//
	// 返回值:
	//   error - 操作错误
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Del 删除缓存值
	//
	// 参数:
	//   ctx - 上下文，用于超时控制和取消操作
	//   key - 要删除的缓存键
	//
	// 返回值:
	//   error - 操作错误
	Del(ctx context.Context, key string) error

	// Close 关闭缓存连接
	//
	// 返回值:
	//   error - 关闭过程中的错误
	Close() error
}

// ============================================================================
// 缓存工具类 - 键构建器和序列化支持
// ============================================================================

// CacheKeyBuilder 缓存键构建器 - 统一的缓存键命名规范
//
// 命名规范:
//   - 使用项目名作为前缀
//   - 不同类型的数据使用不同的命名空间
//   - 键名清晰表达数据含义
type CacheKeyBuilder struct{}

// RuleKey 构建规则缓存键
//
// 参数:
//   bizCode - 业务码
//
// 返回值:
//   string - 格式化的缓存键
//
// 格式: runehammer:rule:{bizCode}
func (CacheKeyBuilder) RuleKey(bizCode string) string {
	return "runehammer:rule:" + bizCode
}

// MetaKey 构建元数据缓存键
//
// 参数:
//   bizCode - 业务码
//
// 返回值:
//   string - 格式化的缓存键
//
// 格式: runehammer:meta:{bizCode}
func (CacheKeyBuilder) MetaKey(bizCode string) string {
	return "runehammer:meta:" + bizCode
}

// ============================================================================
// 缓存数据结构 - 规则缓存项的序列化支持
// ============================================================================

// RuleCacheItem 规则缓存项 - 用于缓存规则数据的结构体
//
// 功能:
//   - 包装规则数据和元信息
//   - 支持JSON序列化和反序列化
//   - 版本控制和更新时间跟踪
type RuleCacheItem struct {
	Rules     []*Rule `json:"rules"`      // 规则列表
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
	Version   int       `json:"version"`    // 版本号
}

// ToBytes 序列化为字节数组 - 将结构体转换为可存储的字节数据
//
// 返回值:
//   []byte - 序列化后的字节数据
//   error  - 序列化过程中的错误
func (r *RuleCacheItem) ToBytes() ([]byte, error) {
	return json.Marshal(r)
}

// FromBytes 从字节数组反序列化 - 将字节数据转换回结构体
//
// 参数:
//   data - 序列化的字节数据
//
// 返回值:
//   error - 反序列化过程中的错误
func (r *RuleCacheItem) FromBytes(data []byte) error {
	return json.Unmarshal(data, r)
}