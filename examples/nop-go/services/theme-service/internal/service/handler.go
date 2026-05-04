// Package service 涓婚鏈嶅姟HTTP澶勭悊灞?
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/theme-service/internal/biz"
	"nop-go/services/theme-service/internal/models"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// ThemeService 涓婚HTTP鏈嶅姟
type ThemeService struct {
	themeUC          *biz.ThemeUseCase
	configUC         *biz.ThemeConfigurationUseCase
	customerUC       *biz.CustomerThemeUseCase
	fileUC           *biz.ThemeFileUseCase
	jwtService       securitycontract.JWTService
}

// NewThemeService 鍒涘缓涓婚鏈嶅姟
func NewThemeService(
	themeUC *biz.ThemeUseCase,
	configUC *biz.ThemeConfigurationUseCase,
	customerUC *biz.CustomerThemeUseCase,
	fileUC *biz.ThemeFileUseCase,
	jwtService securitycontract.JWTService,
) *ThemeService {
	return &ThemeService{
		themeUC:    themeUC,
		configUC:   configUC,
		customerUC: customerUC,
		fileUC:     fileUC,
		jwtService: jwtService,
	}
}

// RegisterRoutes 娉ㄥ唽璺敱
func (s *ThemeService) RegisterRoutes(r *gin.Engine) {
	// 涓婚绠＄悊璺敱锛堥渶瑕佽璇侊級
	themes := r.Group("/api/themes")
	themes.Use(s.authMiddleware())
	{
		themes.GET("", s.ListThemes)                // 涓婚鍒楄〃
		themes.GET("/active", s.ListActiveThemes)   // 婵€娲荤殑涓婚
		themes.GET("/default", s.GetDefaultTheme)   // 榛樿涓婚
		themes.POST("", s.CreateTheme)              // 鍒涘缓涓婚
		themes.GET("/:id", s.GetTheme)              // 涓婚璇︽儏
		themes.GET("/:id/preview", s.GetThemePreview) // 涓婚棰勮锛堝惈鍙橀噺锛?
		themes.PUT("/:id", s.UpdateTheme)           // 鏇存柊涓婚
		themes.DELETE("/:id", s.DeleteTheme)        // 鍒犻櫎涓婚
	}

	// 涓婚鍙橀噺璺敱
	variables := r.Group("/api/themes/:themeId/variables")
	variables.Use(s.authMiddleware())
	{
		variables.GET("", s.GetVariables)           // 鍙橀噺鍒楄〃
		variables.POST("", s.CreateVariable)        // 鍒涘缓鍙橀噺
		variables.PUT("/:id", s.UpdateVariable)     // 鏇存柊鍙橀噺
		variables.DELETE("/:id", s.DeleteVariable)  // 鍒犻櫎鍙橀噺
	}

	// 涓婚閰嶇疆璺敱锛堝簵閾虹骇鍒級
	configs := r.Group("/api/themes/configurations")
	configs.Use(s.authMiddleware())
	{
		configs.GET("/store/:storeId", s.GetStoreConfigurations) // 搴楅摵閰嶇疆鍒楄〃
		configs.GET("/:themeId/store/:storeId", s.GetThemeConfiguration) // 鐗瑰畾閰嶇疆
		configs.PUT("", s.UpdateConfiguration)                   // 鏇存柊閰嶇疆
		configs.DELETE("/:id", s.DeleteConfiguration)            // 鍒犻櫎閰嶇疆
	}

	// 瀹㈡埛涓婚璁剧疆璺敱
	customer := r.Group("/api/themes/customer")
	customer.Use(s.authMiddleware())
	{
		customer.GET("/:customerId", s.GetCustomerTheme)       // 瀹㈡埛涓婚
		customer.PUT("", s.SetCustomerTheme)                   // 璁剧疆瀹㈡埛涓婚
		customer.DELETE("/:customerId", s.DeleteCustomerTheme) // 鍒犻櫎瀹㈡埛涓婚
	}

	// 涓婚鏂囦欢璺敱
	files := r.Group("/api/themes/:themeId/files")
	files.Use(s.authMiddleware())
	{
		files.GET("", s.GetThemeFiles)            // 鏂囦欢鍒楄〃
		files.GET("/:id", s.GetThemeFile)         // 鏂囦欢璇︽儏
		files.GET("/path", s.GetThemeFileByPath)  // 閫氳繃璺緞鑾峰彇鏂囦欢
		files.POST("", s.CreateThemeFile)         // 鍒涘缓鏂囦欢
		files.PUT("/:id", s.UpdateThemeFile)      // 鏇存柊鏂囦欢
		files.DELETE("/:id", s.DeleteThemeFile)   // 鍒犻櫎鏂囦欢
	}
}

// authMiddleware JWT璁よ瘉涓棿浠?
func (s *ThemeService) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缂哄皯璁よ瘉浠ょ墝"})
			c.Abort()
			return
		}

		// 鍘婚櫎 Bearer 鍓嶇紑
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}

		claims, err := s.jwtService.Verify(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "浠ょ墝鏃犳晥"})
			c.Abort()
			return
		}

		c.Set("subject_id", claims.SubjectID)
		c.Set("claims", claims)
		c.Next()
	}
}

// ========== 涓婚绠＄悊 ==========

// ListThemes 涓婚鍒楄〃
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

// ListActiveThemes 婵€娲荤殑涓婚鍒楄〃
func (s *ThemeService) ListActiveThemes(c *gin.Context) {
	themes, err := s.themeUC.ListActiveThemes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": themes})
}

// GetDefaultTheme 鑾峰彇榛樿涓婚
func (s *ThemeService) GetDefaultTheme(c *gin.Context) {
	theme, err := s.themeUC.GetDefaultTheme(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": theme})
}

// CreateTheme 鍒涘缓涓婚
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

// GetTheme 鑾峰彇涓婚璇︽儏
func (s *ThemeService) GetTheme(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
		return
	}

	theme, err := s.themeUC.GetTheme(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": theme})
}

// GetThemePreview 鑾峰彇涓婚棰勮锛堝惈鍙橀噺锛?
func (s *ThemeService) GetThemePreview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
		return
	}

	preview, err := s.themeUC.GetThemeWithVariables(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": preview})
}

// UpdateTheme 鏇存柊涓婚
func (s *ThemeService) UpdateTheme(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
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

// DeleteTheme 鍒犻櫎涓婚
func (s *ThemeService) DeleteTheme(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
		return
	}

	if err := s.themeUC.DeleteTheme(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "涓婚宸插垹闄?})
}

// ========== 涓婚鍙橀噺 ==========

// GetVariables 鑾峰彇涓婚鍙橀噺鍒楄〃
func (s *ThemeService) GetVariables(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
		return
	}

	variables, err := s.themeUC.GetVariablesByThemeID(c.Request.Context(), uint(themeId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": variables})
}

// CreateVariable 鍒涘缓涓婚鍙橀噺
func (s *ThemeService) CreateVariable(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
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

// UpdateVariable 鏇存柊涓婚鍙橀噺
func (s *ThemeService) UpdateVariable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫彉閲廔D"})
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

// DeleteVariable 鍒犻櫎涓婚鍙橀噺
func (s *ThemeService) DeleteVariable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫彉閲廔D"})
		return
	}

	if err := s.themeUC.DeleteVariable(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "鍙橀噺宸插垹闄?})
}

// ========== 涓婚閰嶇疆 ==========

// GetStoreConfigurations 鑾峰彇搴楅摵鐨勪富棰橀厤缃?
func (s *ThemeService) GetStoreConfigurations(c *gin.Context) {
	storeId, err := strconv.ParseUint(c.Param("storeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫簵閾篒D"})
		return
	}

	configs, err := s.configUC.GetStoreConfiguration(c.Request.Context(), uint(storeId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": configs})
}

// GetThemeConfiguration 鑾峰彇鐗瑰畾涓婚鍜屽簵閾虹殑閰嶇疆
func (s *ThemeService) GetThemeConfiguration(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
		return
	}

	storeId, err := strconv.ParseUint(c.Param("storeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫簵閾篒D"})
		return
	}

	config, err := s.configUC.GetThemeConfiguration(c.Request.Context(), uint(themeId), uint(storeId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// UpdateConfiguration 鏇存柊涓婚閰嶇疆
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

// DeleteConfiguration 鍒犻櫎涓婚閰嶇疆
func (s *ThemeService) DeleteConfiguration(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勯厤缃甀D"})
		return
	}

	if err := s.configUC.DeleteConfiguration(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "閰嶇疆宸插垹闄?})
}

// ========== 瀹㈡埛涓婚璁剧疆 ==========

// GetCustomerTheme 鑾峰彇瀹㈡埛涓婚璁剧疆
func (s *ThemeService) GetCustomerTheme(c *gin.Context) {
	customerId, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫鎴稩D"})
		return
	}

	setting, err := s.customerUC.GetCustomerTheme(c.Request.Context(), uint(customerId))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": setting})
}

// SetCustomerTheme 璁剧疆瀹㈡埛涓婚
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

// DeleteCustomerTheme 鍒犻櫎瀹㈡埛涓婚璁剧疆
func (s *ThemeService) DeleteCustomerTheme(c *gin.Context) {
	customerId, err := strconv.ParseUint(c.Param("customerId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫鎴稩D"})
		return
	}

	if err := s.customerUC.DeleteCustomerTheme(c.Request.Context(), uint(customerId)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "瀹㈡埛涓婚璁剧疆宸插垹闄?})
}

// ========== 涓婚鏂囦欢 ==========

// GetThemeFiles 鑾峰彇涓婚鏂囦欢鍒楄〃
func (s *ThemeService) GetThemeFiles(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
		return
	}

	files, err := s.fileUC.GetThemeFiles(c.Request.Context(), uint(themeId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": files})
}

// GetThemeFile 鑾峰彇涓婚鏂囦欢璇︽儏
func (s *ThemeService) GetThemeFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勬枃浠禝D"})
		return
	}

	file, err := s.fileUC.GetThemeFile(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": file})
}

// GetThemeFileByPath 閫氳繃璺緞鑾峰彇涓婚鏂囦欢
func (s *ThemeService) GetThemeFileByPath(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缂哄皯鏂囦欢璺緞"})
		return
	}

	file, err := s.fileUC.GetThemeFileByPath(c.Request.Context(), uint(themeId), filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": file})
}

// CreateThemeFile 鍒涘缓涓婚鏂囦欢
func (s *ThemeService) CreateThemeFile(c *gin.Context) {
	themeId, err := strconv.ParseUint(c.Param("themeId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勪富棰業D"})
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

// UpdateThemeFile 鏇存柊涓婚鏂囦欢
func (s *ThemeService) UpdateThemeFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勬枃浠禝D"})
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

// DeleteThemeFile 鍒犻櫎涓婚鏂囦欢
func (s *ThemeService) DeleteThemeFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勬枃浠禝D"})
		return
	}

	if err := s.fileUC.DeleteThemeFile(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "鏂囦欢宸插垹闄?})
}