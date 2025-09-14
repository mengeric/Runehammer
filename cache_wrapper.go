package runehammer

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// 缓存包装器 - 支持二级缓存和故障降级的缓存方案
// ============================================================================

// CacheWrapper 缓存包装器 - 实现二级缓存机制
//
// 二级缓存策略:
//  1. 优先使用主缓存（通常是Redis）
//  2. 主缓存失败时降级到备用缓存（通常是内存缓存）
//  3. 写操作同时写入两级缓存
//  4. 提供故障自动降级和恢复
type CacheWrapper struct {
	primary   Cache         // 主缓存（如Redis）
	secondary Cache         // 备用缓存（如内存缓存）
	logger    Logger        // 日志记录器 - 最小化接口
}

// NewCacheWrapper 创建缓存包装器
//
// 参数:
//
//	primary   - 主缓存实例，通常为Redis
//	secondary - 备用缓存实例，通常为内存缓存
//	logger    - 日志记录器，可为nil
//
// 返回值:
//
//	Cache - 缓存接口实例
//
// 使用场景:
//   - 需要高可用性的缓存方案
//   - Redis故障时的自动降级
//   - 混合缓存策略（远程+本地）
func NewCacheWrapper(primary, secondary Cache, logger Logger) Cache {
	return &CacheWrapper{
		primary:   primary,
		secondary: secondary,
		logger:    logger,
	}
}

// Get 获取缓存值 - 支持二级缓存读取和故障降级
//
// 读取策略:
//  1. 优先从主缓存读取
//  2. 主缓存失败时从备用缓存读取
//  3. 记录缓存命中和失败情况
func (w *CacheWrapper) Get(ctx context.Context, key string) ([]byte, error) {
	// 尝试主缓存
	if w.primary != nil {
		data, err := w.primary.Get(ctx, key)
		if err == nil {
			return data, nil
		}

		// 记录主缓存失败日志
		if w.logger != nil {
			w.logger.Debugf(ctx, "主缓存读取失败，尝试备用缓存", "key", key, "error", err)
		}
	}

	// 降级到备用缓存
	if w.secondary != nil {
		return w.secondary.Get(ctx, key)
	}

	return nil, fmt.Errorf("cache key not found")
}

// Set 设置缓存值 - 支持双写策略和部分失败处理
//
// 写入策略:
//  1. 同时写入主缓存和备用缓存
//  2. 记录写入失败的情况
//  3. 只要有一个缓存写入成功就返回成功
func (w *CacheWrapper) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var errs []error

	// 设置主缓存
	if w.primary != nil {
		if err := w.primary.Set(ctx, key, value, ttl); err != nil {
			errs = append(errs, err)
			if w.logger != nil {
				w.logger.Warnf(ctx, "主缓存写入失败", "key", key, "error", err)
			}
		}
	}

	// 设置备用缓存
	if w.secondary != nil {
		if err := w.secondary.Set(ctx, key, value, ttl); err != nil {
			errs = append(errs, err)
			if w.logger != nil {
				w.logger.Warnf(ctx, "备用缓存写入失败", "key", key, "error", err)
			}
		}
	}

	// 如果两个缓存都失败，返回主缓存的错误
	if len(errs) == 2 {
		return errs[0]
	}

	return nil
}

// Del 删除缓存值 - 从所有级别的缓存中删除
//
// 删除策略:
//  1. 同时从主缓存和备用缓存删除
//  2. 忽略删除过程中的错误
//  3. 确保数据一致性
func (w *CacheWrapper) Del(ctx context.Context, key string) error {
	// 从主缓存删除
	if w.primary != nil {
		if err := w.primary.Del(ctx, key); err != nil && w.logger != nil {
			w.logger.Debugf(ctx, "主缓存删除失败", "key", key, "error", err)
		}
	}

	// 从备用缓存删除
	if w.secondary != nil {
		if err := w.secondary.Del(ctx, key); err != nil && w.logger != nil {
			w.logger.Debugf(ctx, "备用缓存删除失败", "key", key, "error", err)
		}
	}

	return nil
}

// Close 关闭所有缓存连接 - 释放所有级别的缓存资源
//
// 关闭策略:
//  1. 依次关闭主缓存和备用缓存
//  2. 记录关闭过程中的错误
//  3. 确保所有资源都得到释放
func (w *CacheWrapper) Close() error {
	var errs []error

	// 关闭主缓存
	if w.primary != nil {
		if err := w.primary.Close(); err != nil {
			errs = append(errs, err)
			if w.logger != nil {
				w.logger.Warnf(context.Background(), "主缓存关闭失败", "error", err)
			}
		}
	}

	// 关闭备用缓存
	if w.secondary != nil {
		if err := w.secondary.Close(); err != nil {
			errs = append(errs, err)
			if w.logger != nil {
				w.logger.Warnf(context.Background(), "备用缓存关闭失败", "error", err)
			}
		}
	}

	// 返回第一个错误
	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}
