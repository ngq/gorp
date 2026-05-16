// Package gin provides Gin-based HTTP server implementation for gorp framework.
// This file implements engine-level middleware attachment and container injection.
// Mounts tracing, metadata propagation, and optional service auth middleware.
//
// Gin HTTP 服务包，提供基于 Gin 的 HTTP 服务器实现，用于 gorp 框架。
// 本文件实现 engine 级中间件挂载和容器注入。
// 挂载 tracing、元数据传播和可选的服务认证中间件。
package gin

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/container"
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
// Falls back to nil when config service is not bound or type mismatch.
//
// getConfig 在可用时从容器中加载 config 契约。
// 配置服务未绑定或类型不匹配时返回 nil。
func getConfig(c runtimecontract.Container) datacontract.Config {
	if !c.IsBind(datacontract.ConfigKey) {
		return nil
	}
	cfg, err := container.MakeWith[datacontract.Config](c, datacontract.ConfigKey)
	if err != nil {
		return nil
	}
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

// newGinFirstEngine 创建 Gin-first 模式下的 Engine。
// 只注入 injectRequestContainer（框架容器注入），不自动挂载治理 preset。
// 用户获取 *gin.Engine 后通过 engine.Use(AdaptMiddleware(...)) 按需手动挂载治理 middleware。
//
// newGinFirstEngine creates the Engine for Gin-first mode.
// Only injects container middleware; governance presets are NOT auto-mounted.
// Users obtain *gin.Engine and manually mount governance middleware via engine.Use(AdaptMiddleware(...)).
func newGinFirstEngine(c runtimecontract.Container) *gin.Engine {
	engine := gin.New()
	engine.Use(injectRequestContainer(c))
	return engine
}

// isGinFirstMode detects whether the current governance mode is gin-first.
//
// isGinFirstMode 检测当前治理模式是否为 gin-first。
func isGinFirstMode(c runtimecontract.Container) bool {
	cfg := getConfig(c)
	if cfg == nil {
		return false
	}
	modeStr := strings.TrimSpace(cfg.GetString("governance.mode"))
	return strings.EqualFold(modeStr, "gin-first") || strings.EqualFold(modeStr, "ginfirst")
}
