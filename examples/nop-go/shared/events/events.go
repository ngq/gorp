// Package events 定义跨服务事件
//
// 中文说明：
// - 定义消息队列事件格式；
// - 用于服务间异步通信。
package events

import (
	"fmt"
	"time"
)

// Event 基础事件
type Event struct {
	ID        string    `json:"id"`         // 事件ID
	Type      string    `json:"type"`       // 事件类型
	Source    string    `json:"source"`     // 来源服务
	Timestamp time.Time `json:"timestamp"`  // 事件时间
	Payload   any       `json:"payload"`    // 事件负载
}

// NewEvent 创建事件
func NewEvent(eventType, source string, payload any) *Event {
	return &Event{
		ID:        generateID(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// 事件类型常量
const (
	// 订单事件
	EventOrderCreated   = "order.created"
	EventOrderCancelled = "order.cancelled"
	EventOrderPaid      = "order.paid"
	EventOrderShipped   = "order.shipped"
	EventOrderDelivered = "order.delivered"

	// 支付事件
	EventPaymentCreated   = "payment.created"
	EventPaymentCompleted = "payment.completed"
	EventPaymentFailed    = "payment.failed"
	EventPaymentRefunded  = "payment.refunded"

	// 库存事件
	EventInventoryReserved  = "inventory.reserved"
	EventInventoryReleased  = "inventory.released"
	EventInventoryConfirmed = "inventory.confirmed"
	EventInventoryLowStock  = "inventory.low_stock"

	// 客户事件
	EventCustomerRegistered = "customer.registered"
	EventCustomerLogin      = "customer.login"

	// 配送事件
	EventShipmentCreated   = "shipment.created"
	EventShipmentUpdated   = "shipment.updated"
	EventShipmentDelivered = "shipment.delivered"
)

// OrderCreatedPayload 订单创建事件负载
type OrderCreatedPayload struct {
	OrderID    uint64 `json:"order_id"`
	OrderNumber string `json:"order_number"`
	CustomerID uint64 `json:"customer_id"`
	Total      string `json:"total"`
	Items      []OrderItemPayload `json:"items"`
}

// OrderItemPayload 订单项负载
type OrderItemPayload struct {
	ProductID uint64 `json:"product_id"`
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
}

// PaymentCompletedPayload 支付完成事件负载
type PaymentCompletedPayload struct {
	OrderID       uint64 `json:"order_id"`
	PaymentID     uint64 `json:"payment_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	PaidAt        string `json:"paid_at"`
}

// InventoryReservedPayload 库存预留事件负载
type InventoryReservedPayload struct {
	ReservationID string                 `json:"reservation_id"`
	Items         []InventoryItemPayload `json:"items"`
	OrderID       uint64                 `json:"order_id"`
}

// InventoryItemPayload 库存项负载
type InventoryItemPayload struct {
	ProductID   uint64 `json:"product_id"`
	WarehouseID uint64 `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
}

// generateID 生成事件ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}