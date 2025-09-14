package runehammer

import (
	"context"
	"fmt"
)

// ============================================================================
// 生命周期管理 - 处理引擎的启动、同步和关闭
// ============================================================================

// startSync 启动同步任务 - 定期同步规则缓存
//
// 同步功能:
//  1. 定期清理过期缓存
//  2. 预热热点规则
//  3. 检查规则更新
//
// 返回值:
//
//	error - 启动过程中的错误
func (e *engineImpl[T]) StartSync() error {
	if e.config.GetSyncInterval() <= 0 {
		// 未配置同步间隔，不启动同步任务
		return nil
	}

	// 添加同步任务到定时调度器
	_, err := e.cron.AddFunc(fmt.Sprintf("@every %s", e.config.GetSyncInterval()), func() {
		if err := e.syncRules(); err != nil && e.logger != nil {
			e.logger.Errorf(context.Background(), "规则同步失败", "error", err)
		}
	})

	if err != nil {
		return fmt.Errorf("添加同步任务失败: %w", err)
	}

	// 启动定时调度器
	e.cron.Start()

	if e.logger != nil {
		e.logger.Infof(context.Background(), "同步任务已启动", "interval", e.config.GetSyncInterval())
	}

	return nil
}

// syncRules 同步规则 - 执行实际的同步逻辑
//
// 同步策略:
//  1. 获取所有活跃的业务码
//  2. 检查规则是否有更新
//  3. 清理过期的编译缓存
//  4. 预热重要规则
//
// 返回值:
//
//	error - 同步过程中的错误
func (e *engineImpl[T]) syncRules() error {
	ctx := context.Background()

	if e.logger != nil {
		e.logger.Debugf(ctx, "开始执行规则同步")
	}

	// 这里可以实现具体的同步逻辑
	// 例如：
	// 1. 获取数据库中所有规则的更新时间
	// 2. 与缓存中的版本进行比较
	// 3. 更新变化的规则缓存
	// 4. 清理已删除规则的缓存

	// 示例：清理编译缓存（可以根据实际需求调整）
	e.clearExpiredKnowledgeBases()

	if e.logger != nil {
		e.logger.Debugf(ctx, "规则同步完成")
	}

	return nil
}

// clearExpiredKnowledgeBases 清理过期的编译缓存
//
// 清理策略:
//   - 定期清理编译后的知识库缓存
//   - 释放内存空间
//   - 强制重新编译以获取最新规则
func (e *engineImpl[T]) clearExpiredKnowledgeBases() {
	// 清理所有编译缓存，强制重新编译
	// 这是一个简单的实现，生产环境中可以更智能地决定清理策略
	e.knowledgeBases.Range(func(key, value interface{}) bool {
		e.knowledgeBases.Delete(key)
		return true
	})
}

// refreshCache 刷新指定业务码的缓存
//
// 参数:
//
//	bizCode - 业务码
//
// 功能:
//  1. 从数据库重新加载规则
//  2. 更新缓存
//  3. 清理编译缓存
//
// 返回值:
//
//	error - 刷新过程中的错误
func (e *engineImpl[T]) refreshCache(bizCode string) error {
	ctx := context.Background()

	// 清理编译缓存
	e.knowledgeBases.Delete(bizCode)

	// 清理规则缓存
	if e.cache != nil {
		cacheKey := e.cacheKeys.RuleKey(bizCode)
		if err := e.cache.Del(ctx, cacheKey); err != nil && e.logger != nil {
			e.logger.Warnf(ctx, "清理规则缓存失败", "bizCode", bizCode, "error", err)
		}
	}

	// 预热：重新加载规则到缓存
	_, err := e.getRules(ctx, bizCode)
	if err != nil {
		return fmt.Errorf("预热规则缓存失败: %w", err)
	}

	if e.logger != nil {
		e.logger.Infof(ctx, "缓存刷新完成", "bizCode", bizCode)
	}

	return nil
}

// getStats 获取引擎统计信息
//
// 返回值:
//
//	map[string]interface{} - 统计信息
//
// 统计项目:
//   - 编译缓存条目数
//   - 引擎状态
//   - 运行时长等
func (e *engineImpl[T]) getStats() map[string]interface{} {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	// 统计编译缓存条目数
	kbCount := 0
	e.knowledgeBases.Range(func(key, value interface{}) bool {
		kbCount++
		return true
	})

	return map[string]interface{}{
		"closed":          e.closed,
		"knowledge_bases": kbCount,
		"sync_interval":   e.config.GetSyncInterval(),
		"cache_enabled":   e.cache != nil,
		"logger_enabled":  e.logger != nil,
	}
}
