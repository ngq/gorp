// Package gorp provides the root-package application startup surface for gorp framework.
package gorp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ngq/gorp/framework/bootstrap"
	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	featurecontract "github.com/ngq/gorp/framework/contract/feature"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	securitycontract "github.com/ngq/gorp/framework/contract/security"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/httpx"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
	featureprovider "github.com/ngq/gorp/framework/provider/feature"
	gingin "github.com/ngq/gorp/framework/provider/gin"
	gormpkg "gorm.io/gorm"
)

// Engine is the high-ergonomics application engine wrapper for Gin Plus mode.
// Provides direct Gin routing, config-driven resource access, and graceful shutdown.
type Engine struct {
	runtime   *bootstrap.HTTPServiceRuntime
	ginEngine *gin.Engine
}

// Default creates a pre-configured gorp Engine with standard middleware,
// config-driven database/redis bindings, structured logger, and graceful shutdown.
func Default() *Engine {
	opts := bootstrap.HTTPServiceOptions{
		HTTPMode: string(resiliencecontract.HTTPModeGin),
	}
	rt, err := bootstrap.NewHTTPServiceRuntime(opts)
	if err != nil {
		panic(fmt.Sprintf("gorp: failed to initialize default engine: %v", err))
	}
	engine, ok := gingin.MustEngine(rt.Container)
	if !ok || engine == nil {
		engine = gin.New()
	}
	app := &Engine{
		runtime:   rt,
		ginEngine: engine,
	}
	app.EnableDebugDashboard()
	return app
}

// New creates a minimal gorp Engine.
func New() *Engine {
	return Default()
}

// Gin returns the underlying native *gin.Engine.
func (e *Engine) Gin() *gin.Engine {
	return e.ginEngine
}

// Container returns the underlying runtime dependency injection container.
func (e *Engine) Container() runtimecontract.Container {
	if e.runtime == nil {
		return nil
	}
	return e.runtime.Container
}

// Config returns the configuration capability.
func (e *Engine) Config() datacontract.Config {
	if e.runtime == nil {
		return nil
	}
	return e.runtime.Config
}

// Logger returns the structured logger capability.
func (e *Engine) Logger() observabilitycontract.Logger {
	if e.runtime == nil {
		return nil
	}
	return e.runtime.Logger
}

// DB returns the GORM database instance if configured, or nil.
func (e *Engine) DB() *gormpkg.DB {
	if e.runtime == nil {
		return nil
	}
	return e.runtime.DB
}

// SQLX returns the SQLX database instance if available, or nil.
func (e *Engine) SQLX() *sqlx.DB {
	if e.runtime == nil || e.runtime.Container == nil {
		return nil
	}
	db, _ := container.MakeWith[*sqlx.DB](e.runtime.Container, datacontract.SQLXKey)
	return db
}

// Redis returns the Redis capability if configured, or nil.
func (e *Engine) Redis() datacontract.Redis {
	if e.runtime == nil {
		return nil
	}
	return e.runtime.Redis
}

// JWT returns the JWT security capability.
func (e *Engine) JWT() securitycontract.JWTService {
	if e.runtime == nil {
		return nil
	}
	return e.runtime.JWT
}

// RPCClient returns the declarative RPC client instance if available.
func (e *Engine) RPCClient() transportcontract.RPCClient {
	if e.runtime == nil || e.runtime.Container == nil {
		return nil
	}
	client, _ := container.MakeRPCClient(e.runtime.Container)
	return client
}

// Feature returns the FeatureManager for dynamic feature flags and canary rollouts.
func (e *Engine) Feature() featurecontract.FeatureManager {
	if e.runtime != nil && e.runtime.Container != nil && e.runtime.Container.IsBind(featurecontract.FeatureKey) {
		if fmAny, err := e.runtime.Container.Make(featurecontract.FeatureKey); err == nil {
			if fm, ok := fmAny.(featurecontract.FeatureManager); ok {
				return fm
			}
		}
	}
	return featureprovider.NewService()
}

// Swagger mounts an interactive Swagger UI endpoint at the specified path prefix (e.g. "/swagger").
func (e *Engine) Swagger(pathPrefix string, specPath ...string) {
	if e.ginEngine == nil {
		return
	}
	targetSpec := "docs/openapi.json"
	if len(specPath) > 0 && specPath[0] != "" {
		targetSpec = specPath[0]
	}
	handler := httpx.GinSwaggerHandler(targetSpec)
	prefix := strings.TrimSuffix(pathPrefix, "/")
	e.ginEngine.GET(prefix, handler)
	e.ginEngine.GET(prefix+"/*any", handler)
}

// EnableDebugDashboard mounts the /debug/gorp Web Console Dashboard and API endpoints.
func (e *Engine) EnableDebugDashboard(pathPrefix ...string) {
	if e.ginEngine == nil {
		return
	}
	prefix := "/debug/gorp"
	if len(pathPrefix) > 0 && pathPrefix[0] != "" {
		prefix = pathPrefix[0]
	}
	var containerRef runtimecontract.Container
	if e.runtime != nil {
		containerRef = e.runtime.Container
	}
	httpx.MountDebugDashboard(e.ginEngine, containerRef, prefix)
}

// OpenAPI returns the auto-exported OpenAPI 3.0 document JSON string.
func (e *Engine) OpenAPI(title, version string) string {
	return httpx.ExportOpenAPI3JSON(e.ginEngine, title, version)
}

// Runtime returns the underlying HTTPServiceRuntime.
func (e *Engine) Runtime() *bootstrap.HTTPServiceRuntime {
	return e.runtime
}

// --- Direct Gin Routing Delegation ---

func (e *Engine) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.GET(relativePath, handlers...)
}

func (e *Engine) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.POST(relativePath, handlers...)
}

func (e *Engine) PUT(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.PUT(relativePath, handlers...)
}

func (e *Engine) DELETE(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.DELETE(relativePath, handlers...)
}

func (e *Engine) PATCH(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.PATCH(relativePath, handlers...)
}

func (e *Engine) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.OPTIONS(relativePath, handlers...)
}

func (e *Engine) HEAD(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.HEAD(relativePath, handlers...)
}

func (e *Engine) Any(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.Any(relativePath, handlers...)
}

func (e *Engine) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return e.ginEngine.Group(relativePath, handlers...)
}

func (e *Engine) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.Use(middleware...)
}

func (e *Engine) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return e.ginEngine.Handle(httpMethod, relativePath, handlers...)
}

func (e *Engine) Static(relativePath, root string) gin.IRoutes {
	return e.ginEngine.Static(relativePath, root)
}

func (e *Engine) StaticFile(relativePath, filepath string) gin.IRoutes {
	return e.ginEngine.StaticFile(relativePath, filepath)
}

func (e *Engine) StaticFS(relativePath string, fs http.FileSystem) gin.IRoutes {
	return e.ginEngine.StaticFS(relativePath, fs)
}

// Run starts the HTTP server with deterministic graceful shutdown and container cleanup.
func (e *Engine) Run(addr ...string) error {
	return e.RunContext(context.Background(), addr...)
}

// RunContext starts the HTTP server with parent context support.
func (e *Engine) RunContext(parent context.Context, addr ...string) (retErr error) {
	if e.runtime != nil && e.runtime.Container != nil {
		defer func() {
			_ = e.runtime.Container.Destroy()
		}()
	}

	listenAddr := ":8080"
	if len(addr) > 0 && addr[0] != "" {
		listenAddr = addr[0]
	} else if e.runtime != nil && e.runtime.Config != nil {
		if cfgAddr := e.runtime.Config.GetString("app.address"); cfgAddr != "" {
			listenAddr = cfgAddr
		} else if port := e.runtime.Config.GetInt("app.port"); port > 0 {
			listenAddr = fmt.Sprintf(":%d", port)
		}
	}

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: e.ginEngine,
	}

	if e.runtime != nil && e.runtime.Logger != nil {
		e.runtime.Logger.Info("starting http server",
			observabilitycontract.Field{Key: "addr", Value: listenAddr},
			observabilitycontract.Field{Key: "service", Value: e.runtime.ServiceName},
		)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(parent, sigs...)
	defer stop()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		if e.runtime != nil && e.runtime.Logger != nil {
			e.runtime.Logger.Info("shutdown signal received")
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	if e.runtime != nil && e.runtime.Logger != nil {
		e.runtime.Logger.Info("http server stopped gracefully")
	}
	return nil
}

// UseSecurityGuard enables SQLi/XSS/Path Traversal guard middleware.
func (e *Engine) UseSecurityGuard(opts ...httpmiddleware.SecurityGuardOption) {
	if e.ginEngine != nil {
		e.ginEngine.Use(gingin.AdaptMiddleware(httpmiddleware.SecurityGuardMiddleware(opts...)))
	}
}

// UseK8sProbes registers K8s standard three-color probes (/startupz, /readyz, /livez, /healthz).
func (e *Engine) UseK8sProbes() {
	if e.ginEngine != nil {
		e.ginEngine.GET("/startupz", gingin.AdaptHandler(httpmiddleware.StartupHandler(nil)))
		e.ginEngine.GET("/readyz", gingin.AdaptHandler(httpmiddleware.ReadinessHandlerFromContainer(e.Container())))
		e.ginEngine.GET("/livez", gingin.AdaptHandler(httpmiddleware.LivenessHandler()))
		e.ginEngine.GET("/healthz", gingin.AdaptHandler(httpmiddleware.HealthCheckHandlerFromContainer(e.Container())))
	}
}
