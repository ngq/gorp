// Package models AI服务数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// AIConversation AI对话
type AIConversation struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CustomerID   uint           `gorm:"index" json:"customer_id"`               // 客户ID
	SessionID    string         `gorm:"size:128;index" json:"session_id"`       // 会话ID
	Title        string         `gorm:"size:256" json:"title"`                   // 对话标题
	Status       string         `gorm:"size:32;default:'active'" json:"status"` // 状态: active/archived
	ModelUsed    string         `gorm:"size:64" json:"model_used"`               // 使用的AI模型
	TokensUsed   int            `gorm:"default:0" json:"tokens_used"`            // 消耗的tokens数
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AIConversation) TableName() string {
	return "ai_conversations"
}

// AIMessage AI消息
type AIMessage struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ConversationID uint      `gorm:"not null;index" json:"conversation_id"` // 对话ID
	Role           string    `gorm:"size:32;not null" json:"role"`           // 角色: user/assistant/system
	Content        string    `gorm:"type:text;not null" json:"content"`      // 消息内容
	Tokens         int       `gorm:"default:0" json:"tokens"`                // 消息tokens数
	CreatedAt      time.Time `json:"created_at"`
}

// TableName 指定表名
func (AIMessage) TableName() string {
	return "ai_messages"
}

// AIRecommendation AI推荐记录
type AIRecommendation struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CustomerID  uint           `gorm:"index" json:"customer_id"`               // 客户ID
	ProductID   uint           `gorm:"not null;index" json:"product_id"`       // 商品ID
	Reason      string         `gorm:"type:text" json:"reason"`                // 推荐理由
	Score       float64        `gorm:"default:0" json:"score"`                 // 推荐分数
	Source      string         `gorm:"size:32" json:"source"`                  // 来源: ai/collaborative/content
	IsClicked   bool           `gorm:"default:false" json:"is_clicked"`        // 是否点击
	IsPurchased bool           `gorm:"default:false" json:"is_purchased"`      // 是否购买
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AIRecommendation) TableName() string {
	return "ai_recommendations"
}

// AISearchSuggestion AI搜索建议
type AISearchSuggestion struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Query       string    `gorm:"size:256;not null;uniqueIndex" json:"query"` // 搜索词
	Suggestion  string    `gorm:"size:512;not null" json:"suggestion"`        // 建议内容
	Type        string    `gorm:"size:32" json:"type"`                        // 类型: correction/expansion/related
	ClickCount  int       `gorm:"default:0" json:"click_count"`               // 点击次数
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AISearchSuggestion) TableName() string {
	return "ai_search_suggestions"
}

// AIGeneratedContent AI生成的内容
type AIGeneratedContent struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Type        string         `gorm:"size:32;not null;index" json:"type"`       // 类型: description/review/email/banner
	EntityID    uint           `gorm:"index" json:"entity_id"`                  // 关联实体ID
	EntityType  string         `gorm:"size:32" json:"entity_type"`              // 实体类型: product/category/customer
	Content     string         `gorm:"type:longtext;not null" json:"content"`   // 生成内容
	ModelUsed   string         `gorm:"size:64" json:"model_used"`               // 使用模型
	TokensUsed  int            `gorm:"default:0" json:"tokens_used"`            // 消耗tokens
	IsApproved  bool           `gorm:"default:false" json:"is_approved"`        // 是否审核通过
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (AIGeneratedContent) TableName() string {
	return "ai_generated_contents"
}

// AIModelConfig AI模型配置
type AIModelConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64;not null;uniqueIndex" json:"name"` // 模型名称
	Provider    string    `gorm:"size:32;not null" json:"provider"`         // 提供商: openai/azure/custom
	ModelType   string    `gorm:"size:32" json:"model_type"`                // 类型: chat/embedding/image
	MaxTokens   int       `gorm:"default:4096" json:"max_tokens"`          // 最大tokens
	Temperature float64   `gorm:"default:0.7" json:"temperature"`          // 温度参数
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (AIModelConfig) TableName() string {
	return "ai_model_configs"
}

// ========== DTO ==========

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID  string   `json:"session_id"`                    // 会话ID（可选）
	Message    string   `json:"message" binding:"required"`   // 用户消息
	Model      string   `json:"model"`                         // 模型（可选）
	Context    []string `json:"context"`                       // 上下文信息
	MaxTokens  int      `json:"max_tokens"`                    // 最大tokens
}

// ChatResponse 聊天响应
type ChatResponse struct {
	SessionID     string `json:"session_id"`       // 会话ID
	Response      string `json:"response"`         // AI回复
	TokensUsed    int    `json:"tokens_used"`      // 消耗tokens
	ModelUsed     string `json:"model_used"`       // 使用模型
	ConversationID uint   `json:"conversation_id"`  // 对话ID
}

// RecommendRequest 推荐请求
type RecommendRequest struct {
	CustomerID   uint     `json:"customer_id"`              // 客户ID
	ProductIDs   []uint   `json:"product_ids"`              // 参考商品ID（可选）
	CategoryID   uint     `json:"category_id"`              // 分类ID（可选）
	Query        string   `json:"query"`                    // 搜索词（可选）
	Limit        int      `json:"limit"`                    // 返回数量
	IncludeReason bool    `json:"include_reason"`           // 是否包含推荐理由
}

// RecommendResponse 推荐响应
type RecommendResponse struct {
	Products []RecommendedProduct `json:"products"`
}

// RecommendedProduct 推荐商品
type RecommendedProduct struct {
	ProductID uint    `json:"product_id"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason,omitempty"`
}

// SearchSuggestRequest 搜索建议请求
type SearchSuggestRequest struct {
	Query  string `json:"query" binding:"required"` // 搜索词
	Limit  int    `json:"limit"`                     // 返回数量
}

// SearchSuggestResponse 搜索建议响应
type SearchSuggestResponse struct {
	Suggestions []SearchSuggestion `json:"suggestions"`
}

// SearchSuggestion 搜索建议
type SearchSuggestion struct {
	Query      string `json:"query"`
	Suggestion string `json:"suggestion"`
	Type       string `json:"type"`
}

// GenerateContentRequest 内容生成请求
type GenerateContentRequest struct {
	Type       string                 `json:"type" binding:"required"`      // 类型
	EntityID   uint                   `json:"entity_id"`                    // 实体ID
	EntityType string                 `json:"entity_type"`                  // 实体类型
	Params     map[string]interface{} `json:"params"`                       // 生成参数
	Model      string                 `json:"model"`                        // 模型
}

// GenerateContentResponse 内容生成响应
type GenerateContentResponse struct {
	ContentID   uint   `json:"content_id"`
	Content     string `json:"content"`
	TokensUsed  int    `json:"tokens_used"`
	ModelUsed   string `json:"model_used"`
}