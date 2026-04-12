// Package data 库存服务数据访问层
package data

import (
	"context"
	"errors"
	"time"

	"nop-go/services/inventory-service/internal/models"

	"gorm.io/gorm"
)

type InventoryRepository interface {
	Create(ctx context.Context, inventory *models.Inventory) error
	GetByID(ctx context.Context, id uint64) (*models.Inventory, error)
	GetByProductAndWarehouse(ctx context.Context, productID, warehouseID uint64) (*models.Inventory, error)
	GetByProductID(ctx context.Context, productID uint64) ([]*models.Inventory, error)
	Update(ctx context.Context, inventory *models.Inventory) error
	List(ctx context.Context, page, pageSize int) ([]*models.Inventory, int64, error)
	UpdateQuantity(ctx context.Context, productID, warehouseID uint64, quantity int) error
	ReserveStock(ctx context.Context, productID, warehouseID uint64, quantity int) error
	ReleaseStock(ctx context.Context, productID, warehouseID uint64, quantity int) error
}

type WarehouseRepository interface {
	Create(ctx context.Context, warehouse *models.Warehouse) error
	GetByID(ctx context.Context, id uint64) (*models.Warehouse, error)
	GetByCode(ctx context.Context, code string) (*models.Warehouse, error)
	Update(ctx context.Context, warehouse *models.Warehouse) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.Warehouse, error)
}

type StockReservationRepository interface {
	Create(ctx context.Context, reservation *models.StockReservation) error
	GetByOrderID(ctx context.Context, orderID uint64) (*models.StockReservation, error)
	Confirm(ctx context.Context, orderID uint64) error
	Release(ctx context.Context, orderID uint64) error
}

type InventoryLogRepository interface {
	Create(ctx context.Context, log *models.InventoryLog) error
	GetByProductID(ctx context.Context, productID uint64, limit int) ([]*models.InventoryLog, error)
}

type inventoryRepo struct{ db *gorm.DB }

func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepo{db: db}
}

func (r *inventoryRepo) Create(ctx context.Context, i *models.Inventory) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *inventoryRepo) GetByID(ctx context.Context, id uint64) (*models.Inventory, error) {
	var i models.Inventory
	err := r.db.WithContext(ctx).First(&i, id).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *inventoryRepo) GetByProductAndWarehouse(ctx context.Context, productID, warehouseID uint64) (*models.Inventory, error) {
	var i models.Inventory
	err := r.db.WithContext(ctx).Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).First(&i).Error
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *inventoryRepo) GetByProductID(ctx context.Context, productID uint64) ([]*models.Inventory, error) {
	var list []*models.Inventory
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&list).Error
	return list, err
}

func (r *inventoryRepo) Update(ctx context.Context, i *models.Inventory) error {
	return r.db.WithContext(ctx).Save(i).Error
}

func (r *inventoryRepo) List(ctx context.Context, page, pageSize int) ([]*models.Inventory, int64, error) {
	var list []*models.Inventory
	var total int64
	db := r.db.WithContext(ctx).Model(&models.Inventory{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *inventoryRepo) UpdateQuantity(ctx context.Context, productID, warehouseID uint64, quantity int) error {
	return r.db.WithContext(ctx).Model(&models.Inventory{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Update("quantity", gorm.Expr("quantity + ?", quantity)).Error
}

func (r *inventoryRepo) ReserveStock(ctx context.Context, productID, warehouseID uint64, quantity int) error {
	result := r.db.WithContext(ctx).Model(&models.Inventory{}).
		Where("product_id = ? AND warehouse_id = ? AND quantity - reserved_quantity >= ?", productID, warehouseID, quantity).
		Update("reserved_quantity", gorm.Expr("reserved_quantity + ?", quantity))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("insufficient stock")
	}
	return nil
}

func (r *inventoryRepo) ReleaseStock(ctx context.Context, productID, warehouseID uint64, quantity int) error {
	return r.db.WithContext(ctx).Model(&models.Inventory{}).
		Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		Update("reserved_quantity", gorm.Expr("reserved_quantity - ?", quantity)).Error
}

type warehouseRepo struct{ db *gorm.DB }

func NewWarehouseRepository(db *gorm.DB) WarehouseRepository {
	return &warehouseRepo{db: db}
}

func (r *warehouseRepo) Create(ctx context.Context, w *models.Warehouse) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *warehouseRepo) GetByID(ctx context.Context, id uint64) (*models.Warehouse, error) {
	var w models.Warehouse
	err := r.db.WithContext(ctx).First(&w, id).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *warehouseRepo) GetByCode(ctx context.Context, code string) (*models.Warehouse, error) {
	var w models.Warehouse
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&w).Error
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *warehouseRepo) Update(ctx context.Context, w *models.Warehouse) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *warehouseRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.Warehouse{}, id).Error
}

func (r *warehouseRepo) List(ctx context.Context) ([]*models.Warehouse, error) {
	var list []*models.Warehouse
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&list).Error
	return list, err
}

type stockReservationRepo struct{ db *gorm.DB }

func NewStockReservationRepository(db *gorm.DB) StockReservationRepository {
	return &stockReservationRepo{db: db}
}

func (r *stockReservationRepo) Create(ctx context.Context, s *models.StockReservation) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *stockReservationRepo) GetByOrderID(ctx context.Context, orderID uint64) (*models.StockReservation, error) {
	var s models.StockReservation
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *stockReservationRepo) Confirm(ctx context.Context, orderID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.StockReservation{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":       "confirmed",
			"confirmed_at": now,
		}).Error
}

func (r *stockReservationRepo) Release(ctx context.Context, orderID uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.StockReservation{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":      "released",
			"released_at": now,
		}).Error
}

type inventoryLogRepo struct{ db *gorm.DB }

func NewInventoryLogRepository(db *gorm.DB) InventoryLogRepository {
	return &inventoryLogRepo{db: db}
}

func (r *inventoryLogRepo) Create(ctx context.Context, l *models.InventoryLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *inventoryLogRepo) GetByProductID(ctx context.Context, productID uint64, limit int) ([]*models.InventoryLog, error) {
	var list []*models.InventoryLog
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).
		Order("created_at DESC").Limit(limit).Find(&list).Error
	return list, err
}