//go:build wireinject

package main

import (
	"nop-go/services/cms-service/internal/biz"
	"nop-go/services/cms-service/internal/data"
	"nop-go/services/cms-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// wireCMSService 使用 Wire 生成 cms-service 的装配代码。
func wireCMSService(db *gorm.DB) (*service.CMSService, error) {
	panic(wire.Build(
		data.NewBlogPostRepository,
		data.NewBlogCategoryRepository,
		data.NewNewsRepository,
		data.NewTopicRepository,
		data.NewForumRepository,
		data.NewMenuRepository,
		data.NewMenuItemRepository,
		data.NewPollRepository,
		data.NewPollAnswerRepository,
		data.NewPollVotingRecordRepository,
		data.NewHtmlBodyRepository,
		biz.NewBlogUseCase,
		biz.NewNewsUseCase,
		biz.NewTopicUseCase,
		biz.NewForumUseCase,
		biz.NewMenuUseCase,
		biz.NewPollUseCase,
		biz.NewHtmlBodyUseCase,
		service.NewCMSService,
	))
}
