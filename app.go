package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/facade"
)

const DBRuntimeKey = data.DBRuntimeKey

var (
	ErrServiceNameRequired         = facade.ErrServiceNameRequired
	ErrNoServiceDeclared           = facade.ErrNoServiceDeclared
	ErrHTTPRouteRegistrationFailed = facade.ErrHTTPRouteRegistrationFailed
	ErrHTTPRuntimeUnavailable      = facade.ErrHTTPRuntimeUnavailable
	ErrSetupFailed                 = facade.ErrSetupFailed
	ErrMigrateFailed               = facade.ErrMigrateFailed
	ErrStartupCanceled             = facade.ErrStartupCanceled
	ErrHTTPServiceRunFailed        = facade.ErrHTTPServiceRunFailed
	ErrHTTPRuntimeBuildFailed      = facade.ErrHTTPRuntimeBuildFailed
)

type HTTPRuntime = facade.HTTPRuntime
type HTTPServiceOptions = facade.HTTPServiceOptions
type ServiceProvider = runtime.ServiceProvider
type MigrateFunc = facade.MigrateFunc
type SetupFunc = facade.SetupFunc
type HTTPRouteRegistrar = facade.HTTPRouteRegistrar
type Option = facade.Option

func Run(serviceName string, options ...Option) error {
	return facade.Run(serviceName, options...)
}

func Start(serviceName string, options ...Option) error {
	return facade.Start(serviceName, options...)
}

func RunContext(ctx context.Context, serviceName string, options ...Option) error {
	return facade.RunContext(ctx, serviceName, options...)
}

func BuildHTTPRuntime(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return facade.BuildHTTPRuntime(serviceName, options...)
}

func Build(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return facade.Build(serviceName, options...)
}

func HTTP(opts ...HTTPServiceOptions) Option {
	return facade.HTTP(opts...)
}

func WithoutHTTP() Option {
	return facade.WithoutHTTP()
}

func Module(providers ...ServiceProvider) Option {
	return facade.Module(providers...)
}

func Modules(groups ...[]ServiceProvider) Option {
	return facade.Modules(groups...)
}

func WithModule(providers ...ServiceProvider) Option {
	return facade.WithModule(providers...)
}

func WithProviders(providers ...ServiceProvider) Option {
	return facade.WithProviders(providers...)
}

func WithMigrate(fn MigrateFunc) Option {
	return facade.WithMigrate(fn)
}

func WithSetup(fn SetupFunc) Option {
	return facade.WithSetup(fn)
}

func WithHTTPRoutes(register HTTPRouteRegistrar) Option {
	return facade.WithHTTPRoutes(register)
}
