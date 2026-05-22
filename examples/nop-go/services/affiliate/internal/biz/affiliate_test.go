// Package biz_test 联盟服务业务逻辑层单元测试。
//
// 使用 mock 仓储隔离依赖，测试 AffiliateUseCase 的核心业务逻辑。
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nop-go/services/affiliate/internal/biz"
)

// ============================================================
// Mock 仓储实现
// ============================================================

// MockAffiliateRepository 联盟仓储 mock 实现。
type MockAffiliateRepository struct {
	Affiliates map[uint]*biz.Affiliate
	NextID     uint
}

// NewMockAffiliateRepository 创建 mock 联盟仓储。
func NewMockAffiliateRepository() *MockAffiliateRepository {
	return &MockAffiliateRepository{
		Affiliates: make(map[uint]*biz.Affiliate),
		NextID:     1,
	}
}

func (m *MockAffiliateRepository) Create(ctx context.Context, aff *biz.Affiliate) error {
	aff.ID = m.NextID
	m.NextID++
	m.Affiliates[aff.ID] = aff
	return nil
}

func (m *MockAffiliateRepository) GetByID(ctx context.Context, id uint) (*biz.Affiliate, error) {
	aff, ok := m.Affiliates[id]
	if !ok {
		return nil, errors.New("affiliate not found")
	}
	return aff, nil
}

func (m *MockAffiliateRepository) List(ctx context.Context, page, size int) ([]*biz.Affiliate, int64, error) {
	var result []*biz.Affiliate
	for _, aff := range m.Affiliates {
		result = append(result, aff)
	}
	return result, int64(len(result)), nil
}

func (m *MockAffiliateRepository) Update(ctx context.Context, aff *biz.Affiliate) error {
	m.Affiliates[aff.ID] = aff
	return nil
}

func (m *MockAffiliateRepository) Delete(ctx context.Context, id uint) error {
	delete(m.Affiliates, id)
	return nil
}

// MockAffiliateOrderRepository 联盟订单仓储 mock 实现。
type MockAffiliateOrderRepository struct {
	Orders map[uint]*biz.AffiliateOrder
	NextID uint
}

// NewMockAffiliateOrderRepository 创建 mock 联盟订单仓储。
func NewMockAffiliateOrderRepository() *MockAffiliateOrderRepository {
	return &MockAffiliateOrderRepository{
		Orders: make(map[uint]*biz.AffiliateOrder),
		NextID: 1,
	}
}

func (m *MockAffiliateOrderRepository) ListByAffiliateID(ctx context.Context, affiliateID uint, page, size int) ([]*biz.AffiliateOrder, int64, error) {
	var result []*biz.AffiliateOrder
	for _, order := range m.Orders {
		if order.AffiliateID == affiliateID {
			result = append(result, order)
		}
	}
	return result, int64(len(result)), nil
}

// MockAffiliateCustomerRepository 联盟客户仓储 mock 实现。
type MockAffiliateCustomerRepository struct {
	Customers map[uint]*biz.AffiliateCustomer
	NextID    uint
}

// NewMockAffiliateCustomerRepository 创建 mock 联盟客户仓储。
func NewMockAffiliateCustomerRepository() *MockAffiliateCustomerRepository {
	return &MockAffiliateCustomerRepository{
		Customers: make(map[uint]*biz.AffiliateCustomer),
		NextID:    1,
	}
}

func (m *MockAffiliateCustomerRepository) ListByAffiliateID(ctx context.Context, affiliateID uint, page, size int) ([]*biz.AffiliateCustomer, int64, error) {
	var result []*biz.AffiliateCustomer
	for _, customer := range m.Customers {
		if customer.AffiliateID == affiliateID {
			result = append(result, customer)
		}
	}
	return result, int64(len(result)), nil
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestUseCase 创建测试用的 AffiliateUseCase。
func newTestUseCase() (*biz.AffiliateUseCase, *MockAffiliateRepository, *MockAffiliateOrderRepository, *MockAffiliateCustomerRepository) {
	affiliateRepo := NewMockAffiliateRepository()
	orderRepo := NewMockAffiliateOrderRepository()
	customerRepo := NewMockAffiliateCustomerRepository()

	uc := biz.NewAffiliateUseCase(affiliateRepo, orderRepo, customerRepo)
	return uc, affiliateRepo, orderRepo, customerRepo
}

// ============================================================
// 联盟 CRUD 测试
// ============================================================

func TestCreateAffiliate_Success(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	aff, err := uc.Create(ctx, "测试联盟", "https://example.com/affiliate", true)
	assert.NoError(t, err)
	assert.NotNil(t, aff)
	assert.Equal(t, "测试联盟", aff.Name)
	assert.Equal(t, "https://example.com/affiliate", aff.Url)
	assert.True(t, aff.Active)
	assert.NotZero(t, aff.ID)
}

func TestGetAffiliateByID_Success(t *testing.T) {
	uc, repo, _, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个联盟
	testAff := &biz.Affiliate{
		Name:   "Test Affiliate",
		Url:    "https://test.com",
		Active: true,
	}
	require.NoError(t, repo.Create(ctx, testAff))

	// 获取联盟
	aff, err := uc.GetByID(ctx, testAff.ID)
	assert.NoError(t, err)
	assert.NotNil(t, aff)
	assert.Equal(t, "Test Affiliate", aff.Name)
}

func TestGetAffiliateByID_NotFound(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	aff, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, aff)
	assert.Contains(t, err.Error(), "not found")
}

func TestListAffiliates_Success(t *testing.T) {
	uc, repo, _, _ := newTestUseCase()
	ctx := context.Background()

	// 创建多个联盟
	aff1 := &biz.Affiliate{Name: "联盟1", Url: "https://aff1.com", Active: true}
	aff2 := &biz.Affiliate{Name: "联盟2", Url: "https://aff2.com", Active: true}
	aff3 := &biz.Affiliate{Name: "联盟3", Url: "https://aff3.com", Active: false}
	require.NoError(t, repo.Create(ctx, aff1))
	require.NoError(t, repo.Create(ctx, aff2))
	require.NoError(t, repo.Create(ctx, aff3))

	// 获取列表
	affiliates, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, affiliates, 3)
	assert.Equal(t, int64(3), total)
}

func TestListAffiliates_Empty(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	affiliates, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, affiliates)
	assert.Equal(t, int64(0), total)
}

func TestUpdateAffiliate_Success(t *testing.T) {
	uc, repo, _, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个联盟
	testAff := &biz.Affiliate{
		Name:   "Test Affiliate",
		Url:    "https://test.com",
		Active: true,
	}
	require.NoError(t, repo.Create(ctx, testAff))

	// 更新联盟
	updated, err := uc.Update(ctx, testAff.ID, "更新后的联盟", "https://updated.com", false)
	assert.NoError(t, err)
	assert.NotNil(t, updated)
	assert.Equal(t, "更新后的联盟", updated.Name)
	assert.Equal(t, "https://updated.com", updated.Url)
	assert.False(t, updated.Active)
}

func TestUpdateAffiliate_NotFound(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	updated, err := uc.Update(ctx, 999, "Test", "https://test.com", true)
	assert.Error(t, err)
	assert.Nil(t, updated)
}

func TestDeleteAffiliate_Success(t *testing.T) {
	uc, repo, _, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个联盟
	testAff := &biz.Affiliate{
		Name:   "Test Affiliate",
		Url:    "https://test.com",
		Active: true,
	}
	require.NoError(t, repo.Create(ctx, testAff))

	// 删除联盟
	err := uc.Delete(ctx, testAff.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = repo.GetByID(ctx, testAff.ID)
	assert.Error(t, err)
}

func TestDeleteAffiliate_NotFound(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 删除不存在的联盟（mock 实现返回 nil）
	err := uc.Delete(ctx, 999)
	assert.NoError(t, err)
}

// ============================================================
// 联盟订单测试
// ============================================================

func TestListOrders_Success(t *testing.T) {
	uc, _, orderRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 创建多个订单
	order1 := &biz.AffiliateOrder{
		AffiliateID: 1,
		OrderNo:     "ORD001",
		CustomerID:  100,
		TotalAmount: 100.00,
		Status:      "completed",
	}
	order2 := &biz.AffiliateOrder{
		AffiliateID: 1,
		OrderNo:     "ORD002",
		CustomerID:  101,
		TotalAmount: 200.00,
		Status:      "pending",
	}
	order3 := &biz.AffiliateOrder{
		AffiliateID: 2,
		OrderNo:     "ORD003",
		CustomerID:  102,
		TotalAmount: 300.00,
		Status:      "completed",
	}
	order1.ID = 1
	order2.ID = 2
	order3.ID = 3
	orderRepo.Orders[1] = order1
	orderRepo.Orders[2] = order2
	orderRepo.Orders[3] = order3

	// 获取联盟1的订单
	orders, total, err := uc.ListOrders(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, orders, 2)
	assert.Equal(t, int64(2), total)
}

func TestListOrders_Empty(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	orders, total, err := uc.ListOrders(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, orders)
	assert.Equal(t, int64(0), total)
}

func TestListOrders_MultipleAffiliates(t *testing.T) {
	uc, _, orderRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 为不同联盟创建订单
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 5; j++ {
			order := &biz.AffiliateOrder{
				ID:          uint(i*10 + j),
				AffiliateID: uint(i),
				OrderNo:     "ORD" + string(rune(i*10+j)),
				TotalAmount: float64(j * 100),
			}
			orderRepo.Orders[order.ID] = order
		}
	}

	// 获取联盟2的订单
	orders, total, err := uc.ListOrders(ctx, 2, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, orders, 5)
	assert.Equal(t, int64(5), total)
}

// ============================================================
// 联盟客户测试
// ============================================================

func TestListCustomers_Success(t *testing.T) {
	uc, _, _, customerRepo := newTestUseCase()
	ctx := context.Background()

	// 创建多个客户
	customer1 := &biz.AffiliateCustomer{
		ID:          1,
		AffiliateID: 1,
		Username:    "user1",
		Email:       "user1@example.com",
	}
	customer2 := &biz.AffiliateCustomer{
		ID:          2,
		AffiliateID: 1,
		Username:    "user2",
		Email:       "user2@example.com",
	}
	customer3 := &biz.AffiliateCustomer{
		ID:          3,
		AffiliateID: 2,
		Username:    "user3",
		Email:       "user3@example.com",
	}
	customerRepo.Customers[1] = customer1
	customerRepo.Customers[2] = customer2
	customerRepo.Customers[3] = customer3

	// 获取联盟1的客户
	customers, total, err := uc.ListCustomers(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, customers, 2)
	assert.Equal(t, int64(2), total)
}

func TestListCustomers_Empty(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	customers, total, err := uc.ListCustomers(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, customers)
	assert.Equal(t, int64(0), total)
}

func TestListCustomers_MultipleAffiliates(t *testing.T) {
	uc, _, _, customerRepo := newTestUseCase()
	ctx := context.Background()

	// 为不同联盟创建客户
	for i := 1; i <= 3; i++ {
		for j := 1; j <= 5; j++ {
			customer := &biz.AffiliateCustomer{
				ID:          uint(i*10 + j),
				AffiliateID: uint(i),
				Username:    "user" + string(rune(i*10+j)),
				Email:       "user" + string(rune(i*10+j)) + "@example.com",
			}
			customerRepo.Customers[customer.ID] = customer
		}
	}

	// 获取联盟3的客户
	customers, total, err := uc.ListCustomers(ctx, 3, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, customers, 5)
	assert.Equal(t, int64(5), total)
}

// ============================================================
// 时间戳测试
// ============================================================

func TestAffiliate_Timestamps(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	beforeCreate := time.Now()

	// 创建联盟
	aff, err := uc.Create(ctx, "测试联盟", "https://test.com", true)
	assert.NoError(t, err)

	// 验证时间戳已设置
	assert.False(t, aff.CreatedAt.IsZero())
	assert.False(t, aff.UpdatedAt.IsZero())
	assert.True(t, aff.CreatedAt.After(beforeCreate) || aff.CreatedAt.Equal(beforeCreate))

	// 更新联盟
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now()
	updated, err := uc.Update(ctx, aff.ID, "更新后的联盟", "https://updated.com", false)
	assert.NoError(t, err)

	// 验证更新时间已更新
	assert.True(t, updated.UpdatedAt.After(beforeUpdate) || updated.UpdatedAt.Equal(beforeUpdate))
}

func TestAffiliateOrder_Timestamp(t *testing.T) {
	// 验证订单实体时间戳
	order := &biz.AffiliateOrder{
		AffiliateID: 1,
		OrderNo:     "ORD001",
		CustomerID:  100,
		TotalAmount: 100.00,
		Status:      "completed",
		CreatedAt:   time.Now(),
	}
	assert.False(t, order.CreatedAt.IsZero())
}

func TestAffiliateCustomer_Timestamp(t *testing.T) {
	// 验证客户实体时间戳
	customer := &biz.AffiliateCustomer{
		ID:          1,
		AffiliateID: 1,
		Username:    "testuser",
		Email:       "test@example.com",
		CreatedAt:   time.Now(),
	}
	assert.False(t, customer.CreatedAt.IsZero())
}

// ============================================================
// 边界条件测试
// ============================================================

func TestCreateAffiliate_EmptyFields(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 测试空字段
	aff, err := uc.Create(ctx, "", "", false)
	assert.NoError(t, err)
	assert.NotNil(t, aff)
	assert.Equal(t, "", aff.Name)
	assert.Equal(t, "", aff.Url)
	assert.False(t, aff.Active)
}

func TestCreateAffiliate_Inactive(t *testing.T) {
	uc, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 测试创建未激活的联盟
	aff, err := uc.Create(ctx, "未激活联盟", "https://inactive.com", false)
	assert.NoError(t, err)
	assert.False(t, aff.Active)
}

func TestUpdateAffiliate_ActiveToggle(t *testing.T) {
	uc, repo, _, _ := newTestUseCase()
	ctx := context.Background()

	// 先创建一个激活的联盟
	testAff := &biz.Affiliate{
		Name:   "Test Affiliate",
		Url:    "https://test.com",
		Active: true,
	}
	require.NoError(t, repo.Create(ctx, testAff))

	// 停用联盟
	updated, err := uc.Update(ctx, testAff.ID, testAff.Name, testAff.Url, false)
	assert.NoError(t, err)
	assert.False(t, updated.Active)

	// 再次激活
	updated, err = uc.Update(ctx, testAff.ID, testAff.Name, testAff.Url, true)
	assert.NoError(t, err)
	assert.True(t, updated.Active)
}

func TestListOrders_Pagination(t *testing.T) {
	uc, _, orderRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 创建大量订单
	for i := 1; i <= 50; i++ {
		order := &biz.AffiliateOrder{
			ID:          uint(i),
			AffiliateID: 1,
			OrderNo:     "ORD" + string(rune(i)),
			TotalAmount: float64(i * 10),
		}
		orderRepo.Orders[order.ID] = order
	}

	// 获取第一页
	orders, total, err := uc.ListOrders(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(50), total)
	// mock 实现返回所有记录，不做分页
	assert.Len(t, orders, 50)
}

func TestListCustomers_Pagination(t *testing.T) {
	uc, _, _, customerRepo := newTestUseCase()
	ctx := context.Background()

	// 创建大量客户
	for i := 1; i <= 50; i++ {
		customer := &biz.AffiliateCustomer{
			ID:          uint(i),
			AffiliateID: 1,
			Username:    "user" + string(rune(i)),
			Email:       "user" + string(rune(i)) + "@example.com",
		}
		customerRepo.Customers[customer.ID] = customer
	}

	// 获取客户列表
	customers, total, err := uc.ListCustomers(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(50), total)
	// mock 实现返回所有记录，不做分页
	assert.Len(t, customers, 50)
}

func TestAffiliateOrder_DifferentStatuses(t *testing.T) {
	uc, _, orderRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 创建不同状态的订单
	statuses := []string{"pending", "processing", "completed", "cancelled", "refunded"}
	for i, status := range statuses {
		order := &biz.AffiliateOrder{
			ID:          uint(i + 1),
			AffiliateID: 1,
			OrderNo:     "ORD" + string(rune(i+1)),
			Status:      status,
		}
		orderRepo.Orders[order.ID] = order
	}

	// 获取订单
	orders, _, err := uc.ListOrders(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, orders, 5)
}

func TestAffiliateOrder_DifferentAmounts(t *testing.T) {
	uc, _, orderRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 创建不同金额的订单
	amounts := []float64{0.01, 10.50, 100.00, 1000.00, 10000.99}
	for i, amount := range amounts {
		order := &biz.AffiliateOrder{
			ID:          uint(i + 1),
			AffiliateID: 1,
			OrderNo:     "ORD" + string(rune(i+1)),
			TotalAmount: amount,
		}
		orderRepo.Orders[order.ID] = order
	}

	// 获取订单
	orders, _, err := uc.ListOrders(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, orders, 5)
}
