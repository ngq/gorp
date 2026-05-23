// Package biz 定义 GDPR 相关的业务领域层。
// 本文件合并了原 gdpr 服务的核心业务逻辑：GDPR 数据删除请求。
package biz

import (
	"context"
	"time"
)

// Gdpr GDPR 数据删除请求领域模型
// 用于记录用户发起的个人数据删除请求及其处理状态
type Gdpr struct {
	ID           uint       // 请求唯一标识
	UserID       uint       // 所属用户 ID
	RequestType  string     // 请求类型：delete-数据删除, export-数据导出
	Status       string     // 状态：pending-待处理, processing-处理中, completed-已完成, rejected-已拒绝
	Reason       string     // 请求原因
	ReviewedBy   *uint      // 审核人 ID，nil 表示未审核
	ReviewedAt   *time.Time // 审核时间，nil 表示未审核
	CompletedAt  *time.Time // 完成时间，nil 表示未完成
	CreatedAt    time.Time  // 创建时间
	UpdatedAt    time.Time  // 更新时间
}

// ======================== 仓储接口 ========================

// GdprRepository GDPR 仓储接口，定义 GDPR 请求数据的持久化操作
type GdprRepository interface {
	// Create 创建 GDPR 请求
	Create(ctx context.Context, gdpr *Gdpr) (*Gdpr, error)
	// Update 更新 GDPR 请求
	Update(ctx context.Context, gdpr *Gdpr) (*Gdpr, error)
	// Delete 删除 GDPR 请求
	Delete(ctx context.Context, id uint) error
	// GetByID 根据 ID 获取 GDPR 请求
	GetByID(ctx context.Context, id uint) (*Gdpr, error)
	// ListByUserID 获取用户的 GDPR 请求列表
	ListByUserID(ctx context.Context, userID uint) ([]*Gdpr, error)
	// List 获取 GDPR 请求列表（管理端）
	List(ctx context.Context, offset, limit int) ([]*Gdpr, int64, error)
}

// ======================== 用例 ========================

// GdprUseCase GDPR 业务用例，编排 GDPR 相关的业务流程
type GdprUseCase struct {
	gdprRepo GdprRepository
}

// NewGdprUseCase 创建 GDPR 用例实例
func NewGdprUseCase(gdprRepo GdprRepository) *GdprUseCase {
	return &GdprUseCase{
		gdprRepo: gdprRepo,
	}
}

// CreateGdpr 创建 GDPR 请求
func (uc *GdprUseCase) CreateGdpr(ctx context.Context, gdpr *Gdpr) (*Gdpr, error) {
	// 新创建的请求默认为 pending 状态
	gdpr.Status = "pending"
	created, err := uc.gdprRepo.Create(ctx, gdpr)
	if err != nil {
		return nil, err
	}
	return created, nil
}

// UpdateGdpr 更新 GDPR 请求
func (uc *GdprUseCase) UpdateGdpr(ctx context.Context, gdpr *Gdpr) (*Gdpr, error) {
	_, err := uc.gdprRepo.GetByID(ctx, gdpr.ID)
	if err != nil {
		return nil, err
	}
	updated, err := uc.gdprRepo.Update(ctx, gdpr)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteGdpr 删除 GDPR 请求
func (uc *GdprUseCase) DeleteGdpr(ctx context.Context, id uint) error {
	_, err := uc.gdprRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return uc.gdprRepo.Delete(ctx, id)
}

// GetGdpr 根据 ID 获取 GDPR 请求
func (uc *GdprUseCase) GetGdpr(ctx context.Context, id uint) (*Gdpr, error) {
	return uc.gdprRepo.GetByID(ctx, id)
}

// ListGdprsByUserID 获取用户的 GDPR 请求列表
func (uc *GdprUseCase) ListGdprsByUserID(ctx context.Context, userID uint) ([]*Gdpr, error) {
	return uc.gdprRepo.ListByUserID(ctx, userID)
}

// ListGdprs 获取 GDPR 请求列表（管理端分页查询）
func (uc *GdprUseCase) ListGdprs(ctx context.Context, offset, limit int) ([]*Gdpr, int64, error) {
	return uc.gdprRepo.List(ctx, offset, limit)
}

// ReviewGdpr 审核 GDPR 请求（批准或拒绝）
func (uc *GdprUseCase) ReviewGdpr(ctx context.Context, id uint, status string, reviewedBy uint) (*Gdpr, error) {
	gdpr, err := uc.gdprRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 更新审核信息
	gdpr.Status = status
	gdpr.ReviewedBy = &reviewedBy
	now := time.Now()
	gdpr.ReviewedAt = &now

	updated, err := uc.gdprRepo.Update(ctx, gdpr)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// CompleteGdpr 完成 GDPR 请求
func (uc *GdprUseCase) CompleteGdpr(ctx context.Context, id uint) (*Gdpr, error) {
	gdpr, err := uc.gdprRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 更新完成状态
	gdpr.Status = "completed"
	now := time.Now()
	gdpr.CompletedAt = &now

	updated, err := uc.gdprRepo.Update(ctx, gdpr)
	if err != nil {
		return nil, err
	}
	return updated, nil
}
