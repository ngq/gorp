package transport

import (
	"context"
	"net/http"
)

const HTTPKey = "framework.http"

type HTTP interface {
	Router() HTTPRouter
	Server() *http.Server

	Run() error
	Shutdown(ctx context.Context) error
}
