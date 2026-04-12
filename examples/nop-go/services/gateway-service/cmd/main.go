// Package main API网关服务入口
package main

import (
	"fmt"
	"os"

	"nop-go/shared/bootstrap"
)

type routeConfig struct {
	Path   string `mapstructure:"path"`
	Target string `mapstructure:"target"`
}

func main() {
	if err := bootstrap.BootHTTPService("gateway-service", bootstrap.Options{
		DisableGorm:  true,
		DisableRedis: true,
	}, nil, setup); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func setup(rt *bootstrap.HTTPServiceRuntime) error {
	routes, err := gatewayRoutesFromRuntime(rt)
	if err != nil {
		return err
	}

	gatewayService, err := wireGatewayService(routes)
	if err != nil {
		return err
	}

	gatewayService.RegisterRoutes(rt.Engine)
	return nil
}
