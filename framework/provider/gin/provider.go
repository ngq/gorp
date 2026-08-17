// Package gin provides Gin-based HTTP server implementation for gorp framework.
// Implements framework-level HTTP service contract with Gin engine and net/http server.
// Includes container injection, default responder, Prometheus metrics, and graceful shutdown.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 实现框架级 HTTP 服务契约，包含 Gin engine 和 net/http server。
// 包括容器注入、默认响应器、Prometheus 指标和优雅关闭。
//
// Eg:
//
//	// 注册 Provider
//	app.Register(gin.NewProvider())
//
//	// 获取 HTTP 服务
//	httpSvc := c.MustMake(transportcontract.HTTPKey).(transportcontract.HTTP)
//	httpSvc.Router().GET("/hello", helloHandler)
//	httpSvc.Run()
package gin

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
	"github.com/ngq/gorp/framework/http/serverconfig"
)

const httpEngineKey = "framework.http.engine"

// HTTPEngineKey is the public container key for the underlying Gin engine.
// Use this key to retrieve *gin.Engine from the framework container.
//
// HTTPEngineKey 是底层 Gin engine 的公开容器键。
// 使用此键从框架容器获取 *gin.Engine。
const HTTPEngineKey = httpEngineKey

// Provider wires the Gin-based HTTP server into the framework container.
//
// Provider 将基于 Gin 的 HTTP 服务接入框架容器。
type Provider struct{}

// NewProvider creates a Gin HTTP provider instance.
//
// NewProvider 创建一个 Gin HTTP Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name used by the runtime container.
//
// Name 返回运行时容器使用的 provider 名称。
func (p *Provider) Name() string { return "http.gin" }

// IsDefer reports whether provider registration should be deferred.
//
// IsDefer 表示该 provider 是否需要延迟注册。
func (p *Provider) IsDefer() bool { return false }

// Provides declares the services exported by this provider.
//
// Provides 声明该 provider 对外提供的服务键。
func (p *Provider) Provides() []string {
	return []string{transportcontract.HTTPKey, transportcontract.HTTPRegistryKey, httpEngineKey, transportcontract.MiddlewareRegistryKey}
}

// DependsOn returns the keys this provider depends on.
// Gin provider depends on Config and Log.
//
// DependsOn 返回该 provider 依赖的 key。
// Gin provider 依赖 Config 和 Log。
func (p *Provider) DependsOn() []string {
	return []string{datacontract.ConfigKey, observabilitycontract.LogKey}
}

// Register binds the Gin engine, HTTP service, and middleware registry into the container.
//
// Register 将 Gin engine、HTTP service 和中间件注册表绑定到容器。
//
// Example:
//
//	container.RegisterProviders(ginprovider.NewProvider())
func (p *Provider) Register(c runtimecontract.Container) error {
	if !c.IsBind(transportcontract.HTTPResponderKey) {
		c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
			return httpmiddleware.NewDefaultResponder(), nil
		}, true)
	}

	// 注册中间件注册表，供 proto 注解驱动的自动中间件挂载使用。
	// 用户通过 registry.Register("auth", jwtMiddleware) 注册具名中间件，
	// 生成的路由代码通过 registry.Lookup("auth") 自动查找。
	if !c.IsBind(transportcontract.MiddlewareRegistryKey) {
		c.Bind(transportcontract.MiddlewareRegistryKey, func(runtimecontract.Container) (any, error) {
			return httpmiddleware.NewMiddlewareRegistry(), nil
		}, true)
	}

	c.Bind(transportcontract.HTTPRegistryKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := container.MakeWith[datacontract.Config](c, datacontract.ConfigKey)
		if err != nil {
			return nil, err
		}
		configs, err := serverconfig.Parse(cfg)
		if err != nil {
			return nil, fmt.Errorf("parse HTTP services: %w", err)
		}
		entries := make([]transportcontract.HTTPServiceEntry, 0, len(configs))
		for _, serviceConfig := range configs {
			httpService := newService(c, serviceConfig)
			entries = append(entries, transportcontract.HTTPServiceEntry{
				Name: serviceConfig.Name, Service: httpService, Enabled: serviceConfig.Enabled,
			})
		}
		return newHTTPRegistry(entries), nil
	}, true)

	c.Bind(transportcontract.HTTPKey, func(c runtimecontract.Container) (any, error) {
		registry, err := container.MakeWith[transportcontract.HTTPRegistry](c, transportcontract.HTTPRegistryKey)
		if err != nil {
			return nil, err
		}
		service, ok := registry.Default()
		if !ok {
			return nil, fmt.Errorf("default HTTP service is not configured; use HTTPRegistry for named services")
		}
		return service, nil
	}, true)

	c.Bind(httpEngineKey, func(c runtimecontract.Container) (any, error) {
		httpService, err := container.MakeWith[transportcontract.HTTP](c, transportcontract.HTTPKey)
		if err != nil {
			return nil, err
		}
		provider, ok := httpService.(transportcontract.GINEngineProvider)
		if !ok {
			return nil, fmt.Errorf("default HTTP service does not expose a Gin engine")
		}
		engine, ok := provider.GINEngine().(*gin.Engine)
		if !ok || engine == nil {
			return nil, fmt.Errorf("default HTTP service returned an invalid Gin engine")
		}
		return engine, nil
	}, true)

	return nil
}

// ginModeOnce 保证 gin.SetMode 只生效一次。gin 的 mode 是进程级全局状态，
// 多 HTTP 服务逐个调用 SetMode 会互相覆盖（最终只留最后创建的服务配置，
// 并发创建时还存在对全局变量的并发写）。
var ginModeOnce sync.Once

func newService(c runtimecontract.Container, cfg serverconfig.Service) transportcontract.HTTP {
	if cfg.Mode != "" {
		ginModeOnce.Do(func() { gin.SetMode(cfg.Mode) })
	}
	engine := gin.New()
	engine.ContextWithFallback = true
	engine.Use(injectRequestContainer(c))
	logger := getLogger(c).With(
		observabilitycontract.Field{Key: "http_service", Value: cfg.Name},
		observabilitycontract.Field{Key: "addr", Value: cfg.Addr},
	)
	if cfg.Mode != "" && gin.Mode() != cfg.Mode {
		// 第一个服务的 mode 已全局生效；提示配置冲突，避免排查困难。
		logger.Info(fmt.Sprintf("gin mode %q requested but %q already active globally (mode is process-wide; first service wins)", cfg.Mode, gin.Mode()))
	}
	engine.Use(adaptMiddleware(httpmiddleware.DefaultMiddleware(logger)))
	attachHTTPTransportMiddleware(engine, c, cfg.Name)
	router := newRouter(&engine.RouterGroup, engine)
	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      engine,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	logger.Info("http server initialized", observabilitycontract.Field{Key: "enabled", Value: cfg.Enabled})
	return &service{
		routeSurface:    router,
		srv:             server,
		engine:          engine,
		log:             logger,
		shutdownTimeout: cfg.ShutdownTimeout,
	}
}

// Boot warms up optional dependencies needed at startup.
//
// Boot 预热启动阶段需要的可选依赖。
func (p *Provider) Boot(c runtimecontract.Container) error {
	if c.IsBind(observabilitycontract.LogKey) {
		_, _ = c.Make(observabilitycontract.LogKey)
	}
	return nil
}
