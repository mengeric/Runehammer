package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================================================
// Redis缓存实现 - 基于Redis的分布式缓存方案
// ============================================================================

// RedisCache Redis缓存实现 - 适用于分布式部署和大规模缓存需求
//
// 特性:
//   - 分布式缓存支持
//   - Redis原生TTL机制
//   - 高性能和高可用
//   - 支持集群和哨兵模式
type RedisCache struct {
	client *redis.Client // Redis客户端连接
}

// NewRedisCache 创建Redis缓存实例
//
// 参数:
//   client - 已配置的Redis客户端实例
//
// 返回值:
//   Cache - 缓存接口实例
//
// 使用场景:
//   - 生产环境分布式部署
//   - 大规模缓存需求
//   - 多服务共享缓存
//   - 需要持久化的缓存
//
// 注意事项:
//   - 确保Redis服务可用
//   - 合理配置连接池
//   - 注意网络延迟影响
func NewRedisCache(client *redis.Client) Cache {
	return &RedisCache{
		client: client,
	}
}

// Get 获取缓存值 - 从Redis获取指定键的值
//
// 参数:
//   ctx - 上下文，用于超时控制和取消操作
//   key - 缓存键
//
// 返回值:
//   []byte - 缓存的字节数据
//   error  - 操作错误，键不存在时返回ErrCacheNotFound
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	result := r.client.Get(ctx, key)
	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nil, fmt.Errorf("cache key not found")
		}
		return nil, result.Err()
	}
	
	return result.Bytes()
}

// Set 设置缓存值 - 将键值对存储到Redis，支持TTL
//
// 参数:
//   ctx   - 上下文，用于超时控制和取消操作  
//   key   - 缓存键
//   value - 缓存值（字节数据）
//   ttl   - 生存时间，过期后自动删除
//
// 返回值:
//   error - 操作错误
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Del 删除缓存值 - 从Redis删除指定键
//
// 参数:
//   ctx - 上下文，用于超时控制和取消操作
//   key - 要删除的缓存键
//
// 返回值:
//   error - 操作错误
func (r *RedisCache) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Close 关闭Redis连接 - 释放客户端连接资源
//
// 返回值:
//   error - 关闭过程中的错误
//
// 注意:
//   - 关闭后不能再进行任何操作
//   - 建议在应用程序退出时调用
func (r *RedisCache) Close() error {
	return r.client.Close()
}