// Package service 服务层。
// 封装业务用例调用，提供面向 handler 的服务接口。
// 负责领域实体与响应结构体之间的转换。
package service

import (
	"context"

	"nop-go/services/product/internal/biz"
	"nop-go/services/product/internal/data"

	"gorm.io/gorm"
)

// Services 聚合所有服务实例，供路由注册时统一注入。
type Services struct {
	Product     *ProductService
	Category    *CategoryService
	Manufacturer *ManufacturerService
	Review      *ProductReviewService
}

// NewServices 创建所有服务实例。
// 依次初始化数据层仓储 -> 业务用例 -> 服务层，完成依赖注入链路。
func NewServices(db *gorm.DB) *Services {
	// 初始化仓储
	productRepo := data.NewProductRepo(db)
	categoryRepo := data.NewCategoryRepo(db)
	manufacturerRepo := data.NewManufacturerRepo(db)
	reviewRepo := data.NewProductReviewRepo(db)
	recentlyViewedRepo := data.NewRecentlyViewedRepo(db)

	// 初始化业务用例
	productUC := biz.NewProductUseCase(productRepo, recentlyViewedRepo)
	categoryUC := biz.NewCategoryUseCase(categoryRepo)
	manufacturerUC := biz.NewManufacturerUseCase(manufacturerRepo)
	reviewUC := biz.NewProductReviewUseCase(reviewRepo)

	return &Services{
		Product:     &ProductService{uc: productUC},
		Category:    &CategoryService{uc: categoryUC},
		Manufacturer: &ManufacturerService{uc: manufacturerUC},
		Review:      &ProductReviewService{uc: reviewUC},
	}
}

// ---------------------------------------------------------------------------
// 商品服务
// ---------------------------------------------------------------------------

// ProductService 商品服务，封装商品相关业务逻辑。
type ProductService struct {
	uc *biz.ProductUseCase
}

// CreateProductRequest 创建商品请求（服务层）。
type CreateProductRequest struct {
	Name           string  `json:"name" binding:"required"`
	ShortDesc      string  `json:"short_desc"`
	FullDesc       string  `json:"full_desc"`
	SKU            string  `json:"sku"`
	Price          float64 `json:"price" binding:"required"`
	OldPrice       float64 `json:"old_price"`
	Cost           float64 `json:"cost"`
	Stock          int     `json:"stock"`
	CategoryID     uint    `json:"category_id"`
	ManufacturerID uint    `json:"manufacturer_id"`
	IsPublished    bool    `json:"is_published"`
}

// UpdateProductRequest 更新商品请求（服务层）。
type UpdateProductRequest struct {
	Name           *string  `json:"name"`
	ShortDesc      *string  `json:"short_desc"`
	FullDesc       *string  `json:"full_desc"`
	SKU            *string  `json:"sku"`
	Price          *float64 `json:"price"`
	OldPrice       *float64 `json:"old_price"`
	Cost           *float64 `json:"cost"`
	Stock          *int     `json:"stock"`
	CategoryID     *uint    `json:"category_id"`
	ManufacturerID *uint    `json:"manufacturer_id"`
	IsPublished    *bool    `json:"is_published"`
}

// ProductResponse 商品响应（服务层）。
type ProductResponse struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	ShortDesc      string  `json:"short_desc"`
	FullDesc       string  `json:"full_desc"`
	SKU            string  `json:"sku"`
	Price          float64 `json:"price"`
	OldPrice       float64 `json:"old_price"`
	Cost           float64 `json:"cost"`
	Stock          int     `json:"stock"`
	CategoryID     uint    `json:"category_id"`
	ManufacturerID uint    `json:"manufacturer_id"`
	IsPublished    bool    `json:"is_published"`
	CreatedAt      int64   `json:"created_at"`
	UpdatedAt      int64   `json:"updated_at"`
}

// List 获取商品列表，支持按分类、制造商、关键词筛选。
func (s *ProductService) List(ctx context.Context, page, size int, categoryID, manufacturerID uint, keyword string) ([]ProductResponse, int64, error) {
	products, total, err := s.uc.List(ctx, page, size, categoryID, manufacturerID, keyword)
	if err != nil {
		return nil, 0, err
	}

	items := make([]ProductResponse, len(products))
	for i, p := range products {
		items[i] = toProductResponse(p)
	}
	return items, total, nil
}

// GetByID 根据ID获取商品详情。
func (s *ProductService) GetByID(ctx context.Context, id uint) (*ProductResponse, error) {
	product, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toProductResponse(product)
	return &resp, nil
}

// Create 创建商品。
func (s *ProductService) Create(ctx context.Context, req CreateProductRequest) (*ProductResponse, error) {
	product, err := s.uc.Create(ctx, &biz.Product{
		Name:           req.Name,
		ShortDesc:      req.ShortDesc,
		FullDesc:       req.FullDesc,
		SKU:            req.SKU,
		Price:          req.Price,
		OldPrice:       req.OldPrice,
		Cost:           req.Cost,
		Stock:          req.Stock,
		CategoryID:     req.CategoryID,
		ManufacturerID: req.ManufacturerID,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		return nil, err
	}
	resp := toProductResponse(product)
	return &resp, nil
}

// Update 更新商品信息。
// 仅更新请求中非 nil 的字段，实现部分更新语义。
func (s *ProductService) Update(ctx context.Context, id uint, req UpdateProductRequest) (*ProductResponse, error) {
	// 先获取现有商品
	product, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 仅更新非 nil 字段
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.ShortDesc != nil {
		product.ShortDesc = *req.ShortDesc
	}
	if req.FullDesc != nil {
		product.FullDesc = *req.FullDesc
	}
	if req.SKU != nil {
		product.SKU = *req.SKU
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.OldPrice != nil {
		product.OldPrice = *req.OldPrice
	}
	if req.Cost != nil {
		product.Cost = *req.Cost
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.CategoryID != nil {
		product.CategoryID = *req.CategoryID
	}
	if req.ManufacturerID != nil {
		product.ManufacturerID = *req.ManufacturerID
	}
	if req.IsPublished != nil {
		product.IsPublished = *req.IsPublished
	}

	if err := s.uc.Update(ctx, product); err != nil {
		return nil, err
	}
	resp := toProductResponse(product)
	return &resp, nil
}

// Delete 删除商品。
func (s *ProductService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// Search 搜索商品。
func (s *ProductService) Search(ctx context.Context, keyword string, page, size int) ([]ProductResponse, int64, error) {
	products, total, err := s.uc.Search(ctx, keyword, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]ProductResponse, len(products))
	for i, p := range products {
		items[i] = toProductResponse(p)
	}
	return items, total, nil
}

// GetRecentlyViewed 获取最近浏览商品列表。
func (s *ProductService) GetRecentlyViewed(ctx context.Context, customerID uint, limit int) ([]ProductResponse, error) {
	products, err := s.uc.GetRecentlyViewed(ctx, customerID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]ProductResponse, len(products))
	for i, p := range products {
		items[i] = toProductResponse(p)
	}
	return items, nil
}

// CompareProducts 商品对比。
func (s *ProductService) CompareProducts(ctx context.Context, productIDs []uint) ([]ProductResponse, error) {
	products, err := s.uc.CompareProducts(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	items := make([]ProductResponse, len(products))
	for i, p := range products {
		items[i] = toProductResponse(p)
	}
	return items, nil
}

// toProductResponse 将领域实体转换为服务层响应。
func toProductResponse(p *biz.Product) ProductResponse {
	return ProductResponse{
		ID:             p.ID,
		Name:           p.Name,
		ShortDesc:      p.ShortDesc,
		FullDesc:       p.FullDesc,
		SKU:            p.SKU,
		Price:          p.Price,
		OldPrice:       p.OldPrice,
		Cost:           p.Cost,
		Stock:          p.Stock,
		CategoryID:     p.CategoryID,
		ManufacturerID: p.ManufacturerID,
		IsPublished:    p.IsPublished,
		CreatedAt:      p.CreatedAt.Unix(),
		UpdatedAt:      p.UpdatedAt.Unix(),
	}
}

// ---------------------------------------------------------------------------
// 分类服务
// ---------------------------------------------------------------------------

// CategoryService 分类服务。
type CategoryService struct {
	uc *biz.CategoryUseCase
}

// CategoryResponse 分类响应。
type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    uint   `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// List 获取分类列表。
func (s *CategoryService) List(ctx context.Context, page, size int) ([]CategoryResponse, int64, error) {
	categories, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		items[i] = toCategoryResponse(c)
	}
	return items, total, nil
}

// GetByID 根据ID获取分类详情。
func (s *CategoryService) GetByID(ctx context.Context, id uint) (*CategoryResponse, error) {
	category, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toCategoryResponse(category)
	return &resp, nil
}

// Create 创建分类。
func (s *CategoryService) Create(ctx context.Context, name, description string, parentID uint, sortOrder int, isPublished bool) (*CategoryResponse, error) {
	category, err := s.uc.Create(ctx, &biz.Category{
		Name:        name,
		Description: description,
		ParentID:    parentID,
		SortOrder:   sortOrder,
		IsPublished: isPublished,
	})
	if err != nil {
		return nil, err
	}
	resp := toCategoryResponse(category)
	return &resp, nil
}

func toCategoryResponse(c *biz.Category) CategoryResponse {
	return CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		ParentID:    c.ParentID,
		SortOrder:   c.SortOrder,
		IsPublished: c.IsPublished,
		CreatedAt:   c.CreatedAt.Unix(),
		UpdatedAt:   c.UpdatedAt.Unix(),
	}
}

// ---------------------------------------------------------------------------
// 制造商服务
// ---------------------------------------------------------------------------

// ManufacturerService 制造商服务。
type ManufacturerService struct {
	uc *biz.ManufacturerUseCase
}

// ManufacturerResponse 制造商响应。
type ManufacturerResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// List 获取制造商列表。
func (s *ManufacturerService) List(ctx context.Context, page, size int) ([]ManufacturerResponse, int64, error) {
	manufacturers, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]ManufacturerResponse, len(manufacturers))
	for i, m := range manufacturers {
		items[i] = toManufacturerResponse(m)
	}
	return items, total, nil
}

// GetByID 根据ID获取制造商详情。
func (s *ManufacturerService) GetByID(ctx context.Context, id uint) (*ManufacturerResponse, error) {
	manufacturer, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toManufacturerResponse(manufacturer)
	return &resp, nil
}

// Create 创建制造商。
func (s *ManufacturerService) Create(ctx context.Context, name, description string, isPublished bool) (*ManufacturerResponse, error) {
	manufacturer, err := s.uc.Create(ctx, &biz.Manufacturer{
		Name:        name,
		Description: description,
		IsPublished: isPublished,
	})
	if err != nil {
		return nil, err
	}
	resp := toManufacturerResponse(manufacturer)
	return &resp, nil
}

func toManufacturerResponse(m *biz.Manufacturer) ManufacturerResponse {
	return ManufacturerResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		IsPublished: m.IsPublished,
		CreatedAt:   m.CreatedAt.Unix(),
		UpdatedAt:   m.UpdatedAt.Unix(),
	}
}

// ---------------------------------------------------------------------------
// 商品评论服务
// ---------------------------------------------------------------------------

// ProductReviewService 商品评论服务。
type ProductReviewService struct {
	uc *biz.ProductReviewUseCase
}

// ProductReviewResponse 商品评论响应。
type ProductReviewResponse struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"product_id"`
	CustomerID   uint   `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	Title        string `json:"title"`
	Content      string `json:"content"`
	Rating       int    `json:"rating"`
	IsApproved   bool   `json:"is_approved"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// ListByProductID 获取指定商品的评论列表。
func (s *ProductReviewService) ListByProductID(ctx context.Context, productID uint, page, size int) ([]ProductReviewResponse, int64, error) {
	reviews, total, err := s.uc.ListByProductID(ctx, productID, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]ProductReviewResponse, len(reviews))
	for i, r := range reviews {
		items[i] = toProductReviewResponse(r)
	}
	return items, total, nil
}

// Create 创建商品评论。
func (s *ProductReviewService) Create(ctx context.Context, productID, customerID uint, customerName, title, content string, rating int) (*ProductReviewResponse, error) {
	review, err := s.uc.Create(ctx, &biz.ProductReview{
		ProductID:    productID,
		CustomerID:   customerID,
		CustomerName: customerName,
		Title:        title,
		Content:      content,
		Rating:       rating,
	})
	if err != nil {
		return nil, err
	}
	resp := toProductReviewResponse(review)
	return &resp, nil
}

func toProductReviewResponse(r *biz.ProductReview) ProductReviewResponse {
	return ProductReviewResponse{
		ID:           r.ID,
		ProductID:    r.ProductID,
		CustomerID:   r.CustomerID,
		CustomerName: r.CustomerName,
		Title:        r.Title,
		Content:      r.Content,
		Rating:       r.Rating,
		IsApproved:   r.IsApproved,
		CreatedAt:    r.CreatedAt.Unix(),
		UpdatedAt:    r.UpdatedAt.Unix(),
	}
}
