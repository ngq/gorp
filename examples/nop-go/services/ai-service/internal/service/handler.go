// Package service AI服务HTTP处理层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/ai-service/internal/biz"
	"nop-go/services/ai-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
)

// AIService AI HTTP服务
type AIService struct {
	aiUC        *biz.AIUseCase
	modelUC     *biz.AIModelConfigUseCase
	jwtService  contract.JWTService
}

// NewAIService 创建AI服务
func NewAIService(
	aiUC *biz.AIUseCase,
	modelUC *biz.AIModelConfigUseCase,
	jwtService contract.JWTService,
) *AIService {
	return &AIService{
		aiUC:       aiUC,
		modelUC:    modelUC,
		jwtService: jwtService,
	}
}

// RegisterRoutes 注册路由
func (s *AIService) RegisterRoutes(r *gin.Engine) {
	// AI聊天路由（需要认证）
	chat := r.Group("/api/ai/chat")
	chat.Use(s.authMiddleware())
	{
		chat.POST("", s.Chat)                      // 发送消息
		chat.GET("/:id", s.GetConversation)        // 获取对话
		chat.GET("/list", s.ListConversations)     // 对话列表
		chat.DELETE("/:id", s.DeleteConversation)  // 删除对话
	}

	// 商品推荐路由（需要认证）
	recommend := r.Group("/api/ai/recommend")
	recommend.Use(s.authMiddleware())
	{
		recommend.POST("", s.GetRecommendations)               // 获取推荐
		recommend.POST("/:id/click", s.MarkRecommendClicked)   // 标记点击
		recommend.POST("/:id/purchase", s.MarkRecommendPurchased) // 标记购买
	}

	// 搜索建议路由（需要认证）
	suggest := r.Group("/api/ai/suggest")
	suggest.Use(s.authMiddleware())
	{
		suggest.POST("", s.GetSearchSuggestions) // 搜索建议
	}

	// 内容生成路由（需要认证）
	content := r.Group("/api/ai/content")
	content.Use(s.authMiddleware())
	{
		content.POST("", s.GenerateContent)                   // 生成内容
		content.GET("/:id", s.GetGeneratedContent)            // 获取内容
		content.GET("/entity/:entityId/:entityType", s.GetContentByEntity) // 按实体获取
		content.POST("/:id/approve", s.ApproveContent)        // 审核通过
	}

	// AI模型配置路由（需要认证）
	models := r.Group("/api/ai/models")
	models.Use(s.authMiddleware())
	{
		models.GET("", s.ListModels)              // 模型列表
		models.POST("", s.CreateModelConfig)      // 创建配置
		models.GET("/:id", s.GetModelConfig)      // 获取配置
		models.PUT("/:id", s.UpdateModelConfig)   // 更新配置
		models.DELETE("/:id", s.DeleteModelConfig) // 删除配置
	}
}

// authMiddleware JWT认证中间件
func (s *AIService) authMiddleware() gin.HandlerFunc {
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

// ========== AI聊天 ==========

// Chat 发送消息
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

// GetConversation 获取对话
func (s *AIService) GetConversation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID"})
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

// ListConversations 对话列表
func (s *AIService) ListConversations(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Query("customer_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的客户ID"})
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

// DeleteConversation 删除对话
func (s *AIService) DeleteConversation(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的对话ID"})
		return
	}

	if err := s.aiUC.DeleteConversation(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "对话已删除"})
}

// ========== 商品推荐 ==========

// GetRecommendations 获取推荐
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

// MarkRecommendClicked 标记推荐点击
func (s *AIService) MarkRecommendClicked(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的推荐ID"})
		return
	}

	if err := s.aiUC.MarkRecommendationClicked(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记点击"})
}

// MarkRecommendPurchased 标记推荐购买
func (s *AIService) MarkRecommendPurchased(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的推荐ID"})
		return
	}

	if err := s.aiUC.MarkRecommendationPurchased(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记购买"})
}

// ========== 搜索建议 ==========

// GetSearchSuggestions 搜索建议
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

// ========== 内容生成 ==========

// GenerateContent 生成内容
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

// GetGeneratedContent 获取生成内容
func (s *AIService) GetGeneratedContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的内容ID"})
		return
	}

	content, err := s.aiUC.GetGeneratedContent(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": content})
}

// GetContentByEntity 按实体获取内容
func (s *AIService) GetContentByEntity(c *gin.Context) {
	entityID, err := strconv.ParseUint(c.Param("entityId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的实体ID"})
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

// ApproveContent 审核通过内容
func (s *AIService) ApproveContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的内容ID"})
		return
	}

	if err := s.aiUC.ApproveContent(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "内容已审核通过"})
}

// ========== AI模型配置 ==========

// ListModels 模型列表
func (s *AIService) ListModels(c *gin.Context) {
	models, err := s.aiUC.ListModels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": models})
}

// CreateModelConfig 创建模型配置
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

// GetModelConfig 获取模型配置
func (s *AIService) GetModelConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	config, err := s.modelUC.GetModelConfig(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// UpdateModelConfig 更新模型配置
func (s *AIService) UpdateModelConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
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

// DeleteModelConfig 删除模型配置
func (s *AIService) DeleteModelConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的配置ID"})
		return
	}

	if err := s.modelUC.DeleteModelConfig(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "配置已删除"})
}