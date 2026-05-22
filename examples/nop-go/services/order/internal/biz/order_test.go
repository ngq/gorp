// Package biz_test 订单服务业务逻辑层单元测试。
//
// 使用 mock 仓储隔离依赖，测试所有用例的核心业务逻辑。
package biz_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nop-go/services/order/internal/biz"
)

// ============================================================
// Mock 仓储实现 - 订单仓储
// ============================================================

// MockOrderRepository 订单仓储 mock 实现。
type MockOrderRepository struct {
	Orders    map[uint]*biz.Order
	OrderNo   map[string]*biz.Order
	Items     map[uint][]*biz.OrderItem
	Shipments map[uint]*biz.Shipment // key: orderID*1000+shipmentID
	NextID    uint
}

// NewMockOrderRepository 创建 mock 订单仓储。
func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		Orders:    make(map[uint]*biz.Order),
		OrderNo:   make(map[string]*biz.Order),
		Items:     make(map[uint][]*biz.OrderItem),
		Shipments: make(map[uint]*biz.Shipment),
		NextID:    1,
	}
}

func (m *MockOrderRepository) Create(ctx context.Context, order *biz.Order) error {
	order.ID = m.NextID
	m.NextID++
	m.Orders[order.ID] = order
	m.OrderNo[order.OrderNo] = order
	return nil
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uint) (*biz.Order, error) {
	order, ok := m.Orders[id]
	if !ok {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (m *MockOrderRepository) List(ctx context.Context, userID uint, page, size int) ([]*biz.Order, int64, error) {
	var result []*biz.Order
	for _, order := range m.Orders {
		if userID == 0 || order.UserID == userID {
			result = append(result, order)
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	order, ok := m.Orders[id]
	if !ok {
		return errors.New("order not found")
	}
	order.Status = status
	order.UpdatedAt = time.Now()
	return nil
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uint) error {
	order, ok := m.Orders[id]
	if !ok {
		return errors.New("order not found")
	}
	delete(m.Orders, id)
	delete(m.OrderNo, order.OrderNo)
	return nil
}

func (m *MockOrderRepository) GetOrderItems(ctx context.Context, orderID uint) ([]*biz.OrderItem, error) {
	items, ok := m.Items[orderID]
	if !ok {
		return []*biz.OrderItem{}, nil
	}
	return items, nil
}

func (m *MockOrderRepository) GetShipment(ctx context.Context, orderID, shipmentID uint) (*biz.Shipment, error) {
	key := orderID*1000 + shipmentID
	shipment, ok := m.Shipments[key]
	if !ok {
		return nil, errors.New("shipment not found")
	}
	return shipment, nil
}

// AddOrderItem 测试辅助方法：添加订单项。
func (m *MockOrderRepository) AddOrderItem(orderID uint, item *biz.OrderItem) {
	item.OrderID = orderID
	m.Items[orderID] = append(m.Items[orderID], item)
}

// AddShipment 测试辅助方法：添加配送信息。
func (m *MockOrderRepository) AddShipment(shipment *biz.Shipment) {
	key := shipment.OrderID*1000 + shipment.ID
	m.Shipments[key] = shipment
}

// ============================================================
// Mock 仓储实现 - 购物车仓储
// ============================================================

// MockCartRepository 购物车仓储 mock 实现。
type MockCartRepository struct {
	CartItems map[uint]*biz.ShoppingCartItem // key: item ID
	UserItems map[uint][]uint                // key: user ID, value: item IDs
	NextID    uint
}

// NewMockCartRepository 创建 mock 购物车仓储。
func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{
		CartItems: make(map[uint]*biz.ShoppingCartItem),
		UserItems: make(map[uint][]uint),
		NextID:    1,
	}
}

func (m *MockCartRepository) GetCart(ctx context.Context, userID uint) ([]*biz.ShoppingCartItem, error) {
	var result []*biz.ShoppingCartItem
	itemIDs, ok := m.UserItems[userID]
	if !ok {
		return result, nil
	}
	for _, id := range itemIDs {
		if item, exists := m.CartItems[id]; exists {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *MockCartRepository) AddCartItem(ctx context.Context, item *biz.ShoppingCartItem) error {
	item.ID = m.NextID
	m.NextID++
	m.CartItems[item.ID] = item
	m.UserItems[item.UserID] = append(m.UserItems[item.UserID], item.ID)
	return nil
}

func (m *MockCartRepository) UpdateCartItem(ctx context.Context, id uint, quantity int) error {
	item, ok := m.CartItems[id]
	if !ok {
		return errors.New("cart item not found")
	}
	item.Quantity = quantity
	item.UpdatedAt = time.Now()
	return nil
}

func (m *MockCartRepository) DeleteCartItem(ctx context.Context, id uint) error {
	item, ok := m.CartItems[id]
	if !ok {
		return errors.New("cart item not found")
	}
	// 从用户列表中移除
	userItems := m.UserItems[item.UserID]
	for i, itemID := range userItems {
		if itemID == id {
			m.UserItems[item.UserID] = append(userItems[:i], userItems[i+1:]...)
			break
		}
	}
	delete(m.CartItems, id)
	return nil
}

func (m *MockCartRepository) ClearCart(ctx context.Context, userID uint) error {
	itemIDs, ok := m.UserItems[userID]
	if ok {
		for _, id := range itemIDs {
			delete(m.CartItems, id)
		}
		delete(m.UserItems, userID)
	}
	return nil
}

// ============================================================
// Mock 仓储实现 - 愿望清单仓储
// ============================================================

// MockWishlistRepository 愿望清单仓储 mock 实现。
type MockWishlistRepository struct {
	Items     map[uint]*biz.WishlistItem // key: item ID
	UserItems map[uint][]uint            // key: user ID, value: item IDs
	NextID    uint
}

// NewMockWishlistRepository 创建 mock 愿望清单仓储。
func NewMockWishlistRepository() *MockWishlistRepository {
	return &MockWishlistRepository{
		Items:     make(map[uint]*biz.WishlistItem),
		UserItems: make(map[uint][]uint),
		NextID:    1,
	}
}

func (m *MockWishlistRepository) GetWishlist(ctx context.Context, userID uint) ([]*biz.WishlistItem, error) {
	var result []*biz.WishlistItem
	itemIDs, ok := m.UserItems[userID]
	if !ok {
		return result, nil
	}
	for _, id := range itemIDs {
		if item, exists := m.Items[id]; exists {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *MockWishlistRepository) AddWishlistItem(ctx context.Context, item *biz.WishlistItem) error {
	item.ID = m.NextID
	m.NextID++
	m.Items[item.ID] = item
	m.UserItems[item.UserID] = append(m.UserItems[item.UserID], item.ID)
	return nil
}

func (m *MockWishlistRepository) DeleteWishlistItem(ctx context.Context, id uint) error {
	item, ok := m.Items[id]
	if !ok {
		return errors.New("wishlist item not found")
	}
	// 从用户列表中移除
	userItems := m.UserItems[item.UserID]
	for i, itemID := range userItems {
		if itemID == id {
			m.UserItems[item.UserID] = append(userItems[:i], userItems[i+1:]...)
			break
		}
	}
	delete(m.Items, id)
	return nil
}

// ============================================================
// Mock 仓储实现 - 退货请求仓储
// ============================================================

// MockReturnRequestRepository 退货请求仓储 mock 实现。
type MockReturnRequestRepository struct {
	ReturnRequests map[uint]*biz.ReturnRequest
	UserRequests   map[uint][]uint // key: user ID, value: request IDs
	NextID         uint
}

// NewMockReturnRequestRepository 创建 mock 退货请求仓储。
func NewMockReturnRequestRepository() *MockReturnRequestRepository {
	return &MockReturnRequestRepository{
		ReturnRequests: make(map[uint]*biz.ReturnRequest),
		UserRequests:   make(map[uint][]uint),
		NextID:         1,
	}
}

func (m *MockReturnRequestRepository) Create(ctx context.Context, rr *biz.ReturnRequest) error {
	rr.ID = m.NextID
	m.NextID++
	m.ReturnRequests[rr.ID] = rr
	m.UserRequests[rr.UserID] = append(m.UserRequests[rr.UserID], rr.ID)
	return nil
}

func (m *MockReturnRequestRepository) List(ctx context.Context, userID uint, page, size int) ([]*biz.ReturnRequest, int64, error) {
	var result []*biz.ReturnRequest
	if userID == 0 {
		// 返回所有退货请求
		for _, rr := range m.ReturnRequests {
			result = append(result, rr)
		}
	} else {
		// 按用户过滤
		requestIDs, ok := m.UserRequests[userID]
		if ok {
			for _, id := range requestIDs {
				if rr, exists := m.ReturnRequests[id]; exists {
					result = append(result, rr)
				}
			}
		}
	}
	return result, int64(len(result)), nil
}

func (m *MockReturnRequestRepository) GetByID(ctx context.Context, id uint) (*biz.ReturnRequest, error) {
	rr, ok := m.ReturnRequests[id]
	if !ok {
		return nil, errors.New("return request not found")
	}
	return rr, nil
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestOrderUseCase 创建测试用的 OrderUseCase。
func newTestOrderUseCase() (*biz.OrderUseCase, *MockOrderRepository) {
	repo := NewMockOrderRepository()
	uc := biz.NewOrderUseCase(repo)
	return uc, repo
}

// newTestCartUseCase 创建测试用的 CartUseCase。
func newTestCartUseCase() (*biz.CartUseCase, *MockCartRepository) {
	repo := NewMockCartRepository()
	uc := biz.NewCartUseCase(repo)
	return uc, repo
}

// newTestWishlistUseCase 创建测试用的 WishlistUseCase。
func newTestWishlistUseCase() (*biz.WishlistUseCase, *MockWishlistRepository) {
	repo := NewMockWishlistRepository()
	uc := biz.NewWishlistUseCase(repo)
	return uc, repo
}

// newTestReturnRequestUseCase 创建测试用的 ReturnRequestUseCase。
func newTestReturnRequestUseCase() (*biz.ReturnRequestUseCase, *MockReturnRequestRepository) {
	repo := NewMockReturnRequestRepository()
	uc := biz.NewReturnRequestUseCase(repo)
	return uc, repo
}

// ============================================================
// 订单用例测试 - 创建订单
// ============================================================

func TestOrderCreate_Success(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{
		UserID:         1,
		SubTotal:       100.00,
		ShippingTotal:  10.00,
		TaxTotal:       5.00,
		OrderTotal:     115.00,
		CurrencyCode:   "CNY",
		ShippingMethod: "express",
		PaymentMethod:  "alipay",
	}

	created, err := uc.Create(ctx, order)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.NotEmpty(t, created.OrderNo)
	assert.Equal(t, "pending", created.Status)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
}

func TestOrderCreate_WithCoupon(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{
		UserID:        1,
		SubTotal:      100.00,
		DiscountTotal: 10.00,
		OrderTotal:    90.00,
		CouponCode:    "DISCOUNT10",
	}

	created, err := uc.Create(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, "DISCOUNT10", created.CouponCode)
}

func TestOrderCreate_WithGiftCard(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{
		UserID:       1,
		SubTotal:     100.00,
		OrderTotal:   100.00,
		GiftCardCode: "GIFTCARD50",
	}

	created, err := uc.Create(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, "GIFTCARD50", created.GiftCardCode)
}

// ============================================================
// 订单用例测试 - 获取订单
// ============================================================

func TestOrderGetByID_Success(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	// 先创建订单
	order := &biz.Order{UserID: 1, OrderTotal: 100.00}
	created, _ := uc.Create(ctx, order)

	// 获取订单
	found, err := uc.GetByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.OrderNo, found.OrderNo)
}

func TestOrderGetByID_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	// 测试获取不存在的订单
	found, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, found)
}

// ============================================================
// 订单用例测试 - 订单列表
// ============================================================

func TestOrderList_AllOrders(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	// 创建多个订单
	for i := 1; i <= 3; i++ {
		order := &biz.Order{UserID: uint(i), OrderTotal: float64(i) * 100}
		_, _ = uc.Create(ctx, order)
	}

	// 获取所有订单
	orders, total, err := uc.List(ctx, 0, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, orders, 3)
	assert.Equal(t, int64(3), total)
}

func TestOrderList_ByUserID(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	// 创建订单，属于用户1和用户2
	order1 := &biz.Order{UserID: 1, OrderTotal: 100}
	order2 := &biz.Order{UserID: 1, OrderTotal: 200}
	order3 := &biz.Order{UserID: 2, OrderTotal: 150}
	_, _ = uc.Create(ctx, order1)
	_, _ = uc.Create(ctx, order2)
	_, _ = uc.Create(ctx, order3)

	// 获取用户1的订单
	orders, total, err := uc.List(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, int64(2), total)
}

func TestOrderList_Empty(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	orders, total, err := uc.List(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, orders)
	assert.Equal(t, int64(0), total)
}

// ============================================================
// 订单用例测试 - 取消订单
// ============================================================

func TestOrderCancel_PendingStatus(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	// 默认状态是 pending

	err := uc.CancelOrder(ctx, created.ID)
	assert.NoError(t, err)

	// 验证状态已更新
	found, _ := uc.GetByID(ctx, created.ID)
	assert.Equal(t, "cancelled", found.Status)
}

func TestOrderCancel_ProcessingStatus(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	// 手动设置为 processing
	repo.UpdateStatus(ctx, created.ID, "processing")

	err := uc.CancelOrder(ctx, created.ID)
	assert.NoError(t, err)

	found, _ := uc.GetByID(ctx, created.ID)
	assert.Equal(t, "cancelled", found.Status)
}

func TestOrderCancel_CompletedStatus(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	// 设置为 completed
	repo.UpdateStatus(ctx, created.ID, "completed")

	err := uc.CancelOrder(ctx, created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法取消")
}

func TestOrderCancel_CancelledStatus(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	// 设置为 cancelled
	repo.UpdateStatus(ctx, created.ID, "cancelled")

	err := uc.CancelOrder(ctx, created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法取消")
}

func TestOrderCancel_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	err := uc.CancelOrder(ctx, 999)
	assert.Error(t, err)
}

// ============================================================
// 订单用例测试 - 退款
// ============================================================

func TestOrderRefund_ProcessingStatus(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	repo.UpdateStatus(ctx, created.ID, "processing")

	err := uc.RefundOrder(ctx, created.ID)
	assert.NoError(t, err)

	found, _ := uc.GetByID(ctx, created.ID)
	assert.Equal(t, "refunded", found.Status)
}

func TestOrderRefund_CompletedStatus(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	repo.UpdateStatus(ctx, created.ID, "completed")

	err := uc.RefundOrder(ctx, created.ID)
	assert.NoError(t, err)

	found, _ := uc.GetByID(ctx, created.ID)
	assert.Equal(t, "refunded", found.Status)
}

func TestOrderRefund_PendingStatus(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	// 默认状态是 pending

	err := uc.RefundOrder(ctx, created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法退款")
}

func TestOrderRefund_RefundedStatus(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	repo.UpdateStatus(ctx, created.ID, "refunded")

	err := uc.RefundOrder(ctx, created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法退款")
}

func TestOrderRefund_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	err := uc.RefundOrder(ctx, 999)
	assert.Error(t, err)
}

// ============================================================
// 订单用例测试 - 删除订单
// ============================================================

func TestOrderDelete_Success(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)

	err := uc.Delete(ctx, created.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = uc.GetByID(ctx, created.ID)
	assert.Error(t, err)
}

func TestOrderDelete_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	err := uc.Delete(ctx, 999)
	assert.Error(t, err)
}

// ============================================================
// 订单用例测试 - 获取订单项
// ============================================================

func TestGetOrderItems_Success(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)

	// 添加订单项
	repo.AddOrderItem(created.ID, &biz.OrderItem{
		ProductID:   1,
		ProductName: "商品A",
		SKU:         "SKU001",
		Quantity:    2,
		UnitPrice:   50.00,
		TotalPrice:  100.00,
	})
	repo.AddOrderItem(created.ID, &biz.OrderItem{
		ProductID:   2,
		ProductName: "商品B",
		SKU:         "SKU002",
		Quantity:    1,
		UnitPrice:   30.00,
		TotalPrice:  30.00,
	})

	items, err := uc.GetOrderItems(ctx, created.ID)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetOrderItems_Empty(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)

	items, err := uc.GetOrderItems(ctx, created.ID)
	assert.NoError(t, err)
	assert.Empty(t, items)
}

// ============================================================
// 订单用例测试 - 获取配送信息
// ============================================================

func TestGetShipment_Success(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)

	// 添加配送信息
	now := time.Now()
	shipment := &biz.Shipment{
		ID:             1,
		OrderID:        created.ID,
		TrackingNumber: "SF123456789",
		ShippingMethod: "顺丰快递",
		Status:         "shipped",
		ShippedAt:      &now,
	}
	repo.AddShipment(shipment)

	found, err := uc.GetShipment(ctx, created.ID, 1)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "SF123456789", found.TrackingNumber)
	assert.Equal(t, "顺丰快递", found.ShippingMethod)
}

func TestGetShipment_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	found, err := uc.GetShipment(ctx, 999, 1)
	assert.Error(t, err)
	assert.Nil(t, found)
}

// ============================================================
// 订单用例测试 - 重新下单
// ============================================================

func TestReorder_Success(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	// 创建原始订单
	original := &biz.Order{
		UserID:         1,
		SubTotal:       100.00,
		ShippingTotal:  10.00,
		TaxTotal:       5.00,
		OrderTotal:     115.00,
		CurrencyCode:   "CNY",
		ShippingMethod: "express",
		PaymentMethod:  "alipay",
	}
	created, _ := uc.Create(ctx, original)

	// 重新下单
	newOrder, err := uc.Reorder(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, newOrder)
	assert.NotZero(t, newOrder.ID)
	assert.NotEqual(t, created.ID, newOrder.ID)
	// 验证订单编号有效（格式为 ORD-时间戳）
	assert.NotEmpty(t, newOrder.OrderNo)
	assert.Contains(t, newOrder.OrderNo, "ORD-")
	assert.Equal(t, "pending", newOrder.Status)
	assert.Equal(t, original.UserID, newOrder.UserID)
	assert.Equal(t, original.SubTotal, newOrder.SubTotal)
	assert.Equal(t, original.ShippingMethod, newOrder.ShippingMethod)
}

func TestReorder_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	newOrder, err := uc.Reorder(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, newOrder)
}

// ============================================================
// 订单用例测试 - 重新提交支付
// ============================================================

func TestRePostPayment_PendingStatus(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	// 默认状态是 pending

	err := uc.RePostPayment(ctx, created.ID)
	assert.NoError(t, err)
}

func TestRePostPayment_ProcessingStatus(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)
	repo.UpdateStatus(ctx, created.ID, "processing")

	err := uc.RePostPayment(ctx, created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无法重新提交支付")
}

func TestRePostPayment_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	err := uc.RePostPayment(ctx, 999)
	assert.Error(t, err)
}

// ============================================================
// 订单用例测试 - 获取PDF发票
// ============================================================

func TestGetPDFInvoice_Success(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)

	pdf, err := uc.GetPDFInvoice(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, pdf)
	assert.Contains(t, string(pdf), "PDF invoice placeholder")
	assert.Contains(t, string(pdf), fmt.Sprintf("%d", created.ID))
}

func TestGetPDFInvoice_NotFound(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	pdf, err := uc.GetPDFInvoice(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, pdf)
}

// ============================================================
// 购物车用例测试 - 获取购物车
// ============================================================

func TestGetCart_Success(t *testing.T) {
	uc, cartRepo := newTestCartUseCase()
	ctx := context.Background()

	// 添加购物车项
	item1 := &biz.ShoppingCartItem{UserID: 1, ProductID: 1, ProductName: "商品A", Quantity: 2}
	item2 := &biz.ShoppingCartItem{UserID: 1, ProductID: 2, ProductName: "商品B", Quantity: 1}
	require.NoError(t, cartRepo.AddCartItem(ctx, item1))
	require.NoError(t, cartRepo.AddCartItem(ctx, item2))

	items, err := uc.GetCart(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetCart_Empty(t *testing.T) {
	uc, _ := newTestCartUseCase()
	ctx := context.Background()

	items, err := uc.GetCart(ctx, 1)
	assert.NoError(t, err)
	assert.Empty(t, items)
}

// ============================================================
// 购物车用例测试 - 添加到购物车
// ============================================================

func TestAddToCart_Success(t *testing.T) {
	uc, _ := newTestCartUseCase()
	ctx := context.Background()

	item := &biz.ShoppingCartItem{
		UserID:      1,
		ProductID:   1,
		ProductName: "测试商品",
		SKU:         "SKU001",
		Quantity:    2,
		UnitPrice:   99.99,
	}

	err := uc.AddToCart(ctx, item)
	assert.NoError(t, err)
	assert.NotZero(t, item.ID)
	assert.NotZero(t, item.CreatedAt)
	assert.NotZero(t, item.UpdatedAt)
}

func TestAddToCart_MultipleItems(t *testing.T) {
	uc, cartRepo := newTestCartUseCase()
	ctx := context.Background()

	// 添加多个商品
	for i := 1; i <= 3; i++ {
		item := &biz.ShoppingCartItem{
			UserID:      1,
			ProductID:   uint(i),
			ProductName: fmt.Sprintf("商品%d", i),
			Quantity:    i,
		}
		require.NoError(t, uc.AddToCart(ctx, item))
	}

	// 验证购物车中有3个商品
	items, _ := cartRepo.GetCart(ctx, 1)
	assert.Len(t, items, 3)
}

// ============================================================
// 购物车用例测试 - 更新购物车
// ============================================================

func TestUpdateCart_Success(t *testing.T) {
	uc, cartRepo := newTestCartUseCase()
	ctx := context.Background()

	item := &biz.ShoppingCartItem{UserID: 1, ProductID: 1, ProductName: "商品A", Quantity: 2}
	require.NoError(t, cartRepo.AddCartItem(ctx, item))

	err := uc.UpdateCart(ctx, item.ID, 5)
	assert.NoError(t, err)

	// 验证数量已更新
	items, _ := cartRepo.GetCart(ctx, 1)
	assert.Equal(t, 5, items[0].Quantity)
}

func TestUpdateCart_NotFound(t *testing.T) {
	uc, _ := newTestCartUseCase()
	ctx := context.Background()

	err := uc.UpdateCart(ctx, 999, 5)
	assert.Error(t, err)
}

// ============================================================
// 购物车用例测试 - 应用折扣码/礼品卡
// ============================================================

func TestApplyCoupon_Success(t *testing.T) {
	uc, _ := newTestCartUseCase()
	ctx := context.Background()

	err := uc.ApplyCoupon(ctx, 1, "DISCOUNT10")
	assert.NoError(t, err)
}

func TestApplyGiftCard_Success(t *testing.T) {
	uc, _ := newTestCartUseCase()
	ctx := context.Background()

	err := uc.ApplyGiftCard(ctx, 1, "GIFTCARD50")
	assert.NoError(t, err)
}

// ============================================================
// 购物车用例测试 - 清空购物车
// ============================================================

func TestClearCart_Success(t *testing.T) {
	uc, cartRepo := newTestCartUseCase()
	ctx := context.Background()

	// 添加购物车项
	item1 := &biz.ShoppingCartItem{UserID: 1, ProductID: 1, ProductName: "商品A", Quantity: 2}
	item2 := &biz.ShoppingCartItem{UserID: 1, ProductID: 2, ProductName: "商品B", Quantity: 1}
	require.NoError(t, cartRepo.AddCartItem(ctx, item1))
	require.NoError(t, cartRepo.AddCartItem(ctx, item2))

	err := uc.ClearCart(ctx, 1)
	assert.NoError(t, err)

	// 验证购物车已清空
	items, _ := cartRepo.GetCart(ctx, 1)
	assert.Empty(t, items)
}

func TestClearCart_Empty(t *testing.T) {
	uc, _ := newTestCartUseCase()
	ctx := context.Background()

	// 清空空购物车不应该报错
	err := uc.ClearCart(ctx, 1)
	assert.NoError(t, err)
}

// ============================================================
// 愿望清单用例测试 - 获取愿望清单
// ============================================================

func TestGetWishlist_Success(t *testing.T) {
	uc, wishlistRepo := newTestWishlistUseCase()
	ctx := context.Background()

	// 添加愿望清单项
	item1 := &biz.WishlistItem{UserID: 1, ProductID: 1, ProductName: "商品A"}
	item2 := &biz.WishlistItem{UserID: 1, ProductID: 2, ProductName: "商品B"}
	require.NoError(t, wishlistRepo.AddWishlistItem(ctx, item1))
	require.NoError(t, wishlistRepo.AddWishlistItem(ctx, item2))

	items, err := uc.GetWishlist(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestGetWishlist_Empty(t *testing.T) {
	uc, _ := newTestWishlistUseCase()
	ctx := context.Background()

	items, err := uc.GetWishlist(ctx, 1)
	assert.NoError(t, err)
	assert.Empty(t, items)
}

// ============================================================
// 愿望清单用例测试 - 添加到愿望清单
// ============================================================

func TestAddToWishlist_Success(t *testing.T) {
	uc, _ := newTestWishlistUseCase()
	ctx := context.Background()

	item := &biz.WishlistItem{
		UserID:      1,
		ProductID:   1,
		ProductName: "测试商品",
	}

	err := uc.AddToWishlist(ctx, item)
	assert.NoError(t, err)
	assert.NotZero(t, item.ID)
	assert.NotZero(t, item.CreatedAt)
}

func TestAddToWishlist_MultipleItems(t *testing.T) {
	uc, wishlistRepo := newTestWishlistUseCase()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		item := &biz.WishlistItem{
			UserID:      1,
			ProductID:   uint(i),
			ProductName: fmt.Sprintf("商品%d", i),
		}
		require.NoError(t, uc.AddToWishlist(ctx, item))
	}

	items, _ := wishlistRepo.GetWishlist(ctx, 1)
	assert.Len(t, items, 3)
}

// ============================================================
// 退货请求用例测试 - 创建退货请求
// ============================================================

func TestReturnRequestCreate_Success(t *testing.T) {
	uc, _ := newTestReturnRequestUseCase()
	ctx := context.Background()

	rr := &biz.ReturnRequest{
		OrderID: 1,
		UserID:  1,
		Reason:  "商品有质量问题",
	}

	created, err := uc.Create(ctx, rr)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "pending", created.Status)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
}

func TestReturnRequestCreate_WithRefundTotal(t *testing.T) {
	uc, _ := newTestReturnRequestUseCase()
	ctx := context.Background()

	rr := &biz.ReturnRequest{
		OrderID:     1,
		UserID:      1,
		Reason:      "不想要了",
		RefundTotal: 100.00,
	}

	created, err := uc.Create(ctx, rr)
	assert.NoError(t, err)
	assert.Equal(t, 100.00, created.RefundTotal)
}

// ============================================================
// 退货请求用例测试 - 获取退货请求列表
// ============================================================

func TestReturnRequestList_ByUserID(t *testing.T) {
	uc, _ := newTestReturnRequestUseCase()
	ctx := context.Background()

	// 创建退货请求
	rr1 := &biz.ReturnRequest{OrderID: 1, UserID: 1, Reason: "原因1"}
	rr2 := &biz.ReturnRequest{OrderID: 2, UserID: 1, Reason: "原因2"}
	rr3 := &biz.ReturnRequest{OrderID: 3, UserID: 2, Reason: "原因3"}
	_, _ = uc.Create(ctx, rr1)
	_, _ = uc.Create(ctx, rr2)
	_, _ = uc.Create(ctx, rr3)

	// 获取用户1的退货请求
	requests, total, err := uc.List(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, requests, 2)
	assert.Equal(t, int64(2), total)
}

func TestReturnRequestList_All(t *testing.T) {
	uc, _ := newTestReturnRequestUseCase()
	ctx := context.Background()

	// 创建退货请求
	rr1 := &biz.ReturnRequest{OrderID: 1, UserID: 1, Reason: "原因1"}
	rr2 := &biz.ReturnRequest{OrderID: 2, UserID: 2, Reason: "原因2"}
	_, _ = uc.Create(ctx, rr1)
	_, _ = uc.Create(ctx, rr2)

	// 获取所有退货请求（userID=0）
	requests, total, err := uc.List(ctx, 0, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, requests, 2)
	assert.Equal(t, int64(2), total)
}

func TestReturnRequestList_Empty(t *testing.T) {
	uc, _ := newTestReturnRequestUseCase()
	ctx := context.Background()

	requests, total, err := uc.List(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, requests)
	assert.Equal(t, int64(0), total)
}

// ============================================================
// 边界条件测试
// ============================================================

func TestOrderCreate_WithAddresses(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	billingAddrID := uint(1)
	shippingAddrID := uint(2)

	order := &biz.Order{
		UserID:         1,
		OrderTotal:     100.00,
		BillingAddrID:  &billingAddrID,
		ShippingAddrID: &shippingAddrID,
	}

	created, err := uc.Create(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, billingAddrID, *created.BillingAddrID)
	assert.Equal(t, shippingAddrID, *created.ShippingAddrID)
}

func TestOrderCreate_WithCustomerIP(t *testing.T) {
	uc, _ := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{
		UserID:     1,
		OrderTotal: 100.00,
		CustomerIP: "192.168.1.1",
	}

	created, err := uc.Create(ctx, order)
	assert.NoError(t, err)
	assert.Equal(t, "192.168.1.1", created.CustomerIP)
}

func TestShipment_Delivered(t *testing.T) {
	uc, repo := newTestOrderUseCase()
	ctx := context.Background()

	order := &biz.Order{UserID: 1, OrderTotal: 100}
	created, _ := uc.Create(ctx, order)

	now := time.Now()
	shipment := &biz.Shipment{
		ID:          1,
		OrderID:     created.ID,
		Status:      "delivered",
		DeliveredAt: &now,
	}
	repo.AddShipment(shipment)

	found, err := uc.GetShipment(ctx, created.ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, "delivered", found.Status)
	assert.NotNil(t, found.DeliveredAt)
}

func TestCartUpdate_QuantityZero(t *testing.T) {
	uc, cartRepo := newTestCartUseCase()
	ctx := context.Background()

	item := &biz.ShoppingCartItem{UserID: 1, ProductID: 1, ProductName: "商品A", Quantity: 2}
	require.NoError(t, cartRepo.AddCartItem(ctx, item))

	// 更新数量为0（可以视为删除，但具体行为由业务决定）
	err := uc.UpdateCart(ctx, item.ID, 0)
	assert.NoError(t, err)

	items, _ := cartRepo.GetCart(ctx, 1)
	assert.Equal(t, 0, items[0].Quantity)
}
