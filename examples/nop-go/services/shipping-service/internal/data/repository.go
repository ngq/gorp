// Package data 物流服务数据访问层
package data

import (
	"context"

	"nop-go/services/shipping-service/internal/models"

	"gorm.io/gorm"
)

type ShipmentRepository interface {
	Create(ctx context.Context, shipment *models.Shipment) error
	GetByID(ctx context.Context, id uint64) (*models.Shipment, error)
	GetByOrderID(ctx context.Context, orderID uint64) (*models.Shipment, error)
	GetByTrackingNumber(ctx context.Context, trackingNumber string) (*models.Shipment, error)
	Update(ctx context.Context, shipment *models.Shipment) error
	List(ctx context.Context, page, pageSize int) ([]*models.Shipment, int64, error)
}

type ShipmentItemRepository interface {
	Create(ctx context.Context, item *models.ShipmentItem) error
	GetByShipmentID(ctx context.Context, shipmentID uint64) ([]*models.ShipmentItem, error)
}

type ShippingMethodRepository interface {
	Create(ctx context.Context, method *models.ShippingMethod) error
	GetByID(ctx context.Context, id uint64) (*models.ShippingMethod, error)
	GetBySystemName(ctx context.Context, systemName string) (*models.ShippingMethod, error)
	Update(ctx context.Context, method *models.ShippingMethod) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.ShippingMethod, error)
}

type ShipmentTrackingRepository interface {
	Create(ctx context.Context, tracking *models.ShipmentTracking) error
	GetByShipmentID(ctx context.Context, shipmentID uint64) ([]*models.ShipmentTracking, error)
}

type shipmentRepo struct{ db *gorm.DB }

func NewShipmentRepository(db *gorm.DB) ShipmentRepository {
	return &shipmentRepo{db: db}
}

func (r *shipmentRepo) Create(ctx context.Context, s *models.Shipment) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *shipmentRepo) GetByID(ctx context.Context, id uint64) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.WithContext(ctx).Preload("Items").First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shipmentRepo) GetByOrderID(ctx context.Context, orderID uint64) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.WithContext(ctx).Preload("Items").Where("order_id = ?", orderID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shipmentRepo) GetByTrackingNumber(ctx context.Context, trackingNumber string) (*models.Shipment, error) {
	var s models.Shipment
	err := r.db.WithContext(ctx).Where("tracking_number = ?", trackingNumber).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *shipmentRepo) Update(ctx context.Context, s *models.Shipment) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *shipmentRepo) List(ctx context.Context, page, pageSize int) ([]*models.Shipment, int64, error) {
	var list []*models.Shipment
	var total int64
	db := r.db.WithContext(ctx).Model(&models.Shipment{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type shipmentItemRepo struct{ db *gorm.DB }

func NewShipmentItemRepository(db *gorm.DB) ShipmentItemRepository {
	return &shipmentItemRepo{db: db}
}

func (r *shipmentItemRepo) Create(ctx context.Context, i *models.ShipmentItem) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *shipmentItemRepo) GetByShipmentID(ctx context.Context, shipmentID uint64) ([]*models.ShipmentItem, error) {
	var items []*models.ShipmentItem
	err := r.db.WithContext(ctx).Where("shipment_id = ?", shipmentID).Find(&items).Error
	return items, err
}

type shippingMethodRepo struct{ db *gorm.DB }

func NewShippingMethodRepository(db *gorm.DB) ShippingMethodRepository {
	return &shippingMethodRepo{db: db}
}

func (r *shippingMethodRepo) Create(ctx context.Context, m *models.ShippingMethod) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *shippingMethodRepo) GetByID(ctx context.Context, id uint64) (*models.ShippingMethod, error) {
	var m models.ShippingMethod
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *shippingMethodRepo) GetBySystemName(ctx context.Context, systemName string) (*models.ShippingMethod, error) {
	var m models.ShippingMethod
	err := r.db.WithContext(ctx).Where("system_name = ?", systemName).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *shippingMethodRepo) Update(ctx context.Context, m *models.ShippingMethod) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *shippingMethodRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.ShippingMethod{}, id).Error
}

func (r *shippingMethodRepo) List(ctx context.Context) ([]*models.ShippingMethod, error) {
	var list []*models.ShippingMethod
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("display_order").Find(&list).Error
	return list, err
}

type shipmentTrackingRepo struct{ db *gorm.DB }

func NewShipmentTrackingRepository(db *gorm.DB) ShipmentTrackingRepository {
	return &shipmentTrackingRepo{db: db}
}

func (r *shipmentTrackingRepo) Create(ctx context.Context, t *models.ShipmentTracking) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *shipmentTrackingRepo) GetByShipmentID(ctx context.Context, shipmentID uint64) ([]*models.ShipmentTracking, error) {
	var list []*models.ShipmentTracking
	err := r.db.WithContext(ctx).Where("shipment_id = ?", shipmentID).Order("occurred_at DESC").Find(&list).Error
	return list, err
}