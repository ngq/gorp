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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
	configprovider "github.com/ngq/gorp/framework/provider/config"
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
	return []string{transportcontract.HTTPKey, httpEngineKey, transportcontract.MiddlewareRegistryKey}
}

// DependsOn returns the keys this provider depends on.
// Gin provider depends on Config and Log.
//
// DependsOn 返回该 provider 依赖的 key。
// Gin provider 依赖 Config 和 Log。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey, observabilitycontract.LogKey} }

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

	c.Bind(httpEngineKey, func(c runtimecontract.Container) (any, error) {
		var engine *gin.Engine
		if isGinFirstMode(c) {
			// Gin-first 模式：只注入容器中间件，不自动挂载治理 preset
			// 用户通过 engine.Use(AdaptMiddleware(...)) 手动挂载
			engine = newGinFirstEngine(c)
		} else {
			// 抽象主线模式：自动挂载默认治理 preset + transport middleware
			engine = gin.New()
			engine.Use(injectRequestContainer(c))
			engine.Use(adaptMiddleware(httpmiddleware.DefaultMiddleware(getLogger(c))))
			attachHTTPTransportMiddleware(engine, c)
		}
		return engine, nil
	}, true)

	c.Bind(transportcontract.HTTPKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, _ := c.Make(datacontract.ConfigKey)
		cfg, _ := cfgAny.(datacontract.Config)

		addr := ":8080"
		readTimeout := 15 * time.Second
		writeTimeout := 15 * time.Second
		idleTimeout := 60 * time.Second

		if cfg != nil {
			if s := configprovider.GetStringAny(cfg, "server.http.addr", "app.address"); s != "" {
				addr = s
			}
			if n := cfg.GetInt("app.http.read_timeout_sec"); n > 0 {
				readTimeout = time.Duration(n) * time.Second
			}
			if n := cfg.GetInt("app.http.write_timeout_sec"); n > 0 {
				writeTimeout = time.Duration(n) * time.Second
			}
			if n := cfg.GetInt("app.http.idle_timeout_sec"); n > 0 {
				idleTimeout = time.Duration(n) * time.Second
			}
		}

		engineAny, err := c.Make(httpEngineKey)
		if err != nil {
			return nil, err
		}
		engine := engineAny.(*gin.Engine)

		log := getLogger(c)
		srv := &http.Server{
			Addr:         addr,
			Handler:      engine,
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
			IdleTimeout:  idleTimeout,
		}

		log.Info("http server initialized", observabilitycontract.Field{Key: "addr", Value: addr})
		return &service{
			srv:    srv,
			engine: engine,
			router: newRouter(&engine.RouterGroup),
			log:    log,
		}, nil
	}, true)

	return nil
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
