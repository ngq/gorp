package contract

import (
	"context"
	"net/http"
)

const (
	HTTPKey = "framework.http"
)

// HTTP is the web server service.
//
// 中文说明：
// - Router() 是新的 framework 主入口；
// - 新框架主线只暴露 Router / Server 宿主能力，不再把 Gin engine 作为公开契约的一部分；
// - 业务与模板默认应优先依赖 Router()，而不是感知具体 Web engine。
type HTTP interface {
	Router() HTTPRouter
	Server() *http.Server

	Run() error
	Shutdown(ctx context.Context) error
}
