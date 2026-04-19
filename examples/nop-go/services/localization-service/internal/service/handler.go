// Package service 本地化服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	jwtmiddleware "github.com/ngq/gorp/framework/provider/auth/jwt"
	"nop-go/services/localization-service/internal/biz"
	"nop-go/services/localization-service/internal/models"
)

// LocalizationService 本地化服务
type LocalizationService struct {
	locUC *biz.LocalizationUseCase
	jwtSvc contract.JWTService
}

// NewLocalizationService 创建本地化服务
func NewLocalizationService(locUC *biz.LocalizationUseCase, jwtSvc contract.JWTService) *LocalizationService {
	return &LocalizationService{locUC: locUC, jwtSvc: jwtSvc}
}

// RegisterRoutes 注册路由
func (s *LocalizationService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1/localization")
	adminAuth := jwtmiddleware.AuthMiddleware(s.jwtSvc, "admin")
	{
		// 语言管理
		api.POST("/languages", adminAuth, s.CreateLanguage)
		api.GET("/languages", s.ListLanguages)
		api.GET("/languages/published", s.ListPublishedLanguages)
		api.GET("/languages/:id", s.GetLanguage)
		api.GET("/languages/culture/:culture", s.GetLanguageByCulture)
		api.PUT("/languages/:id", adminAuth, s.UpdateLanguage)
		api.DELETE("/languages/:id", adminAuth, s.DeleteLanguage)

		// 本地化资源管理
		api.POST("/resources", adminAuth, s.CreateResource)
		api.GET("/resources", s.ListResourcesByLanguage)
		api.GET("/resources/:id", s.GetResource)
		api.GET("/resources/search", s.SearchResources)
		api.PUT("/resources/:id", adminAuth, s.UpdateResource)
		api.PUT("/resources/batch", adminAuth, s.BatchUpdateResources)
		api.DELETE("/resources/:id", adminAuth, s.DeleteResource)

		// 翻译接口
		api.GET("/translate", s.Translate)
		api.GET("/translate/:language_id/:resource_name", s.GetTranslation)
		api.GET("/translations/:language_id", s.GetAllTranslations)
		api.GET("/groups/:language_id", s.GetResourceGroups)
		api.GET("/groups/:language_id/:group", s.GetResourcesByGroup)

		// 货币管理
		api.POST("/currencies", adminAuth, s.CreateCurrency)
		api.GET("/currencies", s.ListCurrencies)
		api.GET("/currencies/published", s.ListPublishedCurrencies)
		api.GET("/currencies/:id", s.GetCurrency)
		api.GET("/currencies/code/:code", s.GetCurrencyByCode)
		api.PUT("/currencies/:id", adminAuth, s.UpdateCurrency)
		api.DELETE("/currencies/:id", adminAuth, s.DeleteCurrency)
	}
}

// ========== 语言管理接口 ==========

func (s *LocalizationService) CreateLanguage(c *gin.Context) {
	var req models.LanguageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	language, err := s.locUC.CreateLanguage(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, language)
}

func (s *LocalizationService) GetLanguage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	language, err := s.locUC.GetLanguage(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, language)
}

func (s *LocalizationService) GetLanguageByCulture(c *gin.Context) {
	culture := c.Param("culture")
	language, err := s.locUC.GetLanguageByCulture(c.Request.Context(), culture)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, language)
}

func (s *LocalizationService) ListLanguages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	languages, total, err := s.locUC.ListLanguages(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": languages, "total": total})
}

func (s *LocalizationService) ListPublishedLanguages(c *gin.Context) {
	languages, err := s.locUC.ListPublishedLanguages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, languages)
}

func (s *LocalizationService) UpdateLanguage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.LanguageUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	language, err := s.locUC.UpdateLanguage(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, language)
}

func (s *LocalizationService) DeleteLanguage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.locUC.DeleteLanguage(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== 本地化资源接口 ==========

func (s *LocalizationService) CreateResource(c *gin.Context) {
	var req models.ResourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resource, err := s.locUC.CreateResource(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resource)
}

func (s *LocalizationService) GetResource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	resource, err := s.locUC.GetResource(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resource)
}

func (s *LocalizationService) ListResourcesByLanguage(c *gin.Context) {
	languageID, _ := strconv.ParseUint(c.Query("language_id"), 10, 64)
	if languageID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "language_id is required"})
		return
	}

	resources, err := s.locUC.GetResourcesByLanguage(c.Request.Context(), uint(languageID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resources)
}

func (s *LocalizationService) SearchResources(c *gin.Context) {
	languageID, _ := strconv.ParseUint(c.Query("language_id"), 10, 64)
	if languageID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "language_id is required"})
		return
	}

	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	resources, total, err := s.locUC.SearchResources(c.Request.Context(), uint(languageID), keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": resources, "total": total})
}

func (s *LocalizationService) UpdateResource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.ResourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resource, err := s.locUC.UpdateResource(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resource)
}

func (s *LocalizationService) BatchUpdateResources(c *gin.Context) {
	var req models.ResourceBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.locUC.BatchUpdateResources(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resources updated successfully"})
}

func (s *LocalizationService) DeleteResource(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.locUC.DeleteResource(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== 翻译接口 ==========

func (s *LocalizationService) Translate(c *gin.Context) {
	languageID, _ := strconv.ParseUint(c.Query("language_id"), 10, 64)
	resourceName := c.Query("name")

	if languageID == 0 || resourceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "language_id and name are required"})
		return
	}

	translation, err := s.locUC.GetTranslation(c.Request.Context(), uint(languageID), resourceName)
	if err != nil {
		// 未找到翻译时返回原文
		c.JSON(http.StatusOK, gin.H{"value": resourceName})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": translation})
}

func (s *LocalizationService) GetTranslation(c *gin.Context) {
	languageID, _ := strconv.ParseUint(c.Param("language_id"), 10, 64)
	resourceName := c.Param("resource_name")

	translation, err := s.locUC.GetTranslation(c.Request.Context(), uint(languageID), resourceName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": resourceName, "value": translation, "language_id": languageID})
}

func (s *LocalizationService) GetAllTranslations(c *gin.Context) {
	languageID, _ := strconv.ParseUint(c.Param("language_id"), 10, 64)

	translations, err := s.locUC.GetAllTranslations(c.Request.Context(), uint(languageID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, translations)
}

func (s *LocalizationService) GetResourceGroups(c *gin.Context) {
	languageID, _ := strconv.ParseUint(c.Param("language_id"), 10, 64)

	groups, err := s.locUC.GetResourceGroups(c.Request.Context(), uint(languageID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, groups)
}

func (s *LocalizationService) GetResourcesByGroup(c *gin.Context) {
	languageID, _ := strconv.ParseUint(c.Param("language_id"), 10, 64)
	group := c.Param("group")

	resources, err := s.locUC.GetResourcesByGroup(c.Request.Context(), uint(languageID), group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resources)
}

// ========== 货币接口 ==========

func (s *LocalizationService) CreateCurrency(c *gin.Context) {
	var currency models.Currency
	if err := c.ShouldBindJSON(&currency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.locUC.CreateCurrency(c.Request.Context(), &currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (s *LocalizationService) GetCurrency(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	currency, err := s.locUC.GetCurrency(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, currency)
}

func (s *LocalizationService) GetCurrencyByCode(c *gin.Context) {
	code := c.Param("code")
	currency, err := s.locUC.GetCurrencyByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, currency)
}

func (s *LocalizationService) ListCurrencies(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	currencies, total, err := s.locUC.ListCurrencies(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": currencies, "total": total})
}

func (s *LocalizationService) ListPublishedCurrencies(c *gin.Context) {
	currencies, err := s.locUC.ListPublishedCurrencies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, currencies)
}

func (s *LocalizationService) UpdateCurrency(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var currency models.Currency
	if err := c.ShouldBindJSON(&currency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currency.ID = uint(id)
	if err := s.locUC.UpdateCurrency(c.Request.Context(), &currency); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, currency)
}

func (s *LocalizationService) DeleteCurrency(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.locUC.DeleteCurrency(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}