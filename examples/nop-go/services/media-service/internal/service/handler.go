// Package service 媒体服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/auth/jwt"
	"nop-go/services/media-service/internal/biz"
	"nop-go/services/media-service/internal/models"
)

// MediaService 媒体服务
type MediaService struct {
	mediaUC *biz.MediaUseCase
	jwtSvc  contract.JWTService
}

// NewMediaService 创建媒体服务
func NewMediaService(mediaUC *biz.MediaUseCase, jwtSvc contract.JWTService) *MediaService {
	return &MediaService{mediaUC: mediaUC, jwtSvc: jwtSvc}
}

// RegisterRoutes 注册路由
func (s *MediaService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/media")
	adminAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "admin")
	{
		// 图片管理
		api.POST("/pictures", s.UploadPicture)
		api.GET("/pictures", adminAuth, s.ListPictures)
		api.GET("/pictures/:id", s.GetPicture)
		api.DELETE("/pictures/:id", adminAuth, s.DeletePicture)

		// 商品图片
		api.GET("/products/:product_id/pictures", s.GetProductPictures)
		api.DELETE("/products/:product_id/pictures/:picture_id", adminAuth, s.DeleteProductPicture)

		// 文档管理
		api.POST("/documents", adminAuth, s.UploadDocument)
		api.GET("/documents", adminAuth, s.ListDocuments)
		api.GET("/documents/:id", s.GetDocument)
		api.DELETE("/documents/:id", adminAuth, s.DeleteDocument)

		// 静态文件服务
		r.Static("/media", "./uploads")
	}
}

// UploadPicture 上传图片
func (s *MediaService) UploadPicture(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	var req models.PictureUploadRequest
	req.AltAttribute = c.PostForm("alt_attribute")
	req.TitleAttribute = c.PostForm("title_attribute")
	req.SeoFilename = c.PostForm("seo_filename")
	req.EntityType = c.PostForm("entity_type")
	if entityID := c.PostForm("entity_id"); entityID != "" {
		id, _ := strconv.ParseUint(entityID, 10, 64)
		req.EntityID = uint(id)
	}
	if displayOrder := c.PostForm("display_order"); displayOrder != "" {
		req.DisplayOrder, _ = strconv.Atoi(displayOrder)
	}
	req.IsMain = c.PostForm("is_main") == "true"

	result, err := s.mediaUC.UploadPicture(c.Request.Context(), file, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (s *MediaService) GetPicture(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	picture, err := s.mediaUC.GetPicture(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, picture)
}

func (s *MediaService) ListPictures(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	pictures, total, err := s.mediaUC.ListPictures(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": pictures, "total": total})
}

func (s *MediaService) DeletePicture(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.mediaUC.DeletePicture(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *MediaService) GetProductPictures(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("product_id"), 10, 64)
	pictures, err := s.mediaUC.GetProductPictures(c.Request.Context(), uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pictures)
}

func (s *MediaService) DeleteProductPicture(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Param("product_id"), 10, 64)
	pictureID, _ := strconv.ParseUint(c.Param("picture_id"), 10, 64)
	if err := s.mediaUC.DeleteProductPicture(c.Request.Context(), uint(productID), uint(pictureID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *MediaService) UploadDocument(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	var req models.DocumentUploadRequest
	req.Name = c.PostForm("name")
	if req.Name == "" {
		req.Name = file.Filename
	}
	req.Description = c.PostForm("description")

	doc, err := s.mediaUC.UploadDocument(c.Request.Context(), file, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func (s *MediaService) GetDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	doc, err := s.mediaUC.GetDocument(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (s *MediaService) ListDocuments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	docs, total, err := s.mediaUC.ListDocuments(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": docs, "total": total})
}

func (s *MediaService) DeleteDocument(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.mediaUC.DeleteDocument(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}