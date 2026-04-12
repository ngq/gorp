// Package models 库存服务数据模型
package models

import (
	"time"
)

// Inventory 库存实体
type Inventory struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	ProductID        uint64    `gorm:"not null;uniqueIndex:idx_product_warehouse" json:"product_id"`
	WarehouseID      uint64    `gorm:"not null;uniqueIndex:idx_product_warehouse" json:"warehouse_id"`
	Quantity         int       `gorm:"not null;default:0" json:"quantity"`
	ReservedQuantity int       `gorm:"not null;default:0" json:"reserved_quantity"`
	MinStock         int       `gorm:"default:10" json:"min_stock"`
	MaxStock         int       `gorm:"default:1000" json:"max_stock"`
	LowStockAction   string    `gorm:"size:16;default:'notify'" json:"low_stock_action"` // none, notify, disable
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Inventory) TableName() string {
	return "inventories"
}

// AvailableQuantity 可用库存
func (i *Inventory) AvailableQuantity() int {
	return i.Quantity - i.ReservedQuantity
}

// Warehouse 仓库
type Warehouse struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:256;not null" json:"name"`
	Code      string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Address   string    `gorm:"size:512" json:"address"`
	City      string    `gorm:"size:128" json:"city"`
	State     string    `gorm:"size:128" json:"state"`
	Country   string    `gorm:"size:64" json:"country"`
	ZipCode   string    `gorm:"size:16" json:"zip_code"`
	Phone     string    `gorm:"size:32" json:"phone"`
	Email     string    `gorm:"size:128" json:"email"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Warehouse) TableName() string {
	return "warehouses"
}

// InventoryLog 库存变动日志
type InventoryLog struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	ProductID     uint64    `gorm:"not null;index" json:"product_id"`
	WarehouseID   uint64    `gorm:"not null;index" json:"warehouse_id"`
	ChangeType    string    `gorm:"size:16;not null" json:"change_type"` // in, out, reserve, release, adjust
	Quantity      int       `gorm:"not null" json:"quantity"`
	ReferenceType string    `gorm:"size:64" json:"reference_type"` // order, return, adjustment
	ReferenceID   uint64    `json:"reference_id"`
	Remark        string    `gorm:"size:256" json:"remark"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (InventoryLog) TableName() string {
	return "inventory_logs"
}

// StockReservation 库存预留
type StockReservation struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	OrderID      uint64    `gorm:"not null;uniqueIndex" json:"order_id"`
	ProductID    uint64    `gorm:"not null;index" json:"product_id"`
	WarehouseID  uint64    `gorm:"not null" json:"warehouse_id"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	Status       string    `gorm:"size:16;default:'reserved'" json:"status"` // reserved, confirmed, released
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	ConfirmedAt  *time.Time `json:"confirmed_at"`
	ReleasedAt   *time.Time `json:"released_at"`
}

func (StockReservation) TableName() string {
	return "stock_reservations"
}

// TierPrice 阶梯价格
type TierPrice struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	ProductID      uint64     `gorm:"not null;index" json:"product_id"`
	CustomerRoleID *uint64    `json:"customer_role_id"`
	Quantity       int        `gorm:"not null" json:"quantity"`
	Price          float64    `gorm:"type:decimal(10,2);not null" json:"price"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (TierPrice) TableName() string {
	return "tier_prices"
}

// DTO
type InventoryResponse struct {
	ID               uint64 `json:"id"`
	ProductID        uint64 `json:"product_id"`
	WarehouseID      uint64 `json:"warehouse_id"`
	Quantity         int    `json:"quantity"`
	ReservedQuantity int    `json:"reserved_quantity"`
	AvailableQuantity int   `json:"available_quantity"`
	MinStock         int    `json:"min_stock"`
	MaxStock         int    `json:"max_stock"`
	IsLowStock       bool   `json:"is_low_stock"`
}

func ToInventoryResponse(i *Inventory) InventoryResponse {
	return InventoryResponse{
		ID:               i.ID,
		ProductID:        i.ProductID,
		WarehouseID:      i.WarehouseID,
		Quantity:         i.Quantity,
		ReservedQuantity: i.ReservedQuantity,
		AvailableQuantity: i.AvailableQuantity(),
		MinStock:         i.MinStock,
		MaxStock:         i.MaxStock,
		IsLowStock:       i.AvailableQuantity() < i.MinStock,
	}
}

type ReserveStockRequest struct {
	OrderID     uint64 `json:"order_id" binding:"required"`
	ProductID   uint64 `json:"product_id" binding:"required"`
	WarehouseID uint64 `json:"warehouse_id" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
}

type AdjustStockRequest struct {
	ProductID   uint64 `json:"product_id" binding:"required"`
	WarehouseID uint64 `json:"warehouse_id" binding:"required"`
	Quantity    int    `json:"quantity"` // 正数入库，负数出库
	Remark      string `json:"remark"`
}