package service

import (
	"context"

	"nop-go/services/store/internal/biz"
	"nop-go/services/store/internal/data"

	"gorm.io/gorm"
)

// Services 店铺服务集合。
type Services struct {
	Store *StoreService
}

// NewServices 创建店铺服务集合。
func NewServices(db *gorm.DB) *Services {
	storeRepo := data.NewStoreRepo(db)
	storeUC := biz.NewStoreUseCase(storeRepo)
	return &Services{
		Store: &StoreService{uc: storeUC},
	}
}

// StoreService 店铺服务。
type StoreService struct {
	uc *biz.StoreUseCase
}

// CreateStoreRequest 创建店铺请求。
type CreateStoreRequest struct {
	Name         string `json:"name" binding:"required"`           // 店铺名称
	Url          string `json:"url" binding:"required"`            // 店铺URL
	SslEnabled   bool   `json:"ssl_enabled"`                       // 是否启用SSL
	Hosts        string `json:"hosts"`                             // 绑定主机列表
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// UpdateStoreRequest 更新店铺请求。
type UpdateStoreRequest struct {
	Name         string `json:"name" binding:"required"`           // 店铺名称
	Url          string `json:"url" binding:"required"`            // 店铺URL
	SslEnabled   bool   `json:"ssl_enabled"`                       // 是否启用SSL
	Hosts        string `json:"hosts"`                             // 绑定主机列表
	DisplayOrder int    `json:"display_order"`                     // 显示排序
}

// StoreResponse 店铺响应。
type StoreResponse struct {
	ID           uint   `json:"id"`            // 店铺ID
	Name         string `json:"name"`          // 店铺名称
	Url          string `json:"url"`           // 店铺URL
	SslEnabled   bool   `json:"ssl_enabled"`   // 是否启用SSL
	Hosts        string `json:"hosts"`         // 绑定主机列表
	DisplayOrder int    `json:"display_order"` // 显示排序
	CreatedAt    string `json:"created_at"`    // 创建时间
	UpdatedAt    string `json:"updated_at"`    // 更新时间
}

// toResponse 将领域实体转换为响应结构体。
func toStoreResponse(store *biz.Store) *StoreResponse {
	return &StoreResponse{
		ID:           store.ID,
		Name:         store.Name,
		Url:          store.Url,
		SslEnabled:   store.SslEnabled,
		Hosts:        store.Hosts,
		DisplayOrder: store.DisplayOrder,
		CreatedAt:    store.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    store.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// List 获取店铺列表。
func (s *StoreService) List(ctx context.Context, page, size int) ([]StoreResponse, int64, error) {
	stores, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]StoreResponse, len(stores))
	for i, store := range stores {
		items[i] = *toStoreResponse(store)
	}
	return items, total, nil
}

// Create 创建店铺。
func (s *StoreService) Create(ctx context.Context, req CreateStoreRequest) (*StoreResponse, error) {
	store, err := s.uc.Create(ctx, req.Name, req.Url, req.SslEnabled, req.Hosts, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toStoreResponse(store), nil
}

// Update 更新店铺。
func (s *StoreService) Update(ctx context.Context, id uint, req UpdateStoreRequest) (*StoreResponse, error) {
	store, err := s.uc.Update(ctx, id, req.Name, req.Url, req.SslEnabled, req.Hosts, req.DisplayOrder)
	if err != nil {
		return nil, err
	}
	return toStoreResponse(store), nil
}

// Delete 删除店铺。
func (s *StoreService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}