// Package biz 商品目录业务逻辑层
//
// 中文说明:
// - 定义商品、分类、品牌等用例;
// - 处理业务规则和验证;
// - 协调多个仓储完成复杂业务。
package biz

import (
	"context"
	"errors"

	"nop-go/services/catalog-service/internal/data"
	"nop-go/services/catalog-service/internal/models"
	shareErrors "nop-go/shared/errors"
)

// ProductUseCase 商品用例
//
// 中文说明:
// - 封装商品相关的业务逻辑;
// - 包括商品创建、查询、更新、删除等操作;
// - 处理商品分类关联和状态管理。
type ProductUseCase struct {
	productRepo     data.ProductRepository
	categoryRepo    data.CategoryRepository
	manufacturerRepo data.ManufacturerRepository
	pictureRepo     data.ProductPictureRepository
}

// NewProductUseCase 创建商品用例
func NewProductUseCase(
	productRepo data.ProductRepository,
	categoryRepo data.CategoryRepository,
	manufacturerRepo data.ManufacturerRepository,
	pictureRepo data.ProductPictureRepository,
) *ProductUseCase {
	return &ProductUseCase{
		productRepo:     productRepo,
		categoryRepo:    categoryRepo,
		manufacturerRepo: manufacturerRepo,
		pictureRepo:     pictureRepo,
	}
}

// CreateProductRequest 创建商品请求
type CreateProductRequest struct {
	SKU              string   `json:"sku" binding:"required,min=3,max=64"`
	Name             string   `json:"name" binding:"required,min=1,max=256"`
	ShortDescription string   `json:"short_description" binding:"max=512"`
	FullDescription  string   `json:"full_description"`
	ManufacturerID   *uint64  `json:"manufacturer_id"`
	ProductType      string   `json:"product_type"` // simple, grouped
	CategoryIDs      []uint64 `json:"category_ids"`
	PrimaryCategory  *uint64  `json:"primary_category_id"`
	IsPublished      bool     `json:"is_published"`
	ShowOnHomepage   bool     `json:"show_on_homepage"`
	SEOSlug          string   `json:"seo_slug"`
	MetaKeywords     string   `json:"meta_keywords"`
	MetaDescription  string   `json:"meta_description"`
}

// CreateProduct 创建商品
func (uc *ProductUseCase) CreateProduct(ctx context.Context, req *CreateProductRequest) (*models.Product, error) {
	// 检查 SKU 是否已存在
	if existing, _ := uc.productRepo.GetBySKU(ctx, req.SKU); existing != nil {
		return nil, shareErrors.ErrDuplicateSku
	}

	// 验证品牌是否存在
	if req.ManufacturerID != nil {
		if _, err := uc.manufacturerRepo.GetByID(ctx, *req.ManufacturerID); err != nil {
			return nil, shareErrors.ErrManufacturerNotFound
		}
	}

	// 验证分类是否存在
	for _, catID := range req.CategoryIDs {
		if _, err := uc.categoryRepo.GetByID(ctx, catID); err != nil {
			return nil, shareErrors.ErrCategoryNotFound
		}
	}

	product := &models.Product{
		SKU:              req.SKU,
		Name:             req.Name,
		ShortDescription: req.ShortDescription,
		FullDescription:  req.FullDescription,
		ManufacturerID:   req.ManufacturerID,
		ProductType:      req.ProductType,
		IsPublished:      req.IsPublished,
		ShowOnHomepage:   req.ShowOnHomepage,
		SEOSlug:          req.SEOSlug,
		MetaKeywords:     req.MetaKeywords,
		MetaDescription:  req.MetaDescription,
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	// 添加分类关联
	for _, catID := range req.CategoryIDs {
		isPrimary := req.PrimaryCategory != nil && catID == *req.PrimaryCategory
		uc.productRepo.AddCategory(ctx, product.ID, catID, isPrimary)
	}

	return product, nil
}

// GetProduct 根据ID获取商品
func (uc *ProductUseCase) GetProduct(ctx context.Context, id uint64) (*models.Product, error) {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrProductNotFound
	}
	return product, nil
}

// GetProductBySKU 根据SKU获取商品
func (uc *ProductUseCase) GetProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
	product, err := uc.productRepo.GetBySKU(ctx, sku)
	if err != nil {
		return nil, shareErrors.ErrProductNotFound
	}
	return product, nil
}

// UpdateProduct 更新商品
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, id uint64, req *CreateProductRequest) (*models.Product, error) {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrProductNotFound
	}

	// 如果 SKU 改变,检查是否冲突
	if req.SKU != "" && req.SKU != product.SKU {
		if existing, _ := uc.productRepo.GetBySKU(ctx, req.SKU); existing != nil {
			return nil, shareErrors.ErrDuplicateSku
		}
		product.SKU = req.SKU
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.ShortDescription != "" {
		product.ShortDescription = req.ShortDescription
	}
	if req.FullDescription != "" {
		product.FullDescription = req.FullDescription
	}
	if req.ManufacturerID != nil {
		product.ManufacturerID = req.ManufacturerID
	}
	if req.ProductType != "" {
		product.ProductType = req.ProductType
	}
	product.IsPublished = req.IsPublished
	product.ShowOnHomepage = req.ShowOnHomepage
	if req.SEOSlug != "" {
		product.SEOSlug = req.SEOSlug
	}
	if req.MetaKeywords != "" {
		product.MetaKeywords = req.MetaKeywords
	}
	if req.MetaDescription != "" {
		product.MetaDescription = req.MetaDescription
	}

	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// DeleteProduct 删除商品
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, id uint64) error {
	if _, err := uc.productRepo.GetByID(ctx, id); err != nil {
		return shareErrors.ErrProductNotFound
	}
	return uc.productRepo.Delete(ctx, id)
}

// ListProducts 商品列表
func (uc *ProductUseCase) ListProducts(ctx context.Context, req *models.ProductListRequest) ([]*models.Product, int64, error) {
	return uc.productRepo.List(ctx, req)
}

// GetProductsByCategory 根据分类获取商品
func (uc *ProductUseCase) GetProductsByCategory(ctx context.Context, categoryID uint64, page, pageSize int) ([]*models.Product, int64, error) {
	// 验证分类是否存在
	if _, err := uc.categoryRepo.GetByID(ctx, categoryID); err != nil {
		return nil, 0, shareErrors.ErrCategoryNotFound
	}
	return uc.productRepo.GetByCategoryID(ctx, categoryID, page, pageSize)
}

// GetProductsByManufacturer 根据品牌获取商品
func (uc *ProductUseCase) GetProductsByManufacturer(ctx context.Context, manufacturerID uint64, page, pageSize int) ([]*models.Product, int64, error) {
	// 验证品牌是否存在
	if _, err := uc.manufacturerRepo.GetByID(ctx, manufacturerID); err != nil {
		return nil, 0, shareErrors.ErrManufacturerNotFound
	}
	return uc.productRepo.GetByManufacturerID(ctx, manufacturerID, page, pageSize)
}

// GetHomepageProducts 获取首页商品
func (uc *ProductUseCase) GetHomepageProducts(ctx context.Context, limit int) ([]*models.Product, error) {
	if limit <= 0 {
		limit = 10
	}
	return uc.productRepo.GetHomepageProducts(ctx, limit)
}

// PublishProduct 发布商品
func (uc *ProductUseCase) PublishProduct(ctx context.Context, id uint64) error {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return shareErrors.ErrProductNotFound
	}
	product.IsPublished = true
	return uc.productRepo.Update(ctx, product)
}

// UnpublishProduct 下架商品
func (uc *ProductUseCase) UnpublishProduct(ctx context.Context, id uint64) error {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return shareErrors.ErrProductNotFound
	}
	product.IsPublished = false
	return uc.productRepo.Update(ctx, product)
}

// CategoryUseCase 分类用例
type CategoryUseCase struct {
	categoryRepo data.CategoryRepository
}

// NewCategoryUseCase 创建分类用例
func NewCategoryUseCase(categoryRepo data.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{categoryRepo: categoryRepo}
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name         string  `json:"name" binding:"required,min=1,max=256"`
	ParentID     *uint64 `json:"parent_id"`
	DisplayOrder int     `json:"display_order"`
	IsPublished  bool    `json:"is_published"`
	SEOSlug      string  `json:"seo_slug"`
	Description  string  `json:"description"`
}

// CreateCategory 创建分类
func (uc *CategoryUseCase) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*models.Category, error) {
	// 计算层级
	level := 1
	if req.ParentID != nil {
		parent, err := uc.categoryRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, shareErrors.ErrCategoryNotFound
		}
		level = parent.Level + 1
	}

	category := &models.Category{
		Name:         req.Name,
		ParentID:     req.ParentID,
		Level:        level,
		DisplayOrder: req.DisplayOrder,
		IsPublished:  req.IsPublished,
		SEOSlug:      req.SEOSlug,
		Description:  req.Description,
	}

	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory 根据ID获取分类
func (uc *CategoryUseCase) GetCategory(ctx context.Context, id uint64) (*models.Category, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrCategoryNotFound
	}
	return category, nil
}

// UpdateCategory 更新分类
func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, id uint64, req *CreateCategoryRequest) (*models.Category, error) {
	category, err := uc.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrCategoryNotFound
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.DisplayOrder >= 0 {
		category.DisplayOrder = req.DisplayOrder
	}
	category.IsPublished = req.IsPublished
	if req.SEOSlug != "" {
		category.SEOSlug = req.SEOSlug
	}
	if req.Description != "" {
		category.Description = req.Description
	}

	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory 删除分类
func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id uint64) error {
	// 检查是否有子分类
	children, err := uc.categoryRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return errors.New("分类下存在子分类,无法删除")
	}

	return uc.categoryRepo.Delete(ctx, id)
}

// ListCategories 分类列表
func (uc *CategoryUseCase) ListCategories(ctx context.Context) ([]*models.Category, error) {
	return uc.categoryRepo.List(ctx)
}

// GetCategoryTree 获取分类树
func (uc *CategoryUseCase) GetCategoryTree(ctx context.Context) ([]*models.Category, error) {
	return uc.categoryRepo.GetTree(ctx)
}

// ManufacturerUseCase 品牌用例
type ManufacturerUseCase struct {
	manufacturerRepo data.ManufacturerRepository
}

// NewManufacturerUseCase 创建品牌用例
func NewManufacturerUseCase(manufacturerRepo data.ManufacturerRepository) *ManufacturerUseCase {
	return &ManufacturerUseCase{manufacturerRepo: manufacturerRepo}
}

// CreateManufacturerRequest 创建品牌请求
type CreateManufacturerRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=256"`
	Description  string `json:"description"`
	LogoURL      string `json:"logo_url"`
	IsPublished  bool   `json:"is_published"`
	DisplayOrder int    `json:"display_order"`
	SEOSlug      string `json:"seo_slug"`
}

// CreateManufacturer 创建品牌
func (uc *ManufacturerUseCase) CreateManufacturer(ctx context.Context, req *CreateManufacturerRequest) (*models.Manufacturer, error) {
	manufacturer := &models.Manufacturer{
		Name:         req.Name,
		Description:  req.Description,
		LogoURL:      req.LogoURL,
		IsPublished:  req.IsPublished,
		DisplayOrder: req.DisplayOrder,
		SEOSlug:      req.SEOSlug,
	}

	if err := uc.manufacturerRepo.Create(ctx, manufacturer); err != nil {
		return nil, err
	}

	return manufacturer, nil
}

// GetManufacturer 根据ID获取品牌
func (uc *ManufacturerUseCase) GetManufacturer(ctx context.Context, id uint64) (*models.Manufacturer, error) {
	manufacturer, err := uc.manufacturerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrManufacturerNotFound
	}
	return manufacturer, nil
}

// UpdateManufacturer 更新品牌
func (uc *ManufacturerUseCase) UpdateManufacturer(ctx context.Context, id uint64, req *CreateManufacturerRequest) (*models.Manufacturer, error) {
	manufacturer, err := uc.manufacturerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrManufacturerNotFound
	}

	if req.Name != "" {
		manufacturer.Name = req.Name
	}
	if req.Description != "" {
		manufacturer.Description = req.Description
	}
	if req.LogoURL != "" {
		manufacturer.LogoURL = req.LogoURL
	}
	manufacturer.IsPublished = req.IsPublished
	manufacturer.DisplayOrder = req.DisplayOrder
	if req.SEOSlug != "" {
		manufacturer.SEOSlug = req.SEOSlug
	}

	if err := uc.manufacturerRepo.Update(ctx, manufacturer); err != nil {
		return nil, err
	}

	return manufacturer, nil
}

// DeleteManufacturer 删除品牌
func (uc *ManufacturerUseCase) DeleteManufacturer(ctx context.Context, id uint64) error {
	return uc.manufacturerRepo.Delete(ctx, id)
}

// ListManufacturers 品牌列表
func (uc *ManufacturerUseCase) ListManufacturers(ctx context.Context) ([]*models.Manufacturer, error) {
	return uc.manufacturerRepo.List(ctx)
}

// ProductPictureUseCase 商品图片用例
type ProductPictureUseCase struct {
	pictureRepo  data.ProductPictureRepository
	productRepo  data.ProductRepository
}

// NewProductPictureUseCase 创建商品图片用例
func NewProductPictureUseCase(pictureRepo data.ProductPictureRepository, productRepo data.ProductRepository) *ProductPictureUseCase {
	return &ProductPictureUseCase{pictureRepo: pictureRepo, productRepo: productRepo}
}

// AddPictureRequest 添加图片请求
type AddPictureRequest struct {
	ProductID    uint64 `json:"product_id" binding:"required"`
	PictureURL   string `json:"picture_url" binding:"required"`
	AltText      string `json:"alt_text"`
	DisplayOrder int    `json:"display_order"`
	IsMain       bool   `json:"is_main"`
}

// AddPicture 添加商品图片
func (uc *ProductPictureUseCase) AddPicture(ctx context.Context, req *AddPictureRequest) (*models.ProductPicture, error) {
	// 验证商品是否存在
	if _, err := uc.productRepo.GetByID(ctx, req.ProductID); err != nil {
		return nil, shareErrors.ErrProductNotFound
	}

	picture := &models.ProductPicture{
		ProductID:    req.ProductID,
		PictureURL:   req.PictureURL,
		AltText:      req.AltText,
		DisplayOrder: req.DisplayOrder,
		IsMain:       req.IsMain,
	}

	if err := uc.pictureRepo.Create(ctx, picture); err != nil {
		return nil, err
	}

	// 如果设置为主图,清除其他主图
	if req.IsMain {
		uc.pictureRepo.SetMain(ctx, req.ProductID, picture.ID)
	}

	return picture, nil
}

// GetPictures 获取商品图片列表
func (uc *ProductPictureUseCase) GetPictures(ctx context.Context, productID uint64) ([]*models.ProductPicture, error) {
	return uc.pictureRepo.GetByProductID(ctx, productID)
}

// DeletePicture 删除商品图片
func (uc *ProductPictureUseCase) DeletePicture(ctx context.Context, id uint64) error {
	return uc.pictureRepo.Delete(ctx, id)
}

// SetMainPicture 设置主图
func (uc *ProductPictureUseCase) SetMainPicture(ctx context.Context, productID, pictureID uint64) error {
	return uc.pictureRepo.SetMain(ctx, productID, pictureID)
}

// ProductReviewUseCase 商品评论用例
type ProductReviewUseCase struct {
	reviewRepo  data.ProductReviewRepository
	productRepo data.ProductRepository
}

// NewProductReviewUseCase 创建商品评论用例
func NewProductReviewUseCase(reviewRepo data.ProductReviewRepository, productRepo data.ProductRepository) *ProductReviewUseCase {
	return &ProductReviewUseCase{reviewRepo: reviewRepo, productRepo: productRepo}
}

// CreateReviewRequest 创建评论请求
type CreateReviewRequest struct {
	ProductID  uint64 `json:"product_id" binding:"required"`
	CustomerID uint64 `json:"customer_id" binding:"required"`
	Rating     int    `json:"rating" binding:"required,min=1,max=5"`
	Title      string `json:"title"`
	ReviewText string `json:"review_text" binding:"required"`
}

// CreateReview 创建商品评论
func (uc *ProductReviewUseCase) CreateReview(ctx context.Context, req *CreateReviewRequest) (*models.ProductReview, error) {
	// 验证商品是否存在
	if _, err := uc.productRepo.GetByID(ctx, req.ProductID); err != nil {
		return nil, shareErrors.ErrProductNotFound
	}

	// 验证评分范围
	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.New("评分必须在 1-5 之间")
	}

	review := &models.ProductReview{
		ProductID:  req.ProductID,
		CustomerID: req.CustomerID,
		Rating:     req.Rating,
		Title:      req.Title,
		ReviewText: req.ReviewText,
		IsApproved: false, // 默认需要审核
	}

	if err := uc.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	return review, nil
}

// GetReviews 获取商品评论列表
func (uc *ProductReviewUseCase) GetReviews(ctx context.Context, productID uint64, page, pageSize int) ([]*models.ProductReview, int64, error) {
	return uc.reviewRepo.GetByProductID(ctx, productID, page, pageSize)
}

// ApproveReview 批准评论
func (uc *ProductReviewUseCase) ApproveReview(ctx context.Context, id uint64) error {
	return uc.reviewRepo.Approve(ctx, id)
}

// DeleteReview 删除评论
func (uc *ProductReviewUseCase) DeleteReview(ctx context.Context, id uint64) error {
	return uc.reviewRepo.Delete(ctx, id)
}

// MarkReviewHelpful 标记评论有用
func (uc *ProductReviewUseCase) MarkReviewHelpful(ctx context.Context, id uint64, helpful bool) error {
	return uc.reviewRepo.MarkHelpful(ctx, id, helpful)
}