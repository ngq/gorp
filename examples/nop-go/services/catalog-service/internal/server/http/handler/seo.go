// Package handler SEO 相关 HTTP 处理器。
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/catalog-service/internal/server/http/request"
	"nop-go/services/catalog-service/internal/server/http/response"
	"nop-go/services/catalog-service/internal/service"
)

// SeoHandler SEO HTTP 处理器。
type SeoHandler struct {
	seo *service.SeoService
}

// NewSeoHandler 创建 SEO 处理器。
func NewSeoHandler(seo *service.SeoService) *SeoHandler {
	return &SeoHandler{seo: seo}
}

// List 获取 SEO 元数据列表。
// 路由：GET /seo
func (h *SeoHandler) List(c gorp.Context) {
	var req request.ListSeoRequest
	req.Page = 1
	req.Size = 10
	if err := c.BindQuery(&req); err != nil {
		gorp.BadRequest(c, "无效的查询参数: "+err.Error())
		return
	}

	items, total, err := h.seo.List(c.Context(), req.Page, req.Size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Seo, len(items))
	for i, item := range items {
		respItems[i] = response.Seo{
			ID:       item.ID,
			Username: item.Username,
			Email:    item.Email,
		}
	}

	gorp.Success(c, response.SeoList{
		Items: respItems,
		Total: total,
		Page:  req.Page,
		Size:  req.Size,
	})
}

// GetByID 根据ID获取 SEO 元数据。
// 路由：GET /seo/:id
func (h *SeoHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的SEO元数据ID")
		return
	}

	seo, err := h.seo.GetByID(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Seo{
		ID:       seo.ID,
		Username: seo.Username,
		Email:    seo.Email,
	})
}

// Create 创建 SEO 元数据。
// 路由：POST /seo
func (h *SeoHandler) Create(c gorp.Context) {
	var req request.CreateSeo
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	seo, err := h.seo.Create(c.Context(), service.CreateSeoRequest{
		Username: req.Username, Email: req.Email,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Seo{
		ID:       seo.ID,
		Username: seo.Username,
		Email:    seo.Email,
	})
}

// Delete 删除 SEO 元数据。
// 路由：DELETE /seo/:id
func (h *SeoHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的SEO元数据ID")
		return
	}

	if err := h.seo.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"deleted": true})
}