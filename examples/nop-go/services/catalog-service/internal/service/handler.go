// Package service 商品目录HTTP层
//
// 中文说明:
// - 定义商品、分类、品牌等HTTP处理器;
// - 处理请求参数绑定和响应转换;
// - 注册路由到 Gin Engine。
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/catalog-service/internal/biz"
	"nop-go/services/catalog-service/internal/models"
	shareModels "nop-go/shared/models"

	"github.com/gin-gonic/gin"
)

// CatalogService 商品目录服务
type CatalogService struct {
	productUC      *biz.ProductUseCase
	categoryUC     *biz.CategoryUseCase
	manufacturerUC *biz.ManufacturerUseCase
	pictureUC      *biz.ProductPictureUseCase
	reviewUC       *biz.ProductReviewUseCase
	enableMedia    bool
	enableReview   bool
}

// Options 控制 CatalogService 的轻量/完整模式。
type Options struct {
	EnableMedia  bool
	EnableReview bool
}

// NewCatalogService 创建商品目录服务
func NewCatalogService(
	productUC *biz.ProductUseCase,
	categoryUC *biz.CategoryUseCase,
	manufacturerUC *biz.ManufacturerUseCase,
	pictureUC *biz.ProductPictureUseCase,
	reviewUC *biz.ProductReviewUseCase,
	opts Options,
) *CatalogService {
	return &CatalogService{
		productUC:      productUC,
		categoryUC:     categoryUC,
		manufacturerUC: manufacturerUC,
		pictureUC:      pictureUC,
		reviewUC:       reviewUC,
		enableMedia:    opts.EnableMedia,
		enableReview:   opts.EnableReview,
	}
}

// RegisterRoutes 注册路由
//
// 中文说明:
// - 将业务路由挂载到 /api/v1;
// - 分为商品、分类、品牌、图片、评论等模块。
func (s *CatalogService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		// 商品管理
		products := api.Group("/products")
		{
			products.GET("", s.ListProducts)
			products.GET("/homepage", s.GetHomepageProducts)
			products.POST("", s.CreateProduct)
			products.GET("/:id", s.GetProduct)
			products.PUT("/:id", s.UpdateProduct)
			products.DELETE("/:id", s.DeleteProduct)
			products.PUT("/:id/publish", s.PublishProduct)
			products.PUT("/:id/unpublish", s.UnpublishProduct)

			// 商品图片（可选）
			if s.enableMedia {
				products.GET("/:id/pictures", s.GetProductPictures)
				products.POST("/:id/pictures", s.AddProductPicture)
				products.DELETE("/:id/pictures/:picture_id", s.DeleteProductPicture)
				products.PUT("/:id/pictures/:picture_id/main", s.SetMainPicture)
			}

			// 商品评论（可选）
			if s.enableReview {
				products.GET("/:id/reviews", s.GetProductReviews)
				products.POST("/:id/reviews", s.CreateProductReview)
			}
		}

		// 分类管理
		categories := api.Group("/categories")
		{
			categories.GET("", s.ListCategories)
			categories.GET("/tree", s.GetCategoryTree)
			categories.POST("", s.CreateCategory)
			categories.GET("/:id", s.GetCategory)
			categories.PUT("/:id", s.UpdateCategory)
			categories.DELETE("/:id", s.DeleteCategory)
			categories.GET("/:id/products", s.GetProductsByCategory)
		}

		// 品牌管理
		manufacturers := api.Group("/manufacturers")
		{
			manufacturers.GET("", s.ListManufacturers)
			manufacturers.POST("", s.CreateManufacturer)
			manufacturers.GET("/:id", s.GetManufacturer)
			manufacturers.PUT("/:id", s.UpdateManufacturer)
			manufacturers.DELETE("/:id", s.DeleteManufacturer)
			manufacturers.GET("/:id/products", s.GetProductsByManufacturer)
		}
	}
}

// ListProducts 商品列表
func (s *CatalogService) ListProducts(c *gin.Context) {
	var req models.ProductListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认分页
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	products, total, err := s.productUC.ListProducts(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.ProductResponse, len(products))
	for i, p := range products {
		items[i] = models.ToProductResponse(p)
	}

	c.JSON(http.StatusOK, models.NewPagingResponse(items, total, req.Page, req.PageSize))
}

// GetHomepageProducts 获取首页商品
func (s *CatalogService) GetHomepageProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	products, err := s.productUC.GetHomepageProducts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.ProductResponse, len(products))
	for i, p := range products {
		items[i] = models.ToProductResponse(p)
	}

	c.JSON(http.StatusOK, items)
}

// CreateProduct 创建商品
func (s *CatalogService) CreateProduct(c *gin.Context) {
	var req biz.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := s.productUC.CreateProduct(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.ToProductResponse(product))
}

// GetProduct 获取商品详情
func (s *CatalogService) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	product, err := s.productUC.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ToProductResponse(product))
}

// UpdateProduct 更新商品
func (s *CatalogService) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req biz.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := s.productUC.UpdateProduct(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ToProductResponse(product))
}

// DeleteProduct 删除商品
func (s *CatalogService) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.productUC.DeleteProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// PublishProduct 发布商品
func (s *CatalogService) PublishProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.productUC.PublishProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product published"})
}

// UnpublishProduct 下架商品
func (s *CatalogService) UnpublishProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.productUC.UnpublishProduct(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product unpublished"})
}

// GetProductPictures 获取商品图片
func (s *CatalogService) GetProductPictures(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	pictures, err := s.pictureUC.GetPictures(c.Request.Context(), productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.ProductPictureInfo, len(pictures))
	for i, p := range pictures {
		items[i] = models.ProductPictureInfo{
			ID:         p.ID,
			PictureURL: p.PictureURL,
			AltText:    p.AltText,
			IsMain:     p.IsMain,
		}
	}

	c.JSON(http.StatusOK, items)
}

// AddProductPicture 添加商品图片
func (s *CatalogService) AddProductPicture(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req biz.AddPictureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductID = productID

	picture, err := s.pictureUC.AddPicture(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.ProductPictureInfo{
		ID:         picture.ID,
		PictureURL: picture.PictureURL,
		AltText:    picture.AltText,
		IsMain:     picture.IsMain,
	})
}

// DeleteProductPicture 删除商品图片
func (s *CatalogService) DeleteProductPicture(c *gin.Context) {
	pictureID, err := strconv.ParseUint(c.Param("picture_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid picture id"})
		return
	}

	if err := s.pictureUC.DeletePicture(c.Request.Context(), pictureID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// SetMainPicture 设置主图
func (s *CatalogService) SetMainPicture(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	pictureID, err := strconv.ParseUint(c.Param("picture_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid picture id"})
		return
	}

	if err := s.pictureUC.SetMainPicture(c.Request.Context(), productID, pictureID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "main picture set"})
}

// GetProductReviews 获取商品评论
func (s *CatalogService) GetProductReviews(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	reviews, total, err := s.reviewUC.GetReviews(c.Request.Context(), productID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.ProductReviewResponse, len(reviews))
	for i, r := range reviews {
		items[i] = models.ToProductReviewResponse(r)
	}

	c.JSON(http.StatusOK, shareModels.NewPagingResponse(items, total, page, pageSize))
}

// CreateProductReview 创建商品评论
func (s *CatalogService) CreateProductReview(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req biz.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ProductID = productID

	// TODO: 从 JWT Token 获取 customer_id
	// 暂时从请求中获取
	if req.CustomerID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id required"})
		return
	}

	review, err := s.reviewUC.CreateReview(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.ToProductReviewResponse(review))
}

// ListCategories 分类列表
func (s *CatalogService) ListCategories(c *gin.Context) {
	categories, err := s.categoryUC.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.CategoryTreeResponse, len(categories))
	for i, c := range categories {
		items[i] = models.ToCategoryTreeResponse(c)
	}

	c.JSON(http.StatusOK, items)
}

// GetCategoryTree 获取分类树
func (s *CatalogService) GetCategoryTree(c *gin.Context) {
	categories, err := s.categoryUC.GetCategoryTree(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.CategoryTreeResponse, len(categories))
	for i, c := range categories {
		items[i] = models.ToCategoryTreeResponse(c)
	}

	c.JSON(http.StatusOK, items)
}

// CreateCategory 创建分类
func (s *CatalogService) CreateCategory(c *gin.Context) {
	var req biz.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := s.categoryUC.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.ToCategoryTreeResponse(category))
}

// GetCategory 获取分类详情
func (s *CatalogService) GetCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	category, err := s.categoryUC.GetCategory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ToCategoryTreeResponse(category))
}

// UpdateCategory 更新分类
func (s *CatalogService) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req biz.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := s.categoryUC.UpdateCategory(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ToCategoryTreeResponse(category))
}

// DeleteCategory 删除分类
func (s *CatalogService) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.categoryUC.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProductsByCategory 根据分类获取商品
func (s *CatalogService) GetProductsByCategory(c *gin.Context) {
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	products, total, err := s.productUC.GetProductsByCategory(c.Request.Context(), categoryID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.ProductResponse, len(products))
	for i, p := range products {
		items[i] = models.ToProductResponse(p)
	}

	c.JSON(http.StatusOK, models.NewPagingResponse(items, total, page, pageSize))
}

// ListManufacturers 品牌列表
func (s *CatalogService) ListManufacturers(c *gin.Context) {
	manufacturers, err := s.manufacturerUC.ListManufacturers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.ManufacturerInfo, len(manufacturers))
	for i, m := range manufacturers {
		items[i] = models.ManufacturerInfo{
			ID:   m.ID,
			Name: m.Name,
		}
	}

	c.JSON(http.StatusOK, items)
}

// CreateManufacturer 创建品牌
func (s *CatalogService) CreateManufacturer(c *gin.Context) {
	var req biz.CreateManufacturerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manufacturer, err := s.manufacturerUC.CreateManufacturer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          manufacturer.ID,
		"name":        manufacturer.Name,
		"logo_url":    manufacturer.LogoURL,
		"is_published": manufacturer.IsPublished,
	})
}

// GetManufacturer 获取品牌详情
func (s *CatalogService) GetManufacturer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	manufacturer, err := s.manufacturerUC.GetManufacturer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          manufacturer.ID,
		"name":        manufacturer.Name,
		"description": manufacturer.Description,
		"logo_url":    manufacturer.LogoURL,
		"is_published": manufacturer.IsPublished,
		"seo_slug":    manufacturer.SEOSlug,
	})
}

// UpdateManufacturer 更新品牌
func (s *CatalogService) UpdateManufacturer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req biz.CreateManufacturerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manufacturer, err := s.manufacturerUC.UpdateManufacturer(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          manufacturer.ID,
		"name":        manufacturer.Name,
		"logo_url":    manufacturer.LogoURL,
	})
}

// DeleteManufacturer 删除品牌
func (s *CatalogService) DeleteManufacturer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := s.manufacturerUC.DeleteManufacturer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProductsByManufacturer 根据品牌获取商品
func (s *CatalogService) GetProductsByManufacturer(c *gin.Context) {
	manufacturerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manufacturer id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	products, total, err := s.productUC.GetProductsByManufacturer(c.Request.Context(), manufacturerID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.ProductResponse, len(products))
	for i, p := range products {
		items[i] = models.ToProductResponse(p)
	}

	c.JSON(http.StatusOK, models.NewPagingResponse(items, total, page, pageSize))
}