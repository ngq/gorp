// Package handler 包含交易服务 HTTP 处理器
// tax.go 定义税务相关 handler
// 注意：原 Provider/Category handler 已重命名为 TaxProvider/TaxCategory handler
package handler

import (
	"net/http"

	gorp "github.com/ngq/gorp"
	"nop-go/services/trade-service/internal/biz"
	"nop-go/services/trade-service/internal/server/http/request"
	"nop-go/services/trade-service/internal/server/http/response"
	"nop-go/services/trade-service/internal/service"
)

// TaxHandler 税务处理器，通过 Services 容器获取子服务
type TaxHandler struct {
	svc *service.Services
}

// NewTaxHandler 创建税务处理器
func NewTaxHandler(svc *service.Services) *TaxHandler {
	return &TaxHandler{svc: svc}
}

// --- 税务服务商 ---

// ListTaxProviders 获取税务服务商列表
// GET /api/v1/tax/providers
func (h *TaxHandler) ListTaxProviders(c gorp.Context) {
	providers, err := h.svc.Tax.UC.ListTaxProviders(c.Context())
	if err != nil {
		gorp.Error(c, err)
		return
	}
	items := make([]response.TaxProviderResponse, len(providers))
	for i, p := range providers {
		items[i] = toTaxProviderResponse(p)
	}
	gorp.Success(c, items)
}

// CreateTaxProvider 创建税务服务商
// POST /api/v1/tax/providers
func (h *TaxHandler) CreateTaxProvider(c gorp.Context) {
	var req request.CreateTaxProviderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	provider := &biz.TaxProvider{
		Name: req.Name, Code: req.Code,
		Description: req.Description, IsActive: req.IsActive,
	}
	if err := h.svc.Tax.UC.CreateTaxProvider(c.Context(), provider); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toTaxProviderResponse(provider))
}

// UpdateTaxProvider 更新税务服务商
// PUT /api/v1/tax/providers/:id
func (h *TaxHandler) UpdateTaxProvider(c gorp.Context) {
	id := c.Param("id")
	var req request.UpdateTaxProviderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	provider, err := h.svc.Tax.UC.GetTaxProvider(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	provider.Name = req.Name
	provider.Description = req.Description
	provider.IsActive = req.IsActive
	if err := h.svc.Tax.UC.UpdateTaxProvider(c.Context(), provider); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toTaxProviderResponse(provider))
}

// DeleteTaxProvider 删除税务服务商
// DELETE /api/v1/tax/providers/:id
func (h *TaxHandler) DeleteTaxProvider(c gorp.Context) {
	id := c.Param("id")
	if err := h.svc.Tax.UC.DeleteTaxProvider(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 税种分类 ---

// ListTaxCategories 获取税种分类列表
// GET /api/v1/tax/categories
func (h *TaxHandler) ListTaxCategories(c gorp.Context) {
	categories, err := h.svc.Tax.UC.ListTaxCategories(c.Context())
	if err != nil {
		gorp.Error(c, err)
		return
	}
	items := make([]response.TaxCategoryResponse, len(categories))
	for i, cat := range categories {
		items[i] = toTaxCategoryResponse(cat)
	}
	gorp.Success(c, items)
}

// CreateTaxCategory 创建税种分类
// POST /api/v1/tax/categories
func (h *TaxHandler) CreateTaxCategory(c gorp.Context) {
	var req request.CreateTaxCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	category := &biz.TaxCategory{
		Name: req.Name, Code: req.Code,
		Description: req.Description, Rate: req.Rate, IsActive: req.IsActive,
	}
	if err := h.svc.Tax.UC.CreateTaxCategory(c.Context(), category); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toTaxCategoryResponse(category))
}

// UpdateTaxCategory 更新税种分类
// PUT /api/v1/tax/categories/:id
func (h *TaxHandler) UpdateTaxCategory(c gorp.Context) {
	id := c.Param("id")
	var req request.UpdateTaxCategoryRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	category, err := h.svc.Tax.UC.GetTaxCategory(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	category.Name = req.Name
	category.Description = req.Description
	category.Rate = req.Rate
	category.IsActive = req.IsActive
	if err := h.svc.Tax.UC.UpdateTaxCategory(c.Context(), category); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toTaxCategoryResponse(category))
}

// DeleteTaxCategory 删除税种分类
// DELETE /api/v1/tax/categories/:id
func (h *TaxHandler) DeleteTaxCategory(c gorp.Context) {
	id := c.Param("id")
	if err := h.svc.Tax.UC.DeleteTaxCategory(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 税率 ---

// ListTaxRates 获取税率列表
// GET /api/v1/tax/rates/:categoryId
func (h *TaxHandler) ListTaxRates(c gorp.Context) {
	categoryID := c.Param("categoryId")
	rates, err := h.svc.Tax.UC.ListTaxRates(c.Context(), categoryID)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	items := make([]response.TaxRateResponse, len(rates))
	for i, r := range rates {
		items[i] = toTaxRateResponse(r)
	}
	gorp.Success(c, items)
}

// CreateTaxRate 创建税率
// POST /api/v1/tax/rates
func (h *TaxHandler) CreateTaxRate(c gorp.Context) {
	var req request.CreateTaxRateRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	rate := &biz.TaxRate{
		TaxCategoryID: req.TaxCategoryID, Region: req.Region,
		Rate: req.Rate, IsActive: req.IsActive,
	}
	if err := h.svc.Tax.UC.CreateTaxRate(c.Context(), rate); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toTaxRateResponse(rate))
}

// ============================================================================
// 响应转换辅助函数
// ============================================================================

func toTaxProviderResponse(p *biz.TaxProvider) response.TaxProviderResponse {
	return response.TaxProviderResponse{
		ID: p.ID, Name: p.Name, Code: p.Code,
		Description: p.Description, IsActive: p.IsActive,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toTaxCategoryResponse(c *biz.TaxCategory) response.TaxCategoryResponse {
	return response.TaxCategoryResponse{
		ID: c.ID, Name: c.Name, Code: c.Code,
		Description: c.Description, Rate: c.Rate, IsActive: c.IsActive,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func toTaxRateResponse(r *biz.TaxRate) response.TaxRateResponse {
	return response.TaxRateResponse{
		ID: r.ID, TaxCategoryID: r.TaxCategoryID, Region: r.Region,
		Rate: r.Rate, IsActive: r.IsActive,
		EffectiveFrom: r.EffectiveFrom, EffectiveTo: r.EffectiveTo,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}
