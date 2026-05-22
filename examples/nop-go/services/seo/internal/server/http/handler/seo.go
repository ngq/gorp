package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/seo/internal/server/http/request"
	"nop-go/services/seo/internal/server/http/response"
	"nop-go/services/seo/internal/service"
)

type SeoHandler struct {
	seo *service.SeoService
}

func NewSeoHandler(seo *service.SeoService) *SeoHandler {
	return &SeoHandler{seo: seo}
}

func (h *SeoHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.seo.List(c, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	seoItems := make([]response.Seo, len(items))
	for i, item := range items {
		seoItems[i] = response.Seo{ID: item.ID, Username: item.Username, Email: item.Email}
	}

	gorp.Success(c, response.SeoList{
		Items: seoItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func (h *SeoHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	seo, err := h.seo.GetByID(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "seo not found"})
		return
	}

	gorp.Success(c, response.Seo{ID: seo.ID, Username: seo.Username, Email: seo.Email})
}

func (h *SeoHandler) Create(c gorp.Context) {
	var req request.CreateSeo
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	seo, err := h.seo.Create(c, service.CreateSeoRequest{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Seo{ID: seo.ID, Username: seo.Username, Email: seo.Email})
}

func (h *SeoHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	if err := h.seo.Delete(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
