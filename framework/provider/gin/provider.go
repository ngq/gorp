package gin

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	providerlog "github.com/ngq/gorp/framework/provider/log"
	metadatamw "github.com/ngq/gorp/framework/provider/metadata/middleware"
	tracingmw "github.com/ngq/gorp/framework/provider/tracing/middleware"
)

const httpEngineKey = "framework.http.engine"

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

type router struct {
	group *gin.RouterGroup
}

type ginHTTPContext struct {
	*transportcontract.DefaultHTTPContext
	gin *gin.Context
}

func (c *ginHTTPContext) GinContext() *gin.Context {
	if c == nil {
		return nil
	}
	return c.gin
}

func (p *Provider) Name() string { return "http.gin" }

func (p *Provider) IsDefer() bool { return false }

func (p *Provider) Provides() []string { return []string{transportcontract.HTTPKey, httpEngineKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	if !c.IsBind(transportcontract.HTTPResponderKey) {
		c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
			return NewDefaultResponder(), nil
		}, true)
	}

	c.Bind(httpEngineKey, func(c runtimecontract.Container) (any, error) {
		engine := gin.New()
		engine.Use(injectRequestContainer(c))
		engine.Use(adaptMiddleware(RequestID()))
		engine.Use(adaptMiddleware(TraceID()))
		engine.Use(adaptMiddleware(LoggingMiddleware(getLogger(c))))
		engine.Use(adaptMiddleware(RecoveryMiddleware()))
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

func (p *Provider) Boot(c runtimecontract.Container) error {
	if c.IsBind(observabilitycontract.LogKey) {
		_, _ = c.Make(observabilitycontract.LogKey)
	}
	return nil
}

func injectRequestContainer(c runtimecontract.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if c != nil && ctx != nil && ctx.Request != nil {
			ctx.Request = ctx.Request.WithContext(supportcontract.NewContainerContext(ctx.Request.Context(), c))
		}
		ctx.Next()
	}
}

func attachHTTPTransportMiddleware(engine *gin.Engine, c runtimecontract.Container) {
	if engine == nil {
		return
	}

	if c.IsBind(observabilitycontract.TracerKey) {
		if tracerAny, err := c.Make(observabilitycontract.TracerKey); err == nil {
			if tracer, ok := tracerAny.(observabilitycontract.Tracer); ok {
				serviceName := configprovider.GetStringAny(getConfig(c), "service.name", "tracing.service_name")
				if serviceName == "" {
					serviceName = "http-service"
				}
				engine.Use(adaptMiddleware(tracingmw.TracingMiddleware(tracer, serviceName)))
			}
		}
	}

	if c.IsBind(transportcontract.MetadataPropagatorKey) {
		if propagatorAny, err := c.Make(transportcontract.MetadataPropagatorKey); err == nil {
			if propagator, ok := propagatorAny.(transportcontract.MetadataPropagator); ok {
				engine.Use(adaptMiddleware(metadatamw.MetadataMiddleware(propagator)))
			}
		}
	}

	if c.IsBind(securitycontract.ServiceAuthKey) {
		if authAny, err := c.Make(securitycontract.ServiceAuthKey); err == nil {
			authenticator, _ := authAny.(securitycontract.ServiceAuthenticator)
			engine.Use(func(authenticator securitycontract.ServiceAuthenticator) gin.HandlerFunc {
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
								ctx = securitycontract.NewServiceIdentityContext(ctx, identity)
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

func getConfig(c runtimecontract.Container) datacontract.Config {
	if !c.IsBind(datacontract.ConfigKey) {
		return nil
	}
	v, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil
	}
	cfg, _ := v.(datacontract.Config)
	return cfg
}

type service struct {
	srv    *http.Server
	engine *gin.Engine
	router transportcontract.HTTPRouter
	log    observabilitycontract.Logger
}

func (s *service) Router() transportcontract.HTTPRouter { return s.router }
func (s *service) Server() *http.Server                 { return s.srv }
func (s *service) Run() error                           { return s.srv.ListenAndServe() }
func (s *service) Shutdown(ctx context.Context) error   { return s.srv.Shutdown(ctx) }

func getLogger(c runtimecontract.Container) observabilitycontract.Logger {
	v, err := c.Make(observabilitycontract.LogKey)
	if err != nil {
		l, _ := providerlog.NewDefaultLogger()
		return l
	}
	if logger, ok := v.(observabilitycontract.Logger); ok {
		return logger
	}
	l, _ := providerlog.NewDefaultLogger()
	return l
}

func newRouter(group *gin.RouterGroup) transportcontract.HTTPRouter {
	return &router{group: group}
}

func newHTTPContext(ctx *gin.Context) transportcontract.HTTPContext {
	base := transportcontract.NewDefaultHTTPContext(nil, nil)
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
	}, func(status int, body string) {
		if ctx == nil {
			return
		}
		ctx.String(status, body)
	}, func(status int, body any) {
		if ctx == nil {
			return
		}
		ctx.XML(status, body)
	}, func(status int, contentType string, body []byte) {
		if ctx == nil {
			return
		}
		ctx.Data(status, contentType, body)
	}, func(status int, location string) {
		if ctx == nil {
			return
		}
		ctx.Redirect(status, location)
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

func unwrapGinContext(c transportcontract.HTTPContext) (*gin.Context, bool) {
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

func adaptMiddleware(middleware transportcontract.HTTPMiddleware) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if middleware == nil {
			ctx.Next()
			return
		}
		httpCtx := newHTTPContext(ctx)
		next := func(c transportcontract.HTTPContext) {
			if c != nil && c.Request() != nil {
				ctx.Request = c.Request()
			}
			ctx.Next()
			httpCtx.SetRequest(ctx.Request)
		}
		wrapped := middleware(next)
		if wrapped == nil {
			ctx.Next()
			return
		}
		wrapped(httpCtx)
		ctx.Request = httpCtx.Request()
	}
}

func adaptHandler(handler transportcontract.HTTPHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if handler == nil {
			return
		}
		handler(newHTTPContext(ctx))
	}
}

func (r *router) Use(middleware ...transportcontract.HTTPMiddleware) {
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

func (r *router) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
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

func (r *router) Handle(method, path string, handler transportcontract.HTTPHandler) {
	if r == nil || r.group == nil || handler == nil {
		return
	}
	r.group.Handle(method, path, adaptHandler(handler))
}

func (r *router) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {
	if handlerFunc == nil {
		return
	}
	r.Handle(method, path, handlerFunc)
}

func (r *router) GET(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodGet, path, handler)
}

func (r *router) POST(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPost, path, handler)
}

func (r *router) PUT(path string, handler transportcontract.HTTPHandler) {
	r.Handle(http.MethodPut, path, handler)
}

func (r *router) DELETE(path string, handler transportcontract.HTTPHandler) {
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
