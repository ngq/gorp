//go:build wireinject

package main

import (
	"nop-go/services/customer-service/internal/biz"
	"nop-go/services/customer-service/internal/data"
	"nop-go/services/customer-service/internal/service"

	"github.com/google/wire"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	"gorm.io/gorm"
)

func wireCustomerService(db *gorm.DB, jwtSvc securitycontract.JWTService) (*service.CustomerService, error) {
	panic(wire.Build(
		data.NewCustomerRepository,
		data.NewAddressRepository,
		data.NewCustomerRoleRepository,
		data.NewGdprConsentRepository,
		data.NewGdprLogRepository,
		data.NewGdprRequestRepository,
		data.NewCustomerConsentRepository,
		biz.NewCustomerUseCase,
		biz.NewAddressUseCase,
		biz.NewGdprUseCase,
		service.NewCustomerService,
	))
}
