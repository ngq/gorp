//go:build wireinject

package main

import (
	"github.com/ngq/gorp"
	"grpc-demo/services/order/internal/service"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func wireOrderServices(db *gorm.DB, userConnFactory gorp.GRPCConnFactory, dlock gorp.DistributedLock) (*service.Services, error) {
	panic(wire.Build(
		service.NewServices,
	))
}
