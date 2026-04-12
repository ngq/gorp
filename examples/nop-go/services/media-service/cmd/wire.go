//go:build wireinject

package main

import (
	"nop-go/services/media-service/internal/biz"
	"nop-go/services/media-service/internal/data"
	"nop-go/services/media-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

// wireMediaService 使用 Wire 生成 media-service 的装配代码。
func wireMediaService(db *gorm.DB, jwtSvc contract.JWTService, localPath, urlPrefix, storageType string) (*service.MediaService, error) {
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
