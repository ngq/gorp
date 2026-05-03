package gin

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	providerlog "github.com/ngq/gorp/framework/provider/log"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"

	"github.com/gin-gonic/gin"
)

const httpEngineKey = "framework.http.engine"

// Provider 把 Gin Engine 与 HTTP Server 一起注册进容器。
//
// 中文说明：
// - 这里同时提供两个 key：
//  1. provider 内部 engine key：底层 *gin.Engine
//  2. contract.HTTPKey：对外统一的 HTTP 服务抽象
//
// - 这样路由注册与服务启动可以解耦。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

type router struct {
	group *gin.RouterGroup
}

// ginHTTPContext 是 provider/gin 对 framework HTTPContext 的默认适配实现。
//
// 中文说明：
// - 它把底层 `*gin.Context` 适配成 framework 默认 HTTPContext；
// - Gin 仍可继续作为当前默认 provider，但 Gin 细节只停留在 provider 边界内部；
// - framework 其余层只应依赖 `contract.HTTPContext`，不再直接感知 `*gin.Context`。
type ginHTTPContext struct {
	*contract.DefaultHTTPContext
	gin *gin.Context
}

func (c *ginHTTPContext) GinContext() *gin.Context {
	if c == nil {
		return nil
	}
	return c.gin
}

// Name 返回 provider 名称。
func (p *Provider) Name() string { return "http.gin" }

// IsDefer 表示 gin provider 不走延迟加载。
func (p *Provider) IsDefer() bool {
	return false
}

// Provides 返回 gin provider 暴露的能力 key。
func (p *Provider) Provides() []string { return []string{contract.HTTPKey, httpEngineKey} }

// Register 绑定 Gin Engine 与统一 HTTP 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(httpEngineKey, func(c contract.Container) (any, error) {
		engine := gin.New()
		engine.Use(gin.Recovery())
		engine.Use(injectRequestContainer(c))
		engine.Use(adaptMiddleware(RequestID()))
		engine.Use(adaptMiddleware(TraceID()))
		engine.Use(injectRequestLogger(c))
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

		log.Info("http server initialized", contract.Field{Key: "addr", Value: addr})
		return &service{
			srv:    srv,
			engine: engine,
			router: newRouter(&engine.RouterGroup),
			log:    log,
		}, nil
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

func injectRequestContainer(c contract.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if c != nil && ctx != nil && ctx.Request != nil {
			ctx.Request = ctx.Request.WithContext(contract.NewContainerContext(ctx.Request.Context(), c))
		}
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
				engine.Use(adaptMiddleware(tracingmw.TracingMiddleware(tracer, serviceName)))
			}
		}
	}

	if c.IsBind(contract.MetadataPropagatorKey) {
		if propagatorAny, err := c.Make(contract.MetadataPropagatorKey); err == nil {
			if propagator, ok := propagatorAny.(contract.MetadataPropagator); ok {
				engine.Use(adaptMiddleware(metadatamw.MetadataMiddleware(propagator)))
			}
		}
	}

	if c.IsBind(contract.ServiceAuthKey) {
		if authAny, err := c.Make(contract.ServiceAuthKey); err == nil {
			authenticator, _ := authAny.(contract.ServiceAuthenticator)
			engine.Use(func(authenticator contract.ServiceAuthenticator) gin.HandlerFunc {
				return func(c *gin.Context) {
					ctx := c.Request.Context()
					if auth := strings.TrimSpace(c.GetHeader("Authorization")); auth != "" {
						ctx = context.WithValue(ctx, "authorization", auth)
					}
					if token := strings.TrimSpace(c.GetHeader("X-Service-Token")); token != "" {
						ctx = context.WithValue(ctx, "x-service-token", token)
					}
					if authenticator != nil {
						hasToken := strings.TrimSpace(c.GetHeader("X-Service-Token")) != "" ||
							strings.TrimSpace(c.GetHeader("Authorization")) != ""
						if hasToken {
							identity, err := authenticator.Authenticate(ctx)
							if err != nil {
								c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "service authentication failed"})
								return
							}
							if identity != nil {
								ctx = contract.NewServiceIdentityContext(ctx, identity)
							}
						}
					}
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				}
			}(authenticator))
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
	router contract.HTTPRouter
	log    contract.Logger
}

func (s *service) Router() contract.HTTPRouter { return s.router }
func (s *service) Server() *http.Server        { return s.srv }

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
		l, _ := providerlog.NewDefaultLogger()
		return l
	}
	if logger, ok := v.(contract.Logger); ok {
		return logger
	}
	l, _ := providerlog.NewDefaultLogger()
	return l
}

func newRouter(group *gin.RouterGroup) contract.HTTPRouter {
	return &router{group: group}
}

func newHTTPContext(ctx *gin.Context) contract.HTTPContext {
	base := contract.NewDefaultHTTPContext(nil, nil)
	if ctx != nil {
		base.SetRequest(ctx.Request)
	}
	base.SetParamFunc(func(key string) string {
		if ctx == nil {
			return ""
		}
		return ctx.Param(key)
	})
	base.SetQueryFunc(func(key string) string {
		if ctx == nil {
			return ""
		}
		return ctx.Query(key)
	})
	base.SetDefaultQueryFunc(func(key, defaultValue string) string {
		if ctx == nil {
			return defaultValue
		}
		return ctx.DefaultQuery(key, defaultValue)
	})
	base.SetHeaderFuncs(func(key string) string {
		if ctx == nil {
			return ""
		}
		return ctx.GetHeader(key)
	}, func(key, value string) {
		if ctx == nil {
			return
		}
		ctx.Header(key, value)
	})
	base.SetBindFuncs(func(obj any) error {
		if ctx == nil {
			return nil
		}
		return ctx.ShouldBindJSON(obj)
	}, func(obj any) error {
		if ctx == nil {
			return nil
		}
		return ctx.ShouldBindQuery(obj)
	}, func(obj any) error {
		if ctx == nil {
			return nil
		}
		return ctx.ShouldBind(obj)
	})
	base.SetResponseFuncs(func(status int, body any) {
		if ctx == nil {
			return
		}
		ctx.JSON(status, body)
	}, func(code int) {
		if ctx == nil {
			return
		}
		ctx.Status(code)
	}, func() int {
		if ctx == nil {
			return 0
		}
		return ctx.Writer.Status()
	})
	base.SetRoutePathFunc(func() string {
		if ctx == nil {
			return ""
		}
		return ctx.FullPath()
	})
	return &ginHTTPContext{DefaultHTTPContext: base, gin: ctx}
}

func unwrapGinContext(c contract.HTTPContext) (*gin.Context, bool) {
	type ginContextProvider interface {
		GinContext() *gin.Context
	}
	provider, ok := c.(ginContextProvider)
	if !ok {
		return nil, false
	}
	gc := provider.GinContext()
	if gc == nil {
		return nil, false
	}
	return gc, true
}

func adaptMiddleware(middleware contract.HTTPMiddleware) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if middleware == nil {
			ctx.Next()
			return
		}
		httpCtx := newHTTPContext(ctx)
		middleware(httpCtx, func() {
			ctx.Request = httpCtx.Request()
			ctx.Next()
			httpCtx.SetRequest(ctx.Request)
		})
		ctx.Request = httpCtx.Request()
	}
}

func adaptHandler(handler contract.HTTPHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if handler == nil {
			return
		}
		handler(newHTTPContext(ctx))
	}
}

func (r *router) Use(middleware ...contract.HTTPMiddleware) {
	if r == nil || r.group == nil || len(middleware) == 0 {
		return
	}
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		adapted = append(adapted, adaptMiddleware(mw))
	}
	if len(adapted) == 0 {
		return
	}
	r.group.Use(adapted...)
}

func (r *router) Group(prefix string, middleware ...contract.HTTPMiddleware) contract.HTTPRouter {
	if r == nil || r.group == nil {
		return &router{}
	}
	adapted := make([]gin.HandlerFunc, 0, len(middleware))
	for _, mw := range middleware {
		if mw == nil {
			continue
		}
		adapted = append(adapted, adaptMiddleware(mw))
	}
	return newRouter(r.group.Group(prefix, adapted...))
}

func (r *router) Handle(method, path string, handler contract.HTTPHandler) {
	if r == nil || r.group == nil || handler == nil {
		return
	}
	r.group.Handle(method, path, adaptHandler(handler))
}

func (r *router) HandleFunc(method, path string, handlerFunc contract.HTTPHandler) {
	if handlerFunc == nil {
		return
	}
	r.Handle(method, path, handlerFunc)
}

func (r *router) GET(path string, handler contract.HTTPHandler) {
	r.Handle(http.MethodGet, path, handler)
}

func (r *router) POST(path string, handler contract.HTTPHandler) {
	r.Handle(http.MethodPost, path, handler)
}

func (r *router) PUT(path string, handler contract.HTTPHandler) {
	r.Handle(http.MethodPut, path, handler)
}

func (r *router) DELETE(path string, handler contract.HTTPHandler) {
	r.Handle(http.MethodDelete, path, handler)
}

func (r *router) Mount(path string, handler http.Handler) {
	if handler == nil {
		return
	}
	h := wrapHTTPHandler(handler)
	r.group.Handle(http.MethodGet, path, h)
	r.group.Handle(http.MethodHead, path, h)
}

func wrapHTTPHandler(handler http.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
