// Package biz 业务逻辑层。
// 定义商品服务（catalog）的领域实体、仓储接口和用例。
// 包含商品、分类、制造商、评论、最近浏览等核心领域。
package biz

import (
	"context"
	"time"
)

// ---------------------------------------------------------------------------
// 商品领域实体
// ---------------------------------------------------------------------------

// Product 商品领域实体。
// 包含商品基本信息、价格、库存、分类归属等核心字段。
type Product struct {
	ID             uint      // 商品ID
	Name           string    // 商品名称
	ShortDesc      string    // 简短描述
	FullDesc       string    // 完整描述（富文本）
	SKU            string    // 库存单位编码
	Price          float64   // 商品价格
	OldPrice       float64   // 原价（用于划线价展示）
	Cost           float64   // 成本价
	Stock          int       // 库存数量
	CategoryID     uint      // 所属分类ID
	ManufacturerID uint      // 制造商ID
	IsPublished    bool      // 是否上架
	CreatedAt      time.Time // 创建时间
	UpdatedAt      time.Time // 更新时间
}

// ProductRepository 商品仓储接口。
// 定义商品持久化的核心操作契约。
type ProductRepository interface {
	// Create 创建商品
	Create(ctx context.Context, product *Product) error
	// GetByID 根据ID获取商品
	GetByID(ctx context.Context, id uint) (*Product, error)
	// Update 更新商品信息
	Update(ctx context.Context, product *Product) error
	// List 获取商品列表，支持按分类、制造商、关键词筛选
	List(ctx context.Context, page, size int, categoryID, manufacturerID uint, keyword string) ([]*Product, int64, error)
	// Delete 删除商品
	Delete(ctx context.Context, id uint) error
}

// ProductUseCase 商品用例。
// 封装商品相关的业务逻辑，包括 CRUD、搜索、最近浏览、对比等。
type ProductUseCase struct {
	repo               ProductRepository
	recentlyViewedRepo RecentlyViewedRepository
}

// NewProductUseCase 创建商品用例。
func NewProductUseCase(repo ProductRepository, recentlyViewedRepo RecentlyViewedRepository) *ProductUseCase {
	return &ProductUseCase{
		repo:               repo,
		recentlyViewedRepo: recentlyViewedRepo,
	}
}

// Create 创建商品。
func (uc *ProductUseCase) Create(ctx context.Context, product *Product) (*Product, error) {
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now
	if err := uc.repo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

// GetByID 根据ID获取商品详情。
func (uc *ProductUseCase) GetByID(ctx context.Context, id uint) (*Product, error) {
	return uc.repo.GetByID(ctx, id)
}

// Update 更新商品信息。
func (uc *ProductUseCase) Update(ctx context.Context, product *Product) error {
	product.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, product)
}

// List 获取商品列表，支持按分类、制造商、关键词筛选。
func (uc *ProductUseCase) List(ctx context.Context, page, size int, categoryID, manufacturerID uint, keyword string) ([]*Product, int64, error) {
	return uc.repo.List(ctx, page, size, categoryID, manufacturerID, keyword)
}

// Delete 删除商品。
func (uc *ProductUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// Search 搜索商品（基于关键词模糊匹配）。
// 内部委托给 List 方法，通过 keyword 参数实现搜索。
func (uc *ProductUseCase) Search(ctx context.Context, keyword string, page, size int) ([]*Product, int64, error) {
	return uc.repo.List(ctx, page, size, 0, 0, keyword)
}

// GetRecentlyViewed 获取指定客户的最近浏览商品列表。
// 先从浏览记录中获取商品ID列表，再逐个查询商品详情。
func (uc *ProductUseCase) GetRecentlyViewed(ctx context.Context, customerID uint, limit int) ([]*Product, error) {
	productIDs, err := uc.recentlyViewedRepo.ListByCustomerID(ctx, customerID, limit)
	if err != nil {
		return nil, err
	}

	products := make([]*Product, 0, len(productIDs))
	for _, pid := range productIDs {
		p, err := uc.repo.GetByID(ctx, pid)
		if err != nil {
			continue // 跳过已删除的商品
		}
		products = append(products, p)
	}
	return products, nil
}

// RecordRecentlyViewed 记录用户浏览商品行为。
func (uc *ProductUseCase) RecordRecentlyViewed(ctx context.Context, customerID, productID uint) error {
	return uc.recentlyViewedRepo.Record(ctx, customerID, productID)
}

// CompareProducts 商品对比。
// 根据传入的商品ID列表获取多个商品详情，用于前端对比展示。
func (uc *ProductUseCase) CompareProducts(ctx context.Context, productIDs []uint) ([]*Product, error) {
	products := make([]*Product, 0, len(productIDs))
	for _, pid := range productIDs {
		p, err := uc.repo.GetByID(ctx, pid)
		if err != nil {
			continue // 跳过不存在的商品
		}
		products = append(products, p)
	}
	return products, nil
}

// ---------------------------------------------------------------------------
// 分类领域实体
// ---------------------------------------------------------------------------

// Category 分类领域实体。
// 支持树形结构（ParentID），用于商品分类管理。
type Category struct {
	ID          uint      // 分类ID
	Name        string    // 分类名称
	Description string    // 分类描述
	ParentID    uint      // 父分类ID，0 表示顶级分类
	SortOrder   int       // 排序权重
	IsPublished bool      // 是否启用
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// CategoryRepository 分类仓储接口。
type CategoryRepository interface {
	// Create 创建分类
	Create(ctx context.Context, category *Category) error
	// GetByID 根据ID获取分类
	GetByID(ctx context.Context, id uint) (*Category, error)
	// List 获取分类列表
	List(ctx context.Context, page, size int) ([]*Category, int64, error)
}

// CategoryUseCase 分类用例。
type CategoryUseCase struct {
	repo CategoryRepository
}

// NewCategoryUseCase 创建分类用例。
func NewCategoryUseCase(repo CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

// Create 创建分类。
func (uc *CategoryUseCase) Create(ctx context.Context, category *Category) (*Category, error) {
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now
	if err := uc.repo.Create(ctx, category); err != nil {
		return nil, err
	}
	return category, nil
}

// GetByID 根据ID获取分类详情。
func (uc *CategoryUseCase) GetByID(ctx context.Context, id uint) (*Category, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取分类列表。
func (uc *CategoryUseCase) List(ctx context.Context, page, size int) ([]*Category, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// ---------------------------------------------------------------------------
// 制造商领域实体
// ---------------------------------------------------------------------------

// Manufacturer 制造商领域实体。
type Manufacturer struct {
	ID          uint      // 制造商ID
	Name        string    // 制造商名称
	Description string    // 制造商描述
	IsPublished bool      // 是否启用
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
}

// ManufacturerRepository 制造商仓储接口。
type ManufacturerRepository interface {
	// Create 创建制造商
	Create(ctx context.Context, manufacturer *Manufacturer) error
	// GetByID 根据ID获取制造商
	GetByID(ctx context.Context, id uint) (*Manufacturer, error)
	// List 获取制造商列表
	List(ctx context.Context, page, size int) ([]*Manufacturer, int64, error)
}

// ManufacturerUseCase 制造商用例。
type ManufacturerUseCase struct {
	repo ManufacturerRepository
}

// NewManufacturerUseCase 创建制造商用例。
func NewManufacturerUseCase(repo ManufacturerRepository) *ManufacturerUseCase {
	return &ManufacturerUseCase{repo: repo}
}

// Create 创建制造商。
func (uc *ManufacturerUseCase) Create(ctx context.Context, manufacturer *Manufacturer) (*Manufacturer, error) {
	now := time.Now()
	manufacturer.CreatedAt = now
	manufacturer.UpdatedAt = now
	if err := uc.repo.Create(ctx, manufacturer); err != nil {
		return nil, err
	}
	return manufacturer, nil
}

// GetByID 根据ID获取制造商详情。
func (uc *ManufacturerUseCase) GetByID(ctx context.Context, id uint) (*Manufacturer, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取制造商列表。
func (uc *ManufacturerUseCase) List(ctx context.Context, page, size int) ([]*Manufacturer, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// ---------------------------------------------------------------------------
// 商品评论领域实体
// ---------------------------------------------------------------------------

// ProductReview 商品评论领域实体。
type ProductReview struct {
	ID           uint      // 评论ID
	ProductID    uint      // 商品ID
	CustomerID   uint      // 评论者客户ID
	CustomerName string    // 评论者名称
	Title        string    // 评论标题
	Content      string    // 评论内容
	Rating       int       // 评分 1-5
	IsApproved   bool      // 是否审核通过
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// ProductReviewRepository 商品评论仓储接口。
type ProductReviewRepository interface {
	// Create 创建商品评论
	Create(ctx context.Context, review *ProductReview) error
	// ListByProductID 获取指定商品的评论列表
	ListByProductID(ctx context.Context, productID uint, page, size int) ([]*ProductReview, int64, error)
}

// ProductReviewUseCase 商品评论用例。
type ProductReviewUseCase struct {
	repo ProductReviewRepository
}

// NewProductReviewUseCase 创建商品评论用例。
func NewProductReviewUseCase(repo ProductReviewRepository) *ProductReviewUseCase {
	return &ProductReviewUseCase{repo: repo}
}

// Create 创建商品评论。
// 新评论默认未审核（IsApproved=false），需后台审核后展示。
func (uc *ProductReviewUseCase) Create(ctx context.Context, review *ProductReview) (*ProductReview, error) {
	now := time.Now()
	review.CreatedAt = now
	review.UpdatedAt = now
	// 新评论默认未审核
	review.IsApproved = false
	if err := uc.repo.Create(ctx, review); err != nil {
		return nil, err
	}
	return review, nil
}

// ListByProductID 获取指定商品的评论列表。
func (uc *ProductReviewUseCase) ListByProductID(ctx context.Context, productID uint, page, size int) ([]*ProductReview, int64, error) {
	return uc.repo.ListByProductID(ctx, productID, page, size)
}

// ---------------------------------------------------------------------------
// 最近浏览领域实体
// ---------------------------------------------------------------------------

// RecentlyViewedRepository 最近浏览仓储接口。
type RecentlyViewedRepository interface {
	// Record 记录一次浏览行为
	Record(ctx context.Context, customerID, productID uint) error
	// ListByCustomerID 获取指定客户的最近浏览商品ID列表
	ListByCustomerID(ctx context.Context, customerID uint, limit int) ([]uint, error)
}
