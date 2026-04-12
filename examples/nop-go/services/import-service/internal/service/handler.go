// Package service 导入导出服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/serviceauth/token"
	"nop-go/services/import-service/internal/biz"
	"nop-go/services/import-service/internal/models"
)

// ImportService 导入导出服务
type ImportService struct {
	importUC *biz.ImportUseCase
	exportUC *biz.ExportUseCase
	jwtSvc   contract.JWTService
}

// NewImportService 创建导入导出服务
func NewImportService(importUC *biz.ImportUseCase, exportUC *biz.ExportUseCase, jwtSvc contract.JWTService) *ImportService {
	return &ImportService{importUC: importUC, exportUC: exportUC, jwtSvc: jwtSvc}
}

// RegisterRoutes 注册路由
func (s *ImportService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/import")
	adminAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "admin")
	{
		// 导入配置管理
		api.POST("/profiles", adminAuth, s.CreateImportProfile)
		api.GET("/profiles", s.ListImportProfiles)
		api.GET("/profiles/:id", s.GetImportProfile)
		api.PUT("/profiles/:id", adminAuth, s.UpdateImportProfile)
		api.DELETE("/profiles/:id", adminAuth, s.DeleteImportProfile)

		// 导入执行
		api.POST("/execute", adminAuth, s.ExecuteImport)
		api.GET("/history", s.ListImportHistory)
		api.GET("/history/:id", s.GetImportHistory)
		api.GET("/history/:id/errors", s.GetImportErrors)

		// 导出配置管理
		api.POST("/export/profiles", adminAuth, s.CreateExportProfile)
		api.GET("/export/profiles", s.ListExportProfiles)
		api.GET("/export/profiles/:id", s.GetExportProfile)
		api.PUT("/export/profiles/:id", adminAuth, s.UpdateExportProfile)
		api.DELETE("/export/profiles/:id", adminAuth, s.DeleteExportProfile)

		// 导出执行
		api.POST("/export/execute", adminAuth, s.ExecuteExport)
		api.GET("/export/history", s.ListExportHistory)
		api.GET("/export/history/:id", s.GetExportHistory)

		// 实体类型
		api.GET("/entity-types", s.GetEntityTypes)
	}
}

// ========== 导入配置接口 ==========

func (s *ImportService) CreateImportProfile(c *gin.Context) {
	var req models.ImportProfileCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := s.importUC.CreateProfile(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, profile)
}

func (s *ImportService) GetImportProfile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	profile, err := s.importUC.GetProfile(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (s *ImportService) ListImportProfiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	profiles, total, err := s.importUC.ListProfiles(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": profiles, "total": total})
}

func (s *ImportService) UpdateImportProfile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.ImportProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := s.importUC.UpdateProfile(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (s *ImportService) DeleteImportProfile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.importUC.DeleteProfile(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *ImportService) ExecuteImport(c *gin.Context) {
	var req models.ImportExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.importUC.ExecuteImport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *ImportService) ListImportHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	history, total, err := s.importUC.ListImportHistory(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": history, "total": total})
}

func (s *ImportService) GetImportHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	history, err := s.importUC.GetImportHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (s *ImportService) GetImportErrors(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	errors, err := s.importUC.GetImportErrors(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, errors)
}

// ========== 导出配置接口 ==========

func (s *ImportService) CreateExportProfile(c *gin.Context) {
	var req models.ExportProfileCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := s.exportUC.CreateProfile(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, profile)
}

func (s *ImportService) GetExportProfile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	profile, err := s.exportUC.GetProfile(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (s *ImportService) ListExportProfiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	profiles, total, err := s.exportUC.ListProfiles(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": profiles, "total": total})
}

func (s *ImportService) UpdateExportProfile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.ExportProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	profile, err := s.exportUC.UpdateProfile(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

func (s *ImportService) DeleteExportProfile(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.exportUC.DeleteProfile(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *ImportService) ExecuteExport(c *gin.Context) {
	var req models.ExportExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := s.exportUC.ExecuteExport(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (s *ImportService) ListExportHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	history, total, err := s.exportUC.ListExportHistory(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": history, "total": total})
}

func (s *ImportService) GetExportHistory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	history, err := s.exportUC.GetExportHistory(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}

func (s *ImportService) GetEntityTypes(c *gin.Context) {
	types := biz.GetEntityTypes()
	c.JSON(http.StatusOK, types)
}