// Package rpc RPC 客户端封装
// 基于框架 contract.RPCClient 能力
package rpc

import (
	"context"

	"github.com/ngq/gorp/framework/contract"

	"nop-go/shared/inventory"
	"nop-go/shared/payment"
)

const (
	// 服务名称常量
	ServiceInventory = "inventory-service"
	ServicePrice     = "price-service"
	ServicePayment   = "payment-service"
)

// InventoryClient 库存服务客户端
type InventoryClient struct {
	client contract.RPCClient
}

// NewInventoryClient 创建库存服务客户端
func NewInventoryClient(client contract.RPCClient) *InventoryClient {
	return &InventoryClient{client: client}
}

// ReserveStock 预留库存
func (c *InventoryClient) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
	resp := &inventory.ReserveStockResponse{}
	err := c.client.Call(ctx, ServiceInventory, "ReserveStock", req, resp)
	if err != nil {
		resp.Success = false
		resp.ErrorMessage = err.Error()
	}
	return resp, err
}

// ConfirmStock 确认库存
func (c *InventoryClient) ConfirmStock(ctx context.Context, req *inventory.ConfirmStockRequest) (*inventory.ConfirmStockResponse, error) {
	resp := &inventory.ConfirmStockResponse{}
	err := c.client.Call(ctx, ServiceInventory, "ConfirmStock", req, resp)
	return resp, err
}

// ReleaseStock 释放库存
func (c *InventoryClient) ReleaseStock(ctx context.Context, req *inventory.ReleaseStockRequest) (*inventory.ReleaseStockResponse, error) {
	resp := &inventory.ReleaseStockResponse{}
	err := c.client.Call(ctx, ServiceInventory, "ReleaseStock", req, resp)
	return resp, err
}

// PriceClient 价格服务客户端
type PriceClient struct {
	client contract.RPCClient
}

// NewPriceClient 创建价格服务客户端
func NewPriceClient(client contract.RPCClient) *PriceClient {
	return &PriceClient{client: client}
}

// PaymentClient 支付服务客户端
type PaymentClient struct {
	client contract.RPCClient
}

// NewPaymentClient 创建支付服务客户端
func NewPaymentClient(client contract.RPCClient) *PaymentClient {
	return &PaymentClient{client: client}
}

// CreatePayment 创建支付
func (c *PaymentClient) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	resp := &payment.CreatePaymentResponse{}
	err := c.client.Call(ctx, ServicePayment, "CreatePayment", req, resp)
	if err != nil {
		resp.Success = false
		resp.ErrorMessage = err.Error()
	}
	return resp, err
}

// CancelPayment 取消支付
func (c *PaymentClient) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*payment.CancelPaymentResponse, error) {
	resp := &payment.CancelPaymentResponse{}
	err := c.client.Call(ctx, ServicePayment, "CancelPayment", req, resp)
	return resp, err
}