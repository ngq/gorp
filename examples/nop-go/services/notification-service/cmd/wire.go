//go:build wireinject

package main

import (
	"nop-go/services/notification-service/internal/biz"
	"nop-go/services/notification-service/internal/data"
	"nop-go/services/notification-service/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireNotificationService(db *gorm.DB) (*service.NotificationService, error) {
	panic(wire.Build(
		data.NewNotificationRepository,
		data.NewNotificationTemplateRepository,
		data.NewSMSRecordRepository,
		biz.NewNotificationUseCase,
		biz.NewTemplateUseCase,
		service.NewNotificationService,
	))
}
