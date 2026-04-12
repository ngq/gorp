// Package models 物流服务数据模型
package models

import (
	"time"
)

// Shipment 发货单
type Shipment struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	OrderID          uint64     `gorm:"not null;uniqueIndex" json:"order_id"`
	TrackingNumber   string     `gorm:"size:64" json:"tracking_number"`
	ShippingMethod   string     `gorm:"size:64;not null" json:"shipping_method"`
	ShippingProvider string     `gorm:"size:64" json:"shipping_provider"` // SF, JD, YTO, etc.
	Status           string     `gorm:"size:16;default:'pending'" json:"status"` // pending, shipped, in_transit, delivered
	ShippedAt        *time.Time `json:"shipped_at"`
	DeliveredAt      *time.Time `json:"delivered_at"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
	Weight           float64    `gorm:"type:decimal(10,3)" json:"weight"`
	ShippingFee      float64    `gorm:"type:decimal(10,2)" json:"shipping_fee"`
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Items []ShipmentItem `gorm:"foreignKey:ShipmentID" json:"items,omitempty"`
}

func (Shipment) TableName() string {
	return "shipments"
}

// ShipmentItem 发货商品
type ShipmentItem struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	ShipmentID   uint64    `gorm:"not null;index" json:"shipment_id"`
	OrderItemID  uint64    `gorm:"not null" json:"order_item_id"`
	ProductID    uint64    `gorm:"not null" json:"product_id"`
	ProductName  string    `gorm:"size:256;not null" json:"product_name"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ShipmentItem) TableName() string {
	return "shipment_items"
}

// ShippingMethod 配送方式
type ShippingMethod struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:64;not null" json:"name"`
	SystemName     string    `gorm:"size:64;uniqueIndex;not null" json:"system_name"`
	Description    string    `gorm:"type:text" json:"description"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	DisplayOrder   int       `gorm:"default:0" json:"display_order"`
	Rate           float64   `gorm:"type:decimal(10,2);default:0" json:"rate"` // 基础运费
	FreeShippingOver float64 `gorm:"type:decimal(10,2)" json:"free_shipping_over"` // 满额免运费
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ShippingMethod) TableName() string {
	return "shipping_methods"
}

// ShipmentTracking 物流跟踪
type ShipmentTracking struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	ShipmentID  uint64    `gorm:"not null;index" json:"shipment_id"`
	Status      string    `gorm:"size:32" json:"status"`
	Location    string    `gorm:"size:256" json:"location"`
	Description string    `gorm:"size:512" json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ShipmentTracking) TableName() string {
	return "shipment_tracking"
}

// DTO
type CreateShipmentRequest struct {
	OrderID          uint64                 `json:"order_id" binding:"required"`
	ShippingMethod   string                 `json:"shipping_method" binding:"required"`
	ShippingProvider string                 `json:"shipping_provider"`
	Items            []ShipmentItemInput    `json:"items" binding:"required"`
}

type ShipmentItemInput struct {
	OrderItemID uint64 `json:"order_item_id" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1"`
}

type UpdateTrackingRequest struct {
	TrackingNumber string `json:"tracking_number"`
	Status         string `json:"status"`
	Location       string `json:"location"`
	Description    string `json:"description"`
}

type ShipmentResponse struct {
	ID               uint64                `json:"id"`
	OrderID          uint64                `json:"order_id"`
	TrackingNumber   string                `json:"tracking_number"`
	ShippingMethod   string                `json:"shipping_method"`
	ShippingProvider string                `json:"shipping_provider"`
	Status           string                `json:"status"`
	ShippedAt        string                `json:"shipped_at"`
	DeliveredAt      string                `json:"delivered_at"`
	Items            []ShipmentItemResponse `json:"items"`
	CreatedAt        string                `json:"created_at"`
}

type ShipmentItemResponse struct {
	ProductID   uint64 `json:"product_id"`
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
}

func ToShipmentResponse(s *Shipment) ShipmentResponse {
	resp := ShipmentResponse{
		ID:               s.ID,
		OrderID:          s.OrderID,
		TrackingNumber:   s.TrackingNumber,
		ShippingMethod:   s.ShippingMethod,
		ShippingProvider: s.ShippingProvider,
		Status:           s.Status,
		CreatedAt:        s.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if s.ShippedAt != nil {
		resp.ShippedAt = s.ShippedAt.Format("2006-01-02 15:04:05")
	}
	if s.DeliveredAt != nil {
		resp.DeliveredAt = s.DeliveredAt.Format("2006-01-02 15:04:05")
	}

	if len(s.Items) > 0 {
		resp.Items = make([]ShipmentItemResponse, len(s.Items))
		for i, item := range s.Items {
			resp.Items[i] = ShipmentItemResponse{
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
			}
		}
	}

	return resp
}