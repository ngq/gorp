package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/gdpr/internal/server/http/request"
	"nop-go/services/gdpr/internal/server/http/response"
	"nop-go/services/gdpr/internal/service"
)

type GdprHandler struct {
	gdpr *service.GdprService
}

func NewGdprHandler(gdpr *service.GdprService) *GdprHandler {
	return &GdprHandler{gdpr: gdpr}
}

func (h *GdprHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.gdpr.List(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gdprItems := make([]response.Gdpr, len(items))
	for i, item := range items {
		gdprItems[i] = response.Gdpr{ID: item.ID, Username: item.Username, Email: item.Email}
	}

	gorp.Success(c, response.GdprList{
		Items: gdprItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func (h *GdprHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	gdpr, err := h.gdpr.GetByID(c.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "gdpr not found"})
		return
	}

	gorp.Success(c, response.Gdpr{ID: gdpr.ID, Username: gdpr.Username, Email: gdpr.Email})
}

func (h *GdprHandler) Create(c gorp.Context) {
	var req request.CreateGdpr
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	gdpr, err := h.gdpr.Create(c.Context(), service.CreateGdprRequest{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Gdpr{ID: gdpr.ID, Username: gdpr.Username, Email: gdpr.Email})
}

func (h *GdprHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	if err := h.gdpr.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
