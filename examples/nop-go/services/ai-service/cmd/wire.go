//go:build wireinject

package main

import (
	"nop-go/services/ai-service/internal/biz"
	"nop-go/services/ai-service/internal/data"
	"nop-go/services/ai-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

func wireAIService(db *gorm.DB, jwtSvc contract.JWTService, config biz.AIConfig) (*service.AIService, error) {
	panic(wire.Build(
		data.NewAIConversationRepository,
		data.NewAIMessageRepository,
		data.NewAIRecommendationRepository,
		data.NewAISearchSuggestionRepository,
		data.NewAIGeneratedContentRepository,
		data.NewAIModelConfigRepository,
		biz.NewAIUseCase,
		biz.NewAIModelConfigUseCase,
		service.NewAIService,
	))
}
