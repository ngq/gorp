// Application scenarios:
// - Host the runtime-facing HTTP service implementation backed by Gin and net/http.
// - Expose both the framework router facade and the underlying net/http server object.
// - Support application lifecycle operations such as run and graceful shutdown.
//
// 适用场景：
// - 承载基于 Gin 与 net/http 的运行时 HTTP 服务实现。
// - 同时暴露框架路由门面与底层 net/http server 对象。
// - 支持启动与优雅关闭等应用生命周期操作。
package gin

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

type service struct {
	srv    *http.Server
	engine *gin.Engine
	router transportcontract.HTTPRouter
	log    observabilitycontract.Logger
}

// service is the runtime HTTP service implementation built on top of Gin.
//
// service 是构建在 Gin 之上的运行时 HTTP 服务实现。
//
// Router returns the framework HTTP router facade backed by Gin.
//
// Router 返回由 Gin 驱动的框架 HTTPRouter 门面。
func (s *service) Router() transportcontract.HTTPRouter { return s.router }

// Server returns the underlying net/http server instance.
//
// Server 返回底层 net/http server 实例。
func (s *service) Server() *http.Server { return s.srv }

// Run starts serving HTTP traffic.
//
// Run 启动 HTTP 服务监听。
func (s *service) Run() error { return s.srv.ListenAndServe() }

// Shutdown gracefully stops the HTTP server.
//
// Shutdown 优雅关闭 HTTP 服务。
func (s *service) Shutdown(ctx context.Context) error { return s.srv.Shutdown(ctx) }
