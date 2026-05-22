package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/plugin/internal/server/http/request"
	"nop-go/services/plugin/internal/server/http/response"
	"nop-go/services/plugin/internal/service"
)

type PluginHandler struct {
	plugin *service.PluginService
}

func NewPluginHandler(plugin *service.PluginService) *PluginHandler {
	return &PluginHandler{plugin: plugin}
}

func (h *PluginHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.plugin.List(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	pluginItems := make([]response.Plugin, len(items))
	for i, item := range items {
		pluginItems[i] = response.Plugin{ID: item.ID, Username: item.Username, Email: item.Email}
	}

	gorp.Success(c, response.PluginList{
		Items: pluginItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func (h *PluginHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	plugin, err := h.plugin.GetByID(c.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "plugin not found"})
		return
	}

	gorp.Success(c, response.Plugin{ID: plugin.ID, Username: plugin.Username, Email: plugin.Email})
}

func (h *PluginHandler) Create(c gorp.Context) {
	var req request.CreatePlugin
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	plugin, err := h.plugin.Create(c.Context(), service.CreatePluginRequest{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Plugin{ID: plugin.ID, Username: plugin.Username, Email: plugin.Email})
}

func (h *PluginHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	if err := h.plugin.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
