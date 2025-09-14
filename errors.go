package runehammer

import "errors"

// ErrNoDatabaseConfig 未配置数据库错误
var ErrNoDatabaseConfig = errors.New("no database configuration provided")

// ErrInvalidConfig 无效配置错误  
var ErrInvalidConfig = errors.New("invalid configuration")