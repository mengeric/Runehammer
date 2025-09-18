package rule

//go:generate mockgen -source=rule.go -destination=rule_mapper_mock.go -package=rule

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// 规则数据模型 - 数据库表结构定义
// ============================================================================

// Rule 规则模型 - 对应数据库中的规则表
//
// 表名：runehammer_rules（可通过配置修改）
// 主要功能：存储GRL规则定义和元数据
type Rule struct {
	// 基础字段
	ID      uint64 `gorm:"primaryKey;autoIncrement" json:"id"`      // 主键ID
	BizCode string `gorm:"size:100;not null;index" json:"biz_code"` // 业务码，用于分组规则
	Name    string `gorm:"size:200;not null" json:"name"`           // 规则名称

	// 规则内容
	GRL string `gorm:"type:text;not null" json:"grl"` // GRL规则内容

	// 版本和状态
	Version int  `gorm:"default:1" json:"version"` // 规则版本号
	Enabled bool `gorm:"not null" json:"enabled"`  // 是否启用

	// 时间戳
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"` // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"` // 更新时间

	// 可选字段
	Description string `gorm:"size:500" json:"description"` // 规则描述
	CreatedBy   string `gorm:"size:100" json:"created_by"`  // 创建者
	UpdatedBy   string `gorm:"size:100" json:"updated_by"`  // 更新者
}

// TableName 自定义表名
func (Rule) TableName() string {
	return "runehammer_rules"
}

// ============================================================================
// 规则数据访问接口 - 统一的数据访问抽象层
// ============================================================================

// RuleMapper 规则数据访问接口 - 定义规则相关的数据库操作
//
// 设计原则:
//   - 接口驱动设计，便于测试和扩展
//   - 支持上下文传递
//   - 简单实用的方法定义
type RuleMapper interface {
	// FindByBizCode 根据业务码查找规则
	//
	// 参数:
	//   ctx     - 上下文，用于超时控制和取消操作
	//   bizCode - 业务码
	//
	// 返回值:
	//   []*Rule - 规则列表
	//   error   - 查询错误
	FindByBizCode(ctx context.Context, bizCode string) ([]*Rule, error)
}

// ============================================================================
// 规则数据访问实现 - GORM实现
// ============================================================================

// ruleMapperImpl 规则数据访问实现
type ruleMapperImpl struct {
	db *gorm.DB // GORM数据库连接
}

// NewRuleMapper 创建规则数据访问实例
//
// 参数:
//
//	db - GORM数据库连接实例
//

// 返回值:
//
//	RuleMapper - 规则数据访问接口
func NewRuleMapper(db *gorm.DB) RuleMapper {
	return &ruleMapperImpl{
		db: db,
	}
}

// FindByBizCode 根据业务码查找规则
func (r *ruleMapperImpl) FindByBizCode(ctx context.Context, bizCode string) ([]*Rule, error) {
	var rules []*Rule

	// 查询启用的规则，按版本号降序排列
	err := r.db.WithContext(ctx).
		Where("biz_code = ? AND enabled = ?", bizCode, true).
		Order("version DESC").
		Find(&rules).Error

	if err != nil {
		return nil, err
	}

	return rules, nil
}
