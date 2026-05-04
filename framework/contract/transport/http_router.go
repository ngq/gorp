package transport

import (
	"context"
	"net/http"
)

type HTTPContext interface {
	Context() context.Context
	SetContext(ctx context.Context)

	Request() *http.Request
	SetRequest(req *http.Request)

	Param(key string) string
	Query(key string) string
	DefaultQuery(key, defaultValue string) string
	GetHeader(key string) string
	Header(key, value string)

	BindJSON(obj any) error
	BindQuery(obj any) error
	Bind(obj any) error

	JSON(status int, body any)
	String(status int, body string)
	XML(status int, body any)
	Data(status int, contentType string, body []byte)
	Redirect(status int, location string)
	Status(code int)

	RoutePath() string
	ResponseStatus() int
}

type HTTPHandler func(HTTPContext)

type HTTPMiddleware func(next HTTPHandler) HTTPHandler

type HTTPRouter interface {
	Use(middleware ...HTTPMiddleware)
	Group(prefix string, middleware ...HTTPMiddleware) HTTPRouter

	Handle(method, path string, handler HTTPHandler)
	HandleFunc(method, path string, handlerFunc HTTPHandler)

	GET(path string, handler HTTPHandler)
	POST(path string, handler HTTPHandler)
	PUT(path string, handler HTTPHandler)
	DELETE(path string, handler HTTPHandler)

	Mount(path string, handler http.Handler)
}
