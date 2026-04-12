// Package saga DTM SAGA 分布式事务实现
//
// 中文说明:
// - 使用 DTM 实现 SAGA 模式分布式事务;
// - 订单创建流程: 预留库存 -> 创建支付 -> 创建订单 -> 确认库存;
// - 补偿流程: 取消支付 -> 释放库存。
package saga

import (
	"context"
	"fmt"

	"nop-go/shared/inventory"
	"nop-go/shared/payment"
	rpc "nop-go/shared/rpc"
)

// OrderCreateSaga 订单创建 SAGA 编排器
type OrderCreateSaga struct {
	inventoryClient *rpc.InventoryClient
	priceClient     *rpc.PriceClient
	paymentClient   *rpc.PaymentClient
	dtmServer       string // DTM 服务器地址
}

// NewOrderCreateSaga 创建订单 SAGA 编排器
func NewOrderCreateSaga(
	inventoryClient *rpc.InventoryClient,
	priceClient *rpc.PriceClient,
	paymentClient *rpc.PaymentClient,
	dtmServer string,
) *OrderCreateSaga {
	return &OrderCreateSaga{
		inventoryClient: inventoryClient,
		priceClient:     priceClient,
		paymentClient:   paymentClient,
		dtmServer:       dtmServer,
	}
}

// Execute 执行订单创建 SAGA
//
// 中文说明:
// - SAGA 流程:
//   1. 验证客户 (只读)
//   2. 计算价格 (只读)
//   3. 预留库存 (Try) -> 释放库存
//   4. 创建支付 -> 取消支付
//   5. 创建订单 (本地事务)
//   6. 确认库存
//
// - 如果任何步骤失败,DTM 会自动触发补偿操作。
func (s *OrderCreateSaga) Execute(ctx context.Context, req *CreateOrderSagaRequest) (*CreateOrderSagaResult, error) {
	// 生成事务 ID
	gid := fmt.Sprintf("ORDER-%d-%d", req.CustomerID, req.OrderID)

	// 步骤 1: 验证客户 (只读,不需要补偿)

	// 步骤 2: 计算价格 (只读,不需要补偿)

	// 步骤 3: 预留库存
	//
	// 中文说明:
	// - 调用 inventory-service 的 ReserveStock;
	// - 补偿操作: ReleaseStock。
	var reservationID string
	for _, item := range req.Items {
		reserveResp, err := s.inventoryClient.ReserveStock(ctx, &inventory.ReserveStockRequest{
			OrderID:       req.OrderID,
			ProductID:     item.ProductID,
			WarehouseID:   item.WarehouseID,
			Quantity:      int32(item.Quantity),
			ReservationID: gid,
		})
		if err != nil {
			// 预留失败,回滚已预留的库存
			s.rollbackInventory(ctx, req.OrderID, gid)
			return nil, fmt.Errorf("reserve stock failed: %w", err)
		}
		if !reserveResp.Success {
			s.rollbackInventory(ctx, req.OrderID, gid)
			return nil, fmt.Errorf("reserve stock failed: %s", reserveResp.ErrorMessage)
		}
		reservationID = reserveResp.ReservationID
	}

	// 步骤 4: 创建支付
	//
	// 中文说明:
	// - 调用 payment-service 的 CreatePayment;
	// - 补偿操作: CancelPayment。
	paymentResp, err := s.paymentClient.CreatePayment(ctx, &payment.CreatePaymentRequest{
		OrderID:       req.OrderID,
		CustomerID:    req.CustomerID,
		Amount:        req.Total,
		Currency:      "CNY",
		PaymentMethod: req.PaymentMethod,
		ReturnURL:     req.ReturnURL,
		NotifyURL:     req.NotifyURL,
		TransactionID: gid,
	})
	if err != nil {
		s.rollbackInventory(ctx, req.OrderID, gid)
		return nil, fmt.Errorf("create payment failed: %w", err)
	}
	if !paymentResp.Success {
		s.rollbackInventory(ctx, req.OrderID, gid)
		return nil, fmt.Errorf("create payment failed: %s", paymentResp.ErrorMessage)
	}

	// 返回结果
	// 注意: 订单创建在本地事务中完成,确认库存通过支付回调触发
	return &CreateOrderSagaResult{
		OrderID:        req.OrderID,
		PaymentID:      paymentResp.PaymentID,
		TransactionID:  paymentResp.TransactionID,
		PayURL:         paymentResp.PayURL,
		ReservationID:  reservationID,
	}, nil
}

// rollbackInventory 回滚库存预留
//
// 中文说明:
// - SAGA 补偿操作;
// - 释放已预留的库存。
func (s *OrderCreateSaga) rollbackInventory(ctx context.Context, orderID uint64, reservationID string) {
	_, err := s.inventoryClient.ReleaseStock(ctx, &inventory.ReleaseStockRequest{
		OrderID:       orderID,
		ReservationID: reservationID,
	})
	if err != nil {
		// 记录日志,但不影响返回的错误
		fmt.Printf("rollback inventory failed: %v\n", err)
	}
}

// ConfirmOrder 确认订单
//
// 中文说明:
// - 支付成功后调用;
// - 确认库存扣减;
// - 更新订单状态。
func (s *OrderCreateSaga) ConfirmOrder(ctx context.Context, orderID uint64, reservationID string) error {
	_, err := s.inventoryClient.ConfirmStock(ctx, &inventory.ConfirmStockRequest{
		OrderID:       orderID,
		ReservationID: reservationID,
	})
	return err
}

// CancelOrder 取消订单
//
// 中文说明:
// - 订单取消时调用;
// - 取消支付;
// - 释放库存。
func (s *OrderCreateSaga) CancelOrder(ctx context.Context, orderID, paymentID uint64, reservationID string) error {
	// 取消支付
	_, err := s.paymentClient.CancelPayment(ctx, &payment.CancelPaymentRequest{
		OrderID:   orderID,
		PaymentID: paymentID,
		Reason:    "order cancelled",
	})
	if err != nil {
		fmt.Printf("cancel payment failed: %v\n", err)
	}

	// 释放库存
	_, err = s.inventoryClient.ReleaseStock(ctx, &inventory.ReleaseStockRequest{
		OrderID:       orderID,
		ReservationID: reservationID,
	})
	if err != nil {
		fmt.Printf("release stock failed: %v\n", err)
	}

	return nil
}

// CreateOrderSagaRequest 创建订单 SAGA 请求
type CreateOrderSagaRequest struct {
	OrderID       uint64
	CustomerID    uint64
	Items         []*SagaOrderItem
	Total         float64
	PaymentMethod string
	ReturnURL     string
	NotifyURL     string
}

// SagaOrderItem SAGA 订单商品项
type SagaOrderItem struct {
	ProductID   uint64
	WarehouseID uint64
	Quantity    int
}

// CreateOrderSagaResult 创建订单 SAGA 结果
type CreateOrderSagaResult struct {
	OrderID        uint64
	PaymentID      uint64
	TransactionID  string
	PayURL         string
	ReservationID  string
}