// Application scenarios:
// - Assemble and run the default HTTP service mainline used by the framework.
// - Build one reusable runtime object that carries app, container, router, config, DB, Redis, and JWT capabilities.
// - Centralize health checks, metrics, pprof, graceful shutdown, and host/direct run behavior.
//
// 适用场景：
// - 装配并运行框架默认 HTTP 服务主线。
// - 构建一个复用型 runtime 对象，统一承载 app、container、router、config、DB、Redis 和 JWT 能力。
// - 集中管理健康检查、指标、pprof、优雅停机以及 host/direct 两种运行模式。
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
	frameworklog "github.com/ngq/gorp/framework/log"
	gingin "github.com/ngq/gorp/framework/provider/gin"
	"github.com/ngq/gorp/framework/provider/host"
	redisProvider "github.com/ngq/gorp/framework/provider/redis"

	gormpkg "gorm.io/gorm"
)

var (
	buildHTTPProvidersFunc                        = buildHTTPProviders
	registerSelectedMicroserviceProvidersWithMode = RegisterSelectedMicroserviceProvidersWithMode
	registerSelectedMicroserviceProvidersWithOptionsFunc = registerSelectedMicroserviceProvidersWithOptions
)

// HTTPServiceOptions describes the bootstrap options for the default HTTP mainline.
//
// HTTPServiceOptions 描述默认 HTTP 主线的 bootstrap 选项。
type HTTPServiceOptions struct {
	ExtraProviders []runtimecontract.ServiceProvider
	DisableRedis   bool
	DisableGorm    bool
	DisableMetrics bool
	EnablePprof    bool
	GovernanceMode string
	GovernanceDisable  []string
	GovernanceProviders map[string]string
}

// HTTPServiceRuntime carries the assembled HTTP runtime state used during startup callbacks.
//
// HTTPServiceRuntime 承载启动回调阶段使用的 HTTP runtime 状态。
type HTTPServiceRuntime struct {
	App         *framework.Application
	Container   runtimecontract.Container
	Logger      observabilitycontract.Logger
	Router      transportcontract.HTTPRouter
	DB          *gormpkg.DB
	Redis       datacontract.Redis
	JWT         securitycontract.JWTService
	Config      datacontract.Config
	ServiceName string
	GovernanceMode    resiliencecontract.GovernanceMode
	GovernanceSummary GovernanceSummary
}

// NewHTTPServiceRuntime builds the default HTTP runtime without starting the server.
//
// NewHTTPServiceRuntime 构建默认 HTTP runtime，但不启动服务。
func NewHTTPServiceRuntime(serviceName string, opts HTTPServiceOptions) (*HTTPServiceRuntime, error) {
	app := framework.NewApplication()
	c := app.Container()

	providers := buildHTTPProvidersFunc(opts)
	if err := c.RegisterProviders(providers...); err != nil {
		return nil, fmt.Errorf("register providers: %w", err)
	}
	if err := registerSelectedMicroserviceProvidersWithOptionsFunc(c, opts.GovernanceMode, opts.GovernanceDisable, opts.GovernanceProviders); err != nil {
		return nil, fmt.Errorf("register selected microservice providers: %w", err)
	}

	rt := &HTTPServiceRuntime{
		App:         app,
		Container:   c,
		Logger:      container.MustMakeLogger(c),
		Router:      container.MustMakeHTTPRouter(c),
		Config:      container.MustMakeConfig(c),
		JWT:         container.MustMakeJWTService(c),
		ServiceName: serviceName,
	}
	frameworklog.SetDefault(rt.Logger)
	effectiveConfig := overlayGovernanceConfig(rt.Config, opts.GovernanceDisable, opts.GovernanceProviders)
	governanceMode := DetectGovernanceMode(effectiveConfig)
	if opts.GovernanceMode != "" {
		governanceMode = NormalizeGovernanceMode(resiliencecontract.GovernanceMode(opts.GovernanceMode))
	}
	governanceSummary := BuildGovernanceSummaryWithModeOverride(effectiveConfig, governanceMode, opts.GovernanceMode)
	rt.GovernanceMode = governanceMode
	rt.GovernanceSummary = governanceSummary
	rt.Logger.Info(FormatGovernanceSummary(governanceSummary))

	if !opts.DisableGorm {
		rt.DB = container.MustMakeGorm(c)
	}
	if !opts.DisableRedis {
		// Redis is optional in this mainline, so keep startup tolerant when the capability is absent.
		// Redis 在这条主线里是可选能力，因此这里保持“缺失不阻断启动”的语义。
		if redisSvc, err := container.MakeRedis(c); err == nil {
			rt.Redis = redisSvc
		}
	}

	return rt, nil
}

// buildHTTPProviders assembles the provider list used by the default HTTP mainline.
//
// buildHTTPProviders 组装默认 HTTP 主线使用的 provider 列表。
func buildHTTPProviders(opts HTTPServiceOptions) []runtimecontract.ServiceProvider {
	providers := make([]runtimecontract.ServiceProvider, 0)
	providers = append(providers, FoundationProviders()...)
	if !opts.DisableGorm {
		providers = append(providers, ORMRuntimeProviders()...)
	}
	providers = append(providers, DefaultCapabilityProviders()...)
	if !opts.DisableRedis {
		providers = append(providers, redisProvider.NewProvider())
	}
	providers = append(providers, opts.ExtraProviders...)
	return providers
}

// BootHTTPService assembles, configures, and runs the default HTTP service.
//
// BootHTTPService 装配、配置并运行默认 HTTP 服务。
func BootHTTPService(serviceName string, opts HTTPServiceOptions, migrate func(*HTTPServiceRuntime) error, setup func(*HTTPServiceRuntime) error) error {
	rt, err := NewHTTPServiceRuntime(serviceName, opts)
	if err != nil {
		return fmt.Errorf("initialize http runtime: %w", err)
	}

	rt.Logger.Info(fmt.Sprintf("%s starting", serviceName))

	if migrate != nil {
		if err := migrate(rt); err != nil {
			return fmt.Errorf("migrate models: %w", err)
		}
		rt.Logger.Info("database migration finished")
	}

	if setup != nil {
		if err := setup(rt); err != nil {
			return fmt.Errorf("setup service: %w", err)
		}
	}

	RegisterHealthCheck(rt.Router, serviceName)
	RegisterGovernanceInspectEndpoints(rt.Router, rt.GovernanceSummary)
	if !opts.DisableMetrics {
		RegisterMetricsEndpoint(rt.Router)
		rt.Router.Use(httpmiddleware.MetricsMiddleware())
	}
	if opts.EnablePprof {
		RegisterPprofEndpoints(rt.Router)
	}

	return RunHTTP(rt.Container, rt.Logger)
}

// AutoMigrateModels runs Gorm auto-migration when DB runtime is available.
//
// AutoMigrateModels 在 DB runtime 可用时执行 Gorm 自动迁移。
func AutoMigrateModels(rt *HTTPServiceRuntime, models ...any) error {
	if rt == nil || rt.DB == nil || len(models) == 0 {
		return nil
	}
	return rt.DB.AutoMigrate(models...)
}

// RegisterHealthCheck registers the default health endpoint.
//
// RegisterHealthCheck 注册默认健康检查端点。
func RegisterHealthCheck(router transportcontract.HTTPRouter, serviceName string) {
	if router == nil {
		return
	}
	router.GET("/healthz", func(c transportcontract.HTTPContext) {
		c.JSON(http.StatusOK, map[string]any{
			"status":  "healthy",
			"service": serviceName,
			"version": "1.0.0",
		})
	})
}

// RegisterMetricsEndpoint registers the default metrics endpoint.
//
// RegisterMetricsEndpoint 注册默认 metrics 端点。
func RegisterMetricsEndpoint(router transportcontract.HTTPRouter) {
	if router == nil {
		return
	}
	gingin.RegisterGoRuntimeMetrics()
	router.Mount("/metrics", gingin.PrometheusHandler())
}

// RegisterPprofEndpoints registers the standard pprof endpoints.
//
// RegisterPprofEndpoints 注册标准 pprof 端点。
func RegisterPprofEndpoints(router transportcontract.HTTPRouter) {
	if router == nil {
		return
	}
	router.Mount("/debug/pprof/", http.HandlerFunc(pprof.Index))
	router.Mount("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	router.Mount("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	router.Mount("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	router.Mount("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
}

// RunHTTP runs the HTTP service through the host capability when present.
//
// RunHTTP 优先通过 host 能力运行 HTTP 服务。
func RunHTTP(c runtimecontract.Container, logger observabilitycontract.Logger) error {
	hostSvc, err := container.MakeHost(c)
	if err != nil {
		// Fall back to direct HTTP run mode when host capability is unavailable.
		// 当 host 能力不可用时，回退到 HTTP 直跑模式。
		return runHTTPDirectly(c, logger)
	}

	httpSvc, err := container.MakeHTTP(c)
	if err != nil {
		return fmt.Errorf("make http service: %w", err)
	}

	httpHostable := host.NewHTTPService("http", httpSvc)
	if err := hostSvc.RegisterService("http", httpHostable); err != nil {
		return fmt.Errorf("register http service to host: %w", err)
	}

	logger.Info("starting http server")
	if err := hostSvc.Start(context.Background()); err != nil {
		return err
	}

	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()
	<-ctx.Done()

	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := hostSvc.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("http server stopped gracefully")
	return nil
}

// runHTTPDirectly runs the HTTP service without the host abstraction.
//
// runHTTPDirectly 在不使用 host 抽象的情况下直接运行 HTTP 服务。
func runHTTPDirectly(c runtimecontract.Container, logger observabilitycontract.Logger) error {
	httpSvc, err := container.MakeHTTP(c)
	if err != nil {
		return fmt.Errorf("make http service: %w", err)
	}

	logger.Info("starting http server (direct mode)")

	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		if err := httpSvc.Run(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				errCh <- nil
				return
			}
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSvc.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	logger.Info("http server stopped gracefully")
	return nil
}
