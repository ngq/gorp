//go:build wireinject

package main

import (
	"nop-go/services/import-service/internal/biz"
	"nop-go/services/import-service/internal/data"
	"nop-go/services/import-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

func wireImportService(db *gorm.DB, jwtSvc contract.JWTService, config biz.ImportConfig) (*service.ImportService, error) {
	panic(wire.Build(
		data.NewImportProfileRepository,
		data.NewExportProfileRepository,
		data.NewImportHistoryRepository,
		data.NewExportHistoryRepository,
		data.NewImportErrorRepository,
		biz.NewImportUseCase,
		biz.NewExportUseCase,
		service.NewImportService,
	))
}
