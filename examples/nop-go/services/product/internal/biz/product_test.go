// Package biz_test 商品服务业务逻辑层单元测试。
//
// 使用 mock 仓储隔离依赖，测试各个 UseCase 的核心业务逻辑。
package biz_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nop-go/services/product/internal/biz"
)

// ============================================================
// Mock 仓储实现
// ============================================================

// MockProductRepository 商品仓储 mock 实现。
type MockProductRepository struct {
	Products map[uint]*biz.Product
	NextID   uint
	// 用于模拟错误场景
	CreateError error
	GetError   error
	UpdateError error
	DeleteError error
	ListError  error
}

// NewMockProductRepository 创建 mock 商品仓储。
func NewMockProductRepository() *MockProductRepository {
	return &MockProductRepository{
		Products: make(map[uint]*biz.Product),
		NextID:   1,
	}
}

func (m *MockProductRepository) Create(ctx context.Context, product *biz.Product) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	product.ID = m.NextID
	m.NextID++
	m.Products[product.ID] = product
	return nil
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uint) (*biz.Product, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	product, ok := m.Products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return product, nil
}

func (m *MockProductRepository) Update(ctx context.Context, product *biz.Product) error {
	if m.UpdateError != nil {
		return m.UpdateError
	}
	m.Products[product.ID] = product
	return nil
}

func (m *MockProductRepository) List(ctx context.Context, page, size int, categoryID, manufacturerID uint, keyword string) ([]*biz.Product, int64, error) {
	if m.ListError != nil {
		return nil, 0, m.ListError
	}
	var result []*biz.Product
	for _, p := range m.Products {
		// 按筛选条件过滤
		if categoryID > 0 && p.CategoryID != categoryID {
			continue
		}
		if manufacturerID > 0 && p.ManufacturerID != manufacturerID {
			continue
		}
		if keyword != "" && !containsIgnoreCase(p.Name, keyword) {
			continue
		}
		result = append(result, p)
	}
	// 简单分页
	total := int64(len(result))
	start := (page - 1) * size
	if start >= len(result) {
		return []*biz.Product{}, total, nil
	}
	end := start + size
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

func (m *MockProductRepository) Delete(ctx context.Context, id uint) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	delete(m.Products, id)
	return nil
}

// containsIgnoreCase 检查字符串是否包含子串（忽略大小写）。
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		anyValidIndex(s, substr))
}

// anyValidIndex 简单的子串匹配辅助函数。
func anyValidIndex(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			subc := substr[j]
			// 简单大小写转换
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if subc >= 'A' && subc <= 'Z' {
				subc += 32
			}
			if sc != subc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// MockCategoryRepository 分类仓储 mock 实现。
type MockCategoryRepository struct {
	Categories map[uint]*biz.Category
	NextID     uint
	CreateError error
	GetError   error
	ListError  error
}

// NewMockCategoryRepository 创建 mock 分类仓储。
func NewMockCategoryRepository() *MockCategoryRepository {
	return &MockCategoryRepository{
		Categories: make(map[uint]*biz.Category),
		NextID:     1,
	}
}

func (m *MockCategoryRepository) Create(ctx context.Context, category *biz.Category) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	category.ID = m.NextID
	m.NextID++
	m.Categories[category.ID] = category
	return nil
}

func (m *MockCategoryRepository) GetByID(ctx context.Context, id uint) (*biz.Category, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	category, ok := m.Categories[id]
	if !ok {
		return nil, errors.New("category not found")
	}
	return category, nil
}

func (m *MockCategoryRepository) List(ctx context.Context, page, size int) ([]*biz.Category, int64, error) {
	if m.ListError != nil {
		return nil, 0, m.ListError
	}
	var result []*biz.Category
	for _, c := range m.Categories {
		result = append(result, c)
	}
	total := int64(len(result))
	start := (page - 1) * size
	if start >= len(result) {
		return []*biz.Category{}, total, nil
	}
	end := start + size
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// MockManufacturerRepository 制造商仓储 mock 实现。
type MockManufacturerRepository struct {
	Manufacturers map[uint]*biz.Manufacturer
	NextID        uint
	CreateError   error
	GetError      error
	ListError     error
}

// NewMockManufacturerRepository 创建 mock 制造商仓储。
func NewMockManufacturerRepository() *MockManufacturerRepository {
	return &MockManufacturerRepository{
		Manufacturers: make(map[uint]*biz.Manufacturer),
		NextID:        1,
	}
}

func (m *MockManufacturerRepository) Create(ctx context.Context, manufacturer *biz.Manufacturer) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	manufacturer.ID = m.NextID
	m.NextID++
	m.Manufacturers[manufacturer.ID] = manufacturer
	return nil
}

func (m *MockManufacturerRepository) GetByID(ctx context.Context, id uint) (*biz.Manufacturer, error) {
	if m.GetError != nil {
		return nil, m.GetError
	}
	manufacturer, ok := m.Manufacturers[id]
	if !ok {
		return nil, errors.New("manufacturer not found")
	}
	return manufacturer, nil
}

func (m *MockManufacturerRepository) List(ctx context.Context, page, size int) ([]*biz.Manufacturer, int64, error) {
	if m.ListError != nil {
		return nil, 0, m.ListError
	}
	var result []*biz.Manufacturer
	for _, mfr := range m.Manufacturers {
		result = append(result, mfr)
	}
	total := int64(len(result))
	start := (page - 1) * size
	if start >= len(result) {
		return []*biz.Manufacturer{}, total, nil
	}
	end := start + size
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// MockProductReviewRepository 商品评论仓储 mock 实现。
type MockProductReviewRepository struct {
	Reviews     map[uint]*biz.ProductReview
	NextID      uint
	CreateError error
	ListError   error
}

// NewMockProductReviewRepository 创建 mock 商品评论仓储。
func NewMockProductReviewRepository() *MockProductReviewRepository {
	return &MockProductReviewRepository{
		Reviews: make(map[uint]*biz.ProductReview),
		NextID:  1,
	}
}

func (m *MockProductReviewRepository) Create(ctx context.Context, review *biz.ProductReview) error {
	if m.CreateError != nil {
		return m.CreateError
	}
	review.ID = m.NextID
	m.NextID++
	m.Reviews[review.ID] = review
	return nil
}

func (m *MockProductReviewRepository) ListByProductID(ctx context.Context, productID uint, page, size int) ([]*biz.ProductReview, int64, error) {
	if m.ListError != nil {
		return nil, 0, m.ListError
	}
	var result []*biz.ProductReview
	for _, r := range m.Reviews {
		if r.ProductID == productID {
			result = append(result, r)
		}
	}
	total := int64(len(result))
	start := (page - 1) * size
	if start >= len(result) {
		return []*biz.ProductReview{}, total, nil
	}
	end := start + size
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total, nil
}

// MockRecentlyViewedRepository 最近浏览仓储 mock 实现。
type MockRecentlyViewedRepository struct {
	// key: customerID, value: []productID (按浏览时间倒序)
	ViewedProducts map[uint][]uint
	RecordError    error
	ListError      error
}

// NewMockRecentlyViewedRepository 创建 mock 最近浏览仓储。
func NewMockRecentlyViewedRepository() *MockRecentlyViewedRepository {
	return &MockRecentlyViewedRepository{
		ViewedProducts: make(map[uint][]uint),
	}
}

func (m *MockRecentlyViewedRepository) Record(ctx context.Context, customerID, productID uint) error {
	if m.RecordError != nil {
		return m.RecordError
	}
	// 移除已存在的相同商品ID（避免重复）
	products := m.ViewedProducts[customerID]
	newProducts := make([]uint, 0, len(products)+1)
	for _, pid := range products {
		if pid != productID {
			newProducts = append(newProducts, pid)
		}
	}
	// 新浏览的商品放在最前面
	newProducts = append([]uint{productID}, newProducts...)
	m.ViewedProducts[customerID] = newProducts
	return nil
}

func (m *MockRecentlyViewedRepository) ListByCustomerID(ctx context.Context, customerID uint, limit int) ([]uint, error) {
	if m.ListError != nil {
		return nil, m.ListError
	}
	products := m.ViewedProducts[customerID]
	if limit > 0 && len(products) > limit {
		return products[:limit], nil
	}
	return products, nil
}

// ============================================================
// 测试辅助函数
// ============================================================

// newTestProductUseCase 创建测试用的 ProductUseCase。
func newTestProductUseCase() (*biz.ProductUseCase, *MockProductRepository, *MockRecentlyViewedRepository) {
	productRepo := NewMockProductRepository()
	recentlyViewedRepo := NewMockRecentlyViewedRepository()
	uc := biz.NewProductUseCase(productRepo, recentlyViewedRepo)
	return uc, productRepo, recentlyViewedRepo
}

// newTestCategoryUseCase 创建测试用的 CategoryUseCase。
func newTestCategoryUseCase() (*biz.CategoryUseCase, *MockCategoryRepository) {
	repo := NewMockCategoryRepository()
	uc := biz.NewCategoryUseCase(repo)
	return uc, repo
}

// newTestManufacturerUseCase 创建测试用的 ManufacturerUseCase。
func newTestManufacturerUseCase() (*biz.ManufacturerUseCase, *MockManufacturerRepository) {
	repo := NewMockManufacturerRepository()
	uc := biz.NewManufacturerUseCase(repo)
	return uc, repo
}

// newTestProductReviewUseCase 创建测试用的 ProductReviewUseCase。
func newTestProductReviewUseCase() (*biz.ProductReviewUseCase, *MockProductReviewRepository) {
	repo := NewMockProductReviewRepository()
	uc := biz.NewProductReviewUseCase(repo)
	return uc, repo
}

// ============================================================
// ProductUseCase 测试
// ============================================================

// -------------------- Create 测试 --------------------

func TestProductCreate_Success(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	product := &biz.Product{
		Name:           "测试商品",
		ShortDesc:      "简短描述",
		FullDesc:       "<p>完整描述</p>",
		SKU:            "SKU-001",
		Price:          99.99,
		OldPrice:       129.99,
		Cost:           50.00,
		Stock:          100,
		CategoryID:     1,
		ManufacturerID: 1,
		IsPublished:    true,
	}

	created, err := uc.Create(ctx, product)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "测试商品", created.Name)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
}

func TestProductCreate_RepositoryError(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	productRepo.CreateError = errors.New("database error")
	product := &biz.Product{Name: "测试商品"}

	created, err := uc.Create(ctx, product)
	assert.Error(t, err)
	assert.Nil(t, created)
}

// -------------------- GetByID 测试 --------------------

func TestProductGetByID_Success(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 先创建一个商品
	product := &biz.Product{Name: "测试商品", Price: 99.99}
	err := productRepo.Create(ctx, product)
	require.NoError(t, err)

	// 再获取
	found, err := uc.GetByID(ctx, product.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, product.ID, found.ID)
	assert.Equal(t, "测试商品", found.Name)
}

func TestProductGetByID_NotFound(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	found, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestProductGetByID_RepositoryError(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	productRepo.GetError = errors.New("database error")
	found, err := uc.GetByID(ctx, 1)
	assert.Error(t, err)
	assert.Nil(t, found)
}

// -------------------- Update 测试 --------------------

func TestProductUpdate_Success(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 先创建
	product := &biz.Product{Name: "原名称", Price: 99.99}
	require.NoError(t, productRepo.Create(ctx, product))

	// 更新
	product.Name = "新名称"
	product.Price = 88.88
	err := uc.Update(ctx, product)
	assert.NoError(t, err)

	// 验证
	updated, _ := productRepo.GetByID(ctx, product.ID)
	assert.Equal(t, "新名称", updated.Name)
	assert.Equal(t, 88.88, updated.Price)
	assert.True(t, updated.UpdatedAt.After(updated.CreatedAt) || updated.UpdatedAt.Equal(updated.CreatedAt))
}

func TestProductUpdate_RepositoryError(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	productRepo.UpdateError = errors.New("database error")
	product := &biz.Product{ID: 1, Name: "测试"}

	err := uc.Update(ctx, product)
	assert.Error(t, err)
}

// -------------------- List 测试 --------------------

func TestProductList_Success(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建多个商品
	for i := 1; i <= 15; i++ {
		product := &biz.Product{
			Name:           "商品" + string(rune('0'+i%10)),
			Price:          float64(i * 10),
			CategoryID:     uint(i % 3 + 1),
			ManufacturerID: uint(i % 2 + 1),
		}
		require.NoError(t, productRepo.Create(ctx, product))
	}

	// 获取第一页
	products, total, err := uc.List(ctx, 1, 10, 0, 0, "")
	assert.NoError(t, err)
	assert.Len(t, products, 10)
	assert.Equal(t, int64(15), total)
}

func TestProductList_WithCategoryFilter(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建不同分类的商品
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "商品A", CategoryID: 1}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "商品B", CategoryID: 2}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "商品C", CategoryID: 1}))

	// 按分类筛选
	products, total, err := uc.List(ctx, 1, 10, 1, 0, "")
	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, int64(2), total)
}

func TestProductList_WithManufacturerFilter(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建不同制造商的商品
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "商品A", ManufacturerID: 1}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "商品B", ManufacturerID: 2}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "商品C", ManufacturerID: 1}))

	// 按制造商筛选
	products, total, err := uc.List(ctx, 1, 10, 0, 1, "")
	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, int64(2), total)
}

func TestProductList_WithKeyword(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建商品
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "苹果手机"}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "华为手机"}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "小米平板"}))

	// 按关键词搜索
	products, total, err := uc.List(ctx, 1, 10, 0, 0, "手机")
	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, int64(2), total)
}

func TestProductList_EmptyResult(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	products, total, err := uc.List(ctx, 1, 10, 0, 0, "")
	assert.NoError(t, err)
	assert.Empty(t, products)
	assert.Equal(t, int64(0), total)
}

func TestProductList_RepositoryError(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	productRepo.ListError = errors.New("database error")
	products, total, err := uc.List(ctx, 1, 10, 0, 0, "")
	assert.Error(t, err)
	assert.Nil(t, products)
	assert.Zero(t, total)
}

// -------------------- Delete 测试 --------------------

func TestProductDelete_Success(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 先创建
	product := &biz.Product{Name: "待删除商品"}
	require.NoError(t, productRepo.Create(ctx, product))

	// 删除
	err := uc.Delete(ctx, product.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = productRepo.GetByID(ctx, product.ID)
	assert.Error(t, err)
}

func TestProductDelete_RepositoryError(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	productRepo.DeleteError = errors.New("database error")
	err := uc.Delete(ctx, 1)
	assert.Error(t, err)
}

// -------------------- Search 测试 --------------------

func TestProductSearch_Success(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建商品
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "iPhone 15 Pro"}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "iPhone 14"}))
	require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "Samsung Galaxy"}))

	// 搜索
	products, total, err := uc.Search(ctx, "iPhone", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, products, 2)
	assert.Equal(t, int64(2), total)
}

func TestProductSearch_NoResults(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	products, total, err := uc.Search(ctx, "不存在的关键词", 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, products)
	assert.Equal(t, int64(0), total)
}

// -------------------- GetRecentlyViewed 测试 --------------------

func TestGetRecentlyViewed_Success(t *testing.T) {
	uc, productRepo, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	// 创建商品
	product1 := &biz.Product{Name: "商品A"}
	product2 := &biz.Product{Name: "商品B"}
	product3 := &biz.Product{Name: "商品C"}
	require.NoError(t, productRepo.Create(ctx, product1))
	require.NoError(t, productRepo.Create(ctx, product2))
	require.NoError(t, productRepo.Create(ctx, product3))

	// 记录浏览
	customerID := uint(1)
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product1.ID))
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product2.ID))
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product3.ID))

	// 获取最近浏览
	products, err := uc.GetRecentlyViewed(ctx, customerID, 10)
	assert.NoError(t, err)
	assert.Len(t, products, 3)
}

func TestGetRecentlyViewed_WithLimit(t *testing.T) {
	uc, productRepo, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	// 创建商品
	product1 := &biz.Product{Name: "商品A"}
	product2 := &biz.Product{Name: "商品B"}
	product3 := &biz.Product{Name: "商品C"}
	require.NoError(t, productRepo.Create(ctx, product1))
	require.NoError(t, productRepo.Create(ctx, product2))
	require.NoError(t, productRepo.Create(ctx, product3))

	// 记录浏览
	customerID := uint(1)
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product1.ID))
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product2.ID))
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product3.ID))

	// 获取最近浏览（限制数量）
	products, err := uc.GetRecentlyViewed(ctx, customerID, 2)
	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestGetRecentlyViewed_SkipDeletedProducts(t *testing.T) {
	uc, productRepo, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	// 创建商品
	product1 := &biz.Product{Name: "商品A"}
	product2 := &biz.Product{Name: "商品B"}
	require.NoError(t, productRepo.Create(ctx, product1))
	require.NoError(t, productRepo.Create(ctx, product2))

	// 记录浏览（包括一个将被删除的商品）
	customerID := uint(1)
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product1.ID))
	require.NoError(t, recentlyViewedRepo.Record(ctx, customerID, product2.ID))

	// 删除一个商品
	require.NoError(t, productRepo.Delete(ctx, product2.ID))

	// 获取最近浏览（已删除的商品应被跳过）
	products, err := uc.GetRecentlyViewed(ctx, customerID, 10)
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, product1.ID, products[0].ID)
}

func TestGetRecentlyViewed_Empty(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	products, err := uc.GetRecentlyViewed(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, products)
}

func TestGetRecentlyViewed_RepositoryError(t *testing.T) {
	uc, _, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	recentlyViewedRepo.ListError = errors.New("database error")
	products, err := uc.GetRecentlyViewed(ctx, 1, 10)
	assert.Error(t, err)
	assert.Nil(t, products)
}

// -------------------- RecordRecentlyViewed 测试 --------------------

func TestRecordRecentlyViewed_Success(t *testing.T) {
	uc, _, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	err := uc.RecordRecentlyViewed(ctx, 1, 100)
	assert.NoError(t, err)

	// 验证记录
	productIDs, _ := recentlyViewedRepo.ListByCustomerID(ctx, 1, 10)
	assert.Contains(t, productIDs, uint(100))
}

func TestRecordRecentlyViewed_RepositoryError(t *testing.T) {
	uc, _, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	recentlyViewedRepo.RecordError = errors.New("database error")
	err := uc.RecordRecentlyViewed(ctx, 1, 100)
	assert.Error(t, err)
}

// -------------------- CompareProducts 测试 --------------------

func TestCompareProducts_Success(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建商品
	product1 := &biz.Product{Name: "商品A", Price: 99.99}
	product2 := &biz.Product{Name: "商品B", Price: 88.88}
	require.NoError(t, productRepo.Create(ctx, product1))
	require.NoError(t, productRepo.Create(ctx, product2))

	// 对比
	products, err := uc.CompareProducts(ctx, []uint{product1.ID, product2.ID})
	assert.NoError(t, err)
	assert.Len(t, products, 2)
}

func TestCompareProducts_SkipNonExistent(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 只创建一个商品
	product := &biz.Product{Name: "商品A"}
	require.NoError(t, productRepo.Create(ctx, product))

	// 对比（包含一个不存在的ID）
	products, err := uc.CompareProducts(ctx, []uint{product.ID, 999})
	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, product.ID, products[0].ID)
}

func TestCompareProducts_EmptyIDs(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	products, err := uc.CompareProducts(ctx, []uint{})
	assert.NoError(t, err)
	assert.Empty(t, products)
}

func TestCompareProducts_AllNonExistent(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	products, err := uc.CompareProducts(ctx, []uint{999, 1000})
	assert.NoError(t, err)
	assert.Empty(t, products)
}

// ============================================================
// CategoryUseCase 测试
// ============================================================

func TestCategoryCreate_Success(t *testing.T) {
	uc, _ := newTestCategoryUseCase()
	ctx := context.Background()

	category := &biz.Category{
		Name:        "电子产品",
		Description: "各类电子产品",
		ParentID:    0,
		SortOrder:   1,
		IsPublished: true,
	}

	created, err := uc.Create(ctx, category)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "电子产品", created.Name)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
}

func TestCategoryCreate_RepositoryError(t *testing.T) {
	uc, repo := newTestCategoryUseCase()
	ctx := context.Background()

	repo.CreateError = errors.New("database error")
	category := &biz.Category{Name: "测试分类"}

	created, err := uc.Create(ctx, category)
	assert.Error(t, err)
	assert.Nil(t, created)
}

func TestCategoryGetByID_Success(t *testing.T) {
	uc, repo := newTestCategoryUseCase()
	ctx := context.Background()

	category := &biz.Category{Name: "测试分类"}
	require.NoError(t, repo.Create(ctx, category))

	found, err := uc.GetByID(ctx, category.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, category.ID, found.ID)
}

func TestCategoryGetByID_NotFound(t *testing.T) {
	uc, _ := newTestCategoryUseCase()
	ctx := context.Background()

	found, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestCategoryList_Success(t *testing.T) {
	uc, repo := newTestCategoryUseCase()
	ctx := context.Background()

	// 创建多个分类
	for i := 1; i <= 5; i++ {
		require.NoError(t, repo.Create(ctx, &biz.Category{Name: "分类" + string(rune('0'+i))}))
	}

	categories, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, categories, 5)
	assert.Equal(t, int64(5), total)
}

func TestCategoryList_Empty(t *testing.T) {
	uc, _ := newTestCategoryUseCase()
	ctx := context.Background()

	categories, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, categories)
	assert.Equal(t, int64(0), total)
}

func TestCategoryList_RepositoryError(t *testing.T) {
	uc, repo := newTestCategoryUseCase()
	ctx := context.Background()

	repo.ListError = errors.New("database error")
	categories, total, err := uc.List(ctx, 1, 10)
	assert.Error(t, err)
	assert.Nil(t, categories)
	assert.Zero(t, total)
}

// ============================================================
// ManufacturerUseCase 测试
// ============================================================

func TestManufacturerCreate_Success(t *testing.T) {
	uc, _ := newTestManufacturerUseCase()
	ctx := context.Background()

	manufacturer := &biz.Manufacturer{
		Name:        "苹果公司",
		Description: "美国科技公司",
		IsPublished: true,
	}

	created, err := uc.Create(ctx, manufacturer)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.Equal(t, "苹果公司", created.Name)
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
}

func TestManufacturerCreate_RepositoryError(t *testing.T) {
	uc, repo := newTestManufacturerUseCase()
	ctx := context.Background()

	repo.CreateError = errors.New("database error")
	manufacturer := &biz.Manufacturer{Name: "测试制造商"}

	created, err := uc.Create(ctx, manufacturer)
	assert.Error(t, err)
	assert.Nil(t, created)
}

func TestManufacturerGetByID_Success(t *testing.T) {
	uc, repo := newTestManufacturerUseCase()
	ctx := context.Background()

	manufacturer := &biz.Manufacturer{Name: "测试制造商"}
	require.NoError(t, repo.Create(ctx, manufacturer))

	found, err := uc.GetByID(ctx, manufacturer.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, manufacturer.ID, found.ID)
}

func TestManufacturerGetByID_NotFound(t *testing.T) {
	uc, _ := newTestManufacturerUseCase()
	ctx := context.Background()

	found, err := uc.GetByID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestManufacturerList_Success(t *testing.T) {
	uc, repo := newTestManufacturerUseCase()
	ctx := context.Background()

	// 创建多个制造商
	for i := 1; i <= 3; i++ {
		require.NoError(t, repo.Create(ctx, &biz.Manufacturer{Name: "制造商" + string(rune('0'+i))}))
	}

	manufacturers, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, manufacturers, 3)
	assert.Equal(t, int64(3), total)
}

func TestManufacturerList_Empty(t *testing.T) {
	uc, _ := newTestManufacturerUseCase()
	ctx := context.Background()

	manufacturers, total, err := uc.List(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, manufacturers)
	assert.Equal(t, int64(0), total)
}

func TestManufacturerList_RepositoryError(t *testing.T) {
	uc, repo := newTestManufacturerUseCase()
	ctx := context.Background()

	repo.ListError = errors.New("database error")
	manufacturers, total, err := uc.List(ctx, 1, 10)
	assert.Error(t, err)
	assert.Nil(t, manufacturers)
	assert.Zero(t, total)
}

// ============================================================
// ProductReviewUseCase 测试
// ============================================================

func TestProductReviewCreate_Success(t *testing.T) {
	uc, _ := newTestProductReviewUseCase()
	ctx := context.Background()

	review := &biz.ProductReview{
		ProductID:    1,
		CustomerID:   1,
		CustomerName: "张三",
		Title:        "好评",
		Content:      "商品质量很好",
		Rating:       5,
	}

	created, err := uc.Create(ctx, review)
	assert.NoError(t, err)
	assert.NotNil(t, created)
	assert.NotZero(t, created.ID)
	assert.False(t, created.IsApproved, "新评论默认未审核")
	assert.NotZero(t, created.CreatedAt)
	assert.NotZero(t, created.UpdatedAt)
}

func TestProductReviewCreate_RepositoryError(t *testing.T) {
	uc, repo := newTestProductReviewUseCase()
	ctx := context.Background()

	repo.CreateError = errors.New("database error")
	review := &biz.ProductReview{ProductID: 1, CustomerID: 1}

	created, err := uc.Create(ctx, review)
	assert.Error(t, err)
	assert.Nil(t, created)
}

func TestProductReviewListByProductID_Success(t *testing.T) {
	uc, repo := newTestProductReviewUseCase()
	ctx := context.Background()

	// 创建评论
	review1 := &biz.ProductReview{ProductID: 1, CustomerID: 1, Rating: 5}
	review2 := &biz.ProductReview{ProductID: 1, CustomerID: 2, Rating: 4}
	review3 := &biz.ProductReview{ProductID: 2, CustomerID: 1, Rating: 3}
	require.NoError(t, repo.Create(ctx, review1))
	require.NoError(t, repo.Create(ctx, review2))
	require.NoError(t, repo.Create(ctx, review3))

	reviews, total, err := uc.ListByProductID(ctx, 1, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, reviews, 2)
	assert.Equal(t, int64(2), total)
}

func TestProductReviewListByProductID_Empty(t *testing.T) {
	uc, _ := newTestProductReviewUseCase()
	ctx := context.Background()

	reviews, total, err := uc.ListByProductID(ctx, 999, 1, 10)
	assert.NoError(t, err)
	assert.Empty(t, reviews)
	assert.Equal(t, int64(0), total)
}

func TestProductReviewListByProductID_RepositoryError(t *testing.T) {
	uc, repo := newTestProductReviewUseCase()
	ctx := context.Background()

	repo.ListError = errors.New("database error")
	reviews, total, err := uc.ListByProductID(ctx, 1, 1, 10)
	assert.Error(t, err)
	assert.Nil(t, reviews)
	assert.Zero(t, total)
}

// ============================================================
// 边界条件和并发测试
// ============================================================

func TestProductCreate_TimeFields(t *testing.T) {
	uc, _, _ := newTestProductUseCase()
	ctx := context.Background()

	beforeCreate := time.Now()
	product := &biz.Product{Name: "测试商品"}
	created, err := uc.Create(ctx, product)
	afterCreate := time.Now()

	assert.NoError(t, err)
	assert.True(t, created.CreatedAt.After(beforeCreate) || created.CreatedAt.Equal(beforeCreate))
	assert.True(t, created.UpdatedAt.After(beforeCreate) || created.UpdatedAt.Equal(beforeCreate))
	assert.True(t, created.CreatedAt.Before(afterCreate) || created.CreatedAt.Equal(afterCreate))
}

func TestProductList_Pagination(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建25个商品
	for i := 1; i <= 25; i++ {
		require.NoError(t, productRepo.Create(ctx, &biz.Product{Name: "商品"}))
	}

	// 第一页
	products, total, err := uc.List(ctx, 1, 10, 0, 0, "")
	assert.NoError(t, err)
	assert.Len(t, products, 10)
	assert.Equal(t, int64(25), total)

	// 第三页
	products, total, err = uc.List(ctx, 3, 10, 0, 0, "")
	assert.NoError(t, err)
	assert.Len(t, products, 5)
	assert.Equal(t, int64(25), total)

	// 超出页码
	products, total, err = uc.List(ctx, 10, 10, 0, 0, "")
	assert.NoError(t, err)
	assert.Empty(t, products)
	assert.Equal(t, int64(25), total)
}

func TestCategoryCreate_SubCategory(t *testing.T) {
	uc, repo := newTestCategoryUseCase()
	ctx := context.Background()

	// 创建父分类
	parentCategory := &biz.Category{Name: "电子产品"}
	require.NoError(t, repo.Create(ctx, parentCategory))

	// 创建子分类
	subCategory := &biz.Category{
		Name:     "手机",
		ParentID: parentCategory.ID,
	}
	created, err := uc.Create(ctx, subCategory)
	assert.NoError(t, err)
	assert.Equal(t, parentCategory.ID, created.ParentID)
}

func TestRecordRecentlyViewed_DuplicateRecord(t *testing.T) {
	uc, _, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	// 重复记录同一商品
	require.NoError(t, uc.RecordRecentlyViewed(ctx, 1, 100))
	require.NoError(t, uc.RecordRecentlyViewed(ctx, 1, 100))
	require.NoError(t, uc.RecordRecentlyViewed(ctx, 1, 100))

	// 验证没有重复
	productIDs, err := recentlyViewedRepo.ListByCustomerID(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, productIDs, 1)
	assert.Equal(t, uint(100), productIDs[0])
}

func TestRecordRecentlyViewed_OrderMatters(t *testing.T) {
	uc, _, recentlyViewedRepo := newTestProductUseCase()
	ctx := context.Background()

	// 按顺序记录浏览
	require.NoError(t, uc.RecordRecentlyViewed(ctx, 1, 100))
	require.NoError(t, uc.RecordRecentlyViewed(ctx, 1, 200))
	require.NoError(t, uc.RecordRecentlyViewed(ctx, 1, 300))

	// 最新浏览的应该在最前面
	productIDs, err := recentlyViewedRepo.ListByCustomerID(ctx, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, productIDs, 3)
	assert.Equal(t, uint(300), productIDs[0], "最新浏览的商品应排在最前")
	assert.Equal(t, uint(200), productIDs[1])
	assert.Equal(t, uint(100), productIDs[2])
}

func TestProductReviewCreate_DefaultNotApproved(t *testing.T) {
	uc, _ := newTestProductReviewUseCase()
	ctx := context.Background()

	// 即使设置 IsApproved=true，也应被覆盖为 false
	review := &biz.ProductReview{
		ProductID:  1,
		CustomerID: 1,
		IsApproved: true, // 尝试设置为已审核
	}

	created, err := uc.Create(ctx, review)
	assert.NoError(t, err)
	assert.False(t, created.IsApproved, "新评论应强制设置为未审核状态")
}

func TestCompareProducts_LargeList(t *testing.T) {
	uc, productRepo, _ := newTestProductUseCase()
	ctx := context.Background()

	// 创建100个商品
	productIDs := make([]uint, 100)
	for i := 0; i < 100; i++ {
		product := &biz.Product{Name: "商品"}
		require.NoError(t, productRepo.Create(ctx, product))
		productIDs[i] = product.ID
	}

	// 对比所有商品
	products, err := uc.CompareProducts(ctx, productIDs)
	assert.NoError(t, err)
	assert.Len(t, products, 100)
}

func TestManufacturerGetByID_RepositoryError(t *testing.T) {
	uc, repo := newTestManufacturerUseCase()
	ctx := context.Background()

	repo.GetError = errors.New("database error")
	found, err := uc.GetByID(ctx, 1)
	assert.Error(t, err)
	assert.Nil(t, found)
}

func TestCategoryGetByID_RepositoryError(t *testing.T) {
	uc, repo := newTestCategoryUseCase()
	ctx := context.Background()

	repo.GetError = errors.New("database error")
	found, err := uc.GetByID(ctx, 1)
	assert.Error(t, err)
	assert.Nil(t, found)
}
