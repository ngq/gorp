// Package biz AI服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nop-go/services/ai-service/internal/data"
	"nop-go/services/ai-service/internal/models"

	"gorm.io/gorm"
)

// AIConfig AI配置
type AIConfig struct {
	Provider    string  // AI提供商: openai/azure/custom
	APIKey      string  // API密钥
	APIEndpoint string  // API端点
	Model       string  // 默认模型
	MaxTokens   int     // 最大tokens
	Temperature float64 // 温度参数
	Timeout     int     // 超时时间(秒)
}

// AIUseCase AI用例
type AIUseCase struct {
	convRepo    data.AIConversationRepository
	msgRepo     data.AIMessageRepository
	recRepo     data.AIRecommendationRepository
	sugRepo     data.AISearchSuggestionRepository
	contentRepo data.AIGeneratedContentRepository
	modelRepo   data.AIModelConfigRepository
	config      AIConfig
}

// NewAIUseCase 创建AI用例
func NewAIUseCase(
	convRepo data.AIConversationRepository,
	msgRepo data.AIMessageRepository,
	recRepo data.AIRecommendationRepository,
	sugRepo data.AISearchSuggestionRepository,
	contentRepo data.AIGeneratedContentRepository,
	modelRepo data.AIModelConfigRepository,
	config AIConfig,
) *AIUseCase {
	return &AIUseCase{
		convRepo:    convRepo,
		msgRepo:     msgRepo,
		recRepo:     recRepo,
		sugRepo:     sugRepo,
		contentRepo: contentRepo,
		modelRepo:   modelRepo,
		config:      config,
	}
}

// ========== AI聊天 ==========

// Chat 处理聊天请求
func (uc *AIUseCase) Chat(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// 获取或创建对话
	var conv *models.AIConversation
	var sessionID string

	if req.SessionID != "" {
		// 使用已有会话
		conv, err := uc.convRepo.GetBySessionID(ctx, req.SessionID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if conv != nil {
			sessionID = req.SessionID
		}
	}

	if conv == nil {
		// 创建新对话
		sessionID = generateSessionID()
		conv = &models.AIConversation{
			SessionID: sessionID,
			Status:    "active",
			ModelUsed: uc.getModelName(req.Model),
		}
		if err := uc.convRepo.Create(ctx, conv); err != nil {
			return nil, err
		}
	}

	// 保存用户消息
	userMsg := &models.AIMessage{
		ConversationID: conv.ID,
		Role:           "user",
		Content:        req.Message,
	}
	if err := uc.msgRepo.Create(ctx, userMsg); err != nil {
		return nil, err
	}

	// 获取对话历史作为上下文
	history, err := uc.msgRepo.GetByConversationID(ctx, conv.ID)
	if err != nil {
		return nil, err
	}

	// 构建AI请求上下文
	context := buildChatContext(history, req.Context)

	// 调用AI服务（这里模拟实现，实际项目需要调用真实AI API）
	response, tokensUsed, err := uc.callAI(ctx, context, req.Message, req.Model, req.MaxTokens)
	if err != nil {
		return nil, err
	}

	// 保存AI回复
	assistantMsg := &models.AIMessage{
		ConversationID: conv.ID,
		Role:           "assistant",
		Content:        response,
		Tokens:         tokensUsed,
	}
	if err := uc.msgRepo.Create(ctx, assistantMsg); err != nil {
		return nil, err
	}

	// 更新对话的tokens统计
	conv.TokensUsed += tokensUsed + estimateTokens(req.Message)
	if err := uc.convRepo.Update(ctx, conv); err != nil {
		return nil, err
	}

	return &models.ChatResponse{
		SessionID:     sessionID,
		Response:      response,
		TokensUsed:    tokensUsed,
		ModelUsed:     conv.ModelUsed,
		ConversationID: conv.ID,
	}, nil
}

// GetConversation 获取对话详情
func (uc *AIUseCase) GetConversation(ctx context.Context, conversationID uint) (*models.AIConversation, []*models.AIMessage, error) {
	conv, err := uc.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, data.ErrConversationNotFound
		}
		return nil, nil, err
	}

	messages, err := uc.msgRepo.GetByConversationID(ctx, conversationID)
	if err != nil {
		return nil, nil, err
	}

	return conv, messages, nil
}

// ListConversations 获取用户对话列表
func (uc *AIUseCase) ListConversations(ctx context.Context, customerID uint, page, pageSize int) ([]*models.AIConversation, int64, error) {
	return uc.convRepo.GetByCustomerID(ctx, customerID, page, pageSize)
}

// DeleteConversation 删除对话
func (uc *AIUseCase) DeleteConversation(ctx context.Context, conversationID uint) error {
	// 删除消息
	if err := uc.msgRepo.DeleteByConversationID(ctx, conversationID); err != nil {
		return err
	}
	// 删除对话
	return uc.convRepo.Delete(ctx, conversationID)
}

// ========== 商品推荐 ==========

// GetRecommendations 获取商品推荐
func (uc *AIUseCase) GetRecommendations(ctx context.Context, req *models.RecommendRequest) (*models.RecommendResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// 获取已有的推荐
	existing, err := uc.recRepo.GetByCustomerID(ctx, req.CustomerID, req.Limit)
	if err != nil {
		return nil, err
	}

	// 如果已有足够推荐，直接返回
	if len(existing) >= req.Limit {
		products := make([]models.RecommendedProduct, len(existing))
		for i, rec := range existing {
			products[i] = models.RecommendedProduct{
				ProductID: rec.ProductID,
				Score:     rec.Score,
			}
			if req.IncludeReason {
				products[i].Reason = rec.Reason
			}
		}
		return &models.RecommendResponse{Products: products}, nil
	}

	// 生成新的推荐（模拟实现，实际项目需要结合用户行为数据和商品数据）
	newRecs := uc.generateRecommendations(ctx, req)

	// 保存新推荐
	if len(newRecs) > 0 {
		recModels := make([]*models.AIRecommendation, len(newRecs))
		for i, rec := range newRecs {
			recModels[i] = &models.AIRecommendation{
				CustomerID: req.CustomerID,
				ProductID:  rec.ProductID,
				Reason:     rec.Reason,
				Score:      rec.Score,
				Source:     "ai",
			}
		}
		if err := uc.recRepo.CreateBatch(ctx, recModels); err != nil {
			return nil, err
		}
	}

	// 合并返回
	allProducts := make([]models.RecommendedProduct, 0)
	for _, rec := range existing {
		allProducts = append(allProducts, models.RecommendedProduct{
			ProductID: rec.ProductID,
			Score:     rec.Score,
		})
	}
	allProducts = append(allProducts, newRecs...)

	return &models.RecommendResponse{Products: allProducts[:min(len(allProducts), req.Limit)]}, nil
}

// generateRecommendations 生成推荐（模拟实现）
func (uc *AIUseCase) generateRecommendations(ctx context.Context, req *models.RecommendRequest) []models.RecommendedProduct {
	// 这里是模拟实现，实际项目需要：
	// 1. 获取用户历史购买/浏览数据
	// 2. 结合商品特征进行相似度计算
	// 3. 或调用外部推荐服务/API
	return []models.RecommendedProduct{}
}

// MarkRecommendationClicked 标记推荐已点击
func (uc *AIUseCase) MarkRecommendationClicked(ctx context.Context, id uint) error {
	rec, err := uc.recRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	rec.IsClicked = true
	return uc.recRepo.Update(ctx, rec)
}

// MarkRecommendationPurchased 标记推荐已购买
func (uc *AIUseCase) MarkRecommendationPurchased(ctx context.Context, id uint) error {
	rec, err := uc.recRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	rec.IsPurchased = true
	return uc.recRepo.Update(ctx, rec)
}

// ========== 搜索建议 ==========

// GetSearchSuggestions 获取搜索建议
func (uc *AIUseCase) GetSearchSuggestions(ctx context.Context, req *models.SearchSuggestRequest) (*models.SearchSuggestResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 5
	}

	// 查询已有建议
	suggestions, err := uc.sugRepo.GetByQuery(ctx, req.Query)
	if err != nil {
		return nil, err
	}

	result := make([]models.SearchSuggestion, min(len(suggestions), req.Limit))
	for i, sug := range suggestions[:min(len(suggestions), req.Limit)] {
		result[i] = models.SearchSuggestion{
			Query:      sug.Query,
			Suggestion: sug.Suggestion,
			Type:       sug.Type,
		}
	}

	// 如果已有建议不足，可以调用AI生成新建议（这里省略实现）

	return &models.SearchSuggestResponse{Suggestions: result}, nil
}

// ========== 内容生成 ==========

// GenerateContent 生成内容
func (uc *AIUseCase) GenerateContent(ctx context.Context, req *models.GenerateContentRequest) (*models.GenerateContentResponse, error) {
	// 构建生成参数
	prompt := uc.buildGeneratePrompt(req)

	// 调用AI生成（模拟实现）
	content, tokensUsed, err := uc.callAI(ctx, []string{}, prompt, req.Model, uc.config.MaxTokens)
	if err != nil {
		return nil, err
	}

	// 保存生成内容
	generated := &models.AIGeneratedContent{
		Type:       req.Type,
		EntityID:   req.EntityID,
		EntityType: req.EntityType,
		Content:    content,
		ModelUsed:  uc.getModelName(req.Model),
		TokensUsed: tokensUsed,
		IsApproved: false,
	}
	if err := uc.contentRepo.Create(ctx, generated); err != nil {
		return nil, err
	}

	return &models.GenerateContentResponse{
		ContentID:  generated.ID,
		Content:    content,
		TokensUsed: tokensUsed,
		ModelUsed:  generated.ModelUsed,
	}, nil
}

// ApproveContent 审核通过内容
func (uc *AIUseCase) ApproveContent(ctx context.Context, contentID uint) error {
	content, err := uc.contentRepo.GetByID(ctx, contentID)
	if err != nil {
		return err
	}
	content.IsApproved = true
	return uc.contentRepo.Update(ctx, content)
}

// GetGeneratedContent 获取生成内容
func (uc *AIUseCase) GetGeneratedContent(ctx context.Context, contentID uint) (*models.AIGeneratedContent, error) {
	return uc.contentRepo.GetByID(ctx, contentID)
}

// GetContentByEntity 获取实体的生成内容
func (uc *AIUseCase) GetContentByEntity(ctx context.Context, entityID uint, entityType string) ([]*models.AIGeneratedContent, error) {
	return uc.contentRepo.GetByEntity(ctx, entityID, entityType)
}

// ========== AI模型配置 ==========

// ListModels 获取可用模型列表
func (uc *AIUseCase) ListModels(ctx context.Context) ([]*models.AIModelConfig, error) {
	return uc.modelRepo.List(ctx)
}

// ========== 辅助方法 ==========

// getModelName 获取模型名称
func (uc *AIUseCase) getModelName(model string) string {
	if model != "" {
		return model
	}
	return uc.config.Model
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

// buildChatContext 构建聊天上下文
func buildChatContext(history []*models.AIMessage, extraContext []string) []string {
	context := extraContext
	for _, msg := range history {
		if msg.Role == "user" || msg.Role == "assistant" {
			context = append(context, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
		}
	}
	return context
}

// estimateTokens 估算tokens数（简化实现）
func estimateTokens(text string) int {
	// 简化估算：平均4个字符约1个token
	return len(text) / 4
}

// callAI 调用AI服务（模拟实现）
func (uc *AIUseCase) callAI(ctx context.Context, context []string, message string, model string, maxTokens int) (string, int, error) {
	// 这里是模拟实现
	// 实际项目需要根据 Provider 调用对应的AI API
	// 例如：OpenAI、Azure OpenAI、或自定义AI服务

	if uc.config.APIKey == "" {
		// 没有配置API Key，返回模拟响应
		return "这是一个模拟的AI响应。实际项目需要配置AI服务。", estimateTokens(message) + 50, nil
	}

	// 实际实现示例（伪代码）：
	// client := NewAIClient(uc.config.Provider, uc.config.APIKey, uc.config.APIEndpoint)
	// request := AIRequest{
	//     Model: uc.getModelName(model),
	//     Messages: append(context, message),
	//     MaxTokens: maxTokens,
	//     Temperature: uc.config.Temperature,
	// }
	// return client.Chat(ctx, request)

	return "AI响应内容", estimateTokens(message) + 100, nil
}

// buildGeneratePrompt 构建生成内容提示词
func (uc *AIUseCase) buildGeneratePrompt(req *models.GenerateContentRequest) string {
	// 根据类型构建不同的提示词
	switch req.Type {
	case "description":
		return fmt.Sprintf("请为商品生成一段吸引人的描述文字，包含商品特点、优势和使用场景。")
	case "review":
		return fmt.Sprintf("请根据商品信息，生成一段真实可信的用户评价。")
	case "email":
		return fmt.Sprintf("请根据用户信息，生成个性化的营销邮件内容。")
	case "banner":
		return fmt.Sprintf("请生成简洁有力的营销标语，适合用于广告Banner。")
	default:
		return fmt.Sprintf("请生成%s类型的内容。", req.Type)
	}
}

// min 返回最小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========== AI模型配置管理 ==========

// AIModelConfigUseCase AI模型配置用例
type AIModelConfigUseCase struct {
	modelRepo data.AIModelConfigRepository
}

// NewAIModelConfigUseCase 创建AI模型配置用例
func NewAIModelConfigUseCase(modelRepo data.AIModelConfigRepository) *AIModelConfigUseCase {
	return &AIModelConfigUseCase{modelRepo: modelRepo}
}

// CreateModelConfig 创建模型配置
func (uc *AIModelConfigUseCase) CreateModelConfig(ctx context.Context, config *models.AIModelConfig) error {
	return uc.modelRepo.Create(ctx, config)
}

// GetModelConfig 获取模型配置
func (uc *AIModelConfigUseCase) GetModelConfig(ctx context.Context, id uint) (*models.AIModelConfig, error) {
	return uc.modelRepo.GetByID(ctx, id)
}

// GetModelByName 获取模型配置ByName
func (uc *AIModelConfigUseCase) GetModelByName(ctx context.Context, name string) (*models.AIModelConfig, error) {
	return uc.modelRepo.GetByName(ctx, name)
}

// UpdateModelConfig 更新模型配置
func (uc *AIModelConfigUseCase) UpdateModelConfig(ctx context.Context, config *models.AIModelConfig) error {
	return uc.modelRepo.Update(ctx, config)
}

// DeleteModelConfig 删除模型配置
func (uc *AIModelConfigUseCase) DeleteModelConfig(ctx context.Context, id uint) error {
	return uc.modelRepo.Delete(ctx, id)
}