// Package data 日志模块数据层 —— 活动日志与系统日志的持久化对象与仓储实现
package data

import (
	"context"

	"nop-go/services/admin-service/internal/biz"

	"gorm.io/gorm"
)

// ==================== 持久化对象（PO） ====================

// ActivityLogPO 活动日志持久化对象 —— 映射数据库 activity_logs 表
type ActivityLogPO struct {
	ID        uint   `gorm:"primaryKey" db:"id"`
	UserID    uint   `gorm:"column:user_id;type:bigint;index;not null" db:"user_id"` // 操作用户ID
	UserName  string `gorm:"column:user_name;type:varchar(50)" db:"user_name"`         // 用户名
	Action    string `gorm:"column:action;type:varchar(50);index" db:"action"`      // 操作类型
	Resource  string `gorm:"column:resource;type:varchar(200)" db:"resource"`         // 操作资源
	IP        string `gorm:"column:ip;type:varchar(50)" db:"ip"`                // IP地址
	UserAgent string `gorm:"column:user_agent;type:varchar(500)" db:"user_agent"`       // 用户代理
	Detail    string `gorm:"column:detail;type:text" db:"detail"`                   // 操作详情
	Status    int    `gorm:"column:status;type:tinyint;default:1" db:"status"`      // 操作结果
	CreatedAt string `gorm:"column:created_at;index" db:"created_at"`
}

// TableName 指定活动日志表名
func (ActivityLogPO) TableName() string { return "activity_logs" }

// SystemLogPO 系统日志持久化对象 —— 映射数据库 system_logs 表
type SystemLogPO struct {
	ID        uint   `gorm:"primaryKey" db:"id"`
	Level     string `gorm:"column:level;type:varchar(20);index" db:"level"`  // 日志级别
	Module    string `gorm:"column:module;type:varchar(50);index" db:"module"` // 所属模块
	Message   string `gorm:"column:message;type:text" db:"message"`              // 日志消息
	Stack     string `gorm:"column:stack;type:text" db:"stack"`                // 堆栈信息
	Hostname  string `gorm:"column:hostname;type:varchar(100)" db:"hostname"`     // 主机名
	CreatedAt string `gorm:"column:created_at;index" db:"created_at"`
}

// TableName 指定系统日志表名
func (SystemLogPO) TableName() string { return "system_logs" }

// ==================== PO ↔ Entity 转换 ====================

// toEntity 将 ActivityLogPO 转换为 biz.ActivityLog 领域实体
func (a *ActivityLogPO) toEntity() *biz.ActivityLog {
	return &biz.ActivityLog{
		ID:        a.ID,
		UserID:    a.UserID,
		UserName:  a.UserName,
		Action:    a.Action,
		Resource:  a.Resource,
		IP:        a.IP,
		UserAgent: a.UserAgent,
		Detail:    a.Detail,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
	}
}

// toEntity 将 SystemLogPO 转换为 biz.SystemLog 领域实体
func (s *SystemLogPO) toEntity() *biz.SystemLog {
	return &biz.SystemLog{
		ID:        s.ID,
		Level:     s.Level,
		Module:    s.Module,
		Message:   s.Message,
		Stack:     s.Stack,
		Hostname:  s.Hostname,
		CreatedAt: s.CreatedAt,
	}
}

// activityLogToPO 将 biz.ActivityLog 领域实体转换为 ActivityLogPO
func activityLogToPO(a *biz.ActivityLog) *ActivityLogPO {
	return &ActivityLogPO{
		ID:        a.ID,
		UserID:    a.UserID,
		UserName:  a.UserName,
		Action:    a.Action,
		Resource:  a.Resource,
		IP:        a.IP,
		UserAgent: a.UserAgent,
		Detail:    a.Detail,
		Status:    a.Status,
		CreatedAt: a.CreatedAt,
	}
}

// systemLogToPO 将 biz.SystemLog 领域实体转换为 SystemLogPO
func systemLogToPO(s *biz.SystemLog) *SystemLogPO {
	return &SystemLogPO{
		ID:        s.ID,
		Level:     s.Level,
		Module:    s.Module,
		Message:   s.Message,
		Stack:     s.Stack,
		Hostname:  s.Hostname,
		CreatedAt: s.CreatedAt,
	}
}

// ==================== 仓储实现 ====================

// activityLogRepo 活动日志仓储实现
type activityLogRepo struct {
	db *gorm.DB
}

// NewActivityLogRepo 创建活动日志仓储
func NewActivityLogRepo(db *gorm.DB) biz.ActivityLogRepo {
	return &activityLogRepo{db: db}
}

func (r *activityLogRepo) Create(ctx context.Context, log *biz.ActivityLog) error {
	po := activityLogToPO(log)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *activityLogRepo) GetByID(ctx context.Context, id uint) (*biz.ActivityLog, error) {
	var po ActivityLogPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *activityLogRepo) List(ctx context.Context, userID uint, action string, page, pageSize int) ([]*biz.ActivityLog, int64, error) {
	var pos []*ActivityLogPO
	var total int64
	q := r.db.WithContext(ctx).Model(&ActivityLogPO{})
	// 用户ID过滤
	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	// 动作类型过滤
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.ActivityLog, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}

// systemLogRepo 系统日志仓储实现
type systemLogRepo struct {
	db *gorm.DB
}

// NewSystemLogRepo 创建系统日志仓储
func NewSystemLogRepo(db *gorm.DB) biz.SystemLogRepo {
	return &systemLogRepo{db: db}
}

func (r *systemLogRepo) Create(ctx context.Context, log *biz.SystemLog) error {
	po := systemLogToPO(log)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *systemLogRepo) GetByID(ctx context.Context, id uint) (*biz.SystemLog, error) {
	var po SystemLogPO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.toEntity(), nil
}

func (r *systemLogRepo) List(ctx context.Context, level, module string, page, pageSize int) ([]*biz.SystemLog, int64, error) {
	var pos []*SystemLogPO
	var total int64
	q := r.db.WithContext(ctx).Model(&SystemLogPO{})
	// 级别过滤
	if level != "" {
		q = q.Where("level = ?", level)
	}
	// 模块过滤
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Order("id DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*biz.SystemLog, 0, len(pos))
	for _, po := range pos {
		result = append(result, po.toEntity())
	}
	return result, total, nil
}
