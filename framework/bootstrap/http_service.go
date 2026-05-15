// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file assembles and runs the default HTTP service mainline.
// Builds reusable runtime carrying app, container, router, config, DB, Redis, JWT.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件装配并运行框架默认 HTTP 服务主线。
// 构建复用型 runtime 对象，统一承载 app、container、router、config、DB、Redis、JWT 能力。
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
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
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
	"google.golang.org/grpc"

	gormpkg "gorm.io/gorm"
	"github.com/jmoiron/sqlx"
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
	GovernanceEnable   []string
	GovernanceProviders map[string]string
}

// HTTPServiceRuntime carries the assembled HTTP runtime state used during startup callbacks.
// In microservice mode, also carries the gRPC server instance for service registration.
//
// HTTPServiceRuntime 承载启动回调阶段使用的 HTTP runtime 状态。
// 在微服务模式下，同时承载 gRPC 服务器实例供服务注册使用。
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

	// GRPCServer 是底层 gRPC 服务器实例，仅在微服务模式下可用。
	// 用户可在 setup 回调中使用此字段注册 proto 服务。
	//
	// GRPCServer is the underlying gRPC server instance, only available in microservice mode.
	// Users can use this field to register proto services in the setup callback.
	GRPCServer *grpc.Server

	// GRPCServerRegistrar 是 gRPC 服务注册器，仅在微服务模式下可用。
	// 提供 RegisterProto 方法用于注册 proto 生成的服务实现。
	//
	// GRPCServerRegistrar is the gRPC server registrar, only available in microservice mode.
	// Provides RegisterProto method for registering protobuf-generated service implementations.
	GRPCServerRegistrar transportcontract.GRPCServerRegistrar
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
	if err := registerSelectedMicroserviceProvidersWithOptionsFunc(c, opts.GovernanceMode, opts.GovernanceDisable, opts.GovernanceEnable, opts.GovernanceProviders); err != nil {
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

	// 启动阶段 fail-fast 校验关键配置：缺失或无效的必填字段立即报错，
	// 避免 viper 零值静默传播导致后续运行时错误难以排查。
	//
	// Fail-fast validation of critical config at startup: missing or invalid
	// required fields error immediately, preventing silent viper zero values
	// from causing hard-to-diagnose runtime errors later.
	if err := ValidateCriticalConfig(rt.Config); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}
	effectiveConfig := overlayGovernanceConfig(rt.Config, opts.GovernanceDisable, opts.GovernanceEnable, opts.GovernanceProviders)
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
		// Redis 在这条主线里是可选能力，因此这里保持”缺失不阻断启动”的语义。
		if redisSvc, err := container.MakeRedis(c); err == nil {
			rt.Redis = redisSvc
		}
	}

	// 在微服务模式下，尝试从容器解析 gRPC 服务注册器
	// 让用户可以在 setup 回调中通过 rt.GRPCServer 注册 proto 服务
	// In microservice mode, try to resolve gRPC server registrar from container
	// Let users register proto services via rt.GRPCServer in the setup callback
	if IsMicroserviceMode(governanceMode) && c.IsBind(transportcontract.GRPCServerRegistrarKey) {
		if registrar, err := container.MakeGRPCServerRegistrar(c); err == nil {
			rt.GRPCServerRegistrar = registrar
			rt.GRPCServer = registrar.Server()
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
	if cronSvc, err := container.MakeCron(rt.Container); err == nil {
		RegisterCronInspectEndpoints(rt.Router, cronSvc)
	}
	if !opts.DisableMetrics {
		RegisterMetricsEndpoint(rt.Router)
		rt.Router.Use(httpmiddleware.MetricsMiddleware())
	}
	if opts.EnablePprof {
		RegisterPprofEndpoints(rt.Router)
	}

	return RunHTTP(rt.Container, rt.Logger)
}

// GetGorm returns the Gorm database handle.
// Panics if Gorm is not available (DisableGorm=true).
//
// GetGorm 返回 Gorm 数据库句柄。
// 如果 Gorm 不可用（DisableGorm=true）则 panic。
func (rt *HTTPServiceRuntime) GetGorm() *gormpkg.DB {
	if rt.DB == nil {
		panic("Gorm database not available: DisableGorm is true")
	}
	return rt.DB
}

// GetSQLX returns the SQLX database handle.
// Panics if SQLX is not available.
//
// GetSQLX 返回 SQLX 数据库句柄。
// 如果 SQLX 不可用则 panic。
func (rt *HTTPServiceRuntime) GetSQLX() *sqlx.DB {
	return container.GetSQLXOrPanic(rt.Container)
}

// GetRedis returns the Redis capability.
// Panics if Redis is not available (DisableRedis=true).
//
// GetRedis 返回 Redis 能力。
// 如果 Redis 不可用（DisableRedis=true）则 panic。
func (rt *HTTPServiceRuntime) GetRedis() datacontract.Redis {
	if rt.Redis == nil {
		panic("Redis not available: DisableRedis is true")
	}
	return rt.Redis
}

// GetDB returns both Gorm and SQLX database handles.
// Panics if either is not available.
//
// GetDB 返回 Gorm 和 SQLX 数据库句柄。
// 如果任一不可用则 panic。
func (rt *HTTPServiceRuntime) GetDB() (*gormpkg.DB, *sqlx.DB) {
	return rt.GetGorm(), rt.GetSQLX()
}

// GetCache returns the cache capability.
// Panics if cache is not available.
//
// GetCache 返回缓存能力。
// 如果缓存不可用则 panic。
func (rt *HTTPServiceRuntime) GetCache() datacontract.Cache {
	return container.GetCacheOrPanic(rt.Container)
}

// GetCron returns the cron capability.
// Panics if cron is not available.
//
// GetCron 返回定时任务能力。
// 如果定时任务不可用则 panic。
func (rt *HTTPServiceRuntime) GetCron() runtimecontract.Cron {
	return container.GetCronOrPanic(rt.Container)
}

// GetValidator returns the validator capability.
// Panics if validator is not available.
//
// GetValidator 返回参数校验能力。
// 如果参数校验不可用则 panic。
func (rt *HTTPServiceRuntime) GetValidator() datacontract.Validator {
	return container.GetValidatorOrPanic(rt.Container)
}

// GetDistributedLock returns the distributed lock capability.
// Panics if distributed lock is not available.
//
// GetDistributedLock 返回分布式锁能力。
// 如果分布式锁不可用则 panic。
func (rt *HTTPServiceRuntime) GetDistributedLock() datacontract.DistributedLock {
	return container.GetDistributedLockOrPanic(rt.Container)
}

// GetRetry returns the retry capability.
// Panics if retry is not available.
//
// GetRetry 返回重试策略能力。
// 如果重试策略不可用则 panic。
func (rt *HTTPServiceRuntime) GetRetry() resiliencecontract.Retry {
	return container.GetRetryOrPanic(rt.Container)
}

// GetMessagePublisher returns the message publisher capability.
// Panics if message publisher is not available.
//
// GetMessagePublisher 返回消息发布能力。
// 如果消息发布不可用则 panic。
func (rt *HTTPServiceRuntime) GetMessagePublisher() integrationcontract.MessagePublisher {
	return container.GetMessagePublisherOrPanic(rt.Container)
}

// GetMessageSubscriber returns the message subscriber capability.
// Panics if message subscriber is not available.
//
// GetMessageSubscriber 返回消息订阅能力。
// 如果消息订阅不可用则 panic。
func (rt *HTTPServiceRuntime) GetMessageSubscriber() integrationcontract.MessageSubscriber {
	return container.GetMessageSubscriberOrPanic(rt.Container)
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
// In microservice mode, also starts the gRPC server alongside HTTP when the container
// has a GRPCServerRegistrar bound.
//
// RunHTTP 优先通过 host 能力运行 HTTP 服务。
// 在微服务模式下，当容器绑定了 GRPCServerRegistrar 时，同时启动 gRPC 服务器。
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

	// 在微服务模式下，尝试将 gRPC 服务器也注册到 host 统一管理
	// In microservice mode, try to register gRPC server to host for unified management
	grpcStarted := registerGRPCToHost(c, hostSvc, logger)

	logger.Info("starting http server")
	if grpcStarted {
		logger.Info("starting grpc server")
	}
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

// registerGRPCToHost tries to register the gRPC server to the host when the container
// has a GRPCServerRegistrar bound. Returns true if gRPC was successfully registered.
//
// registerGRPCToHost 尝试在容器绑定了 GRPCServerRegistrar 时将 gRPC 服务器注册到 host。
// 如果 gRPC 成功注册则返回 true。
func registerGRPCToHost(c runtimecontract.Container, hostSvc runtimecontract.Host, logger observabilitycontract.Logger) bool {
	// 检查容器是否有 gRPC 能力
	// Check if container has gRPC capability
	if !c.IsBind(transportcontract.GRPCServerRegistrarKey) {
		return false
	}

	// 解析 RPC Server
	// Resolve RPC Server
	rpcServerAny, err := c.Make(transportcontract.RPCServerKey)
	if err != nil {
		logger.Info("grpc server not available, skipping grpc host registration")
		return false
	}

	rpcServer, ok := rpcServerAny.(transportcontract.RPCServer)
	if !ok {
		return false
	}

	// 创建 Hostable 适配器并注册到 host
	// Create Hostable adapter and register to host
	grpcHostable, err := newGRPCHostableFromRPCServer(rpcServer)
	if err != nil {
		logger.Info("grpc server adapter creation failed, skipping grpc host registration")
		return false
	}

	if err := hostSvc.RegisterService("grpc", grpcHostable); err != nil {
		logger.Info(fmt.Sprintf("register grpc service to host failed: %v, skipping", err))
		return false
	}

	return true
}

// runHTTPDirectly runs the HTTP service without the host abstraction.
// In microservice mode, also starts the gRPC server alongside HTTP.
//
// runHTTPDirectly 在不使用 host 抽象的情况下直接运行 HTTP 服务。
// 在微服务模式下，同时启动 gRPC 服务器。
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

	// 在直跑模式下，尝试同时启动 gRPC 服务器
	// In direct mode, try to start gRPC server alongside HTTP
	var rpcServer transportcontract.RPCServer
	if c.IsBind(transportcontract.GRPCServerRegistrarKey) {
		if rpcServerAny, rpcErr := c.Make(transportcontract.RPCServerKey); rpcErr == nil {
			if rs, ok := rpcServerAny.(transportcontract.RPCServer); ok {
				if startErr := rs.Start(ctx); startErr == nil {
					rpcServer = rs
					logger.Info("starting grpc server (direct mode)")
				}
			}
		}
	}

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

	// 先停止 gRPC 服务器，再停止 HTTP 服务器
	// Stop gRPC server first, then HTTP server
	if rpcServer != nil {
		if stopErr := rpcServer.Stop(shutdownCtx); stopErr != nil {
			logger.Info(fmt.Sprintf("grpc server stop failed: %v", stopErr))
		} else {
			logger.Info("grpc server stopped gracefully")
		}
	}

	if err := httpSvc.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	logger.Info("http server stopped gracefully")
	return nil
}
