// Package data 订单服务数据访问层
package data

import (
	"context"
	"time"

	"nop-go/services/order-service/internal/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetByID(ctx context.Context, id uint64) (*models.Order, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error)
	GetByCustomerID(ctx context.Context, customerID uint64, page, pageSize int) ([]*models.Order, int64, error)
	Update(ctx context.Context, order *models.Order) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
	List(ctx context.Context, page, pageSize int) ([]*models.Order, int64, error)
}

type OrderItemRepository interface {
	Create(ctx context.Context, item *models.OrderItem) error
	GetByOrderID(ctx context.Context, orderID uint64) ([]*models.OrderItem, error)
}

type OrderAddressRepository interface {
	Create(ctx context.Context, addr *models.OrderAddress) error
	GetByOrderID(ctx context.Context, orderID uint64) ([]*models.OrderAddress, error)
}

type GiftCardRepository interface {
	Create(ctx context.Context, card *models.GiftCard) error
	GetByCode(ctx context.Context, code string) (*models.GiftCard, error)
	Redeem(ctx context.Context, code string, customerID uint64) error
}

type ReturnRequestRepository interface {
	Create(ctx context.Context, req *models.ReturnRequest) error
	GetByID(ctx context.Context, id uint64) (*models.ReturnRequest, error)
	GetByOrderID(ctx context.Context, orderID uint64) ([]*models.ReturnRequest, error)
	Update(ctx context.Context, req *models.ReturnRequest) error
}

type orderRepo struct{ db *gorm.DB }

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) Create(ctx context.Context, o *models.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *orderRepo) GetByID(ctx context.Context, id uint64) (*models.Order, error) {
	var o models.Order
	err := r.db.WithContext(ctx).Preload("Items").First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	var o models.Order
	err := r.db.WithContext(ctx).Preload("Items").Where("order_number = ?", orderNumber).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepo) GetByCustomerID(ctx context.Context, customerID uint64, page, pageSize int) ([]*models.Order, int64, error) {
	var list []*models.Order
	var total int64
	db := r.db.WithContext(ctx).Model(&models.Order{}).Where("customer_id = ?", customerID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *orderRepo) Update(ctx context.Context, o *models.Order) error {
	return r.db.WithContext(ctx).Save(o).Error
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&models.Order{}).
		Where("id = ?", id).Update("order_status", status).Error
}

func (r *orderRepo) List(ctx context.Context, page, pageSize int) ([]*models.Order, int64, error) {
	var list []*models.Order
	var total int64
	db := r.db.WithContext(ctx).Model(&models.Order{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type orderItemRepo struct{ db *gorm.DB }

func NewOrderItemRepository(db *gorm.DB) OrderItemRepository {
	return &orderItemRepo{db: db}
}

func (r *orderItemRepo) Create(ctx context.Context, i *models.OrderItem) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *orderItemRepo) GetByOrderID(ctx context.Context, orderID uint64) ([]*models.OrderItem, error) {
	var items []*models.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	return items, err
}

type orderAddressRepo struct{ db *gorm.DB }

func NewOrderAddressRepository(db *gorm.DB) OrderAddressRepository {
	return &orderAddressRepo{db: db}
}

func (r *orderAddressRepo) Create(ctx context.Context, a *models.OrderAddress) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *orderAddressRepo) GetByOrderID(ctx context.Context, orderID uint64) ([]*models.OrderAddress, error) {
	var addrs []*models.OrderAddress
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&addrs).Error
	return addrs, err
}

type giftCardRepo struct{ db *gorm.DB }

func NewGiftCardRepository(db *gorm.DB) GiftCardRepository {
	return &giftCardRepo{db: db}
}

func (r *giftCardRepo) Create(ctx context.Context, c *models.GiftCard) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *giftCardRepo) GetByCode(ctx context.Context, code string) (*models.GiftCard, error) {
	var c models.GiftCard
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *giftCardRepo) Redeem(ctx context.Context, code string, customerID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.GiftCard{}).
		Where("code = ? AND is_active = ? AND is_redeemed = ?", code, true, false).
		Updates(map[string]interface{}{
			"is_redeemed":  true,
			"customer_id":  customerID,
			"redeemed_at":  now,
		}).Error
}

type returnRequestRepo struct{ db *gorm.DB }

func NewReturnRequestRepository(db *gorm.DB) ReturnRequestRepository {
	return &returnRequestRepo{db: db}
}

func (r *returnRequestRepo) Create(ctx context.Context, req *models.ReturnRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *returnRequestRepo) GetByID(ctx context.Context, id uint64) (*models.ReturnRequest, error) {
	var req models.ReturnRequest
	err := r.db.WithContext(ctx).First(&req, id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *returnRequestRepo) GetByOrderID(ctx context.Context, orderID uint64) ([]*models.ReturnRequest, error) {
	var list []*models.ReturnRequest
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&list).Error
	return list, err
}

func (r *returnRequestRepo) Update(ctx context.Context, req *models.ReturnRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}