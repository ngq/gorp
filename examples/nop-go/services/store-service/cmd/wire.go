//go:build wireinject

package main

import (
	"nop-go/services/store-service/internal/biz"
	"nop-go/services/store-service/internal/data"
	"nop-go/services/store-service/internal/service"

	"github.com/google/wire"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"gorm.io/gorm"
)

// wireStoreService 浣跨敤 Wire 鐢熸垚 store-service 鐨勮閰嶄唬鐮併€?
func wireStoreService(db *gorm.DB, jwtSvc securitycontract.JWTService) (*service.StoreService, error) {
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
