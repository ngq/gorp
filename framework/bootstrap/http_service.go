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

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworklog "github.com/ngq/gorp/framework/log"
	gingin "github.com/ngq/gorp/framework/provider/gin"
	"github.com/ngq/gorp/framework/provider/host"
	redisProvider "github.com/ngq/gorp/framework/provider/redis"

	gormpkg "gorm.io/gorm"
)

type HTTPServiceOptions struct {
	ExtraProviders []runtimecontract.ServiceProvider
	DisableRedis   bool
	DisableGorm    bool
	DisableMetrics bool
	EnablePprof    bool
}

type HTTPServiceRuntime struct {
	App         *framework.Application
	Container   runtimecontract.Container
	Logger      observabilitycontract.Logger
	Router      transportcontract.HTTPRouter // 抽象路由（保留兼容）
	GinRouter   *gin.RouterGroup             // 原生 gin 路由组
	DB          *gormpkg.DB
	Redis       datacontract.Redis
	JWT         securitycontract.JWTService
	Config      datacontract.Config
	ServiceName string
}

func NewHTTPServiceRuntime(serviceName string, opts HTTPServiceOptions) (*HTTPServiceRuntime, error) {
	app := framework.NewApplication()
	c := app.Container()

	providers := buildHTTPProviders(opts)
	if err := c.RegisterProviders(providers...); err != nil {
		return nil, fmt.Errorf("register providers: %w", err)
	}
	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
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

	// 获取原生 gin.Engine 并创建 RouterGroup
	if ginEngine, err := container.MakeGinEngine(c); err == nil && ginEngine != nil {
		rt.GinRouter = &ginEngine.RouterGroup
	}

	frameworklog.SetDefault(rt.Logger)

	if !opts.DisableGorm {
		rt.DB = container.MustMakeGorm(c)
	}
	if !opts.DisableRedis {
		if redisSvc, err := container.MakeRedis(c); err == nil {
			rt.Redis = redisSvc
		}
	}

	return rt, nil
}

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
	if !opts.DisableMetrics {
		RegisterMetricsEndpoint(rt.Router)
		rt.Router.Use(gingin.MetricsMiddleware())
	}
	if opts.EnablePprof {
		RegisterPprofEndpoints(rt.Router)
	}

	return RunHTTP(rt.Container, rt.Logger)
}

func AutoMigrateModels(rt *HTTPServiceRuntime, models ...any) error {
	if rt == nil || rt.DB == nil || len(models) == 0 {
		return nil
	}
	return rt.DB.AutoMigrate(models...)
}

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

func RegisterMetricsEndpoint(router transportcontract.HTTPRouter) {
	if router == nil {
		return
	}
	gingin.RegisterGoRuntimeMetrics()
	router.Mount("/metrics", gingin.PrometheusHandler())
}

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

func RunHTTP(c runtimecontract.Container, logger observabilitycontract.Logger) error {
	hostSvc, err := container.MakeHost(c)
	if err != nil {
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
