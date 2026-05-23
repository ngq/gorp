// Package handler HTTP 处理器层。
// 负责接收 gorp.Context 请求，绑定参数，调用服务层，返回统一响应。
// 所有 handler 使用 gorp.Success/gorp.Error/gorp.BadRequest 统一响应格式。
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/catalog-service/internal/server/http/request"
	"nop-go/services/catalog-service/internal/server/http/response"
	"nop-go/services/catalog-service/internal/service"
)

// ---------------------------------------------------------------------------
// 商品处理器
// ---------------------------------------------------------------------------

// ProductHandler 商品 HTTP 处理器。
type ProductHandler struct {
	product *service.ProductService
}

// NewProductHandler 创建商品处理器。
func NewProductHandler(product *service.ProductService) *ProductHandler {
	return &ProductHandler{product: product}
}

// GetProductByID 获取商品详情。
// 路由：GET /products/:id
func (h *ProductHandler) GetProductByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的商品ID")
		return
	}

	product, err := h.product.GetByID(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Product{
		ID:             product.ID,
		Name:           product.Name,
		ShortDesc:      product.ShortDesc,
		FullDesc:       product.FullDesc,
		SKU:            product.SKU,
		Price:          product.Price,
		OldPrice:       product.OldPrice,
		Cost:           product.Cost,
		Stock:          product.Stock,
		CategoryID:     product.CategoryID,
		ManufacturerID: product.ManufacturerID,
		IsPublished:    product.IsPublished,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	})
}

// ListProducts 获取商品列表。
// 路由：GET /products
func (h *ProductHandler) ListProducts(c gorp.Context) {
	var req request.ListProductRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.product.List(c.Context(), req.Page, req.Size, req.CategoryID, req.ManufacturerID, req.Keyword)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.ProductListItem, len(items))
	for i, item := range items {
		respItems[i] = response.ProductListItem{
			ID:             item.ID,
			Name:           item.Name,
			ShortDesc:      item.ShortDesc,
			SKU:            item.SKU,
			Price:          item.Price,
			OldPrice:       item.OldPrice,
			Stock:          item.Stock,
			CategoryID:     item.CategoryID,
			ManufacturerID: item.ManufacturerID,
			IsPublished:    item.IsPublished,
			CreatedAt:      item.CreatedAt,
		}
	}

	gorp.Success(c, response.ProductList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// CreateProduct 创建商品。
// 路由：POST /products
func (h *ProductHandler) CreateProduct(c gorp.Context) {
	var req request.CreateProduct
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	product, err := h.product.Create(c.Context(), service.CreateProductRequest{
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
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Product{
		ID:             product.ID,
		Name:           product.Name,
		ShortDesc:      product.ShortDesc,
		FullDesc:       product.FullDesc,
		SKU:            product.SKU,
		Price:          product.Price,
		OldPrice:       product.OldPrice,
		Cost:           product.Cost,
		Stock:          product.Stock,
		CategoryID:     product.CategoryID,
		ManufacturerID: product.ManufacturerID,
		IsPublished:    product.IsPublished,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	})
}

// UpdateProduct 更新商品信息。
// 路由：PUT /products/:id
func (h *ProductHandler) UpdateProduct(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的商品ID")
		return
	}

	var req request.UpdateProduct
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	product, err := h.product.Update(c.Context(), uint(id), service.UpdateProductRequest{
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
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Product{
		ID:             product.ID,
		Name:           product.Name,
		ShortDesc:      product.ShortDesc,
		FullDesc:       product.FullDesc,
		SKU:            product.SKU,
		Price:          product.Price,
		OldPrice:       product.OldPrice,
		Cost:           product.Cost,
		Stock:          product.Stock,
		CategoryID:     product.CategoryID,
		ManufacturerID: product.ManufacturerID,
		IsPublished:    product.IsPublished,
		CreatedAt:      product.CreatedAt,
		UpdatedAt:      product.UpdatedAt,
	})
}

// DeleteProduct 删除商品。
// 路由：DELETE /products/:id
func (h *ProductHandler) DeleteProduct(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的商品ID")
		return
	}

	if err := h.product.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"deleted": true})
}

// SearchProducts 商品搜索。
// 路由：GET /products/search
func (h *ProductHandler) SearchProducts(c gorp.Context) {
	var req request.SearchRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.product.Search(c.Context(), req.Keyword, req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.ProductListItem, len(items))
	for i, item := range items {
		respItems[i] = response.ProductListItem{
			ID:             item.ID,
			Name:           item.Name,
			ShortDesc:      item.ShortDesc,
			SKU:            item.SKU,
			Price:          item.Price,
			OldPrice:       item.OldPrice,
			Stock:          item.Stock,
			CategoryID:     item.CategoryID,
			ManufacturerID: item.ManufacturerID,
			IsPublished:    item.IsPublished,
			CreatedAt:      item.CreatedAt,
		}
	}

	gorp.Success(c, response.SearchResult{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// GetRecentlyViewed 获取最近浏览商品列表。
// 路由：GET /products/recently-viewed
func (h *ProductHandler) GetRecentlyViewed(c gorp.Context) {
	var req request.RecentlyViewedRequest
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	items, err := h.product.GetRecentlyViewed(c.Context(), req.CustomerID, limit)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.ProductListItem, len(items))
	for i, item := range items {
		respItems[i] = response.ProductListItem{
			ID:             item.ID,
			Name:           item.Name,
			ShortDesc:      item.ShortDesc,
			SKU:            item.SKU,
			Price:          item.Price,
			OldPrice:       item.OldPrice,
			Stock:          item.Stock,
			CategoryID:     item.CategoryID,
			ManufacturerID: item.ManufacturerID,
			IsPublished:    item.IsPublished,
			CreatedAt:      item.CreatedAt,
		}
	}

	gorp.Success(c, response.RecentlyViewedList{Items: respItems})
}

// CompareProducts 商品对比。
// 路由：POST /products/compare
func (h *ProductHandler) CompareProducts(c gorp.Context) {
	var req request.CompareProductsRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	items, err := h.product.CompareProducts(c.Context(), req.ProductIDs)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Product, len(items))
	for i, item := range items {
		respItems[i] = response.Product{
			ID:             item.ID,
			Name:           item.Name,
			ShortDesc:      item.ShortDesc,
			FullDesc:       item.FullDesc,
			SKU:            item.SKU,
			Price:          item.Price,
			OldPrice:       item.OldPrice,
			Cost:           item.Cost,
			Stock:          item.Stock,
			CategoryID:     item.CategoryID,
			ManufacturerID: item.ManufacturerID,
			IsPublished:    item.IsPublished,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
		}
	}

	gorp.Success(c, response.CompareResult{Items: respItems})
}

// ---------------------------------------------------------------------------
// 分类处理器
// ---------------------------------------------------------------------------

// CategoryHandler 分类 HTTP 处理器。
type CategoryHandler struct {
	category *service.ProductService
}

// NewCategoryHandler 创建分类处理器。
func NewCategoryHandler(category *service.ProductService) *CategoryHandler {
	return &CategoryHandler{category: category}
}

// GetCategoryByID 获取分类详情。
// 路由：GET /categories/:id
func (h *CategoryHandler) GetCategoryByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的分类ID")
		return
	}

	category, err := h.category.GetCategoryByID(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Category{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		ParentID:    category.ParentID,
		SortOrder:   category.SortOrder,
		IsPublished: category.IsPublished,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	})
}

// ListCategories 获取分类列表。
// 路由：GET /categories
func (h *CategoryHandler) ListCategories(c gorp.Context) {
	var req request.ListCategoryRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.category.ListCategories(c.Context(), req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Category, len(items))
	for i, item := range items {
		respItems[i] = response.Category{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			ParentID:    item.ParentID,
			SortOrder:   item.SortOrder,
			IsPublished: item.IsPublished,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	gorp.Success(c, response.CategoryList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// CreateCategory 创建分类。
// 路由：POST /categories
func (h *CategoryHandler) CreateCategory(c gorp.Context) {
	var req request.CreateCategory
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	category, err := h.category.CreateCategory(c.Context(), req.Name, req.Description, req.ParentID, req.SortOrder, req.IsPublished)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Category{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		ParentID:    category.ParentID,
		SortOrder:   category.SortOrder,
		IsPublished: category.IsPublished,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	})
}

// ---------------------------------------------------------------------------
// 制造商处理器
// ---------------------------------------------------------------------------

// ManufacturerHandler 制造商 HTTP 处理器。
type ManufacturerHandler struct {
	manufacturer *service.ProductService
}

// NewManufacturerHandler 创建制造商处理器。
func NewManufacturerHandler(manufacturer *service.ProductService) *ManufacturerHandler {
	return &ManufacturerHandler{manufacturer: manufacturer}
}

// GetManufacturerByID 获取制造商详情。
// 路由：GET /manufacturers/:id
func (h *ManufacturerHandler) GetManufacturerByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的制造商ID")
		return
	}

	manufacturer, err := h.manufacturer.GetManufacturerByID(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Manufacturer{
		ID:          manufacturer.ID,
		Name:        manufacturer.Name,
		Description: manufacturer.Description,
		IsPublished: manufacturer.IsPublished,
		CreatedAt:   manufacturer.CreatedAt,
		UpdatedAt:   manufacturer.UpdatedAt,
	})
}

// ListManufacturers 获取制造商列表。
// 路由：GET /manufacturers
func (h *ManufacturerHandler) ListManufacturers(c gorp.Context) {
	var req request.ListManufacturerRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.manufacturer.ListManufacturers(c.Context(), req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Manufacturer, len(items))
	for i, item := range items {
		respItems[i] = response.Manufacturer{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			IsPublished: item.IsPublished,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	gorp.Success(c, response.ManufacturerList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// CreateManufacturer 创建制造商。
// 路由：POST /manufacturers
func (h *ManufacturerHandler) CreateManufacturer(c gorp.Context) {
	var req request.CreateManufacturer
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	manufacturer, err := h.manufacturer.CreateManufacturer(c.Context(), req.Name, req.Description, req.IsPublished)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Manufacturer{
		ID:          manufacturer.ID,
		Name:        manufacturer.Name,
		Description: manufacturer.Description,
		IsPublished: manufacturer.IsPublished,
		CreatedAt:   manufacturer.CreatedAt,
		UpdatedAt:   manufacturer.UpdatedAt,
	})
}

// ---------------------------------------------------------------------------
// 商品评论处理器
// ---------------------------------------------------------------------------

// ReviewHandler 商品评论 HTTP 处理器。
type ReviewHandler struct {
	product *service.ProductService
}

// NewReviewHandler 创建商品评论处理器。
func NewReviewHandler(product *service.ProductService) *ReviewHandler {
	return &ReviewHandler{product: product}
}

// ListProductReviews 获取商品评论列表。
// 路由：GET /products/:id/reviews
func (h *ReviewHandler) ListProductReviews(c gorp.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的商品ID")
		return
	}

	var req request.ListProductReviewRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.product.ListProductReviews(c.Context(), uint(productID), req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.ProductReview, len(items))
	for i, item := range items {
		respItems[i] = response.ProductReview{
			ID:           item.ID,
			ProductID:    item.ProductID,
			CustomerID:   item.CustomerID,
			CustomerName: item.CustomerName,
			Title:        item.Title,
			Content:      item.Content,
			Rating:       item.Rating,
			IsApproved:   item.IsApproved,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		}
	}

	gorp.Success(c, response.ProductReviewList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// CreateProductReview 提交商品评论。
// 路由：POST /products/:id/reviews
func (h *ReviewHandler) CreateProductReview(c gorp.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的商品ID")
		return
	}

	var req request.CreateProductReview
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	// 从上下文中获取客户信息（实际项目中应从认证中间件获取）
	customerID := uint(0)
	customerName := "匿名用户"
	if cid := c.GetHeader("X-Customer-ID"); cid != "" {
		if parsed, err := strconv.ParseUint(cid, 10, 64); err == nil {
			customerID = uint(parsed)
		}
	}
	if cname := c.GetHeader("X-Customer-Name"); cname != "" {
		customerName = cname
	}

	review, err := h.product.CreateProductReview(c.Context(), uint(productID), customerID, customerName, req.Title, req.Content, req.Rating)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.ProductReview{
		ID:           review.ID,
		ProductID:    review.ProductID,
		CustomerID:   review.CustomerID,
		CustomerName: review.CustomerName,
		Title:        review.Title,
		Content:      review.Content,
		Rating:       review.Rating,
		IsApproved:   review.IsApproved,
		CreatedAt:    review.CreatedAt,
		UpdatedAt:    review.UpdatedAt,
	})
}