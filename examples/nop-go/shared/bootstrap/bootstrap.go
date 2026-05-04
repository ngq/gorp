package bootstrap

import (
	"github.com/gin-gonic/gin"
	frameworkbootstrap "github.com/ngq/gorp/framework/bootstrap"
	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"

	gormpkg "gorm.io/gorm"
)

type HTTPServiceOptions = frameworkbootstrap.HTTPServiceOptions
type HTTPServiceRuntime = frameworkbootstrap.HTTPServiceRuntime
type Options = frameworkbootstrap.HTTPServiceOptions

var (
	NewHTTPServiceRuntime   = frameworkbootstrap.NewHTTPServiceRuntime
	BootHTTPService         = frameworkbootstrap.BootHTTPService
	RegisterHealthCheck     = frameworkbootstrap.RegisterHealthCheck
	RegisterMetricsEndpoint = frameworkbootstrap.RegisterMetricsEndpoint
	RunHTTP                 = frameworkbootstrap.RunHTTP
)

func MustMakeJWTService(c runtimecontract.Container) securitycontract.JWTService {
	return container.MustMakeJWTService(c)
}

func MustMakeValidator(c runtimecontract.Container) datacontract.Validator {
	return container.MustMakeValidator(c)
}

func MustMakeRetry(c runtimecontract.Container) resiliencecontract.Retry {
	return container.MustMakeRetry(c)
}

func MustMakeLogger(c runtimecontract.Container) observabilitycontract.Logger {
	return container.MustMakeLogger(c)
}

func MustMakeGorm(c runtimecontract.Container) *gormpkg.DB {
	return container.MustMakeGorm(c)
}

func MustMakeEngine(c runtimecontract.Container) *gin.Engine {
	return container.MustMakeEngine(c)
}

func MustMakeConfig(c runtimecontract.Container) datacontract.Config {
	return container.MustMakeConfig(c)
}
