// Package service 主题服务HTTP处理层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/theme-service/internal/biz"
	"nop-go/services/theme-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

// ThemeService 主题HTTP服务
type ThemeService struct {
	themeUC          *biz.ThemeUseCase
	configUC         *biz.ThemeConfigurationUseCase
	customerUC       *biz.CustomerThemeUseCase
	fileUC           *biz.ThemeFileUseCase
	jwtService       contract.JWTService
}

// NewThemeService 创建主题服务
func NewThemeService(
	themeUC *biz.ThemeUseCase,
	configUC *biz.ThemeConfigurationUseCase,
	customerUC *biz.CustomerThemeUseCase,
	fileUC *biz.ThemeFileUseCase,
	jwtService contract.JWTService,
) *ThemeService {
	return &ThemeService{
		themeUC:    themeUC,
		configUC:   configUC,
		customerUC: customerUC,
		fileUC:     fileUC,
		jwtService: jwtService,
	}
}

// RegisterRoutes 注册路由
func (s *ThemeService) RegisterRoutes(r *gin.Engine) {
	// 主题管理路由（需要认证）
	themes := r.Group("/api/themes")
	themes.Use(s.authMiddleware())
	{
		themes.GET("", s.ListThemes)                // 主题列表
		themes.GET("/active", s.ListActiveThemes)   // 激活的主题
		themes.GET("/default", s.GetDefaultTheme)   // 默认主题
		themes.POST("", s.CreateTheme)              // 创建主题
		themes.GET("/:id", s.GetTheme)              // 主题详情
		themes.GET("/:id/preview", s.GetThemePreview) // 主题预览（含变量）
		themes.PUT("/:id", s.UpdateTheme)           // 更新主题
		themes.DELETE("/:id", s.DeleteTheme)        // 删除主题
	}

	// 主题变量路由
	variables := r.Group("/api/themes/:themeId/variables")
	variables.Use(s.authMiddleware())
	{
		variables.GET("", s.GetVariables)           // 变量列表
		variables.POST("", s.CreateVariable)        // 创建变量
		variables.PUT("/:id", s.UpdateVariable)     // 更新变量
		variables.DELETE("/:id", s.DeleteVariable)  // 删除变量
	}

	// 主题配置路由（店铺级别）
	configs := r.Group("/api/themes/configurations")
	configs.Use(s.authMiddleware())
	{
		configs.GET("/store/:storeId", s.GetStoreConfigurations) // 店铺配置列表
		configs.GET("/:themeId/store/:storeId", s.GetThemeConfiguration) // 特定配置
		configs.PUT("", s.UpdateConfiguration)                   // 更新配置
		configs.DELETE("/:id", s.DeleteConfiguration)            // 删除配置
	}

	// 客户主题设置路由
	customer := r.Group("/api/themes/customer")
	customer.Use(s.authMiddleware())
	{
		customer.GET("/:customerId", s.GetCustomerTheme)       // 客户主题
		customer.PUT("", s.SetCustomerTheme)                   // 设置客户主题
		customer.DELETE("/:customerId", s.DeleteCustomerTheme) // 删除客户主题
	}

	// 主题文件路由
	files := r.Group("/api/themes/:themeId/files")
	files.Use(s.authMiddleware())
	{
		files.GET("", s.GetThemeFiles)            // 文件列表
		files.GET("/:id", s.GetThemeFile)         // 文件详情
		files.GET("/path", s.GetThemeFileByPath)  // 通过路径获取文件
		files.POST("", s.CreateThemeFile)         // 创建文件
		files.PUT("/:id", s.UpdateThemeFile)      // 更新文件
		files.DELETE("/:id", s.DeleteThemeFile)   // 删除文件
	}
}

// authMiddleware JWT认证中间件
func (s *ThemeService) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		// 去除 Bearer 前缀
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		claims, err := s.jwtService.Verify(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌无效"})
			c.Abort()
			return
		}

		c.Set("subject_id", claims.SubjectID)
		c.Set("claims", claims)
		c.Next()
	}
}

// ========== 主题管理 ==========

// ListThemes 主题列表
func (s *ThemeService) ListThemes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	themes, total, err := s.themeUC.ListThemes(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  themes,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// ListActiveThemes 激活的主题列表
func (s *ThemeService) ListActiveThemes(c *gin.Context) {
	themes, err := s.themeUC.ListActiveThemes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": themes})
}

// GetDefaultTheme 获取默认主题
func (s *ThemeService) GetDefaultTheme(c *gin.Context) {
	theme, err := s.themeUC.GetDefaultTheme(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": theme})
}

// CreateTheme 创建主题
func (s *ThemeService) CreateTheme(c *gin.Context) {
	var req models.ThemeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	theme, err := s.themeUC.CreateTheme(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": theme})
}

// GetTheme 获取主题详情
func (s *ThemeService) GetTheme(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	theme, err := s.themeUC.GetTheme(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": theme})
}

// GetThemePreview 获取主题预览（含变量）
func (s *ThemeService) GetThemePreview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	preview, err := s.themeUC.GetThemeWithVariables(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": preview})
}

// UpdateTheme 更新主题
func (s *ThemeService) UpdateTheme(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	var req models.ThemeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	theme, err := s.themeUC.UpdateTheme(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": theme})
}

// DeleteTheme 删除主题
func (s *ThemeService) DeleteTheme(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	if err := s.themeUC.DeleteTheme(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "主题已删除"})
}

// ========== 主题变量 ==========

// GetVariables 获取主题变量列表
func (s *ThemeService) GetVariables(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	variables, err := s.themeUC.GetVariablesByThemeID(c.Request.Context(), uint(themeId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": variables})
}

// CreateVariable 创建主题变量
func (s *ThemeService) CreateVariable(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	var req models.ThemeVariableCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ThemeID = uint(themeId)

	variable, err := s.themeUC.CreateVariable(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": variable})
}

// UpdateVariable 更新主题变量
func (s *ThemeService) UpdateVariable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的变量ID"})
		return
	}

	var req models.ThemeVariableUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variable, err := s.themeUC.UpdateVariable(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": variable})
}

// DeleteVariable 删除主题变量
func (s *ThemeService) DeleteVariable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的变量ID"})
		return
	}

	if err := s.themeUC.DeleteVariable(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "变量已删除"})
}

// ========== 主题配置 ==========

// GetStoreConfigurations 获取店铺的主题配置
func (s *ThemeService) GetStoreConfigurations(c *gin.Context) {
	storeId, err := strconv.ParseUint(c.Param("storeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的店铺ID"})
		return
	}

	configs, err := s.configUC.GetStoreConfiguration(c.Request.Context(), uint(storeId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": configs})
}

// GetThemeConfiguration 获取特定主题和店铺的配置
func (s *ThemeService) GetThemeConfiguration(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	storeId, err := strconv.ParseUint(c.Param("storeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的店铺ID"})
		return
	}

	config, err := s.configUC.GetThemeConfiguration(c.Request.Context(), uint(themeId), uint(storeId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// UpdateConfiguration 更新主题配置
func (s *ThemeService) UpdateConfiguration(c *gin.Context) {
	var req models.ThemeConfigurationUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config, err := s.configUC.UpdateConfiguration(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// DeleteConfiguration 删除主题配置
func (s *ThemeService) DeleteConfiguration(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	if err := s.configUC.DeleteConfiguration(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已删除"})
}

// ========== 客户主题设置 ==========

// GetCustomerTheme 获取客户主题设置
func (s *ThemeService) GetCustomerTheme(c *gin.Context) {
	customerId, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
		return
	}

	setting, err := s.customerUC.GetCustomerTheme(c.Request.Context(), uint(customerId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": setting})
}

// SetCustomerTheme 设置客户主题
func (s *ThemeService) SetCustomerTheme(c *gin.Context) {
	var req models.CustomerThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setting, err := s.customerUC.SetCustomerTheme(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": setting})
}

// DeleteCustomerTheme 删除客户主题设置
func (s *ThemeService) DeleteCustomerTheme(c *gin.Context) {
	customerId, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
		return
	}

	if err := s.customerUC.DeleteCustomerTheme(c.Request.Context(), uint(customerId)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "客户主题设置已删除"})
}

// ========== 主题文件 ==========

// GetThemeFiles 获取主题文件列表
func (s *ThemeService) GetThemeFiles(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	files, err := s.fileUC.GetThemeFiles(c.Request.Context(), uint(themeId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": files})
}

// GetThemeFile 获取主题文件详情
func (s *ThemeService) GetThemeFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	file, err := s.fileUC.GetThemeFile(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": file})
}

// GetThemeFileByPath 通过路径获取主题文件
func (s *ThemeService) GetThemeFileByPath(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件路径"})
		return
	}

	file, err := s.fileUC.GetThemeFileByPath(c.Request.Context(), uint(themeId), filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": file})
}

// CreateThemeFile 创建主题文件
func (s *ThemeService) CreateThemeFile(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主题ID"})
		return
	}

	var file models.ThemeFile
	if err := c.ShouldBindJSON(&file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file.ThemeID = uint(themeId)

	if err := s.fileUC.CreateThemeFile(c.Request.Context(), &file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": file})
}

// UpdateThemeFile 更新主题文件
func (s *ThemeService) UpdateThemeFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var file models.ThemeFile
	if err := c.ShouldBindJSON(&file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file.ID = uint(id)

	if err := s.fileUC.UpdateThemeFile(c.Request.Context(), &file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": file})
}

// DeleteThemeFile 删除主题文件
func (s *ThemeService) DeleteThemeFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	if err := s.fileUC.DeleteThemeFile(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件已删除"})
}