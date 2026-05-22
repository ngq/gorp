package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/logging/internal/server/http/request"
	"nop-go/services/logging/internal/server/http/response"
	"nop-go/services/logging/internal/service"
)

type LoggingHandler struct {
	logging *service.LoggingService
}

func NewLoggingHandler(logging *service.LoggingService) *LoggingHandler {
	return &LoggingHandler{logging: logging}
}

func (h *LoggingHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.logging.List(c, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	loggingItems := make([]response.Logging, len(items))
	for i, item := range items {
		loggingItems[i] = response.Logging{ID: item.ID, Username: item.Username, Email: item.Email}
	}

	gorp.Success(c, response.LoggingList{
		Items: loggingItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func (h *LoggingHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	logging, err := h.logging.GetByID(c, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "logging not found"})
		return
	}

	gorp.Success(c, response.Logging{ID: logging.ID, Username: logging.Username, Email: logging.Email})
}

func (h *LoggingHandler) Create(c gorp.Context) {
	var req request.CreateLogging
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	logging, err := h.logging.Create(c, service.CreateLoggingRequest{
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Logging{ID: logging.ID, Username: logging.Username, Email: logging.Email})
}

func (h *LoggingHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "invalid id")
		return
	}

	if err := h.logging.Delete(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
