// Application scenarios:
// - Define the transport-layer HTTP service contract shared by providers and bootstrap logic.
// - Keep router access, server access, run, and shutdown semantics stable.
// - Let application code depend on one HTTP abstraction instead of a concrete framework.
//
// 适用场景：
// - 定义 provider 和 bootstrap 共同使用的 transport 层 HTTP 服务契约。
// - 稳定维护路由访问、server 访问、运行和关闭语义。
// - 让应用代码依赖统一 HTTP 抽象，而不是具体框架实现。
package transport

import (
	"context"
	"net/http"
)

// HTTPKey is the container key for the HTTP service capability.
//
// HTTPKey 是 HTTP 服务能力的容器键。
const HTTPKey = "framework.http"

// HTTP defines the transport-layer HTTP service abstraction.
//
// HTTP 定义 transport 层 HTTP 服务抽象。
type HTTP interface {
	// Router returns the framework HTTP router facade.
	//
	// Router 返回框架 HTTP 路由门面。
	Router() HTTPRouter

	// Server returns the underlying net/http server.
	//
	// Server 返回底层 net/http server。
	Server() *http.Server

	// Run starts serving HTTP traffic.
	//
	// Run 启动 HTTP 流量服务。
	Run() error

	// Shutdown gracefully stops the HTTP service.
	//
	// Shutdown 优雅关闭 HTTP 服务。
	Shutdown(ctx context.Context) error
}
