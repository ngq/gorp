// Package biz 包含交易服务的业务逻辑层
// shipping.go 定义物流领域实体、仓库接口与物流用例
// 注意：原 shipping 服务中的 Provider 实体已重命名为 ShippingProvider，避免与 TaxProvider 冲突
package biz

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// 物流领域实体
// ============================================================================

// ShippingProvider 物流服务商实体（原 Provider，重命名以避免与 TaxProvider 冲突）
type ShippingProvider struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Code        string         `json:"code"` // 物流商编码，如 sf、ems、yto
	Description string         `json:"description"`
	IsActive    bool           `json:"isActive"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// ShippingOrder 物流订单实体，关联业务订单与物流服务商
type ShippingOrder struct {
	ID               string            `json:"id"`
	OrderID          string            `json:"orderId"`
	ShippingProviderID string          `json:"shippingProviderId"`
	TrackingNumber   string            `json:"trackingNumber"`
	Status           string            `json:"status"` // pending, picked_up, in_transit, delivered, returned
	ShippingAddress  string            `json:"shippingAddress"`
	EstimatedDelivery time.Time        `json:"estimatedDelivery"`
	ActualDelivery   *time.Time        `json:"actualDelivery"`
	CreatedAt        time.Time         `json:"createdAt"`
	UpdatedAt        time.Time         `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt    `json:"-" gorm:"index"`
}

// ShippingEvent 物流事件实体，记录物流轨迹
type ShippingEvent struct {
	ID              string    `json:"id"`
	ShippingOrderID string    `json:"shippingOrderId"`
	Status          string    `json:"status"`
	Location        string    `json:"location"`
	Description     string    `json:"description"`
	EventTime       time.Time `json:"eventTime"`
	CreatedAt       time.Time `json:"createdAt"`
}

// ShippingRate 运费率实体，按物流商和区域定价
type ShippingRate struct {
	ID                string    `json:"id"`
	ShippingProviderID string   `json:"shippingProviderId"`
	OriginZone        string    `json:"originZone"`
	DestinationZone   string    `json:"destinationZone"`
	WeightMin         float64   `json:"weightMin"`
	WeightMax         float64   `json:"weightMax"`
	Rate              float64   `json:"rate"`
	Currency          string    `json:"currency"`
	EstimatedDays     int       `json:"estimatedDays"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

// ============================================================================
// 仓库接口定义
// ============================================================================

// ShippingProviderRepo 物流服务商数据仓库接口
type ShippingProviderRepo interface {
	// Create 创建物流服务商
	Create(ctx context.Context, provider *ShippingProvider) error
	// GetByID 根据ID获取物流服务商
	GetByID(ctx context.Context, id string) (*ShippingProvider, error)
	// Update 更新物流服务商
	Update(ctx context.Context, provider *ShippingProvider) error
	// Delete 删除物流服务商
	Delete(ctx context.Context, id string) error
	// List 获取物流服务商列表
	List(ctx context.Context) ([]*ShippingProvider, error)
	// GetByCode 根据编码获取物流服务商
	GetByCode(ctx context.Context, code string) (*ShippingProvider, error)
}

// ShippingOrderRepo 物流订单数据仓库接口
type ShippingOrderRepo interface {
	// Create 创建物流订单
	Create(ctx context.Context, order *ShippingOrder) error
	// GetByID 根据ID获取物流订单
	GetByID(ctx context.Context, id string) (*ShippingOrder, error)
	// Update 更新物流订单
	Update(ctx context.Context, order *ShippingOrder) error
	// ListByOrderID 按业务订单ID查询物流订单
	ListByOrderID(ctx context.Context, orderID string) ([]*ShippingOrder, error)
	// ListByTrackingNumber 根据运单号查询物流订单
	ListByTrackingNumber(ctx context.Context, trackingNumber string) (*ShippingOrder, error)
}

// ShippingEventRepo 物流事件数据仓库接口
type ShippingEventRepo interface {
	// Create 创建物流事件
	Create(ctx context.Context, event *ShippingEvent) error
	// ListByShippingOrderID 按物流订单ID查询事件列表
	ListByShippingOrderID(ctx context.Context, shippingOrderID string) ([]*ShippingEvent, error)
}

// ShippingRateRepo 运费率数据仓库接口
type ShippingRateRepo interface {
	// Create 创建运费率
	Create(ctx context.Context, rate *ShippingRate) error
	// GetByID 根据ID获取运费率
	GetByID(ctx context.Context, id string) (*ShippingRate, error)
	// Update 更新运费率
	Update(ctx context.Context, rate *ShippingRate) error
	// Delete 删除运费率
	Delete(ctx context.Context, id string) error
	// ListByProvider 按物流服务商查询运费率列表
	ListByProvider(ctx context.Context, providerID string) ([]*ShippingRate, error)
	// FindRate 查询匹配的运费率
	FindRate(ctx context.Context, providerID, originZone, destZone string, weight float64) (*ShippingRate, error)
}

// ============================================================================
// 用例（UseCase）
// ============================================================================

// ShippingUseCase 物流业务用例，封装物流服务商、物流订单、物流事件、运费率的管理逻辑
// 重构说明：UseCase 方法返回领域实体而非 response DTO，DTO 转换移至 service 层
type ShippingUseCase struct {
	providerRepo ShippingProviderRepo
	orderRepo    ShippingOrderRepo
	eventRepo    ShippingEventRepo
	rateRepo     ShippingRateRepo
}

// NewShippingUseCase 创建物流用例实例
func NewShippingUseCase(
	providerRepo ShippingProviderRepo,
	orderRepo ShippingOrderRepo,
	eventRepo ShippingEventRepo,
	rateRepo ShippingRateRepo,
) *ShippingUseCase {
	return &ShippingUseCase{
		providerRepo: providerRepo,
		orderRepo:    orderRepo,
		eventRepo:    eventRepo,
		rateRepo:     rateRepo,
	}
}

// --- 物流服务商操作 ---

// CreateShippingProvider 创建物流服务商
func (uc *ShippingUseCase) CreateShippingProvider(ctx context.Context, provider *ShippingProvider) error {
	return uc.providerRepo.Create(ctx, provider)
}

// GetShippingProvider 获取物流服务商详情
func (uc *ShippingUseCase) GetShippingProvider(ctx context.Context, id string) (*ShippingProvider, error) {
	provider, err := uc.providerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取物流服务商失败: %w", err)
	}
	return provider, nil
}

// UpdateShippingProvider 更新物流服务商
func (uc *ShippingUseCase) UpdateShippingProvider(ctx context.Context, provider *ShippingProvider) error {
	return uc.providerRepo.Update(ctx, provider)
}

// DeleteShippingProvider 删除物流服务商
func (uc *ShippingUseCase) DeleteShippingProvider(ctx context.Context, id string) error {
	return uc.providerRepo.Delete(ctx, id)
}

// ListShippingProviders 获取物流服务商列表
func (uc *ShippingUseCase) ListShippingProviders(ctx context.Context) ([]*ShippingProvider, error) {
	return uc.providerRepo.List(ctx)
}

// --- 物流订单操作 ---

// CreateShippingOrder 创建物流订单
func (uc *ShippingUseCase) CreateShippingOrder(ctx context.Context, order *ShippingOrder) error {
	return uc.orderRepo.Create(ctx, order)
}

// GetShippingOrder 获取物流订单详情
func (uc *ShippingUseCase) GetShippingOrder(ctx context.Context, id string) (*ShippingOrder, error) {
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取物流订单失败: %w", err)
	}
	return order, nil
}

// UpdateShippingOrder 更新物流订单
func (uc *ShippingUseCase) UpdateShippingOrder(ctx context.Context, order *ShippingOrder) error {
	return uc.orderRepo.Update(ctx, order)
}

// ListShippingOrdersByOrder 按业务订单ID查询物流订单
func (uc *ShippingUseCase) ListShippingOrdersByOrder(ctx context.Context, orderID string) ([]*ShippingOrder, error) {
	return uc.orderRepo.ListByOrderID(ctx, orderID)
}

// GetShippingOrderByTracking 根据运单号查询物流订单
func (uc *ShippingUseCase) GetShippingOrderByTracking(ctx context.Context, trackingNumber string) (*ShippingOrder, error) {
	return uc.orderRepo.ListByTrackingNumber(ctx, trackingNumber)
}

// --- 物流事件操作 ---

// CreateShippingEvent 创建物流事件
func (uc *ShippingUseCase) CreateShippingEvent(ctx context.Context, event *ShippingEvent) error {
	return uc.eventRepo.Create(ctx, event)
}

// ListShippingEvents 按物流订单ID查询事件列表
func (uc *ShippingUseCase) ListShippingEvents(ctx context.Context, shippingOrderID string) ([]*ShippingEvent, error) {
	return uc.eventRepo.ListByShippingOrderID(ctx, shippingOrderID)
}

// --- 运费率操作 ---

// CreateShippingRate 创建运费率
func (uc *ShippingUseCase) CreateShippingRate(ctx context.Context, rate *ShippingRate) error {
	return uc.rateRepo.Create(ctx, rate)
}

// GetShippingRate 获取运费率详情
func (uc *ShippingUseCase) GetShippingRate(ctx context.Context, id string) (*ShippingRate, error) {
	return uc.rateRepo.GetByID(ctx, id)
}

// UpdateShippingRate 更新运费率
func (uc *ShippingUseCase) UpdateShippingRate(ctx context.Context, rate *ShippingRate) error {
	return uc.rateRepo.Update(ctx, rate)
}

// DeleteShippingRate 删除运费率
func (uc *ShippingUseCase) DeleteShippingRate(ctx context.Context, id string) error {
	return uc.rateRepo.Delete(ctx, id)
}

// ListShippingRates 按物流服务商查询运费率列表
func (uc *ShippingUseCase) ListShippingRates(ctx context.Context, providerID string) ([]*ShippingRate, error) {
	return uc.rateRepo.ListByProvider(ctx, providerID)
}

// FindShippingRate 查询匹配的运费率
func (uc *ShippingUseCase) FindShippingRate(ctx context.Context, providerID, originZone, destZone string, weight float64) (*ShippingRate, error) {
	return uc.rateRepo.FindRate(ctx, providerID, originZone, destZone, weight)
}
