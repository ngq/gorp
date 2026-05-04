// Package rpc RPC 瀹㈡埛绔皝瑁?
// 鍩轰簬妗嗘灦 transportcontract.RPCClient 鑳藉姏
package rpc

import (
	"context"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"

	"nop-go/shared/inventory"
	"nop-go/shared/payment"
)

const (
	// 鏈嶅姟鍚嶇О甯搁噺
	ServiceInventory = "inventory-service"
	ServicePrice     = "price-service"
	ServicePayment   = "payment-service"
)

// InventoryClient 搴撳瓨鏈嶅姟瀹㈡埛绔?
type InventoryClient struct {
	client transportcontract.RPCClient
}

// NewInventoryClient 鍒涘缓搴撳瓨鏈嶅姟瀹㈡埛绔?
func NewInventoryClient(client transportcontract.RPCClient) *InventoryClient {
	return &InventoryClient{client: client}
}

// ReserveStock 棰勭暀搴撳瓨
func (c *InventoryClient) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
	resp := &inventory.ReserveStockResponse{}
	err := c.client.Call(ctx, ServiceInventory, "ReserveStock", req, resp)
	if err != nil {
		resp.Success = false
		resp.ErrorMessage = err.Error()
	}
	return resp, err
}

// ConfirmStock 纭搴撳瓨
func (c *InventoryClient) ConfirmStock(ctx context.Context, req *inventory.ConfirmStockRequest) (*inventory.ConfirmStockResponse, error) {
	resp := &inventory.ConfirmStockResponse{}
	err := c.client.Call(ctx, ServiceInventory, "ConfirmStock", req, resp)
	return resp, err
}

// ReleaseStock 閲婃斁搴撳瓨
func (c *InventoryClient) ReleaseStock(ctx context.Context, req *inventory.ReleaseStockRequest) (*inventory.ReleaseStockResponse, error) {
	resp := &inventory.ReleaseStockResponse{}
	err := c.client.Call(ctx, ServiceInventory, "ReleaseStock", req, resp)
	return resp, err
}

// PriceClient 浠锋牸鏈嶅姟瀹㈡埛绔?
type PriceClient struct {
	client transportcontract.RPCClient
}

// NewPriceClient 鍒涘缓浠锋牸鏈嶅姟瀹㈡埛绔?
func NewPriceClient(client transportcontract.RPCClient) *PriceClient {
	return &PriceClient{client: client}
}

// PaymentClient 鏀粯鏈嶅姟瀹㈡埛绔?
type PaymentClient struct {
	client transportcontract.RPCClient
}

// NewPaymentClient 鍒涘缓鏀粯鏈嶅姟瀹㈡埛绔?
func NewPaymentClient(client transportcontract.RPCClient) *PaymentClient {
	return &PaymentClient{client: client}
}

// CreatePayment 鍒涘缓鏀粯
func (c *PaymentClient) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	resp := &payment.CreatePaymentResponse{}
	err := c.client.Call(ctx, ServicePayment, "CreatePayment", req, resp)
	if err != nil {
		resp.Success = false
		resp.ErrorMessage = err.Error()
	}
	return resp, err
}

// CancelPayment 鍙栨秷鏀粯
func (c *PaymentClient) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*payment.CancelPaymentResponse, error) {
	resp := &payment.CancelPaymentResponse{}
	err := c.client.Call(ctx, ServicePayment, "CancelPayment", req, resp)
	return resp, err
}
