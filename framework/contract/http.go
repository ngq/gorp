package contract

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	HTTPKey       = "framework.http"
	HTTPEngineKey = "framework.http.engine"
)

// HTTP is the web server service.
// Note: Engine() returns *gin.Engine directly. We intentionally don't abstract
// the HTTP router layer because Gin is the de-facto standard and replacing it
// is extremely rare in practice.
type HTTP interface {
	Engine() *gin.Engine
	Server() *http.Server

	Run() error
	Shutdown(ctx context.Context) error
}
