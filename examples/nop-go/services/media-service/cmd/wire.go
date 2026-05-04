//go:build wireinject

package main

import (
	"nop-go/services/media-service/internal/biz"
	"nop-go/services/media-service/internal/data"
	"nop-go/services/media-service/internal/service"

	"github.com/google/wire"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"gorm.io/gorm"
)

// wireMediaService 浣跨敤 Wire 鐢熸垚 media-service 鐨勮閰嶄唬鐮併€?
func wireMediaService(db *gorm.DB, jwtSvc securitycontract.JWTService, localPath, urlPrefix, storageType string) (*service.MediaService, error) {
	panic(wire.Build(
		data.NewPictureRepository,
		data.NewProductPictureRepository,
		data.NewCategoryPictureRepository,
		data.NewDocumentRepository,
		provideMediaStorageConfig,
		biz.NewMediaUseCase,
		service.NewMediaService,
	))
}

func provideMediaStorageConfig(localPath, urlPrefix, storageType string) biz.StorageConfig {
	return biz.StorageConfig{
		Type:      storageType,
		LocalPath: localPath,
		URLPrefix: urlPrefix,
	}
}
