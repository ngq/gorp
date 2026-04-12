//go:build wireinject

package main

import (
	"nop-go/services/gateway-service/internal/service"

	"github.com/google/wire"
)

func wireGatewayService(routes []service.Route) (*service.GatewayService, error) {
	panic(wire.Build(
		service.NewGatewayService,
	))
}
