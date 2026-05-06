// Application scenarios:
// - Prepare the Gin engine with framework container, tracing, metadata, and optional service auth hooks.
// - Keep provider-specific lifecycle wiring out of business route registration code.
// - Centralize optional engine-level middleware attachment in one place.
//
// 适用场景：
// - 为 Gin engine 预装框架容器、Tracing、Metadata 与可选服务认证挂载点。
// - 将 provider 专属生命周期装配从业务路由注册代码中隔离出来。
// - 将可选 engine 级中间件挂载逻辑集中收口到一个位置。
package gin

import (
	"context"
	"net/http"
	"strings"

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

// injectRequestContainer writes the framework container into each request context.
//
// injectRequestContainer 将框架容器写入每个请求上下文。
func injectRequestContainer(c runtimecontract.Container) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if c != nil && ctx != nil && ctx.Request != nil {
			ctx.Request = ctx.Request.WithContext(supportcontract.NewContainerContext(ctx.Request.Context(), c))
		}
		ctx.Next()
	}
}

// attachHTTPTransportMiddleware attaches optional transport-related middleware to the engine.
//
// attachHTTPTransportMiddleware 为 engine 挂载可选的 transport 相关中间件。
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

// getConfig loads the config contract from the container when available.
//
// getConfig 在可用时从容器中加载 config 契约。
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

// getLogger loads the logger contract or falls back to the default provider logger.
//
// getLogger 加载 logger 契约，失败时回退到默认 provider logger。
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
