package service

import (
	"context"

	"nop-go/services/vendorsvc/internal/biz"
	"nop-go/services/vendorsvc/internal/data"

	"gorm.io/gorm"
)

// Services 供应商服务集合。
type Services struct {
	Vendor *VendorService
}

// NewServices 创建供应商服务集合。
func NewServices(db *gorm.DB) *Services {
	vendorRepo := data.NewVendorRepo(db)
	applyRepo := data.NewVendorApplyRepo(db)
	vendorUC := biz.NewVendorUseCase(vendorRepo, applyRepo)
	return &Services{
		Vendor: &VendorService{uc: vendorUC},
	}
}

// VendorService 供应商服务。
type VendorService struct {
	uc *biz.VendorUseCase
}

// CreateVendorRequest 创建供应商请求。
type CreateVendorRequest struct {
	Name         string `json:"name" binding:"required"`           // 供应商名称
	Email        string `json:"email" binding:"required,email"`    // 供应商邮箱
	Description  string `json:"description"`                       // 供应商描述
	Active       bool   `json:"active"`                            // 是否启用
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// UpdateVendorRequest 更新供应商请求。
type UpdateVendorRequest struct {
	Name         string `json:"name" binding:"required"`           // 供应商名称
	Email        string `json:"email" binding:"required,email"`    // 供应商邮箱
	Description  string `json:"description"`                       // 供应商描述
	Active       bool   `json:"active"`                            // 是否启用
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// VendorApplyRequest 供应商申请请求。
type VendorApplyRequest struct {
	Name        string `json:"name" binding:"required"`           // 供应商名称
	Email       string `json:"email" binding:"required,email"`    // 供应商邮箱
	Description string `json:"description"`                       // 申请描述
}

// VendorResponse 供应商响应。
type VendorResponse struct {
	ID           uint   `json:"id"`            // 供应商ID
	Name         string `json:"name"`          // 供应商名称
	Email        string `json:"email"`         // 供应商邮箱
	Active       bool   `json:"active"`        // 是否启用
	DisplayOrder int    `json:"display_order"` // 显示排序
	CreatedAt    string `json:"created_at"`    // 创建时间
	UpdatedAt    string `json:"updated_at"`    // 更新时间
}

// VendorApplyResponse 供应商申请响应。
type VendorApplyResponse struct {
	ID          uint   `json:"id"`          // 申请ID
	Name        string `json:"name"`        // 供应商名称
	Email       string `json:"email"`       // 供应商邮箱
	Description string `json:"description"` // 申请描述
	Status      string `json:"status"`      // 申请状态
	CreatedAt   string `json:"created_at"`  // 创建时间
}

// toVendorResponse 将供应商领域实体转换为响应结构体。
func toVendorResponse(v *biz.Vendor) *VendorResponse {
	return &VendorResponse{
		ID:           v.ID,
		Name:         v.Name,
		Email:        v.Email,
		Active:       v.Active,
		DisplayOrder: v.DisplayOrder,
		CreatedAt:    v.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    v.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// toVendorApplyResponse 将供应商申请领域实体转换为响应结构体。
func toVendorApplyResponse(a *biz.VendorApply) *VendorApplyResponse {
	return &VendorApplyResponse{
		ID:          a.ID,
		Name:        a.Name,
		Email:       a.Email,
		Description: a.Description,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// List 获取供应商列表。
func (s *VendorService) List(ctx context.Context, page, size int) ([]VendorResponse, int64, error) {
	vendors, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]VendorResponse, len(vendors))
	for i, v := range vendors {
		items[i] = *toVendorResponse(v)
	}
	return items, total, nil
}

// Create 创建供应商。
func (s *VendorService) Create(ctx context.Context, req CreateVendorRequest) (*VendorResponse, error) {
	vendor, err := s.uc.Create(ctx, req.Name, req.Email, req.Description, req.Active, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toVendorResponse(vendor), nil
}

// Update 更新供应商。
func (s *VendorService) Update(ctx context.Context, id uint, req UpdateVendorRequest) (*VendorResponse, error) {
	vendor, err := s.uc.Update(ctx, id, req.Name, req.Email, req.Description, req.Active, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toVendorResponse(vendor), nil
}

// Delete 删除供应商。
func (s *VendorService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// GetApply 查询供应商申请状态。
func (s *VendorService) GetApply(ctx context.Context, id uint) (*VendorApplyResponse, error) {
	apply, err := s.uc.GetApply(ctx, id)
	if err != nil {
		return nil, err
	}
	return toVendorApplyResponse(apply), nil
}

// SubmitApply 提交供应商申请。
func (s *VendorService) SubmitApply(ctx context.Context, req VendorApplyRequest) (*VendorApplyResponse, error) {
	apply, err := s.uc.SubmitApply(ctx, req.Name, req.Email, req.Description)
	if err != nil {
		return nil, err
	}
	return toVendorApplyResponse(apply), nil
}