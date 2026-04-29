package gin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	logzap "github.com/ngq/gorp/contrib/log/zap"
	"github.com/ngq/gorp/framework/contract"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	serviceauthtoken "github.com/ngq/gorp/framework/provider/serviceauth/token"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"

	"github.com/gin-gonic/gin"
)

// Provider 把 Gin Engine 与 HTTP Server 一起注册进容器。
//
// 中文说明：
// - 这里同时提供两个 key：
//   1. contract.HTTPEngineKey：底层 *gin.Engine
//   2. contract.HTTPKey：对外统一的 HTTP 服务抽象
// - 这样路由注册与服务启动可以解耦。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

// EngineFromContainer 从容器解析 *gin.Engine。
func EngineFromContainer(c contract.Container) (*gin.Engine, error) {
	// 中文说明：
	// - 统一把 HTTPEngineKey -> *gin.Engine 的解析收进 provider/gin 边界；
	// - 上层 app / template 不再重复书写 c.Make + type assert；
	// - 后续如果 HTTP 宿主继续抽象，这里就是最自然的单点收口位置。
	engineAny, err := c.Make(contract.HTTPEngineKey)
	if err != nil {
		return nil, err
	}
	engine, ok := engineAny.(*gin.Engine)
	if !ok {
		return nil, fmt.Errorf("http engine is not *gin.Engine: %T", engineAny)
	}
	return engine, nil
}

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "http.gin" }

// IsDefer 表示 gin provider 不走延迟加载。
func (p *Provider) IsDefer() bool {
	return false
}

// Provides 返回 gin provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.HTTPKey, contract.HTTPEngineKey} }

// Register 绑定 Gin Engine 与统一 HTTP 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.HTTPEngineKey, func(c contract.Container) (any, error) {
		engine := gin.New()
		engine.Use(gin.Recovery())
		engine.Use(RequestID())
		engine.Use(TraceID())
		engine.Use(injectRequestLogger(c))
		engine.Use(MetricsMiddleware())
		attachHTTPTransportMiddleware(engine, c)
		// 中文说明：
		// - 默认挂载基础中间件：Recovery + RequestID + TraceID + 请求级 logger 注入 + Metrics；
		// - 如果主链路能力已绑定，则继续挂 tracing / metadata / serviceauth 的 HTTP 服务端中间件；
		// - 这样 HTTP 宿主层就成为统一 transport 装配入口，而不是让业务项目手工重复注册。
		return engine, nil
	}, true)

	c.Bind(contract.HTTPKey, func(c contract.Container) (any, error) {
		cfgAny, _ := c.Make(contract.ConfigKey)
		cfg, _ := cfgAny.(contract.Config)

		addr := ":8080"
		readTimeout := 15 * time.Second
		writeTimeout := 15 * time.Second
		idleTimeout := 60 * time.Second

		if cfg != nil {
			if s := configprovider.GetStringAny(cfg,
				"server.http.addr",
				"app.address",
			); s != "" {
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

		engineAny, err := c.Make(contract.HTTPEngineKey)
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

		log.Info("http server initialized", contract.Field{Key: "addr", Value: addr})
		return &service{srv: srv, engine: engine, log: log}, nil
	}, true)

	return nil
}

// Boot 提前确保 logger 已经可用。
func (p *Provider) Boot(c contract.Container) error {
	// ensure logger is created early
	if c.IsBind(contract.LogKey) {
		_, _ = c.Make(contract.LogKey)
	}
	return nil
}

// injectRequestLogger 把请求级 logger 注入到 request context。
//
// 中文说明：
// - 基于容器里的基础 logger，再附加 trace_id / request_id；
// - 业务层后续统一通过 `framework/log.Ctx(ctx)` 读取。
func injectRequestLogger(c contract.Container) gin.HandlerFunc {
	base := getLogger(c)
	return func(ctx *gin.Context) {
		logger := base
		fields := make([]contract.Field, 0, 2)
		if traceID := GetTraceID(ctx); traceID != "" {
			fields = append(fields, contract.Field{Key: "trace_id", Value: traceID})
		}
		if requestID := GetRequestID(ctx); requestID != "" {
			fields = append(fields, contract.Field{Key: "request_id", Value: requestID})
		}
		if len(fields) > 0 {
			logger = logger.With(fields...)
		}
		ctx.Request = ctx.Request.WithContext(frameworkbizlog.WithContext(ctx.Request.Context(), logger))
		ctx.Next()
	}
}

// attachHTTPTransportMiddleware 按已注册能力为 Gin 装配传输层中间件。
func attachHTTPTransportMiddleware(engine *gin.Engine, c contract.Container) {
	if engine == nil {
		return
	}

	if c.IsBind(contract.TracerKey) {
		if tracerAny, err := c.Make(contract.TracerKey); err == nil {
			if tracer, ok := tracerAny.(contract.Tracer); ok {
				serviceName := configprovider.GetStringAny(getConfig(c), "service.name", "tracing.service_name")
				if serviceName == "" {
					serviceName = "http-service"
				}
				engine.Use(tracingmw.TracingMiddleware(tracer, serviceName))
			}
		}
	}

	if c.IsBind(contract.MetadataPropagatorKey) {
		if propagatorAny, err := c.Make(contract.MetadataPropagatorKey); err == nil {
			if propagator, ok := propagatorAny.(contract.MetadataPropagator); ok {
				engine.Use(metadatamw.MetadataMiddleware(propagator))
			}
		}
	}

	if c.IsBind(contract.ServiceAuthKey) {
		if authAny, err := c.Make(contract.ServiceAuthKey); err == nil {
			authenticator, _ := authAny.(contract.ServiceAuthenticator)
			engine.Use(serviceauthtoken.ServiceAuthHTTPMiddleware(authenticator))
		}
	}
}

// getConfig 从容器读取配置服务。
func getConfig(c contract.Container) contract.Config {
	if !c.IsBind(contract.ConfigKey) {
		return nil
	}
	v, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil
	}
	cfg, _ := v.(contract.Config)
	return cfg
}

// service 是 contract.HTTP 的默认 Gin 实现。
//
// 中文说明：
// - srv 持有真正的 net/http.Server。
// - engine 供上层注册路由。
// - log 用于启动与关闭阶段输出统一日志。
type service struct {
	srv    *http.Server
	engine *gin.Engine
	log    contract.Logger
}

func (s *service) Engine() *gin.Engine  { return s.engine }
func (s *service) Server() *http.Server { return s.srv }

// Run 启动 HTTP 监听。
func (s *service) Run() error { return s.srv.ListenAndServe() }

// Shutdown 触发优雅关闭。
func (s *service) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }

// getLogger 从容器中解析 logger；若失败则回退到一个临时 console logger。
//
// 中文说明：
// - 让 gin provider 在日志 provider 尚未完全就绪时也能输出最小日志；
// - 回退 logger 只用于宿主初始化阶段，不改变容器中的正式 logger 绑定。
func getLogger(c contract.Container) contract.Logger {
	v, err := c.Make(contract.LogKey)
	if err != nil {
		l, _ := logzap.New("info", "console")
		return l
	}
	if logger, ok := v.(contract.Logger); ok {
		return logger
	}
	l, _ := logzap.New("info", "console")
	return l
}
