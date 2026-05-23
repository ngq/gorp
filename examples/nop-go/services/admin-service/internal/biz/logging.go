// Package biz 日志模块业务层 —— 活动日志与系统日志的核心领域逻辑
package biz

import "context"

// ==================== 领域实体 ====================

// ActivityLog 活动日志实体 —— 记录用户的业务操作行为
type ActivityLog struct {
	ID         uint   `json:"id"`
	UserID     uint   `json:"user_id"`       // 操作用户ID
	UserName   string `json:"user_name"`     // 操作用户名
	Action     string `json:"action"`        // 操作类型，如 "login", "create_order"
	Resource   string `json:"resource"`      // 操作的资源，如 "/api/orders"
	IP         string `json:"ip"`            // 操作IP地址
	UserAgent  string `json:"user_agent"`    // 用户代理
	Detail     string `json:"detail"`        // 操作详情（JSON格式）
	Status     int    `json:"status"`        // 操作结果：0=失败 1=成功
	CreatedAt  string `json:"created_at"`
}

// SystemLog 系统日志实体 —— 记录系统运行状态与异常
type SystemLog struct {
	ID        uint   `json:"id"`
	Level     string `json:"level"`         // 日志级别：debug/info/warn/error/fatal
	Module    string `json:"module"`        // 所属模块
	Message   string `json:"message"`       // 日志消息
	Stack     string `json:"stack"`         // 堆栈信息（仅error级别）
	Hostname  string `json:"hostname"`      // 主机名
	CreatedAt string `json:"created_at"`
}

// ==================== 仓储接口 ====================

// ActivityLogRepo 活动日志仓储接口 —— 定义活动日志数据访问的契约
type ActivityLogRepo interface {
	// Create 创建活动日志
	Create(ctx context.Context, log *ActivityLog) error
	// GetByID 根据ID获取活动日志
	GetByID(ctx context.Context, id uint) (*ActivityLog, error)
	// List 获取活动日志列表，支持用户ID、动作类型过滤
	List(ctx context.Context, userID uint, action string, page, pageSize int) ([]*ActivityLog, int64, error)
}

// SystemLogRepo 系统日志仓储接口 —— 定义系统日志数据访问的契约
type SystemLogRepo interface {
	// Create 创建系统日志
	Create(ctx context.Context, log *SystemLog) error
	// GetByID 根据ID获取系统日志
	GetByID(ctx context.Context, id uint) (*SystemLog, error)
	// List 获取系统日志列表，支持级别和模块过滤
	List(ctx context.Context, level, module string, page, pageSize int) ([]*SystemLog, int64, error)
}

// ==================== 用例 ====================

// LoggingUseCase 日志模块用例 —— 封装活动日志与系统日志的业务逻辑
type LoggingUseCase struct {
	activityRepo ActivityLogRepo // 活动日志仓储
	systemRepo   SystemLogRepo   // 系统日志仓储
}

// NewLoggingUseCase 创建日志模块用例
func NewLoggingUseCase(activityRepo ActivityLogRepo, systemRepo SystemLogRepo) *LoggingUseCase {
	return &LoggingUseCase{
		activityRepo: activityRepo,
		systemRepo:   systemRepo,
	}
}

// ---------- 活动日志用例方法 ----------

// CreateActivityLog 创建活动日志
func (uc *LoggingUseCase) CreateActivityLog(ctx context.Context, log *ActivityLog) error {
	return uc.activityRepo.Create(ctx, log)
}

// GetActivityLog 根据ID获取活动日志
func (uc *LoggingUseCase) GetActivityLog(ctx context.Context, id uint) (*ActivityLog, error) {
	return uc.activityRepo.GetByID(ctx, id)
}

// ListActivityLogs 获取活动日志列表
func (uc *LoggingUseCase) ListActivityLogs(ctx context.Context, userID uint, action string, page, pageSize int) ([]*ActivityLog, int64, error) {
	return uc.activityRepo.List(ctx, userID, action, page, pageSize)
}

// ---------- 系统日志用例方法 ----------

// CreateSystemLog 创建系统日志
func (uc *LoggingUseCase) CreateSystemLog(ctx context.Context, log *SystemLog) error {
	return uc.systemRepo.Create(ctx, log)
}

// GetSystemLog 根据ID获取系统日志
func (uc *LoggingUseCase) GetSystemLog(ctx context.Context, id uint) (*SystemLog, error) {
	return uc.systemRepo.GetByID(ctx, id)
}

// ListSystemLogs 获取系统日志列表
func (uc *LoggingUseCase) ListSystemLogs(ctx context.Context, level, module string, page, pageSize int) ([]*SystemLog, int64, error) {
	return uc.systemRepo.List(ctx, level, module, page, pageSize)
}
