// Package biz_test 折扣服务业务逻辑层单元测试。
//
// 使用 mock 仓储隔离依赖，测试 DiscountUsecase 的核心业务逻辑。
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nop-go/services/discount/internal/biz"
	"nop-go/services/discount/internal/server/http/request"
)

// ============================================================
// Mock 仓储实现
// ============================================================

// MockDiscountRepository 折扣仓储 mock 实现。
type MockDiscountRepository struct {
	Discounts map[uint]*biz.Discount
	NextID    uint
	// 用于模拟错误场景
	ReturnError bool
}

// NewMockDiscountRepository 创建 mock 折扣仓储。
func NewMockDiscountRepository() *MockDiscountRepository {
	return &MockDiscountRepository{
		Discounts: make(map[uint]*biz.Discount),
		NextID:    1,
	}
}

func (m *MockDiscountRepository) List(ctx context.Context, page, pageSize int) ([]*biz.Discount, int64, error) {
	if m.ReturnError {
		return nil, 0, errors.New("database error")
	}

	var result []*biz.Discount
	for _, d := range m.Discounts {
		result = append(result, d)
	}
	return result, int64(len(result)), nil
}

func (m *MockDiscountRepository) Create(ctx context.Context, discount *biz.Discount) (*biz.Discount, error) {
	if m.ReturnError {
		return nil, errors.New("database error")
	}

	discount.ID = m.NextID
	m.NextID++
	m.Discounts[discount.ID] = discount
	return discount, nil
}

func (m *MockDiscountRepository) Update(ctx context.Context, discount *biz.Discount) (*biz.Discount, error) {
	if m.ReturnError {
		return nil, errors.New("database error")
	}

	m.Discounts[discount.ID] = discount
	return discount, nil
}

func (m *MockDiscountRepository) Delete(ctx context.Context, id uint) error {
	if m.ReturnError {
		return errors.New("database error")
	}

	delete(m.Discounts, id)
	return nil
}

// MockDiscountProductRepository 折扣关联商品仓储 mock 实现。
type MockDiscountProductRepository struct {
	Products    map[uint]*biz.DiscountProduct
	NextID      uint
	ReturnError bool
}

// NewMockDiscountProductRepository 创建 mock 折扣关联商品仓储。
func NewMockDiscountProductRepository() *MockDiscountProductRepository {
	return &MockDiscountProductRepository{
		Products: make(map[uint]*biz.DiscountProduct),
		NextID:   1,
	}
}

func (m *MockDiscountProductRepository) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountProduct, int64, error) {
	if m.ReturnError {
		return nil, 0, errors.New("database error")
	}

	var result []*biz.DiscountProduct
	for _, p := range m.Products {
		if p.DiscountID == discountID {
			result = append(result, p)
		}
	}
	return result, int64(len(result)), nil
}

// MockDiscountCategoryRepository 折扣关联分类仓储 mock 实现。
type MockDiscountCategoryRepository struct {
	Categories  map[uint]*biz.DiscountCategory
	NextID      uint
	ReturnError bool
}

// NewMockDiscountCategoryRepository 创建 mock 折扣关联分类仓储。
func NewMockDiscountCategoryRepository() *MockDiscountCategoryRepository {
	return &MockDiscountCategoryRepository{
		Categories: make(map[uint]*biz.DiscountCategory),
		NextID:     1,
	}
}

func (m *MockDiscountCategoryRepository) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountCategory, int64, error) {
	if m.ReturnError {
		return nil, 0, errors.New("database error")
	}

	var result []*biz.DiscountCategory
	for _, c := range m.Categories {
		if c.DiscountID == discountID {
			result = append(result, c)
		}
	}
	return result, int64(len(result)), nil
}

// MockDiscountManufacturerRepository 折扣关联制造商仓储 mock 实现。
type MockDiscountManufacturerRepository struct {
	Manufacturers map[uint]*biz.DiscountManufacturer
	NextID        uint
	ReturnError   bool
}

// NewMockDiscountManufacturerRepository 创建 mock 折扣关联制造商仓储。
func NewMockDiscountManufacturerRepository() *MockDiscountManufacturerRepository {
	return &MockDiscountManufacturerRepository{
		Manufacturers: make(map[uint]*biz.DiscountManufacturer),
		NextID:        1,
	}
}

func (m *MockDiscountManufacturerRepository) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountManufacturer, int64, error) {
	if m.ReturnError {
		return nil, 0, errors.New("database error")
	}

	var result []*biz.DiscountManufacturer
	for _, mfr := range m.Manufacturers {
		if mfr.DiscountID == discountID {
			result = append(result, mfr)
		}
	}
	return result, int64(len(result)), nil
}

// MockDiscountUsageHistoryRepository 折扣使用历史仓储 mock 实现。
type MockDiscountUsageHistoryRepository struct {
	History     map[uint]*biz.DiscountUsageHistory
	NextID      uint
	ReturnError bool
}

// NewMockDiscountUsageHistoryRepository 创建 mock 折扣使用历史仓储。
func NewMockDiscountUsageHistoryRepository() *MockDiscountUsageHistoryRepository {
	return &MockDiscountUsageHistoryRepository{
		History: make(map[uint]*biz.DiscountUsageHistory),
		NextID:  1,
	}
}

func (m *MockDiscountUsageHistoryRepository) ListByDiscountID(ctx context.Context, discountID uint, page, pageSize int) ([]*biz.DiscountUsageHistory, int64, error) {
	if m.ReturnError {
		return nil, 0, errors.New("database error")
	}

	var result []*biz.DiscountUsageHistory
	for _, h := range m.History {
		if h.DiscountID == discountID {
			result = append(result, h)
		}
	}
	return result, int64(len(result)), nil
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestUseCase 创建测试用的 DiscountUsecase 及其 mock 仓储。
func newTestUseCase() (*biz.DiscountUsecase, *MockDiscountRepository, *MockDiscountProductRepository, *MockDiscountCategoryRepository, *MockDiscountManufacturerRepository, *MockDiscountUsageHistoryRepository) {
	discountRepo := NewMockDiscountRepository()
	productRepo := NewMockDiscountProductRepository()
	categoryRepo := NewMockDiscountCategoryRepository()
	manufacturerRepo := NewMockDiscountManufacturerRepository()
	usageHistoryRepo := NewMockDiscountUsageHistoryRepository()

	uc := biz.NewDiscountUsecase(discountRepo, productRepo, categoryRepo, manufacturerRepo, usageHistoryRepo)
	return uc, discountRepo, productRepo, categoryRepo, manufacturerRepo, usageHistoryRepo
}

// ============================================================
// 折扣列表测试
// ============================================================

func TestListDiscounts_Success(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	discountRepo.Discounts[1] = &biz.Discount{
		ID:              1,
		Name:            "夏季大促",
		DiscountType:    "percentage",
		DiscountAmount:  20.0,
		IsActive:        true,
		StartDate:       time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:         time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	discountRepo.Discounts[2] = &biz.Discount{
		ID:              2,
		Name:            "新用户专享",
		DiscountType:    "fixed",
		DiscountAmount:  50.0,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	items, total, err := uc.ListDiscounts(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// 验证响应数据
	assert.Equal(t, uint(1), items[0].ID)
	assert.Equal(t, "夏季大促", items[0].Name)
	assert.Equal(t, "percentage", items[0].DiscountType)
	assert.Equal(t, 20.0, items[0].DiscountAmount)
	assert.Equal(t, "2024-06-01", items[0].StartDate)
	assert.Equal(t, "2024-06-30", items[0].EndDate)
}

func TestListDiscounts_EmptyList(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	items, total, err := uc.ListDiscounts(ctx, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
}

func TestListDiscounts_RepositoryError(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 设置仓储返回错误
	discountRepo.ReturnError = true

	items, total, err := uc.ListDiscounts(ctx, 1, 10)
	require.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Nil(t, items)
	assert.Equal(t, int64(0), total)
}

// ============================================================
// 创建折扣测试
// ============================================================

func TestCreateDiscount_Success(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	req := request.CreateDiscountRequest{
		Name:              "双11大促",
		DiscountType:      "percentage",
		DiscountAmount:    30.0,
		StartDate:         "2024-11-01",
		EndDate:           "2024-11-11",
		RequiresCouponCode: false,
		IsCumulative:      true,
		DisplayOrder:      1,
		IsActive:          true,
		LimitationTimes:   1000,
	}

	resp, err := uc.CreateDiscount(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "双11大促", resp.Name)
	assert.Equal(t, "percentage", resp.DiscountType)
	assert.Equal(t, 30.0, resp.DiscountAmount)
	assert.Equal(t, "2024-11-01", resp.StartDate)
	assert.Equal(t, "2024-11-11", resp.EndDate)
	assert.True(t, resp.IsCumulative)
	assert.True(t, resp.IsActive)
	assert.Equal(t, 1000, resp.LimitationTimes)
	assert.NotZero(t, resp.ID)
}

func TestCreateDiscount_WithCouponCode(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	req := request.CreateDiscountRequest{
		Name:               "优惠券折扣",
		DiscountType:       "fixed",
		DiscountAmount:     100.0,
		RequiresCouponCode: true,
		CouponCode:         "SAVE100",
		IsActive:           true,
	}

	resp, err := uc.CreateDiscount(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, resp.RequiresCouponCode)
	assert.Equal(t, "SAVE100", resp.CouponCode)
}

func TestCreateDiscount_FreeShipping(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	req := request.CreateDiscountRequest{
		Name:           "免运费优惠",
		DiscountType:   "free_shipping",
		DiscountAmount: 0,
		IsActive:       true,
	}

	resp, err := uc.CreateDiscount(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "free_shipping", resp.DiscountType)
	assert.Equal(t, 0.0, resp.DiscountAmount)
}

func TestCreateDiscount_EmptyDates(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	req := request.CreateDiscountRequest{
		Name:           "无日期限制折扣",
		DiscountType:   "percentage",
		DiscountAmount: 10.0,
		// 不设置 StartDate 和 EndDate
	}

	resp, err := uc.CreateDiscount(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.StartDate)
	assert.Empty(t, resp.EndDate)
}

func TestCreateDiscount_RepositoryError(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	discountRepo.ReturnError = true

	req := request.CreateDiscountRequest{
		Name:           "测试折扣",
		DiscountType:   "percentage",
		DiscountAmount: 10.0,
	}

	resp, err := uc.CreateDiscount(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
}

// ============================================================
// 更新折扣测试
// ============================================================

func TestUpdateDiscount_Success(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	existing := &biz.Discount{
		ID:              1,
		Name:            "原始折扣",
		DiscountType:    "percentage",
		DiscountAmount:  10.0,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	discountRepo.Discounts[1] = existing

	req := request.UpdateDiscountRequest{
		ID:              1,
		Name:            "更新后的折扣",
		DiscountType:    "fixed",
		DiscountAmount:  50.0,
		StartDate:       "2024-12-01",
		EndDate:         "2024-12-31",
		IsActive:        true,
		DisplayOrder:    5,
		LimitationTimes: 500,
	}

	resp, err := uc.UpdateDiscount(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "更新后的折扣", resp.Name)
	assert.Equal(t, "fixed", resp.DiscountType)
	assert.Equal(t, 50.0, resp.DiscountAmount)
	assert.Equal(t, "2024-12-01", resp.StartDate)
	assert.Equal(t, "2024-12-31", resp.EndDate)
	assert.Equal(t, 5, resp.DisplayOrder)
	assert.Equal(t, 500, resp.LimitationTimes)
}

func TestUpdateDiscount_ChangeCouponCode(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	existing := &biz.Discount{
		ID:                1,
		Name:              "优惠券折扣",
		DiscountType:      "fixed",
		DiscountAmount:    50.0,
		RequiresCouponCode: true,
		CouponCode:        "OLD_CODE",
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	discountRepo.Discounts[1] = existing

	req := request.UpdateDiscountRequest{
		ID:                1,
		Name:              "优惠券折扣",
		DiscountType:      "fixed",
		DiscountAmount:    50.0,
		RequiresCouponCode: true,
		CouponCode:        "NEW_CODE",
		IsActive:          true,
	}

	resp, err := uc.UpdateDiscount(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "NEW_CODE", resp.CouponCode)
}

func TestUpdateDiscount_RepositoryError(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	discountRepo.ReturnError = true

	req := request.UpdateDiscountRequest{
		ID:              1,
		Name:            "更新折扣",
		DiscountType:    "percentage",
		DiscountAmount:  15.0,
	}

	resp, err := uc.UpdateDiscount(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
}

// ============================================================
// 删除折扣测试
// ============================================================

func TestDeleteDiscount_Success(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	discountRepo.Discounts[1] = &biz.Discount{
		ID:        1,
		Name:      "待删除折扣",
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := uc.DeleteDiscount(ctx, 1)
	require.NoError(t, err)

	// 验证已删除
	_, exists := discountRepo.Discounts[1]
	assert.False(t, exists)
}

func TestDeleteDiscount_NotExist(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 删除不存在的折扣，mock 实现不返回错误
	err := uc.DeleteDiscount(ctx, 999)
	assert.NoError(t, err)
}

func TestDeleteDiscount_RepositoryError(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	discountRepo.ReturnError = true

	err := uc.DeleteDiscount(ctx, 1)
	require.Error(t, err)
	assert.Equal(t, "database error", err.Error())
}

// ============================================================
// 折扣关联商品列表测试
// ============================================================

func TestListDiscountProducts_Success(t *testing.T) {
	uc, _, productRepo, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	productRepo.Products[1] = &biz.DiscountProduct{
		ID:          1,
		DiscountID:  100,
		ProductID:   1,
		ProductName: "iPhone 15 Pro",
		CreatedAt:   now,
	}
	productRepo.Products[2] = &biz.DiscountProduct{
		ID:          2,
		DiscountID:  100,
		ProductID:   2,
		ProductName: "MacBook Pro",
		CreatedAt:   now,
	}
	productRepo.Products[3] = &biz.DiscountProduct{
		ID:          3,
		DiscountID:  200, // 其他折扣的商品
		ProductID:   3,
		ProductName: "AirPods",
		CreatedAt:   now,
	}

	items, total, err := uc.ListDiscountProducts(ctx, 100, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// 验证只返回 discountID=100 的商品
	names := []string{items[0].ProductName, items[1].ProductName}
	assert.Contains(t, names, "iPhone 15 Pro")
	assert.Contains(t, names, "MacBook Pro")
}

func TestListDiscountProducts_EmptyList(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	items, total, err := uc.ListDiscountProducts(ctx, 999, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
}

func TestListDiscountProducts_RepositoryError(t *testing.T) {
	uc, _, productRepo, _, _, _ := newTestUseCase()
	ctx := context.Background()

	productRepo.ReturnError = true

	items, total, err := uc.ListDiscountProducts(ctx, 1, 1, 10)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, int64(0), total)
}

// ============================================================
// 折扣关联分类列表测试
// ============================================================

func TestListDiscountCategories_Success(t *testing.T) {
	uc, _, _, categoryRepo, _, _ := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	categoryRepo.Categories[1] = &biz.DiscountCategory{
		ID:           1,
		DiscountID:   100,
		CategoryID:   1,
		CategoryName: "手机数码",
		CreatedAt:    now,
	}
	categoryRepo.Categories[2] = &biz.DiscountCategory{
		ID:           2,
		DiscountID:   100,
		CategoryID:   2,
		CategoryName: "电脑办公",
		CreatedAt:    now,
	}

	items, total, err := uc.ListDiscountCategories(ctx, 100, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// 验证响应数据
	assert.Equal(t, uint(100), items[0].DiscountID)
	assert.Equal(t, uint(100), items[1].DiscountID)
	categoryNames := []string{items[0].CategoryName, items[1].CategoryName}
	assert.Contains(t, categoryNames, "手机数码")
	assert.Contains(t, categoryNames, "电脑办公")
}

func TestListDiscountCategories_EmptyList(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	items, total, err := uc.ListDiscountCategories(ctx, 999, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
}

func TestListDiscountCategories_RepositoryError(t *testing.T) {
	uc, _, _, categoryRepo, _, _ := newTestUseCase()
	ctx := context.Background()

	categoryRepo.ReturnError = true

	items, total, err := uc.ListDiscountCategories(ctx, 1, 1, 10)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, int64(0), total)
}

// ============================================================
// 折扣关联制造商列表测试
// ============================================================

func TestListDiscountManufacturers_Success(t *testing.T) {
	uc, _, _, _, manufacturerRepo, _ := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	manufacturerRepo.Manufacturers[1] = &biz.DiscountManufacturer{
		ID:               1,
		DiscountID:       100,
		ManufacturerID:   1,
		ManufacturerName: "Apple",
		CreatedAt:        now,
	}
	manufacturerRepo.Manufacturers[2] = &biz.DiscountManufacturer{
		ID:               2,
		DiscountID:       100,
		ManufacturerID:   2,
		ManufacturerName: "Samsung",
		CreatedAt:        now,
	}
	manufacturerRepo.Manufacturers[3] = &biz.DiscountManufacturer{
		ID:               3,
		DiscountID:       200,
		ManufacturerID:   3,
		ManufacturerName: "Sony",
		CreatedAt:        now,
	}

	items, total, err := uc.ListDiscountManufacturers(ctx, 100, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// 验证只返回 discountID=100 的制造商
	manufacturerNames := []string{items[0].ManufacturerName, items[1].ManufacturerName}
	assert.Contains(t, manufacturerNames, "Apple")
	assert.Contains(t, manufacturerNames, "Samsung")
}

func TestListDiscountManufacturers_EmptyList(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	items, total, err := uc.ListDiscountManufacturers(ctx, 999, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
}

func TestListDiscountManufacturers_RepositoryError(t *testing.T) {
	uc, _, _, _, manufacturerRepo, _ := newTestUseCase()
	ctx := context.Background()

	manufacturerRepo.ReturnError = true

	items, total, err := uc.ListDiscountManufacturers(ctx, 1, 1, 10)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, int64(0), total)
}

// ============================================================
// 折扣使用历史列表测试
// ============================================================

func TestListDiscountUsageHistory_Success(t *testing.T) {
	uc, _, _, _, _, usageHistoryRepo := newTestUseCase()
	ctx := context.Background()

	// 预置测试数据
	now := time.Now()
	usedOn := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	usageHistoryRepo.History[1] = &biz.DiscountUsageHistory{
		ID:           1,
		DiscountID:   100,
		OrderID:      1001,
		CustomerID:   500,
		CustomerName: "张三",
		CouponCode:   "SUMMER2024",
		UsedOn:       usedOn,
		CreatedAt:    now,
	}
	usageHistoryRepo.History[2] = &biz.DiscountUsageHistory{
		ID:           2,
		DiscountID:   100,
		OrderID:      1002,
		CustomerID:   501,
		CustomerName: "李四",
		CouponCode:   "SUMMER2024",
		UsedOn:       usedOn,
		CreatedAt:    now,
	}
	usageHistoryRepo.History[3] = &biz.DiscountUsageHistory{
		ID:           3,
		DiscountID:   200,
		OrderID:      2001,
		CustomerID:   502,
		CustomerName: "王五",
		CreatedAt:    now,
	}

	items, total, err := uc.ListDiscountUsageHistory(ctx, 100, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)

	// 验证只返回 discountID=100 的使用记录
	customerNames := []string{items[0].CustomerName, items[1].CustomerName}
	assert.Contains(t, customerNames, "张三")
	assert.Contains(t, customerNames, "李四")
	assert.Equal(t, "2024-06-15", items[0].UsedOn)
	assert.Equal(t, "SUMMER2024", items[0].CouponCode)
}

func TestListDiscountUsageHistory_EmptyList(t *testing.T) {
	uc, _, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	items, total, err := uc.ListDiscountUsageHistory(ctx, 999, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
}

func TestListDiscountUsageHistory_RepositoryError(t *testing.T) {
	uc, _, _, _, _, usageHistoryRepo := newTestUseCase()
	ctx := context.Background()

	usageHistoryRepo.ReturnError = true

	items, total, err := uc.ListDiscountUsageHistory(ctx, 1, 1, 10)
	require.Error(t, err)
	assert.Nil(t, items)
	assert.Equal(t, int64(0), total)
}

// ============================================================
// 转换函数测试（通过 Usecase 方法间接测试）
// ============================================================

// 注：toDiscountResponse, toDiscountProductResponse 等转换函数是 biz 包的私有函数，
// 无法从 biz_test 包直接调用。这些转换函数的逻辑已通过以下 Usecase 方法测试覆盖：
// - ListDiscounts 测试 toDiscountResponse
// - CreateDiscount 测试 toDiscountResponse（包含日期解析）
// - UpdateDiscount 测试 toDiscountResponse（包含日期解析）
// - ListDiscountProducts 测试 toDiscountProductResponse
// - ListDiscountCategories 测试 toDiscountCategoryResponse
// - ListDiscountManufacturers 测试 toDiscountManufacturerResponse
// - ListDiscountUsageHistory 测试 toDiscountUsageHistoryResponse

// TestConversion_DateFormat 测试日期格式化逻辑
func TestConversion_DateFormat(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	// 预置带有日期的折扣数据
	startDate := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
	now := time.Now()

	discountRepo.Discounts[1] = &biz.Discount{
		ID:              1,
		Name:            "带日期的折扣",
		DiscountType:    "percentage",
		DiscountAmount:  20.0,
		StartDate:       startDate,
		EndDate:         endDate,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	items, _, err := uc.ListDiscounts(ctx, 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)

	// 验证日期格式化正确
	assert.Equal(t, "2024-06-01", items[0].StartDate)
	assert.Equal(t, "2024-06-30", items[0].EndDate)
	assert.Equal(t, now.Unix(), items[0].CreatedAt)
	assert.Equal(t, now.Unix(), items[0].UpdatedAt)
}

// TestConversion_EmptyDateFields 测试空日期字段处理
func TestConversion_EmptyDateFields(t *testing.T) {
	uc, discountRepo, _, _, _, _ := newTestUseCase()
	ctx := context.Background()

	now := time.Now()
	discountRepo.Discounts[1] = &biz.Discount{
		ID:              1,
		Name:            "无日期折扣",
		DiscountType:    "fixed",
		DiscountAmount:  50.0,
		StartDate:       time.Time{}, // 零值
		EndDate:         time.Time{}, // 零值
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	items, _, err := uc.ListDiscounts(ctx, 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)

	// 验证零值日期字段返回空字符串
	assert.Empty(t, items[0].StartDate)
	assert.Empty(t, items[0].EndDate)
}

// TestConversion_UsedOnDateFormat 测试使用历史的日期格式化
func TestConversion_UsedOnDateFormat(t *testing.T) {
	uc, _, _, _, _, usageHistoryRepo := newTestUseCase()
	ctx := context.Background()

	now := time.Now()
	usedOn := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)

	usageHistoryRepo.History[1] = &biz.DiscountUsageHistory{
		ID:           1,
		DiscountID:   100,
		OrderID:      5001,
		CustomerID:   888,
		CustomerName: "测试用户",
		CouponCode:   "TEST2024",
		UsedOn:       usedOn,
		CreatedAt:    now,
	}

	items, _, err := uc.ListDiscountUsageHistory(ctx, 100, 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)

	// 验证 UsedOn 日期格式化正确
	assert.Equal(t, "2024-07-15", items[0].UsedOn)
	assert.Equal(t, now.Unix(), items[0].CreatedAt)
}

// TestConversion_EmptyUsedOn 测试使用历史的空日期处理
func TestConversion_EmptyUsedOn(t *testing.T) {
	uc, _, _, _, _, usageHistoryRepo := newTestUseCase()
	ctx := context.Background()

	now := time.Now()
	usageHistoryRepo.History[1] = &biz.DiscountUsageHistory{
		ID:           1,
		DiscountID:   100,
		OrderID:      5001,
		CustomerID:   888,
		CustomerName: "测试用户",
		UsedOn:       time.Time{}, // 零值
		CreatedAt:    now,
	}

	items, _, err := uc.ListDiscountUsageHistory(ctx, 100, 1, 10)
	require.NoError(t, err)
	require.Len(t, items, 1)

	// 验证零值 UsedOn 返回空字符串
	assert.Empty(t, items[0].UsedOn)
}
