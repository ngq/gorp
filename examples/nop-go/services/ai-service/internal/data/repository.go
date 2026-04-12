// Package data AI服务数据访问层
package data

import (
	"context"
	"errors"

	"nop-go/services/ai-service/internal/models"

	"gorm.io/gorm"
)

// AIConversationRepository AI对话仓储接口
type AIConversationRepository interface {
	Create(ctx context.Context, conv *models.AIConversation) error
	GetByID(ctx context.Context, id uint) (*models.AIConversation, error)
	GetBySessionID(ctx context.Context, sessionID string) (*models.AIConversation, error)
	GetByCustomerID(ctx context.Context, customerID uint, page, pageSize int) ([]*models.AIConversation, int64, error)
	Update(ctx context.Context, conv *models.AIConversation) error
	Delete(ctx context.Context, id uint) error
}

type aiConversationRepo struct{ db *gorm.DB }

func NewAIConversationRepository(db *gorm.DB) AIConversationRepository {
	return &aiConversationRepo{db: db}
}

func (r *aiConversationRepo) Create(ctx context.Context, c *models.AIConversation) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *aiConversationRepo) GetByID(ctx context.Context, id uint) (*models.AIConversation, error) {
	var c models.AIConversation
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *aiConversationRepo) GetBySessionID(ctx context.Context, sessionID string) (*models.AIConversation, error) {
	var c models.AIConversation
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *aiConversationRepo) GetByCustomerID(ctx context.Context, customerID uint, page, pageSize int) ([]*models.AIConversation, int64, error) {
	var list []*models.AIConversation
	var total int64
	db := r.db.WithContext(ctx).Model(&models.AIConversation{}).Where("customer_id = ?", customerID)
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *aiConversationRepo) Update(ctx context.Context, c *models.AIConversation) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *aiConversationRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AIConversation{}, id).Error
}

// AIMessageRepository AI消息仓储接口
type AIMessageRepository interface {
	Create(ctx context.Context, msg *models.AIMessage) error
	CreateBatch(ctx context.Context, msgs []*models.AIMessage) error
	GetByID(ctx context.Context, id uint) (*models.AIMessage, error)
	GetByConversationID(ctx context.Context, conversationID uint) ([]*models.AIMessage, error)
	DeleteByConversationID(ctx context.Context, conversationID uint) error
}

type aiMessageRepo struct{ db *gorm.DB }

func NewAIMessageRepository(db *gorm.DB) AIMessageRepository {
	return &aiMessageRepo{db: db}
}

func (r *aiMessageRepo) Create(ctx context.Context, m *models.AIMessage) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *aiMessageRepo) CreateBatch(ctx context.Context, msgs []*models.AIMessage) error {
	return r.db.WithContext(ctx).Create(msgs).Error
}

func (r *aiMessageRepo) GetByID(ctx context.Context, id uint) (*models.AIMessage, error) {
	var m models.AIMessage
	err := r.db.WithContext(ctx).First(&m, id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *aiMessageRepo) GetByConversationID(ctx context.Context, conversationID uint) ([]*models.AIMessage, error) {
	var list []*models.AIMessage
	err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("created_at asc").Find(&list).Error
	return list, err
}

func (r *aiMessageRepo) DeleteByConversationID(ctx context.Context, conversationID uint) error {
	return r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Delete(&models.AIMessage{}).Error
}

// AIRecommendationRepository AI推荐仓储接口
type AIRecommendationRepository interface {
	Create(ctx context.Context, rec *models.AIRecommendation) error
	CreateBatch(ctx context.Context, recs []*models.AIRecommendation) error
	GetByID(ctx context.Context, id uint) (*models.AIRecommendation, error)
	GetByCustomerID(ctx context.Context, customerID uint, limit int) ([]*models.AIRecommendation, error)
	Update(ctx context.Context, rec *models.AIRecommendation) error
	Delete(ctx context.Context, id uint) error
}

type aiRecommendationRepo struct{ db *gorm.DB }

func NewAIRecommendationRepository(db *gorm.DB) AIRecommendationRepository {
	return &aiRecommendationRepo{db: db}
}

func (r *aiRecommendationRepo) Create(ctx context.Context, rec *models.AIRecommendation) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *aiRecommendationRepo) CreateBatch(ctx context.Context, recs []*models.AIRecommendation) error {
	return r.db.WithContext(ctx).Create(recs).Error
}

func (r *aiRecommendationRepo) GetByID(ctx context.Context, id uint) (*models.AIRecommendation, error) {
	var rec models.AIRecommendation
	err := r.db.WithContext(ctx).First(&rec, id).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *aiRecommendationRepo) GetByCustomerID(ctx context.Context, customerID uint, limit int) ([]*models.AIRecommendation, error) {
	var list []*models.AIRecommendation
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("score desc").Limit(limit).Find(&list).Error
	return list, err
}

func (r *aiRecommendationRepo) Update(ctx context.Context, rec *models.AIRecommendation) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *aiRecommendationRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AIRecommendation{}, id).Error
}

// AISearchSuggestionRepository AI搜索建议仓储接口
type AISearchSuggestionRepository interface {
	Create(ctx context.Context, sug *models.AISearchSuggestion) error
	GetByID(ctx context.Context, id uint) (*models.AISearchSuggestion, error)
	GetByQuery(ctx context.Context, query string) ([]*models.AISearchSuggestion, error)
	List(ctx context.Context, page, pageSize int) ([]*models.AISearchSuggestion, int64, error)
	Update(ctx context.Context, sug *models.AISearchSuggestion) error
	Delete(ctx context.Context, id uint) error
}

type aiSearchSuggestionRepo struct{ db *gorm.DB }

func NewAISearchSuggestionRepository(db *gorm.DB) AISearchSuggestionRepository {
	return &aiSearchSuggestionRepo{db: db}
}

func (r *aiSearchSuggestionRepo) Create(ctx context.Context, s *models.AISearchSuggestion) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *aiSearchSuggestionRepo) GetByID(ctx context.Context, id uint) (*models.AISearchSuggestion, error) {
	var s models.AISearchSuggestion
	err := r.db.WithContext(ctx).First(&s, id).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *aiSearchSuggestionRepo) GetByQuery(ctx context.Context, query string) ([]*models.AISearchSuggestion, error) {
	var list []*models.AISearchSuggestion
	err := r.db.WithContext(ctx).Where("query LIKE ?", query+"%").Where("is_active = ?", true).Order("click_count desc").Find(&list).Error
	return list, err
}

func (r *aiSearchSuggestionRepo) List(ctx context.Context, page, pageSize int) ([]*models.AISearchSuggestion, int64, error) {
	var list []*models.AISearchSuggestion
	var total int64
	db := r.db.WithContext(ctx).Model(&models.AISearchSuggestion{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("click_count desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *aiSearchSuggestionRepo) Update(ctx context.Context, s *models.AISearchSuggestion) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *aiSearchSuggestionRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AISearchSuggestion{}, id).Error
}

// AIGeneratedContentRepository AI生成内容仓储接口
type AIGeneratedContentRepository interface {
	Create(ctx context.Context, content *models.AIGeneratedContent) error
	GetByID(ctx context.Context, id uint) (*models.AIGeneratedContent, error)
	GetByEntity(ctx context.Context, entityID uint, entityType string) ([]*models.AIGeneratedContent, error)
	List(ctx context.Context, page, pageSize int) ([]*models.AIGeneratedContent, int64, error)
	Update(ctx context.Context, content *models.AIGeneratedContent) error
	Delete(ctx context.Context, id uint) error
}

type aiGeneratedContentRepo struct{ db *gorm.DB }

func NewAIGeneratedContentRepository(db *gorm.DB) AIGeneratedContentRepository {
	return &aiGeneratedContentRepo{db: db}
}

func (r *aiGeneratedContentRepo) Create(ctx context.Context, c *models.AIGeneratedContent) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *aiGeneratedContentRepo) GetByID(ctx context.Context, id uint) (*models.AIGeneratedContent, error) {
	var c models.AIGeneratedContent
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *aiGeneratedContentRepo) GetByEntity(ctx context.Context, entityID uint, entityType string) ([]*models.AIGeneratedContent, error) {
	var list []*models.AIGeneratedContent
	err := r.db.WithContext(ctx).Where("entity_id = ? AND entity_type = ?", entityID, entityType).Order("created_at desc").Find(&list).Error
	return list, err
}

func (r *aiGeneratedContentRepo) List(ctx context.Context, page, pageSize int) ([]*models.AIGeneratedContent, int64, error) {
	var list []*models.AIGeneratedContent
	var total int64
	db := r.db.WithContext(ctx).Model(&models.AIGeneratedContent{})
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (r *aiGeneratedContentRepo) Update(ctx context.Context, c *models.AIGeneratedContent) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *aiGeneratedContentRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AIGeneratedContent{}, id).Error
}

// AIModelConfigRepository AI模型配置仓储接口
type AIModelConfigRepository interface {
	Create(ctx context.Context, config *models.AIModelConfig) error
	GetByID(ctx context.Context, id uint) (*models.AIModelConfig, error)
	GetByName(ctx context.Context, name string) (*models.AIModelConfig, error)
	List(ctx context.Context) ([]*models.AIModelConfig, error)
	Update(ctx context.Context, config *models.AIModelConfig) error
	Delete(ctx context.Context, id uint) error
}

type aiModelConfigRepo struct{ db *gorm.DB }

func NewAIModelConfigRepository(db *gorm.DB) AIModelConfigRepository {
	return &aiModelConfigRepo{db: db}
}

func (r *aiModelConfigRepo) Create(ctx context.Context, c *models.AIModelConfig) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *aiModelConfigRepo) GetByID(ctx context.Context, id uint) (*models.AIModelConfig, error) {
	var c models.AIModelConfig
	err := r.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *aiModelConfigRepo) GetByName(ctx context.Context, name string) (*models.AIModelConfig, error) {
	var c models.AIModelConfig
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *aiModelConfigRepo) List(ctx context.Context) ([]*models.AIModelConfig, error) {
	var list []*models.AIModelConfig
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("name").Find(&list).Error
	return list, err
}

func (r *aiModelConfigRepo) Update(ctx context.Context, c *models.AIModelConfig) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *aiModelConfigRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.AIModelConfig{}, id).Error
}

// 错误定义
var (
	ErrConversationNotFound  = errors.New("conversation not found")
	ErrMessageNotFound       = errors.New("message not found")
	ErrRecommendationNotFound = errors.New("recommendation not found")
	ErrSuggestionNotFound    = errors.New("suggestion not found")
	ErrContentNotFound       = errors.New("generated content not found")
	ErrModelNotFound         = errors.New("model config not found")
	ErrAIServiceUnavailable  = errors.New("AI service unavailable")
)