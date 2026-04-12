// Package service SEO服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/serviceauth/token"
	"nop-go/services/seo-service/internal/biz"
	"nop-go/services/seo-service/internal/models"
)

// SEOService SEO服务
type SEOService struct {
	seoUC  *biz.SEOUseCase
	jwtSvc contract.JWTService
}

// NewSEOService 创建SEO服务
func NewSEOService(seoUC *biz.SEOUseCase, jwtSvc contract.JWTService) *SEOService {
	return &SEOService{seoUC: seoUC, jwtSvc: jwtSvc}
}

// RegisterRoutes 注册路由
func (s *SEOService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/seo")
	adminAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "admin")
	{
		// URL记录管理
		api.POST("/urls", adminAuth, s.CreateUrlRecord)
		api.GET("/urls", adminAuth, s.ListUrlRecords)
		api.GET("/urls/search", adminAuth, s.SearchUrlRecords)
		api.GET("/urls/:id", s.GetUrlRecord)
		api.GET("/urls/slug/:slug", s.GetUrlBySlug)
		api.PUT("/urls/:id", adminAuth, s.UpdateUrlRecord)
		api.DELETE("/urls/:id", adminAuth, s.DeleteUrlRecord)
		api.POST("/urls/generate-slug", s.GenerateSlug)

		// URL重定向管理
		api.POST("/redirects", adminAuth, s.CreateUrlRedirect)
		api.GET("/redirects", adminAuth, s.ListUrlRedirects)
		api.GET("/redirects/:id", s.GetUrlRedirect)
		api.GET("/redirects/check/:old_slug", s.CheckRedirect)
		api.PUT("/redirects/:id", adminAuth, s.UpdateUrlRedirect)
		api.DELETE("/redirects/:id", adminAuth, s.DeleteUrlRedirect)

		// 元信息管理
		api.POST("/meta", s.CreateMetaInfo)
		api.GET("/meta", adminAuth, s.ListMetaInfo)
		api.GET("/meta/:id", s.GetMetaInfo)
		api.GET("/meta/entity/:entity_type/:entity_id", s.GetMetaByEntity)
		api.PUT("/meta/:id", adminAuth, s.UpdateMetaInfo)
		api.DELETE("/meta/:id", adminAuth, s.DeleteMetaInfo)

		// Sitemap管理
		api.POST("/sitemap/nodes", adminAuth, s.AddSitemapNode)
		api.GET("/sitemap/nodes", s.GetSitemapNodes)
		api.DELETE("/sitemap/nodes/:id", adminAuth, s.DeleteSitemapNode)
		api.DELETE("/sitemap/type/:entity_type", adminAuth, s.ClearSitemapByType)
		api.GET("/sitemap/generate", s.GenerateSitemap)

		// SEO分析
		api.GET("/analyze/:entity_type/:entity_id", s.AnalyzeSEO)
	}
}

// ========== URL记录接口 ==========

func (s *SEOService) CreateUrlRecord(c *gin.Context) {
	var req models.UrlRecordCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := s.seoUC.CreateUrlRecord(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (s *SEOService) GetUrlRecord(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	record, err := s.seoUC.GetUrlRecord(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (s *SEOService) GetUrlBySlug(c *gin.Context) {
	slug := c.Param("slug")
	record, err := s.seoUC.GetUrlBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, record)
}

func (s *SEOService) ListUrlRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	records, total, err := s.seoUC.ListUrlRecords(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": records, "total": total})
}

func (s *SEOService) SearchUrlRecords(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	records, total, err := s.seoUC.SearchUrlRecords(c.Request.Context(), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": records, "total": total})
}

func (s *SEOService) UpdateUrlRecord(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.UrlRecordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := s.seoUC.UpdateUrlRecord(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

func (s *SEOService) DeleteUrlRecord(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.seoUC.DeleteUrlRecord(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *SEOService) GenerateSlug(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	slug := s.seoUC.GenerateSlug(c.Request.Context(), name)
	uniqueSlug := s.seoUC.EnsureUniqueSlug(c.Request.Context(), slug)

	c.JSON(http.StatusOK, gin.H{"slug": uniqueSlug})
}

// ========== URL重定向接口 ==========

func (s *SEOService) CreateUrlRedirect(c *gin.Context) {
	var req models.UrlRedirectCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	redirect, err := s.seoUC.CreateUrlRedirect(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, redirect)
}

func (s *SEOService) GetUrlRedirect(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	redirect, err := s.seoUC.GetUrlRedirect(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, redirect)
}

func (s *SEOService) CheckRedirect(c *gin.Context) {
	oldSlug := c.Param("old_slug")
	redirect, err := s.seoUC.GetRedirectByOldSlug(c.Request.Context(), oldSlug)
	if err != nil || redirect == nil {
		c.JSON(http.StatusOK, gin.H{"redirect": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"redirect": redirect})
}

func (s *SEOService) ListUrlRedirects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	redirects, total, err := s.seoUC.ListUrlRedirects(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": redirects, "total": total})
}

func (s *SEOService) UpdateUrlRedirect(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.UrlRedirectUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	redirect, err := s.seoUC.UpdateUrlRedirect(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, redirect)
}

func (s *SEOService) DeleteUrlRedirect(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.seoUC.DeleteUrlRedirect(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== 元信息接口 ==========

func (s *SEOService) CreateMetaInfo(c *gin.Context) {
	var req models.MetaInfoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meta, err := s.seoUC.CreateMetaInfo(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, meta)
}

func (s *SEOService) GetMetaInfo(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	meta, err := s.seoUC.GetMetaInfo(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meta)
}

func (s *SEOService) GetMetaByEntity(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityID, _ := strconv.ParseUint(c.Param("entity_id"), 10, 64)
	languageID, _ := strconv.ParseUint(c.Query("language_id"), 10, 64)

	meta, err := s.seoUC.GetMetaByEntity(c.Request.Context(), uint(entityID), entityType, uint(languageID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meta)
}

func (s *SEOService) ListMetaInfo(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	metas, total, err := s.seoUC.ListMetaInfo(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": metas, "total": total})
}

func (s *SEOService) UpdateMetaInfo(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.MetaInfoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meta, err := s.seoUC.UpdateMetaInfo(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, meta)
}

func (s *SEOService) DeleteMetaInfo(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.seoUC.DeleteMetaInfo(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== Sitemap接口 ==========

func (s *SEOService) AddSitemapNode(c *gin.Context) {
	var node models.SitemapNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.seoUC.AddSitemapNode(c.Request.Context(), &node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, node)
}

func (s *SEOService) GetSitemapNodes(c *gin.Context) {
	nodes, err := s.seoUC.GetSitemapNodes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nodes)
}

func (s *SEOService) DeleteSitemapNode(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.seoUC.DeleteSitemapNode(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *SEOService) ClearSitemapByType(c *gin.Context) {
	entityType := c.Param("entity_type")
	if err := s.seoUC.ClearSitemapByEntityType(c.Request.Context(), entityType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "sitemap nodes cleared for " + entityType})
}

func (s *SEOService) GenerateSitemap(c *gin.Context) {
	baseURL := c.Query("base_url")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	xmlContent, _, err := s.seoUC.GenerateSitemap(c.Request.Context(), baseURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回XML内容
	c.Data(http.StatusOK, "application/xml", []byte(xmlContent))

	// 也可以返回JSON格式
	// c.JSON(http.StatusOK, gin.H{"xml": xmlContent, "result": result})
}

// ========== SEO分析接口 ==========

func (s *SEOService) AnalyzeSEO(c *gin.Context) {
	entityType := c.Param("entity_type")
	entityID, _ := strconv.ParseUint(c.Param("entity_id"), 10, 64)
	languageID, _ := strconv.ParseUint(c.Query("language_id"), 10, 64)

	result, err := s.seoUC.AnalyzeSEO(c.Request.Context(), uint(entityID), entityType, uint(languageID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}