// Package biz 订单服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nop-go/services/order-service/internal/data"
	"nop-go/services/order-service/internal/models"
	shareErrors "nop-go/shared/errors"
)

type OrderUseCase struct {
	orderRepo     data.OrderRepository
	orderItemRepo data.OrderItemRepository
	addressRepo   data.OrderAddressRepository
	giftCardRepo  data.GiftCardRepository
	returnRepo    data.ReturnRequestRepository
}

func NewOrderUseCase(
	orderRepo data.OrderRepository,
	orderItemRepo data.OrderItemRepository,
	addressRepo data.OrderAddressRepository,
	giftCardRepo data.GiftCardRepository,
	returnRepo data.ReturnRequestRepository,
) *OrderUseCase {
	return &OrderUseCase{
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		addressRepo:   addressRepo,
		giftCardRepo:  giftCardRepo,
		returnRepo:    returnRepo,
	}
}

func (uc *OrderUseCase) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	orderNumber := generateOrderNumber()

	order := &models.Order{
		OrderNumber:   orderNumber,
		CustomerID:    req.CustomerID,
		OrderStatus:   "pending",
		PaymentStatus: "pending",
		ShippingStatus: "not_shipped",
		CustomerNote:  req.CustomerNote,
	}

	var subtotal float64
	for _, item := range req.Items {
		itemTotal := float64(item.Quantity) * 100 // TODO: 获取实际价格
		order.Items = append(order.Items, models.OrderItem{
			ProductID:   item.ProductID,
			ProductName: "Product",
			SKU:         "SKU",
			Quantity:    item.Quantity,
			UnitPrice:   100,
			Total:       itemTotal,
		})
		subtotal += itemTotal
	}

	order.Subtotal = subtotal
	order.Total = subtotal

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	for _, addr := range []struct {
		t string
		a models.AddressInput
	}{
		{"billing", req.BillingAddress},
		{"shipping", req.ShippingAddress},
	} {
		orderAddr := &models.OrderAddress{
			OrderID:     order.ID,
			AddressType: addr.t,
			FirstName:   addr.a.FirstName,
			LastName:    addr.a.LastName,
			Email:       addr.a.Email,
			Phone:       addr.a.Phone,
			Company:     addr.a.Company,
			Country:     addr.a.Country,
			State:       addr.a.State,
			City:        addr.a.City,
			Address1:    addr.a.Address1,
			Address2:    addr.a.Address2,
			ZipCode:     addr.a.ZipCode,
		}
		uc.addressRepo.Create(ctx, orderAddr)
	}

	return order, nil
}

func (uc *OrderUseCase) GetOrder(ctx context.Context, id uint64) (*models.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrOrderNotFound
	}
	return order, nil
}

func (uc *OrderUseCase) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	return uc.orderRepo.GetByOrderNumber(ctx, orderNumber)
}

func (uc *OrderUseCase) GetCustomerOrders(ctx context.Context, customerID uint64, page, pageSize int) ([]*models.Order, int64, error) {
	return uc.orderRepo.GetByCustomerID(ctx, customerID, page, pageSize)
}

func (uc *OrderUseCase) UpdateOrderStatus(ctx context.Context, id uint64, status string) error {
	validStatuses := map[string]bool{
		"pending": true, "processing": true, "complete": true, "cancelled": true, "refunded": true,
	}
	if !validStatuses[status] {
		return shareErrors.ErrInvalidOrderStatus
	}
	return uc.orderRepo.UpdateStatus(ctx, id, status)
}

func (uc *OrderUseCase) CancelOrder(ctx context.Context, id uint64) error {
	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		return shareErrors.ErrOrderNotFound
	}

	if order.OrderStatus != "pending" && order.OrderStatus != "processing" {
		return shareErrors.ErrOrderCannotCancel
	}

	return uc.orderRepo.UpdateStatus(ctx, id, "cancelled")
}

func (uc *OrderUseCase) ListOrders(ctx context.Context, page, pageSize int) ([]*models.Order, int64, error) {
	return uc.orderRepo.List(ctx, page, pageSize)
}

type GiftCardUseCase struct {
	giftCardRepo data.GiftCardRepository
}

func NewGiftCardUseCase(giftCardRepo data.GiftCardRepository) *GiftCardUseCase {
	return &GiftCardUseCase{giftCardRepo: giftCardRepo}
}

func (uc *GiftCardUseCase) ValidateGiftCard(ctx context.Context, code string) (*models.GiftCard, error) {
	card, err := uc.giftCardRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, shareErrors.ErrGiftCardNotFound
	}

	if !card.IsActive {
		return nil, errors.New("gift card is not active")
	}

	if card.IsRedeemed {
		return nil, shareErrors.ErrGiftCardUsed
	}

	if card.ExpiresAt != nil && card.ExpiresAt.Before(time.Now()) {
		return nil, shareErrors.ErrGiftCardExpired
	}

	return card, nil
}

func (uc *GiftCardUseCase) RedeemGiftCard(ctx context.Context, code string, customerID uint64) error {
	return uc.giftCardRepo.Redeem(ctx, code, customerID)
}

type ReturnRequestUseCase struct {
	returnRepo data.ReturnRequestRepository
	orderRepo  data.OrderRepository
}

func NewReturnRequestUseCase(returnRepo data.ReturnRequestRepository, orderRepo data.OrderRepository) *ReturnRequestUseCase {
	return &ReturnRequestUseCase{returnRepo: returnRepo, orderRepo: orderRepo}
}

func (uc *ReturnRequestUseCase) CreateReturnRequest(ctx context.Context, req *models.ReturnRequest) error {
	return uc.returnRepo.Create(ctx, req)
}

func (uc *ReturnRequestUseCase) GetReturnRequest(ctx context.Context, id uint64) (*models.ReturnRequest, error) {
	return uc.returnRepo.GetByID(ctx, id)
}

func (uc *ReturnRequestUseCase) ApproveReturnRequest(ctx context.Context, id uint64, adminID uint64) error {
	req, err := uc.returnRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	now := time.Now()
	req.Status = "approved"
	req.ProcessedBy = adminID
	req.ProcessedAt = &now

	return uc.returnRepo.Update(ctx, req)
}

func generateOrderNumber() string {
	return fmt.Sprintf("ORD%s", time.Now().Format("20060102150405"))
}