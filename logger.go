package runehammer

import (
	"context"
	"fmt"
)

// ============================================================================
// 日志接口定义 - 统一的日志抽象层
// ============================================================================

// Logger 日志记录接口 - 支持结构化日志和不同级别的日志输出
//
// 使用者可以实现此接口来集成自己的日志系统，例如：
//   - logrus
//   - zap
//   - slog
//   - 或任何其他日志库
//
// 设计原则:
//   - 支持上下文传递
//   - 结构化日志输出
//   - 多级别日志控制
//   - 简单易用的API
type Logger interface {
	// Debugf 调试级别日志
	Debugf(ctx context.Context, msg string, keyvals ...any)

	// Infof 信息级别日志
	Infof(ctx context.Context, msg string, keyvals ...any)

	// Warnf 警告级别日志
	Warnf(ctx context.Context, msg string, keyvals ...any)

	// Errorf 错误级别日志
	Errorf(ctx context.Context, msg string, keyvals ...any)
}

// ============================================================================
// 默认日志实现
// ============================================================================

// NoopLogger 空日志记录器 - 不输出任何日志的实现
type NoopLogger struct{}

// NewNoopLogger 创建空日志记录器
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// Debugf 调试日志 - 空实现
func (n *NoopLogger) Debugf(ctx context.Context, msg string, keyvals ...any) {
	// 空实现
}

// Infof 信息日志 - 空实现
func (n *NoopLogger) Infof(ctx context.Context, msg string, keyvals ...any) {
	// 空实现
}

// Warnf 警告日志 - 空实现
func (n *NoopLogger) Warnf(ctx context.Context, msg string, keyvals ...any) {
	// 空实现
}

// Errorf 错误日志 - 空实现
func (n *NoopLogger) Errorf(ctx context.Context, msg string, keyvals ...any) {
	// 空实现
}

// DefaultLogger 默认日志记录器 - 使用fmt.Printf输出
type DefaultLogger struct{}

// NewDefaultLogger 创建默认日志记录器
func NewDefaultLogger() Logger {
	return &DefaultLogger{}
}

// Debugf 调试日志 - 使用fmt.Printf输出
func (n *DefaultLogger) Debugf(ctx context.Context, msg string, keyvals ...any) {
	fmt.Printf("[DEBUG] %s %v\n", msg, keyvals)
}

// Infof 信息日志 - 使用fmt.Printf输出
func (n *DefaultLogger) Infof(ctx context.Context, msg string, keyvals ...any) {
	fmt.Printf("[INFO] %s %v\n", msg, keyvals)
}

// Warnf 警告日志 - 使用fmt.Printf输出
func (n *DefaultLogger) Warnf(ctx context.Context, msg string, keyvals ...any) {
	fmt.Printf("[WARN] %s %v\n", msg, keyvals)
}

// Errorf 错误日志 - 使用fmt.Printf输出
func (n *DefaultLogger) Errorf(ctx context.Context, msg string, keyvals ...any) {
	fmt.Printf("[ERROR] %s %v\n", msg, keyvals)
}