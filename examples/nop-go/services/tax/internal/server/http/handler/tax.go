// Package handler 提供 tax 服务的 HTTP 请求处理器
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/tax/internal/biz"
	"nop-go/services/tax/internal/server/http/request"
	"nop-go/services/tax/internal/server/http/response"
)

// TaxHandler 税务服务 HTTP 处理器
type TaxHandler struct {
	uc *biz.TaxUsecase
}

// NewTaxHandler 创建税务服务处理器
func NewTaxHandler(uc *biz.TaxUsecase) *TaxHandler {
	return &TaxHandler{uc: uc}
}

// ==================== 税务提供者 ====================

// ListProviders 获取税务提供者列表
// GET /api/v1/tax/providers
func (h *TaxHandler) ListProviders(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取税务提供者列表
	providers, total, err := h.uc.ListProviders(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListProvidersResponse{
		Total: total,
		Items: providers,
	})
}

// UpdateProvider 更新税务提供者
// PUT /api/v1/tax/providers/:id
func (h *TaxHandler) UpdateProvider(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析请求体
	var req request.UpdateProviderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = id

	// 调用业务层更新税务提供者
	provider, err := h.uc.UpdateProvider(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, provider)
}

// ==================== 税类别 ====================

// ListCategories 获取税类别列表
// GET /api/v1/tax/categories
func (h *TaxHandler) ListCategories(c gorp.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 调用业务层获取税类别列表
	categories, total, err := h.uc.ListCategories(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.ListCategoriesResponse{
		Total: total,
		Items: categories,
	})
}

// CreateCategory 创建税类别
// POST /api/v1/tax/categories
func (h *TaxHandler) CreateCategory(c gorp.Context) {
	// 解析请求体
	var req request.CreateCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	// 调用业务层创建税类别
	category, err := h.uc.CreateCategory(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, category)
}

// UpdateCategory 更新税类别
// PUT /api/v1/tax/categories/:id
func (h *TaxHandler) UpdateCategory(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 解析请求体
	var req request.UpdateCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	req.ID = id

	// 调用业务层更新税类别
	category, err := h.uc.UpdateCategory(c.Context(), req)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, category)
}

// DeleteCategory 删除税类别
// DELETE /api/v1/tax/categories/:id
func (h *TaxHandler) DeleteCategory(c gorp.Context) {
	// 获取路径参数 ID
	id, err := parseIDParam(c, "id")
	if err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	// 调用业务层删除税类别
	if err := h.uc.DeleteCategory(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
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