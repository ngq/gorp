package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"monolith/app/http/request"
	"monolith/app/http/response"
	"monolith/internal/biz"
	"monolith/internal/service"
)

type DemoHandler struct {
	svc *service.DemoService
}

func NewDemoHandler(svc *service.DemoService) *DemoHandler {
	return &DemoHandler{svc: svc}
}

func (h *DemoHandler) Create(c gorp.HTTPContext) {
	var req request.CreateDemo
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	demo, err := h.svc.Create(c.Context(), req.Name)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, toDemoResponse(demo))
}

func (h *DemoHandler) GetByID(c gorp.HTTPContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	demo, err := h.svc.GetByID(c.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "not found"})
		return
	}

	gorp.Success(c, toDemoResponse(demo))
}

func (h *DemoHandler) List(c gorp.HTTPContext) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	demos, total, err := h.svc.List(c.Context(), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	items := make([]*response.Demo, len(demos))
	for i, demo := range demos {
		items[i] = toDemoResponse(demo)
	}

	gorp.Success(c, &response.DemoList{
		Items: items,
		Total: total,
		Page:  page,
		Size:  pageSize,
	})
}

func (h *DemoHandler) Update(c gorp.HTTPContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	var req request.UpdateDemo
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	demo, err := h.svc.Update(c.Context(), uint(id), req.Name)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, toDemoResponse(demo))
}

func (h *DemoHandler) Delete(c gorp.HTTPContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	if err := h.svc.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func toDemoResponse(demo *biz.Demo) *response.Demo {
	return &response.Demo{
		ID:        demo.ID,
		Name:      demo.Name,
		CreatedAt: demo.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: demo.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
