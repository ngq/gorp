// Package handler 媒体相关 HTTP 处理器。
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/catalog-service/internal/server/http/request"
	"nop-go/services/catalog-service/internal/server/http/response"
	"nop-go/services/catalog-service/internal/service"
)

// MediaHandler 媒体 HTTP 处理器。
type MediaHandler struct {
	media *service.MediaService
}

// NewMediaHandler 创建媒体处理器。
func NewMediaHandler(media *service.MediaService) *MediaHandler {
	return &MediaHandler{media: media}
}

// Upload 异步上传图片。
// 路由：POST /media/upload
func (h *MediaHandler) Upload(c gorp.Context) {
	var req request.UploadMedia
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "无效的请求体: "+err.Error())
		return
	}

	media, err := h.media.Upload(c.Context(), service.UploadMediaRequest{
		FileName: req.FileName, MimeType: req.MimeType, FileSize: req.FileSize,
		FileURL: req.FileURL, AltText: req.AltText,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.Media{
		ID:        media.ID,
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		FileSize:  media.FileSize,
		FileURL:   media.FileURL,
		AltText:   media.AltText,
		CreatedAt: media.CreatedAt,
	})
}

// GetByID 根据ID获取媒体。
// 路由：GET /media/:id
func (h *MediaHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的媒体ID")
		return
	}

	media, err := h.media.GetByID(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.Media{
		ID:        media.ID,
		FileName:  media.FileName,
		MimeType:  media.MimeType,
		FileSize:  media.FileSize,
		FileURL:   media.FileURL,
		AltText:   media.AltText,
		CreatedAt: media.CreatedAt,
	})
}

// Delete 删除媒体。
// 路由：DELETE /media/:id
func (h *MediaHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的媒体ID")
		return
	}

	if err := h.media.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{"deleted": true})
}