// Package models 购物车服务数据模型
package models

import (
	"time"
)

// ShoppingCart 购物车
type ShoppingCart struct {
	ID           uint64        `gorm:"primaryKey" json:"id"`
	CustomerID   uint64        `gorm:"uniqueIndex" json:"customer_id"`
	SessionID    string        `gorm:"size:64;uniqueIndex" json:"session_id"`
	CouponCode   string        `gorm:"size:64" json:"coupon_code"`
	Subtotal     float64       `gorm:"type:decimal(10,2);default:0" json:"subtotal"`
	Discount     float64       `gorm:"type:decimal(10,2);default:0" json:"discount"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time     `gorm:"autoUpdateTime" json:"updated_at"`

	Items []CartItem `gorm:"foreignKey:CartID" json:"items,omitempty"`
}

func (ShoppingCart) TableName() string {
	return "shopping_carts"
}

// CartItem 购物车商品
type CartItem struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	CartID      uint64    `gorm:"not null;index" json:"cart_id"`
	ProductID   uint64    `gorm:"not null" json:"product_id"`
	ProductName string    `gorm:"size:256;not null" json:"product_name"`
	SKU         string    `gorm:"size:64;not null" json:"sku"`
	Quantity    int       `gorm:"not null;default:1" json:"quantity"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	Attributes  string    `gorm:"type:json" json:"attributes"` // 商品属性 JSON
	ImageURL    string    `gorm:"size:512" json:"image_url"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (CartItem) TableName() string {
	return "cart_items"
}

// Wishlist 愿望清单
type Wishlist struct {
	ID         uint64          `gorm:"primaryKey" json:"id"`
	CustomerID uint64          `gorm:"not null;uniqueIndex" json:"customer_id"`
	CreatedAt  time.Time       `gorm:"autoCreateTime" json:"created_at"`

	Items []WishlistItem `gorm:"foreignKey:WishlistID" json:"items,omitempty"`
}

func (Wishlist) TableName() string {
	return "wishlists"
}

// WishlistItem 愿望清单商品
type WishlistItem struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	WishlistID  uint64    `gorm:"not null;index" json:"wishlist_id"`
	ProductID   uint64    `gorm:"not null" json:"product_id"`
	ProductName string    `gorm:"size:256;not null" json:"product_name"`
	ImageURL    string    `gorm:"size:512" json:"image_url"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (WishlistItem) TableName() string {
	return "wishlist_items"
}

// DTO
type AddToCartRequest struct {
	ProductID  uint64 `json:"product_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
	Attributes string `json:"attributes"` // JSON
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type CartResponse struct {
	ID         uint64            `json:"id"`
	Items      []CartItemResponse `json:"items"`
	ItemCount  int               `json:"item_count"`
	Subtotal   float64           `json:"subtotal"`
	Discount   float64           `json:"discount"`
	Total      float64           `json:"total"`
	CouponCode string            `json:"coupon_code"`
}

type CartItemResponse struct {
	ID          uint64  `json:"id"`
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
	ImageURL    string  `json:"image_url"`
	Attributes  string  `json:"attributes"`
}

func ToCartResponse(cart *ShoppingCart) CartResponse {
	resp := CartResponse{
		ID:         cart.ID,
		Subtotal:   cart.Subtotal,
		Discount:   cart.Discount,
		Total:      cart.Subtotal - cart.Discount,
		CouponCode: cart.CouponCode,
	}

	if len(cart.Items) > 0 {
		resp.Items = make([]CartItemResponse, len(cart.Items))
		for i, item := range cart.Items {
			resp.Items[i] = CartItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				SKU:         item.SKU,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Total:       item.UnitPrice * float64(item.Quantity),
				ImageURL:    item.ImageURL,
				Attributes:  item.Attributes,
			}
			resp.ItemCount += item.Quantity
		}
	}

	return resp
}