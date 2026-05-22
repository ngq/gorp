package service

import (
	"context"

	"nop-go/services/security/internal/biz"
	"nop-go/services/security/internal/data"

	"gorm.io/gorm"
)

// Services 安全服务集合。
type Services struct {
	Permission *PermissionService
	ACL        *ACLService
}

// NewServices 创建安全服务集合。
func NewServices(db *gorm.DB) *Services {
	permRepo := data.NewPermissionRepo(db)
	aclRepo := data.NewACLRecordRepo(db)
	permUC := biz.NewPermissionUseCase(permRepo)
	aclUC := biz.NewACLUseCase(aclRepo, permRepo)
	return &Services{
		Permission: &PermissionService{uc: permUC},
		ACL:        &ACLService{uc: aclUC},
	}
}

// PermissionService 权限服务。
type PermissionService struct {
	uc *biz.PermissionUseCase
}

// CreatePermissionRequest 创建权限请求。
type CreatePermissionRequest struct {
	Name         string `json:"name" binding:"required"`           // 权限名称
	SystemName   string `json:"system_name" binding:"required"`    // 权限系统名称
	Category     string `json:"category" binding:"required"`       // 权限分类
	DisplayOrder int    `json:"display_order"`                    // 显示排序
}

// UpdatePermissionRequest 更新权限请求。
type UpdatePermissionRequest struct {
	Name         string `json:"name" binding:"required"`           // 权限名称
	SystemName   string `json:"system_name" binding:"required"`    // 权限系统名称
	Category     string `json:"category" binding:"required"`       // 权限分类
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// PermissionResponse 权限响应。
type PermissionResponse struct {
	ID           uint   `json:"id"`            // 权限ID
	Name         string `json:"name"`          // 权限名称
	SystemName   string `json:"system_name"`   // 权限系统名称
	Category     string `json:"category"`      // 权限分类
	DisplayOrder int    `json:"display_order"` // 显示排序
	CreatedAt    string `json:"created_at"`    // 创建时间
	UpdatedAt    string `json:"updated_at"`    // 更新时间
}

// ACLService ACL服务。
type ACLService struct {
	uc *biz.ACLUseCase
}

// CreateACLRequest 创建ACL记录请求。
type CreateACLRequest struct {
	UserID       uint `json:"user_id" binding:"required"`        // 用户ID
	PermissionID uint `json:"permission_id" binding:"required"`  // 权限ID
}

// ACLResponse ACL记录响应。
type ACLResponse struct {
	ID             uint   `json:"id"`              // ACL记录ID
	UserID         uint   `json:"user_id"`         // 用户ID
	PermissionID   uint   `json:"permission_id"`   // 权限ID
	PermissionName string `json:"permission_name"` // 权限名称
}

// toPermissionResponse 将权限领域实体转换为响应结构体。
func toPermissionResponse(perm *biz.Permission) *PermissionResponse {
	return &PermissionResponse{
		ID:           perm.ID,
		Name:         perm.Name,
		SystemName:   perm.SystemName,
		Category:     perm.Category,
		DisplayOrder: perm.DisplayOrder,
		CreatedAt:    perm.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    perm.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toACLResponse 将ACL记录领域实体转换为响应结构体。
func toACLResponse(record *biz.ACLRecord) *ACLResponse {
	return &ACLResponse{
		ID:             record.ID,
		UserID:         record.UserID,
		PermissionID:   record.PermissionID,
		PermissionName: record.PermissionName,
	}
}

// List 获取权限列表。
func (s *PermissionService) List(ctx context.Context, page, size int) ([]PermissionResponse, int64, error) {
	perms, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]PermissionResponse, len(perms))
	for i, perm := range perms {
		items[i] = *toPermissionResponse(perm)
	}
	return items, total, nil
}

// Create 创建权限。
func (s *PermissionService) Create(ctx context.Context, req CreatePermissionRequest) (*PermissionResponse, error) {
	perm, err := s.uc.Create(ctx, req.Name, req.SystemName, req.Category, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toPermissionResponse(perm), nil
}

// Update 更新权限。
func (s *PermissionService) Update(ctx context.Context, id uint, req UpdatePermissionRequest) (*PermissionResponse, error) {
	perm, err := s.uc.Update(ctx, id, req.Name, req.SystemName, req.Category, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toPermissionResponse(perm), nil
}

// Delete 删除权限。
func (s *PermissionService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// List 获取ACL记录列表。
func (s *ACLService) List(ctx context.Context, page, size int) ([]ACLResponse, int64, error) {
	records, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]ACLResponse, len(records))
	for i, record := range records {
		items[i] = *toACLResponse(record)
	}
	return items, total, nil
}

// Create 创建ACL记录。
func (s *ACLService) Create(ctx context.Context, req CreateACLRequest) (*ACLResponse, error) {
	record, err := s.uc.Create(ctx, req.UserID, req.PermissionID)
	if err != nil {
		return nil, err
	}
	return toACLResponse(record), nil
}

// Delete 删除ACL记录。
func (s *ACLService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}