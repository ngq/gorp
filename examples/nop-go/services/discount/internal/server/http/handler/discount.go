// Package handler 提供 discount 服务的 HTTP 请求处理器
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/discount/internal/biz"
	"nop-go/services/discount/internal/server/http/request"
	"nop-go/services/discount/internal/server/http/response"
)

// DiscountHandler 折扣服务 HTTP 处理器
type DiscountHandler struct {
	uc *biz.DiscountUsecase
}

// NewDiscountHandler 创建折扣服务处理器
func NewDiscountHandler(uc *biz.DiscountUsecase) *DiscountHandler {
	return &DiscountHandler{uc: uc}
}

// ==================== 折扣 CRUD ====================

// ListDiscounts 获取折扣列表
// GET /api/v1/discounts
func (h *DiscountHandler) ListDiscounts(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取折扣列表
	discounts, total, err := h.uc.ListDiscounts(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListDiscountsResponse{
		Total: total,
		Items: discounts,
	})
}

// CreateDiscount 创建折扣
// POST /api/v1/discounts
func (h *DiscountHandler) CreateDiscount(c gorp.Context) {
	// 解析请求体
	var req request.CreateDiscountRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 调用业务层创建折扣
	discount, err := h.uc.CreateDiscount(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, discount)
}

// UpdateDiscount 更新折扣
// PUT /api/v1/discounts/:id
func (h *DiscountHandler) UpdateDiscount(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析请求体
	var req request.UpdateDiscountRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = id

	// 调用业务层更新折扣
	discount, err := h.uc.UpdateDiscount(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, discount)
}

// DeleteDiscount 删除折扣
// DELETE /api/v1/discounts/:id
func (h *DiscountHandler) DeleteDiscount(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 调用业务层删除折扣
	if err := h.uc.DeleteDiscount(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ==================== 折扣关联子资源 ====================

// ListDiscountProducts 获取折扣关联商品列表
// GET /api/v1/discounts/:id/products
func (h *DiscountHandler) ListDiscountProducts(c gorp.Context) {
	// 获取路径参数折扣 ID
	discountID, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取折扣关联商品
	products, total, err := h.uc.ListDiscountProducts(c.Context(), discountID, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListDiscountProductsResponse{
		Total: total,
		Items: products,
	})
}

// ListDiscountCategories 获取折扣关联分类列表
// GET /api/v1/discounts/:id/categories
func (h *DiscountHandler) ListDiscountCategories(c gorp.Context) {
	// 获取路径参数折扣 ID
	discountID, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取折扣关联分类
	categories, total, err := h.uc.ListDiscountCategories(c.Context(), discountID, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListDiscountCategoriesResponse{
		Total: total,
		Items: categories,
	})
}

// ListDiscountManufacturers 获取折扣关联制造商列表
// GET /api/v1/discounts/:id/manufacturers
func (h *DiscountHandler) ListDiscountManufacturers(c gorp.Context) {
	// 获取路径参数折扣 ID
	discountID, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取折扣关联制造商
	manufacturers, total, err := h.uc.ListDiscountManufacturers(c.Context(), discountID, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListDiscountManufacturersResponse{
		Total: total,
		Items: manufacturers,
	})
}

// ListDiscountUsageHistory 获取折扣使用历史列表
// GET /api/v1/discounts/:id/usage-history
func (h *DiscountHandler) ListDiscountUsageHistory(c gorp.Context) {
	// 获取路径参数折扣 ID
	discountID, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取折扣使用历史
	history, total, err := h.uc.ListDiscountUsageHistory(c.Context(), discountID, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListDiscountUsageHistoryResponse{
		Total: total,
		Items: history,
	})
}

// ==================== 内部辅助函数 ====================

// parseIDParam 解析路径参数中的 ID，返回 uint
func parseIDParam(c gorp.Context, key string) (uint, error) {
	idStr := c.Param(key)
	if idStr == "" {
		return 0, errParamRequired(key)
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, errInvalidID(key)
	}
	return uint(id), nil
}

// errParamRequired 返回参数必填错误
type paramRequiredError string

func (e paramRequiredError) Error() string { return string(e) + " 不能为空" }

func errParamRequired(key string) paramRequiredError { return paramRequiredError(key) }

// errInvalidID 返回 ID 无效错误
type invalidIDError string

func (e invalidIDError) Error() string { return "无效的 " + string(e) }

func errInvalidID(key string) invalidIDError { return invalidIDError(key) }