// Package service 定义 trade-service 的 gRPC RPC 接口。
// gorp proto from-service 命令通过解析此文件中的接口定义来生成 .proto 文件。
// 只有其他服务需要通过 gRPC 跨服务调用的方法才定义在此接口中。
package service

import "context"

// TradeRPC 交易服务 gRPC 接口 —— 定义其他服务可能需要跨服务调用的方法。
// 当前暂只有 GetOrder 供 gateway 或其他服务查询订单状态，后续可扩展。
type TradeRPC interface {
	// GetOrder 根据 ID 获取订单信息（供 gateway 查询订单状态）
	GetOrder(ctx context.Context, req *GetOrderReq) (*GetOrderResp, error)
}

// ======================== gRPC 请求/响应类型 ========================

// GetOrderReq 获取订单请求
type GetOrderReq struct {
	ID string `json:"id" remark:"订单ID"`
}

// GetOrderResp 获取订单响应
type GetOrderResp struct {
	ID            string  `json:"id" remark:"订单ID"`
	UserID        string  `json:"user_id" remark:"用户ID"`
	Status        string  `json:"status" remark:"订单状态"`
	TotalAmount   float64 `json:"total_amount" remark:"订单总额"`
	Currency      string  `json:"currency" remark:"货币"`
	ShippingAddr  string  `json:"shipping_addr" remark:"收货地址"`
	PaymentMethod string  `json:"payment_method" remark:"支付方式"`
	CreatedAt     string  `json:"created_at" remark:"创建时间"`
	UpdatedAt     string  `json:"updated_at" remark:"更新时间"`
}