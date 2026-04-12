//go:build wireinject

package main

import (
	"nop-go/services/store-service/internal/biz"
	"nop-go/services/store-service/internal/data"
	"nop-go/services/store-service/internal/service"

	"github.com/google/wire"
	"github.com/ngq/gorp/framework/contract"
	"gorm.io/gorm"
)

// wireStoreService 使用 Wire 生成 store-service 的装配代码。
func wireStoreService(db *gorm.DB, jwtSvc contract.JWTService) (*service.StoreService, error) {
	panic(wire.Build(
		data.NewStoreRepository,
		data.NewVendorRepository,
		data.NewStoreVendorRepository,
		data.NewVendorNoteRepository,
		biz.NewStoreUseCase,
		biz.NewVendorUseCase,
		service.NewStoreService,
	))
}
