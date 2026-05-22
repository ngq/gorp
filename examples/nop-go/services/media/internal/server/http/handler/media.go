package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/media/internal/server/http/request"
	"nop-go/services/media/internal/server/http/response"
	"nop-go/services/media/internal/service"
)

// toMediaResponse 将 service.MediaResponse 转换为 response.Media。
//
// service 层与 response 层均使用 time.Time 表示时间戳，直接赋值即可。
func toMediaResponse(src *service.MediaResponse) response.Media {
	return response.Media{
		ID:        src.ID,
		FileName:  src.FileName,
		MimeType:  src.MimeType,
		FileSize:  src.FileSize,
		FileURL:   src.FileURL,
		AltText:   src.AltText,
		CreatedAt: src.CreatedAt,
	}
}

// MediaHandler 媒体服务 HTTP 处理器。
type MediaHandler struct {
	media *service.MediaService
}

// NewMediaHandler 创建媒体服务处理器。
func NewMediaHandler(media *service.MediaService) *MediaHandler {
	return &MediaHandler{media: media}
}

// Upload 异步上传图片。
// 接收文件上传请求，将图片存入存储并返回媒体ID。
func (h *MediaHandler) Upload(c gorp.Context) {
	var req request.UploadMedia
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "上传参数无效: "+err.Error())
		return
	}

	media, err := h.media.Upload(c.Context(), service.UploadMediaRequest{
		FileName: req.FileName,
		MimeType: req.MimeType,
		FileSize: req.FileSize,
		FileURL:  req.FileURL,
		AltText:  req.AltText,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, toMediaResponse(media))
}

// GetByID 根据ID获取图片信息。
func (h *MediaHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	media, err := h.media.GetByID(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, toMediaResponse(media))
}

// Delete 根据ID删除图片。
func (h *MediaHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.media.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}
