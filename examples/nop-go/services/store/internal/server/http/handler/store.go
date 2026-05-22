package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/store/internal/server/http/request"
	"nop-go/services/store/internal/server/http/response"
	"nop-go/services/store/internal/service"
)

// StoreHandler 店铺服务 HTTP 处理器。
type StoreHandler struct {
	store *service.StoreService
}

// NewStoreHandler 创建店铺服务处理器。
func NewStoreHandler(store *service.StoreService) *StoreHandler {
	return &StoreHandler{store: store}
}

// List 店铺列表。
func (h *StoreHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.store.List(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 将 service 响应转换为 handler 响应
	respItems := make([]response.Store, len(items))
	for i, item := range items {
		respItems[i] = response.Store{
			ID:           item.ID,
			Name:         item.Name,
			Url:          item.Url,
			SslEnabled:   item.SslEnabled,
			Hosts:        item.Hosts,
			DisplayOrder: item.DisplayOrder,
		}
	}

	gorp.Success(c, response.StoreList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建店铺。
func (h *StoreHandler) Create(c gorp.Context) {
	var req request.CreateStore
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	store, err := h.store.Create(c.Context(), service.CreateStoreRequest{
		Name:         req.Name,
		Url:          req.Url,
		SslEnabled:   req.SslEnabled,
		Hosts:        req.Hosts,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, store)
}

// Update 更新店铺。
func (h *StoreHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateStore
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	store, err := h.store.Update(c.Context(), uint(id), service.UpdateStoreRequest{
		Name:         req.Name,
		Url:          req.Url,
		SslEnabled:   req.SslEnabled,
		Hosts:        req.Hosts,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, store)
}

// Delete 删除店铺。
func (h *StoreHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.store.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}