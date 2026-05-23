// Package data 包含交易服务的数据访问层
// shipping.go 定义物流相关 PO 及仓库实现
// 注意：原 shipping 服务中的 Provider PO 已重命名为 ShippingProviderPO
package data

import (
	"context"
	"time"

	"nop-go/services/trade-service/internal/biz"

	"gorm.io/gorm"
)

// ============================================================================
// PO 定义
// ============================================================================

// ShippingProviderPO 物流服务商持久化对象（原 ProviderPO，重命名以避免冲突）
type ShippingProviderPO struct {
	gorm.Model
	Name        string `gorm:"size:128;not null;column:name" db:"name"`
	Code        string `gorm:"size:64;uniqueIndex;not null;column:code" db:"code"`
	Description string `gorm:"size:512;column:description" db:"description"`
	IsActive    bool   `gorm:"default:true;column:is_active" db:"is_active"`
}

// TableName 指定物流服务商表名
func (ShippingProviderPO) TableName() string { return "shipping_providers" }

// ToEntity 转换为物流服务商领域实体
func (po *ShippingProviderPO) ToEntity() *biz.ShippingProvider {
	return &biz.ShippingProvider{
		ID:          fmtID(po.ID),
		Name:        po.Name,
		Code:        po.Code,
		Description: po.Description,
		IsActive:    po.IsActive,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

// ShippingOrderPO 物流订单持久化对象
type ShippingOrderPO struct {
	gorm.Model
	OrderID            string     `gorm:"index;not null;size:64;column:order_id" db:"order_id"`
	ShippingProviderID string     `gorm:"index;not null;size:64;column:shipping_provider_id" db:"shipping_provider_id"`
	TrackingNumber     string     `gorm:"size:128;column:tracking_number" db:"tracking_number"`
	Status             string     `gorm:"size:32;default:'pending';column:status" db:"status"`
	ShippingAddress    string     `gorm:"size:512;column:shipping_address" db:"shipping_address"`
	EstimatedDelivery  time.Time  `gorm:"column:estimated_delivery" db:"estimated_delivery"`
	ActualDelivery     *time.Time `gorm:"column:actual_delivery" db:"actual_delivery"`
}

// TableName 指定物流订单表名
func (ShippingOrderPO) TableName() string { return "shipping_orders" }

// ToEntity 转换为物流订单领域实体
func (po *ShippingOrderPO) ToEntity() *biz.ShippingOrder {
	return &biz.ShippingOrder{
		ID:                 fmtID(po.ID),
		OrderID:            po.OrderID,
		ShippingProviderID: po.ShippingProviderID,
		TrackingNumber:     po.TrackingNumber,
		Status:             po.Status,
		ShippingAddress:    po.ShippingAddress,
		EstimatedDelivery:  po.EstimatedDelivery,
		ActualDelivery:     po.ActualDelivery,
		CreatedAt:          po.CreatedAt,
		UpdatedAt:          po.UpdatedAt,
	}
}

// ShippingEventPO 物流事件持久化对象
type ShippingEventPO struct {
	gorm.Model
	ShippingOrderID string    `gorm:"index;not null;size:64;column:shipping_order_id" db:"shipping_order_id"`
	Status          string    `gorm:"size:32;not null;column:status" db:"status"`
	Location        string    `gorm:"size:256;column:location" db:"location"`
	Description     string    `gorm:"size:512;column:description" db:"description"`
	EventTime       time.Time `gorm:"not null;column:event_time" db:"event_time"`
}

// TableName 指定物流事件表名
func (ShippingEventPO) TableName() string { return "shipping_events" }

// ToEntity 转换为物流事件领域实体
func (po *ShippingEventPO) ToEntity() *biz.ShippingEvent {
	return &biz.ShippingEvent{
		ID:              fmtID(po.ID),
		ShippingOrderID: po.ShippingOrderID,
		Status:          po.Status,
		Location:        po.Location,
		Description:     po.Description,
		EventTime:       po.EventTime,
		CreatedAt:       po.CreatedAt,
	}
}

// ShippingRatePO 运费率持久化对象
type ShippingRatePO struct {
	gorm.Model
	ShippingProviderID string  `gorm:"index;not null;size:64;column:shipping_provider_id" db:"shipping_provider_id"`
	OriginZone         string  `gorm:"size:64;not null;column:origin_zone" db:"origin_zone"`
	DestinationZone    string  `gorm:"size:64;not null;column:destination_zone" db:"destination_zone"`
	WeightMin          float64 `gorm:"type:decimal(10,2);not null;column:weight_min" db:"weight_min"`
	WeightMax          float64 `gorm:"type:decimal(10,2);not null;column:weight_max" db:"weight_max"`
	Rate               float64 `gorm:"type:decimal(12,2);not null;column:rate" db:"rate"`
	Currency           string  `gorm:"size:8;default:'CNY';column:currency" db:"currency"`
	EstimatedDays      int     `gorm:"column:estimated_days" db:"estimated_days"`
}

// TableName 指定运费率表名
func (ShippingRatePO) TableName() string { return "shipping_rates" }

// ToEntity 转换为运费率领域实体
func (po *ShippingRatePO) ToEntity() *biz.ShippingRate {
	return &biz.ShippingRate{
		ID:                 fmtID(po.ID),
		ShippingProviderID: po.ShippingProviderID,
		OriginZone:         po.OriginZone,
		DestinationZone:    po.DestinationZone,
		WeightMin:          po.WeightMin,
		WeightMax:          po.WeightMax,
		Rate:               po.Rate,
		Currency:           po.Currency,
		EstimatedDays:      po.EstimatedDays,
		CreatedAt:          po.CreatedAt,
		UpdatedAt:          po.UpdatedAt,
	}
}

// ============================================================================
// 仓库实现
// ============================================================================

// shippingProviderRepo 物流服务商仓储实现
type shippingProviderRepo struct{ db *gorm.DB }

// NewShippingProviderRepo 创建物流服务商仓储
func NewShippingProviderRepo(db *gorm.DB) biz.ShippingProviderRepo {
	return &shippingProviderRepo{db: db}
}

func (r *shippingProviderRepo) Create(ctx context.Context, provider *biz.ShippingProvider) error {
	po := &ShippingProviderPO{
		Name:        provider.Name,
		Code:        provider.Code,
		Description: provider.Description,
		IsActive:    provider.IsActive,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	provider.ID = fmtID(po.ID)
	return nil
}

func (r *shippingProviderRepo) GetByID(ctx context.Context, id string) (*biz.ShippingProvider, error) {
	var po ShippingProviderPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *shippingProviderRepo) Update(ctx context.Context, provider *biz.ShippingProvider) error {
	return r.db.WithContext(ctx).Model(&ShippingProviderPO{}).Where("id = ?", parseID(provider.ID)).Updates(map[string]interface{}{
		"name":        provider.Name,
		"description": provider.Description,
		"is_active":   provider.IsActive,
	}).Error
}

func (r *shippingProviderRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&ShippingProviderPO{}, parseID(id)).Error
}

func (r *shippingProviderRepo) List(ctx context.Context) ([]*biz.ShippingProvider, error) {
	var pos []ShippingProviderPO
	if err := r.db.WithContext(ctx).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.ShippingProvider, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *shippingProviderRepo) GetByCode(ctx context.Context, code string) (*biz.ShippingProvider, error) {
	var po ShippingProviderPO
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

// shippingOrderRepo 物流订单仓储实现
type shippingOrderRepo struct{ db *gorm.DB }

// NewShippingOrderRepo 创建物流订单仓储
func NewShippingOrderRepo(db *gorm.DB) biz.ShippingOrderRepo { return &shippingOrderRepo{db: db} }

func (r *shippingOrderRepo) Create(ctx context.Context, order *biz.ShippingOrder) error {
	po := &ShippingOrderPO{
		OrderID:            order.OrderID,
		ShippingProviderID: order.ShippingProviderID,
		TrackingNumber:     order.TrackingNumber,
		Status:             order.Status,
		ShippingAddress:    order.ShippingAddress,
		EstimatedDelivery:  order.EstimatedDelivery,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	order.ID = fmtID(po.ID)
	return nil
}

func (r *shippingOrderRepo) GetByID(ctx context.Context, id string) (*biz.ShippingOrder, error) {
	var po ShippingOrderPO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *shippingOrderRepo) Update(ctx context.Context, order *biz.ShippingOrder) error {
	return r.db.WithContext(ctx).Model(&ShippingOrderPO{}).Where("id = ?", parseID(order.ID)).Updates(map[string]interface{}{
		"status":          order.Status,
		"tracking_number": order.TrackingNumber,
		"actual_delivery": order.ActualDelivery,
	}).Error
}

func (r *shippingOrderRepo) ListByOrderID(ctx context.Context, orderID string) ([]*biz.ShippingOrder, error) {
	var pos []ShippingOrderPO
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.ShippingOrder, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *shippingOrderRepo) ListByTrackingNumber(ctx context.Context, trackingNumber string) (*biz.ShippingOrder, error) {
	var po ShippingOrderPO
	if err := r.db.WithContext(ctx).Where("tracking_number = ?", trackingNumber).First(&po).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

// shippingEventRepo 物流事件仓储实现
type shippingEventRepo struct{ db *gorm.DB }

// NewShippingEventRepo 创建物流事件仓储
func NewShippingEventRepo(db *gorm.DB) biz.ShippingEventRepo { return &shippingEventRepo{db: db} }

func (r *shippingEventRepo) Create(ctx context.Context, event *biz.ShippingEvent) error {
	po := &ShippingEventPO{
		ShippingOrderID: event.ShippingOrderID,
		Status:          event.Status,
		Location:        event.Location,
		Description:     event.Description,
		EventTime:       event.EventTime,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	event.ID = fmtID(po.ID)
	return nil
}

func (r *shippingEventRepo) ListByShippingOrderID(ctx context.Context, shippingOrderID string) ([]*biz.ShippingEvent, error) {
	var pos []ShippingEventPO
	if err := r.db.WithContext(ctx).Where("shipping_order_id = ?", shippingOrderID).Order("event_time ASC").Find(&pos).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.ShippingEvent, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

// shippingRateRepo 运费率仓储实现
type shippingRateRepo struct{ db *gorm.DB }

// NewShippingRateRepo 创建运费率仓储
func NewShippingRateRepo(db *gorm.DB) biz.ShippingRateRepo { return &shippingRateRepo{db: db} }

func (r *shippingRateRepo) Create(ctx context.Context, rate *biz.ShippingRate) error {
	po := &ShippingRatePO{
		ShippingProviderID: rate.ShippingProviderID,
		OriginZone:         rate.OriginZone,
		DestinationZone:    rate.DestinationZone,
		WeightMin:          rate.WeightMin,
		WeightMax:          rate.WeightMax,
		Rate:               rate.Rate,
		Currency:           rate.Currency,
		EstimatedDays:      rate.EstimatedDays,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil { return err }
	rate.ID = fmtID(po.ID)
	return nil
}

func (r *shippingRateRepo) GetByID(ctx context.Context, id string) (*biz.ShippingRate, error) {
	var po ShippingRatePO
	if err := r.db.WithContext(ctx).First(&po, parseID(id)).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}

func (r *shippingRateRepo) Update(ctx context.Context, rate *biz.ShippingRate) error {
	return r.db.WithContext(ctx).Model(&ShippingRatePO{}).Where("id = ?", parseID(rate.ID)).Updates(map[string]interface{}{
		"rate":           rate.Rate,
		"estimated_days": rate.EstimatedDays,
	}).Error
}

func (r *shippingRateRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&ShippingRatePO{}, parseID(id)).Error
}

func (r *shippingRateRepo) ListByProvider(ctx context.Context, providerID string) ([]*biz.ShippingRate, error) {
	var pos []ShippingRatePO
	if err := r.db.WithContext(ctx).Where("shipping_provider_id = ?", providerID).Find(&pos).Error; err != nil { return nil, err }
	items := make([]*biz.ShippingRate, len(pos))
	for i, po := range pos { items[i] = po.ToEntity() }
	return items, nil
}

func (r *shippingRateRepo) FindRate(ctx context.Context, providerID, originZone, destZone string, weight float64) (*biz.ShippingRate, error) {
	var po ShippingRatePO
	if err := r.db.WithContext(ctx).Where(
		"shipping_provider_id = ? AND origin_zone = ? AND destination_zone = ? AND weight_min <= ? AND weight_max >= ?",
		providerID, originZone, destZone, weight, weight,
	).First(&po).Error; err != nil { return nil, err }
	return po.ToEntity(), nil
}
