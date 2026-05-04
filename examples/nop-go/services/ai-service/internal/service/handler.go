// Package service AI鏈嶅姟HTTP澶勭悊灞?
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/ai-service/internal/biz"
	"nop-go/services/ai-service/internal/models"

	"github.com/gin-gonic/gin"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
)

// AIService AI HTTP鏈嶅姟
type AIService struct {
	aiUC        *biz.AIUseCase
	modelUC     *biz.AIModelConfigUseCase
	jwtService  securitycontract.JWTService
}

// NewAIService 鍒涘缓AI鏈嶅姟
func NewAIService(
	aiUC *biz.AIUseCase,
	modelUC *biz.AIModelConfigUseCase,
	jwtService securitycontract.JWTService,
) *AIService {
	return &AIService{
		aiUC:       aiUC,
		modelUC:    modelUC,
		jwtService: jwtService,
	}
}

// RegisterRoutes 娉ㄥ唽璺敱
func (s *AIService) RegisterRoutes(r *gin.Engine) {
	// AI鑱婂ぉ璺敱锛堥渶瑕佽璇侊級
	chat := r.Group("/api/ai/chat")
	chat.Use(s.authMiddleware())
	{
		chat.POST("", s.Chat)                      // 鍙戦€佹秷鎭?
		chat.GET("/:id", s.GetConversation)        // 鑾峰彇瀵硅瘽
		chat.GET("/list", s.ListConversations)     // 瀵硅瘽鍒楄〃
		chat.DELETE("/:id", s.DeleteConversation)  // 鍒犻櫎瀵硅瘽
	}

	// 鍟嗗搧鎺ㄨ崘璺敱锛堥渶瑕佽璇侊級
	recommend := r.Group("/api/ai/recommend")
	recommend.Use(s.authMiddleware())
	{
		recommend.POST("", s.GetRecommendations)               // 鑾峰彇鎺ㄨ崘
		recommend.POST("/:id/click", s.MarkRecommendClicked)   // 鏍囪鐐瑰嚮
		recommend.POST("/:id/purchase", s.MarkRecommendPurchased) // 鏍囪璐拱
	}

	// 鎼滅储寤鸿璺敱锛堥渶瑕佽璇侊級
	suggest := r.Group("/api/ai/suggest")
	suggest.Use(s.authMiddleware())
	{
		suggest.POST("", s.GetSearchSuggestions) // 鎼滅储寤鸿
	}

	// 鍐呭鐢熸垚璺敱锛堥渶瑕佽璇侊級
	content := r.Group("/api/ai/content")
	content.Use(s.authMiddleware())
	{
		content.POST("", s.GenerateContent)                   // 鐢熸垚鍐呭
		content.GET("/:id", s.GetGeneratedContent)            // 鑾峰彇鍐呭
		content.GET("/entity/:entityId/:entityType", s.GetContentByEntity) // 鎸夊疄浣撹幏鍙?
		content.POST("/:id/approve", s.ApproveContent)        // 瀹℃牳閫氳繃
	}

	// AI妯″瀷閰嶇疆璺敱锛堥渶瑕佽璇侊級
	models := r.Group("/api/ai/models")
	models.Use(s.authMiddleware())
	{
		models.GET("", s.ListModels)              // 妯″瀷鍒楄〃
		models.POST("", s.CreateModelConfig)      // 鍒涘缓閰嶇疆
		models.GET("/:id", s.GetModelConfig)      // 鑾峰彇閰嶇疆
		models.PUT("/:id", s.UpdateModelConfig)   // 鏇存柊閰嶇疆
		models.DELETE("/:id", s.DeleteModelConfig) // 鍒犻櫎閰嶇疆
	}
}

// authMiddleware JWT璁よ瘉涓棿浠?
func (s *AIService) authMiddleware() gin.HandlerFunc {
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

// ========== AI鑱婂ぉ ==========

// Chat 鍙戦€佹秷鎭?
func (s *AIService) Chat(c *gin.Context) {
	var req models.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := s.aiUC.Chat(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetConversation 鑾峰彇瀵硅瘽
func (s *AIService) GetConversation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫璇滻D"})
		return
	}

	conv, messages, err := s.aiUC.GetConversation(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": conv,
		"messages":     messages,
	})
}

// ListConversations 瀵硅瘽鍒楄〃
func (s *AIService) ListConversations(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Query("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫鎴稩D"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	convs, total, err := s.aiUC.ListConversations(c.Request.Context(), uint(customerID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      convs,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteConversation 鍒犻櫎瀵硅瘽
func (s *AIService) DeleteConversation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫璇滻D"})
		return
	}

	if err := s.aiUC.DeleteConversation(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "瀵硅瘽宸插垹闄?})
}

// ========== 鍟嗗搧鎺ㄨ崘 ==========

// GetRecommendations 鑾峰彇鎺ㄨ崘
func (s *AIService) GetRecommendations(c *gin.Context) {
	var req models.RecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := s.aiUC.GetRecommendations(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// MarkRecommendClicked 鏍囪鎺ㄨ崘鐐瑰嚮
func (s *AIService) MarkRecommendClicked(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勬帹鑽怚D"})
		return
	}

	if err := s.aiUC.MarkRecommendationClicked(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "宸叉爣璁扮偣鍑?})
}

// MarkRecommendPurchased 鏍囪鎺ㄨ崘璐拱
func (s *AIService) MarkRecommendPurchased(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勬帹鑽怚D"})
		return
	}

	if err := s.aiUC.MarkRecommendationPurchased(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "宸叉爣璁拌喘涔?})
}

// ========== 鎼滅储寤鸿 ==========

// GetSearchSuggestions 鎼滅储寤鸿
func (s *AIService) GetSearchSuggestions(c *gin.Context) {
	var req models.SearchSuggestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := s.aiUC.GetSearchSuggestions(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// ========== 鍐呭鐢熸垚 ==========

// GenerateContent 鐢熸垚鍐呭
func (s *AIService) GenerateContent(c *gin.Context) {
	var req models.GenerateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := s.aiUC.GenerateContent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetGeneratedContent 鑾峰彇鐢熸垚鍐呭
func (s *AIService) GetGeneratedContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫唴瀹笽D"})
		return
	}

	content, err := s.aiUC.GetGeneratedContent(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": content})
}

// GetContentByEntity 鎸夊疄浣撹幏鍙栧唴瀹?
func (s *AIService) GetContentByEntity(c *gin.Context) {
	entityID, err := strconv.ParseUint(c.Param("entityId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫疄浣揑D"})
		return
	}

	entityType := c.Param("entityType")

 contents, err := s.aiUC.GetContentByEntity(c.Request.Context(), uint(entityID), entityType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contents})
}

// ApproveContent 瀹℃牳閫氳繃鍐呭
func (s *AIService) ApproveContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勫唴瀹笽D"})
		return
	}

	if err := s.aiUC.ApproveContent(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "鍐呭宸插鏍搁€氳繃"})
}

// ========== AI妯″瀷閰嶇疆 ==========

// ListModels 妯″瀷鍒楄〃
func (s *AIService) ListModels(c *gin.Context) {
	models, err := s.aiUC.ListModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models})
}

// CreateModelConfig 鍒涘缓妯″瀷閰嶇疆
func (s *AIService) CreateModelConfig(c *gin.Context) {
	var config models.AIModelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.modelUC.CreateModelConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": config})
}

// GetModelConfig 鑾峰彇妯″瀷閰嶇疆
func (s *AIService) GetModelConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勯厤缃甀D"})
		return
	}

	config, err := s.modelUC.GetModelConfig(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// UpdateModelConfig 鏇存柊妯″瀷閰嶇疆
func (s *AIService) UpdateModelConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勯厤缃甀D"})
		return
	}

	var config models.AIModelConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.ID = uint(id)

	if err := s.modelUC.UpdateModelConfig(c.Request.Context(), &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// DeleteModelConfig 鍒犻櫎妯″瀷閰嶇疆
func (s *AIService) DeleteModelConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "鏃犳晥鐨勯厤缃甀D"})
		return
	}

	if err := s.modelUC.DeleteModelConfig(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "閰嶇疆宸插垹闄?})
}