package gorp

import (
	"context"
	"net/http"
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
	"github.com/stretchr/testify/require"
)

type captureRouter struct {
	used []HTTPMiddleware
}

func (r *captureRouter) Use(middleware ...HTTPMiddleware) {
	r.used = append(r.used, middleware...)
}
func (r *captureRouter) Group(prefix string, middleware ...HTTPMiddleware) HTTPRouter {
	r.used = append(r.used, middleware...)
	return r
}
func (r *captureRouter) Handle(method, path string, handler HTTPHandler)         {}
func (r *captureRouter) HandleFunc(method, path string, handlerFunc HTTPHandler) {}
func (r *captureRouter) GET(path string, handler HTTPHandler)                    {}
func (r *captureRouter) POST(path string, handler HTTPHandler)                   {}
func (r *captureRouter) PUT(path string, handler HTTPHandler)                    {}
func (r *captureRouter) DELETE(path string, handler HTTPHandler)                 {}
func (r *captureRouter) Mount(path string, handler http.Handler)                 {}

func TestMiddlewareWrapsNextHandler(t *testing.T) {
	var calls []string

	mw := Middleware(func(ctx HTTPContext, next HTTPHandler) {
		calls = append(calls, "before")
		require.NotNil(t, ctx)
		if next != nil {
			next(ctx)
		}
		calls = append(calls, "after")
	})

	handler := mw(func(ctx HTTPContext) {
		calls = append(calls, "handler")
	})

	handler(newTestHTTPContext())

	require.Equal(t, []string{"before", "handler", "after"}, calls)
}

func TestMiddlewareNilFuncFallsBackToNext(t *testing.T) {
	var called bool

	handler := Middleware(nil)(func(ctx HTTPContext) {
		called = true
	})

	handler(newTestHTTPContext())

	require.True(t, called)
}

func TestMiddlewareCanShortCircuitWithoutCallingNext(t *testing.T) {
	var called bool

	handler := Middleware(func(ctx HTTPContext, next HTTPHandler) {
		ctx.Status(204)
	})(func(ctx HTTPContext) {
		called = true
	})

	httpCtx := newTestHTTPContext()
	handler(httpCtx)

	require.False(t, called)
	require.Equal(t, 204, httpCtx.ResponseStatus())
}

func TestChainWrapsMiddlewareInOrder(t *testing.T) {
	var calls []string

	m1 := Middleware(func(ctx HTTPContext, next HTTPHandler) {
		calls = append(calls, "m1-before")
		if next != nil {
			next(ctx)
		}
		calls = append(calls, "m1-after")
	})
	m2 := Middleware(func(ctx HTTPContext, next HTTPHandler) {
		calls = append(calls, "m2-before")
		if next != nil {
			next(ctx)
		}
		calls = append(calls, "m2-after")
	})

	handler := Chain(m1, m2)(func(ctx HTTPContext) {
		calls = append(calls, "handler")
	})

	handler(newTestHTTPContext())

	require.Equal(t, []string{
		"m1-before",
		"m2-before",
		"handler",
		"m2-after",
		"m1-after",
	}, calls)
}

func TestChainSkipsNilMiddleware(t *testing.T) {
	var calls []string

	m1 := Middleware(func(ctx HTTPContext, next HTTPHandler) {
		calls = append(calls, "m1")
		if next != nil {
			next(ctx)
		}
	})
	m2 := Middleware(func(ctx HTTPContext, next HTTPHandler) {
		calls = append(calls, "m2")
		if next != nil {
			next(ctx)
		}
	})

	handler := Chain(m1, nil, m2)(func(ctx HTTPContext) {
		calls = append(calls, "handler")
	})

	handler(newTestHTTPContext())

	require.Equal(t, []string{"m1", "m2", "handler"}, calls)
}

func TestChainWithNoMiddlewareFallsBackToNext(t *testing.T) {
	var called bool

	handler := Chain()(func(ctx HTTPContext) {
		called = true
	})

	handler(newTestHTTPContext())

	require.True(t, called)
}

func TestDefaultHTTPServiceGovernanceDefaultsMatchesMainlineTransportPreset(t *testing.T) {
	got := DefaultHTTPServiceGovernanceDefaults()
	expected := httpmiddleware.DefaultHTTPServiceGovernanceDefaults()

	require.Equal(t, expected.MaxConcurrent, got.MaxConcurrent)
	require.Equal(t, expected.API.Timeout, got.API.Timeout)
	require.Equal(t, expected.API.BodyLimitBytes, got.API.BodyLimitBytes)
	require.Equal(t, expected.API.EnableMetrics, got.API.EnableMetrics)
	require.Equal(t, expected.API.EnableCompression, got.API.EnableCompression)
	require.Nil(t, got.API.CORS)
	require.NotNil(t, got.API.SecurityHeaders)
	require.NotNil(t, got.API.Locale)
}

func TestDefaultHTTPServiceGovernanceSetMatchesMainlineCardinality(t *testing.T) {
	got := DefaultHTTPServiceGovernanceSet(nil, DefaultHTTPServiceGovernanceOptions{})
	expected := httpmiddleware.DefaultHTTPServiceGovernanceSet(nil, httpmiddleware.DefaultHTTPServiceGovernanceOptions{})
	require.Len(t, got, len(expected))
}

func TestUseDefaultHTTPServiceGovernanceForwardsMiddlewareSet(t *testing.T) {
	router := &captureRouter{}
	opts := DefaultHTTPServiceGovernanceOptions{
		MaxConcurrent: 8,
	}

	UseDefaultHTTPServiceGovernance(router, nil, opts)

	expected := DefaultHTTPServiceGovernanceSet(nil, opts)
	require.Len(t, router.used, len(expected))
}

func newTestHTTPContext() *transportcontract.DefaultHTTPContext {
	httpCtx := transportcontract.NewDefaultHTTPContext(context.Background(), nil)
	var status int
	httpCtx.SetResponseFuncs(func(code int, body any) {
		status = code
	}, func(code int, body string) {
		status = code
	}, func(code int, body any) {
		status = code
	}, func(code int, contentType string, body []byte) {
		status = code
	}, func(code int, location string) {
		status = code
	}, func(code int) {
		status = code
	}, func() int {
		return status
	})
	return httpCtx
}
