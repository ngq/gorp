// Application scenarios:
// - Register Gin as the framework's HTTP provider implementation.
// - Wire the default responder, Gin engine, and runtime HTTP service into the container.
// - Build a production-ready HTTP entrypoint with configuration-driven server settings.
//
// 适用场景：
// - 将 Gin 注册为框架的 HTTP provider 实现。
// - 把默认 responder、Gin engine 和运行时 HTTP service 装配进容器。
// - 以配置驱动的方式构建可用于生产的 HTTP 入口。
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
func (p *Provider) Provides() []string { return []string{transportcontract.HTTPKey, httpEngineKey} }

// Register binds the Gin engine and HTTP service into the container.
//
// Register 将 Gin engine 和 HTTP service 绑定到容器。
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

	c.Bind(httpEngineKey, func(c runtimecontract.Container) (any, error) {
		engine := gin.New()
		engine.Use(injectRequestContainer(c))
		engine.Use(adaptMiddleware(httpmiddleware.DefaultMiddleware(getLogger(c))))
		attachHTTPTransportMiddleware(engine, c)
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
