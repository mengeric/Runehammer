package runehammer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// 内存缓存实现 - 基于内存的缓存方案，支持过期和容量限制
// ============================================================================

// MemoryCache 内存缓存实现 - 适用于单机部署或小规模缓存需求
//
// 特性:
//   - 支持TTL过期机制
//   - 支持容量限制和LRU清理
//   - 异步清理过期项
//   - 线程安全操作
type MemoryCache struct {
	data     map[string]*cacheItem // 缓存数据存储
	mutex    sync.RWMutex         // 读写锁保护
	maxSize  int                  // 最大缓存条目数
	stopChan chan struct{}        // 停止信号通道
}

// cacheItem 缓存项 - 包含值和过期时间的数据结构
type cacheItem struct {
	Value     []byte    // 缓存的实际数据
	ExpiresAt time.Time // 过期时间
}

// NewMemoryCache 创建内存缓存实例
//
// 参数:
//   maxSize - 最大缓存条目数，超过时会触发清理机制
//
// 返回值:
//   Cache - 缓存接口实例
//
// 使用场景:
//   - 单机部署环境
//   - 开发测试环境
//   - Redis不可用时的降级方案
func NewMemoryCache(maxSize int) Cache {
	cache := &MemoryCache{
		data:     make(map[string]*cacheItem),
		maxSize:  maxSize,
		stopChan: make(chan struct{}),
	}

	// 启动后台清理goroutine
	go cache.cleanup()
	
	return cache
}

// Get 获取缓存值 - 支持过期检查和异步清理
func (m *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("cache key not found")
	}

	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		// 异步删除过期项，避免阻塞读操作
		go m.asyncDelete(key)
		return nil, fmt.Errorf("cache key not found")
	}

	return item.Value, nil
}

// Set 设置缓存值 - 支持容量管理和过期时间
func (m *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查容量限制并清理
	if len(m.data) >= m.maxSize {
		m.evictItems()
	}

	// 设置缓存项
	m.data[key] = &cacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Del 删除缓存值
func (m *MemoryCache) Del(ctx context.Context, key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.data, key)
	return nil
}

// Close 关闭缓存 - 停止后台清理任务
func (m *MemoryCache) Close() error {
	// 防止重复关闭channel
	select {
	case <-m.stopChan:
		// 已经关闭
	default:
		close(m.stopChan)
	}
	return nil
}

// asyncDelete 异步删除指定键 - 避免在读操作中阻塞
func (m *MemoryCache) asyncDelete(key string) {
	m.mutex.Lock()
	delete(m.data, key)
	m.mutex.Unlock()
}

// evictItems 清理部分缓存项 - 优先清理过期项，然后随机清理
//
// 清理策略:
//   1. 优先清理已过期的项
//   2. 如果仍超出限制，随机删除10%的项
func (m *MemoryCache) evictItems() {
	now := time.Now()
	
	// 第一轮：清理过期项
	for key, item := range m.data {
		if now.After(item.ExpiresAt) {
			delete(m.data, key)
		}
	}

	// 第二轮：如果仍然超出限制，随机删除一些项
	if len(m.data) >= m.maxSize {
		count := 0
		deleteCount := m.maxSize / 10 // 删除10%
		if deleteCount == 0 {
			deleteCount = 1 // 至少删除1个
		}
		
		for key := range m.data {
			delete(m.data, key)
			count++
			if count >= deleteCount {
				break
			}
		}
	}
}

// cleanup 定期清理过期项 - 后台任务，每5分钟执行一次
//
// 功能:
//   - 遍历所有缓存项
//   - 删除已过期的项
//   - 释放内存空间
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performCleanup()
		case <-m.stopChan:
			return
		}
	}
}

// performCleanup 执行清理操作
func (m *MemoryCache) performCleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	now := time.Now()
	cleanedCount := 0
	
	for key, item := range m.data {
		if now.After(item.ExpiresAt) {
			delete(m.data, key)
			cleanedCount++
		}
	}
	
	// 可以添加清理日志，但需要logger支持
	// if cleanedCount > 0 {
	//     log.Printf("Cleaned %d expired cache items", cleanedCount)
	// }
}