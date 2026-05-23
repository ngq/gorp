// Package service 服务层。
// 封装业务用例调用，提供面向 handler 的服务接口。
// 合并 product + directory + media + seo 四个服务，
// 统一通过 Services 结构对外暴露。
package service

import (
	"context"
	"time"

	"nop-go/services/catalog-service/internal/biz"
	"nop-go/services/catalog-service/internal/data"

	"gorm.io/gorm"
)

// Services 聚合所有服务实例，供路由注册时统一注入。
// 合并四个子服务：
//   - Product: 商品、分类、制造商、评论、最近浏览、对比、搜索
//   - Directory: 国家、省/州、货币
//   - Media: 媒体文件上传/查询/删除
//   - Seo: SEO 元数据管理
type Services struct {
	Product   *ProductService     // 商品服务（含分类、制造商、评论）
	Directory *DirectoryService   // 目录服务（含国家、省/州、货币）
	Media     *MediaService       // 媒体服务
	Seo       *SeoService         // SEO 服务
}

// NewServices 创建所有服务实例。
// 依次初始化数据层仓储 -> 业务用例 -> 服务层，完成依赖注入链路。
func NewServices(db *gorm.DB) *Services {
	// ==================== Product 子服务 ====================
	productRepo := data.NewProductRepo(db)
	categoryRepo := data.NewCategoryRepo(db)
	manufacturerRepo := data.NewManufacturerRepo(db)
	reviewRepo := data.NewProductReviewRepo(db)
	recentlyViewedRepo := data.NewRecentlyViewedRepo(db)

	productUC := biz.NewProductUseCase(productRepo, recentlyViewedRepo)
	categoryUC := biz.NewCategoryUseCase(categoryRepo)
	manufacturerUC := biz.NewManufacturerUseCase(manufacturerRepo)
	reviewUC := biz.NewProductReviewUseCase(reviewRepo)

	// ==================== Directory 子服务 ====================
	countryRepo := data.NewCountryRepo(db)
	stateRepo := data.NewStateRepo(db)
	currencyRepo := data.NewCurrencyRepo(db)

	countryUC := biz.NewCountryUseCase(countryRepo)
	stateUC := biz.NewStateUseCase(stateRepo)
	currencyUC := biz.NewCurrencyUseCase(currencyRepo)

	// ==================== Media 子服务 ====================
	mediaRepo := data.NewMediaRepo(db)
	mediaUC := biz.NewMediaUseCase(mediaRepo)

	// ==================== Seo 子服务 ====================
	seoMetaRepo := data.NewSeoMetaRepo(db)
	seoUC := biz.NewSeoUseCase(seoMetaRepo)

	return &Services{
		Product: &ProductService{
			productUC:       productUC,
			categoryUC:      categoryUC,
			manufacturerUC:  manufacturerUC,
			reviewUC:        reviewUC,
		},
		Directory: &DirectoryService{
			countryUC:  countryUC,
			stateUC:    stateUC,
			currencyUC: currencyUC,
		},
		Media: &MediaService{
			uc: mediaUC,
		},
		Seo: &SeoService{
			uc: seoUC,
		},
	}
}

	// ===========================================================================
// Product 子服务
// ===========================================================================

// ProductService 商品服务，封装商品相关业务逻辑。
// 包含商品 CRUD、搜索、最近浏览、对比，以及分类和制造商管理。
type ProductService struct {
	productUC      *biz.ProductUseCase
	categoryUC     *biz.CategoryUseCase
	manufacturerUC *biz.ManufacturerUseCase
	reviewUC       *biz.ProductReviewUseCase
}

// --- 商品请求/响应 ---

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

// --- 商品方法 ---

// List 获取商品列表，支持按分类、制造商、关键词筛选。
func (s *ProductService) List(ctx context.Context, page, size int, categoryID, manufacturerID uint, keyword string) ([]ProductResponse, int64, error) {
	products, total, err := s.productUC.List(ctx, page, size, categoryID, manufacturerID, keyword)
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
	product, err := s.productUC.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toProductResponse(product)
	return &resp, nil
}

// Create 创建商品。
func (s *ProductService) Create(ctx context.Context, req CreateProductRequest) (*ProductResponse, error) {
	product, err := s.productUC.Create(ctx, &biz.Product{
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

// Update 更新商品信息。仅更新请求中非 nil 的字段，实现部分更新语义。
func (s *ProductService) Update(ctx context.Context, id uint, req UpdateProductRequest) (*ProductResponse, error) {
	product, err := s.productUC.GetByID(ctx, id)
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
	if err := s.productUC.Update(ctx, product); err != nil {
		return nil, err
	}
	resp := toProductResponse(product)
	return &resp, nil
}

// Delete 删除商品。
func (s *ProductService) Delete(ctx context.Context, id uint) error {
	return s.productUC.Delete(ctx, id)
}

// Search 搜索商品。
func (s *ProductService) Search(ctx context.Context, keyword string, page, size int) ([]ProductResponse, int64, error) {
	products, total, err := s.productUC.Search(ctx, keyword, page, size)
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
	products, err := s.productUC.GetRecentlyViewed(ctx, customerID, limit)
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
	products, err := s.productUC.CompareProducts(ctx, productIDs)
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

// --- 分类请求/响应 ---

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

// --- 分类方法 ---

// ListCategories 获取分类列表。
func (s *ProductService) ListCategories(ctx context.Context, page, size int) ([]CategoryResponse, int64, error) {
	categories, total, err := s.categoryUC.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]CategoryResponse, len(categories))
	for i, c := range categories {
		items[i] = toCategoryResponse(c)
	}
	return items, total, nil
}

// GetCategoryByID 根据ID获取分类详情。
func (s *ProductService) GetCategoryByID(ctx context.Context, id uint) (*CategoryResponse, error) {
	category, err := s.categoryUC.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toCategoryResponse(category)
	return &resp, nil
}

// CreateCategory 创建分类。
func (s *ProductService) CreateCategory(ctx context.Context, name, description string, parentID uint, sortOrder int, isPublished bool) (*CategoryResponse, error) {
	category, err := s.categoryUC.Create(ctx, &biz.Category{
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

// --- 制造商请求/响应 ---

// ManufacturerResponse 制造商响应。
type ManufacturerResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublished bool   `json:"is_published"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// --- 制造商方法 ---

// ListManufacturers 获取制造商列表。
func (s *ProductService) ListManufacturers(ctx context.Context, page, size int) ([]ManufacturerResponse, int64, error) {
	manufacturers, total, err := s.manufacturerUC.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]ManufacturerResponse, len(manufacturers))
	for i, m := range manufacturers {
		items[i] = toManufacturerResponse(m)
	}
	return items, total, nil
}

// GetManufacturerByID 根据ID获取制造商详情。
func (s *ProductService) GetManufacturerByID(ctx context.Context, id uint) (*ManufacturerResponse, error) {
	manufacturer, err := s.manufacturerUC.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toManufacturerResponse(manufacturer)
	return &resp, nil
}

// CreateManufacturer 创建制造商。
func (s *ProductService) CreateManufacturer(ctx context.Context, name, description string, isPublished bool) (*ManufacturerResponse, error) {
	manufacturer, err := s.manufacturerUC.Create(ctx, &biz.Manufacturer{
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

// --- 商品评论请求/响应 ---

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

// --- 商品评论方法 ---

// ListProductReviews 获取指定商品的评论列表。
func (s *ProductService) ListProductReviews(ctx context.Context, productID uint, page, size int) ([]ProductReviewResponse, int64, error) {
	reviews, total, err := s.reviewUC.ListByProductID(ctx, productID, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]ProductReviewResponse, len(reviews))
	for i, r := range reviews {
		items[i] = toProductReviewResponse(r)
	}
	return items, total, nil
}

// CreateProductReview 创建商品评论。
func (s *ProductService) CreateProductReview(ctx context.Context, productID, customerID uint, customerName, title, content string, rating int) (*ProductReviewResponse, error) {
	review, err := s.reviewUC.Create(ctx, &biz.ProductReview{
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

// ===========================================================================
// Directory 子服务
// ===========================================================================

// DirectoryService 目录服务，聚合国家/省/州/货币。
type DirectoryService struct {
	countryUC  *biz.CountryUseCase
	stateUC    *biz.StateUseCase
	currencyUC *biz.CurrencyUseCase
}

// --- 国家请求/响应 ---

// CreateCountryRequest 创建国家请求。
type CreateCountryRequest struct {
	Name             string `json:"name" binding:"required"`
	IsoCode2         string `json:"iso_code2" binding:"required"`
	IsoCode3         string `json:"iso_code3"`
	AddressFormat    string `json:"address_format"`
	PostcodeRequired bool   `json:"postcode_required"`
}

// UpdateCountryRequest 更新国家请求。
type UpdateCountryRequest struct {
	Name             string `json:"name" binding:"required"`
	IsoCode2         string `json:"iso_code2" binding:"required"`
	IsoCode3         string `json:"iso_code3"`
	AddressFormat    string `json:"address_format"`
	PostcodeRequired bool   `json:"postcode_required"`
}

// CountryResponse 国家响应。
type CountryResponse struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	IsoCode2         string `json:"iso_code2"`
	IsoCode3         string `json:"iso_code3"`
	AddressFormat    string `json:"address_format"`
	PostcodeRequired bool   `json:"postcode_required"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// --- 省/州请求/响应 ---

// CreateStateRequest 创建省/州请求。
type CreateStateRequest struct {
	CountryID uint   `json:"country_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	IsoCode   string `json:"iso_code"`
}

// UpdateStateRequest 更新省/州请求。
type UpdateStateRequest struct {
	Name    string `json:"name" binding:"required"`
	IsoCode string `json:"iso_code"`
}

// StateResponse 省/州响应。
type StateResponse struct {
	ID        uint   `json:"id"`
	CountryID uint   `json:"country_id"`
	Name      string `json:"name"`
	IsoCode   string `json:"iso_code"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// --- 货币请求/响应 ---

// CreateCurrencyRequest 创建货币请求。
type CreateCurrencyRequest struct {
	Name     string  `json:"name" binding:"required"`
	Code     string  `json:"code" binding:"required"`
	Symbol   string  `json:"symbol"`
	Rate     float64 `json:"rate"`
	IsActive bool    `json:"is_active"`
}

// UpdateCurrencyRequest 更新货币请求。
type UpdateCurrencyRequest struct {
	Name     string  `json:"name" binding:"required"`
	Code     string  `json:"code" binding:"required"`
	Symbol   string  `json:"symbol"`
	Rate     float64 `json:"rate"`
	IsActive bool    `json:"is_active"`
}

// CurrencyResponse 货币响应。
type CurrencyResponse struct {
	ID        uint    `json:"id"`
	Name      string  `json:"name"`
	Code      string  `json:"code"`
	Symbol    string  `json:"symbol"`
	Rate      float64 `json:"rate"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CurrencyRateItem 汇率更新项。
type CurrencyRateItem struct {
	CurrencyID uint    `json:"currency_id"`
	Rate       float64 `json:"rate"`
}

// --- 国家方法 ---

// ListCountries 获取国家列表。
func (s *DirectoryService) ListCountries(ctx context.Context, page, size int) ([]CountryResponse, int64, error) {
	countries, total, err := s.countryUC.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]CountryResponse, len(countries))
	for i, c := range countries {
		items[i] = CountryResponse{
			ID: c.ID, Name: c.Name, IsoCode2: c.IsoCode2, IsoCode3: c.IsoCode3,
			AddressFormat: c.AddressFormat, PostcodeRequired: c.PostcodeRequired,
			CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}

// CreateCountry 创建国家。
func (s *DirectoryService) CreateCountry(ctx context.Context, req CreateCountryRequest) (*CountryResponse, error) {
	c, err := s.countryUC.Create(ctx, req.Name, req.IsoCode2, req.IsoCode3, req.AddressFormat, req.PostcodeRequired)
	if err != nil {
		return nil, err
	}
	return &CountryResponse{
		ID: c.ID, Name: c.Name, IsoCode2: c.IsoCode2, IsoCode3: c.IsoCode3,
		AddressFormat: c.AddressFormat, PostcodeRequired: c.PostcodeRequired,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateCountry 更新国家。
func (s *DirectoryService) UpdateCountry(ctx context.Context, id uint, req UpdateCountryRequest) (*CountryResponse, error) {
	c, err := s.countryUC.Update(ctx, id, req.Name, req.IsoCode2, req.IsoCode3, req.AddressFormat, req.PostcodeRequired)
	if err != nil {
		return nil, err
	}
	return &CountryResponse{
		ID: c.ID, Name: c.Name, IsoCode2: c.IsoCode2, IsoCode3: c.IsoCode3,
		AddressFormat: c.AddressFormat, PostcodeRequired: c.PostcodeRequired,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: c.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteCountry 删除国家。
func (s *DirectoryService) DeleteCountry(ctx context.Context, id uint) error {
	return s.countryUC.Delete(ctx, id)
}

// --- 省/州方法 ---

// ListStates 获取指定国家下的省/州列表。
func (s *DirectoryService) ListStates(ctx context.Context, countryID uint, page, size int) ([]StateResponse, int64, error) {
	states, total, err := s.stateUC.ListByCountryID(ctx, countryID, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]StateResponse, len(states))
	for i, st := range states {
		items[i] = StateResponse{
			ID: st.ID, CountryID: st.CountryID, Name: st.Name, IsoCode: st.IsoCode,
			CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}

// CreateState 创建省/州。
func (s *DirectoryService) CreateState(ctx context.Context, req CreateStateRequest) (*StateResponse, error) {
	st, err := s.stateUC.Create(ctx, req.CountryID, req.Name, req.IsoCode)
	if err != nil {
		return nil, err
	}
	return &StateResponse{
		ID: st.ID, CountryID: st.CountryID, Name: st.Name, IsoCode: st.IsoCode,
		CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateState 更新省/州。
func (s *DirectoryService) UpdateState(ctx context.Context, id uint, req UpdateStateRequest) (*StateResponse, error) {
	st, err := s.stateUC.Update(ctx, id, req.Name, req.IsoCode)
	if err != nil {
		return nil, err
	}
	return &StateResponse{
		ID: st.ID, CountryID: st.CountryID, Name: st.Name, IsoCode: st.IsoCode,
		CreatedAt: st.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: st.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteState 删除省/州。
func (s *DirectoryService) DeleteState(ctx context.Context, id uint) error {
	return s.stateUC.Delete(ctx, id)
}

// --- 货币方法 ---

// ListCurrencies 获取货币列表。
func (s *DirectoryService) ListCurrencies(ctx context.Context, page, size int) ([]CurrencyResponse, int64, error) {
	currencies, total, err := s.currencyUC.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]CurrencyResponse, len(currencies))
	for i, cu := range currencies {
		items[i] = CurrencyResponse{
			ID: cu.ID, Name: cu.Name, Code: cu.Code, Symbol: cu.Symbol,
			Rate: cu.Rate, IsActive: cu.IsActive,
			CreatedAt: cu.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: cu.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	return items, total, nil
}

// CreateCurrency 创建货币。
func (s *DirectoryService) CreateCurrency(ctx context.Context, req CreateCurrencyRequest) (*CurrencyResponse, error) {
	cu, err := s.currencyUC.Create(ctx, req.Name, req.Code, req.Symbol, req.Rate, req.IsActive)
	if err != nil {
		return nil, err
	}
	return &CurrencyResponse{
		ID: cu.ID, Name: cu.Name, Code: cu.Code, Symbol: cu.Symbol,
		Rate: cu.Rate, IsActive: cu.IsActive,
		CreatedAt: cu.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: cu.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// UpdateCurrency 更新货币。
func (s *DirectoryService) UpdateCurrency(ctx context.Context, id uint, req UpdateCurrencyRequest) (*CurrencyResponse, error) {
	cu, err := s.currencyUC.Update(ctx, id, req.Name, req.Code, req.Symbol, req.Rate, req.IsActive)
	if err != nil {
		return nil, err
	}
	return &CurrencyResponse{
		ID: cu.ID, Name: cu.Name, Code: cu.Code, Symbol: cu.Symbol,
		Rate: cu.Rate, IsActive: cu.IsActive,
		CreatedAt: cu.CreatedAt.Format("2006-01-02 15:04:05"), UpdatedAt: cu.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// DeleteCurrency 删除货币。
func (s *DirectoryService) DeleteCurrency(ctx context.Context, id uint) error {
	return s.currencyUC.Delete(ctx, id)
}

// ApplyRates 批量应用汇率更新。
func (s *DirectoryService) ApplyRates(ctx context.Context, rates []CurrencyRateItem) error {
	bizRates := make([]biz.CurrencyRateItem, len(rates))
	for i, r := range rates {
		bizRates[i] = biz.CurrencyRateItem{CurrencyID: r.CurrencyID, Rate: r.Rate}
	}
	return s.currencyUC.ApplyRates(ctx, bizRates)
}

// ===========================================================================
// Media 子服务
// ===========================================================================

// MediaService 媒体服务。
type MediaService struct {
	uc *biz.MediaUseCase
}

// UploadMediaRequest 上传图片请求。
type UploadMediaRequest struct {
	FileName string `json:"file_name" binding:"required"`
	MimeType string `json:"mime_type" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required"`
	FileURL  string `json:"file_url" binding:"required"`
	AltText  string `json:"alt_text"`
}

// MediaResponse 媒体响应。
type MediaResponse struct {
	ID        uint      `json:"id"`
	FileName  string    `json:"file_name"`
	MimeType  string    `json:"mime_type"`
	FileSize  int64     `json:"file_size"`
	FileURL   string    `json:"file_url"`
	AltText   string    `json:"alt_text"`
	CreatedAt time.Time `json:"created_at"`
}

// Upload 异步上传图片。
func (s *MediaService) Upload(ctx context.Context, req UploadMediaRequest) (*MediaResponse, error) {
	media, err := s.uc.Upload(ctx, req.FileName, req.MimeType, req.FileSize, req.FileURL, req.AltText)
	if err != nil {
		return nil, err
	}
	return &MediaResponse{
		ID:        media.ID,
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		FileSize:  media.FileSize,
		FileURL:   media.FileURL,
		AltText:   media.AltText,
		CreatedAt: media.CreatedAt,
	}, nil
}

// GetByID 根据ID获取媒体。
func (s *MediaService) GetByID(ctx context.Context, id uint) (*MediaResponse, error) {
	media, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &MediaResponse{
		ID:        media.ID,
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		FileSize:  media.FileSize,
		FileURL:   media.FileURL,
		AltText:   media.AltText,
		CreatedAt: media.CreatedAt,
	}, nil
}

// Delete 删除媒体。
func (s *MediaService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// ===========================================================================
// Seo 子服务
// ===========================================================================

// SeoService SEO 服务。
type SeoService struct {
	uc *biz.SeoUseCase
}

// CreateSeoRequest 创建 SEO 元数据请求。
type CreateSeoRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// SeoResponse SEO 元数据响应。
type SeoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// List 获取 SEO 元数据列表。
func (s *SeoService) List(ctx context.Context, page, size int) ([]SeoResponse, int64, error) {
	seos, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}
	items := make([]SeoResponse, len(seos))
	for i, u := range seos {
		items[i] = SeoResponse{ID: u.ID, Username: u.Username, Email: u.Email}
	}
	return items, total, nil
}

// GetByID 根据ID获取 SEO 元数据。
func (s *SeoService) GetByID(ctx context.Context, id uint) (*SeoResponse, error) {
	seo, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SeoResponse{ID: seo.ID, Username: seo.Username, Email: seo.Email}, nil
}

// Create 创建 SEO 元数据。
func (s *SeoService) Create(ctx context.Context, req CreateSeoRequest) (*SeoResponse, error) {
	seo, err := s.uc.Create(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	return &SeoResponse{ID: seo.ID, Username: seo.Username, Email: seo.Email}, nil
}

// Delete 删除 SEO 元数据。
func (s *SeoService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}
