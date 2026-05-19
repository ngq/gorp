package gorp

import (
	"net/http"
	"testing"

	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
	"github.com/stretchr/testify/require"
)

type captureRouter struct {
	used []Middleware
}

func (r *captureRouter) Use(middleware ...Middleware) {
	r.used = append(r.used, middleware...)
}
func (r *captureRouter) Group(prefix string, middleware ...Middleware) Router {
	r.used = append(r.used, middleware...)
	return r
}
func (r *captureRouter) Handle(method, path string, handler Handler)         {}
func (r *captureRouter) HandleFunc(method, path string, handlerFunc Handler) {}
func (r *captureRouter) GET(path string, handler Handler)                    {}
func (r *captureRouter) POST(path string, handler Handler)                   {}
func (r *captureRouter) PUT(path string, handler Handler)                    {}
func (r *captureRouter) DELETE(path string, handler Handler)                 {}
func (r *captureRouter) Mount(path string, handler http.Handler)             {}

func TestAdaptMiddlewareWrapsNextHandler(t *testing.T) {
	var calls []string

	mw := AdaptMiddleware(func(ctx Context, next Handler) {
		calls = append(calls, "before")
		if next != nil {
			next(ctx)
		}
		calls = append(calls, "after")
	})

	handler := mw(func(ctx Context) {
		calls = append(calls, "handler")
	})

	handler(nil)

	require.Equal(t, []string{"before", "handler", "after"}, calls)
}

func TestAdaptMiddlewareNilFuncFallsBackToNext(t *testing.T) {
	var called bool

	handler := AdaptMiddleware(nil)(func(ctx Context) {
		called = true
	})

	handler(nil)

	require.True(t, called)
}

func TestChainWrapsMiddlewareInOrder(t *testing.T) {
	var calls []string

	m1 := AdaptMiddleware(func(ctx Context, next Handler) {
		calls = append(calls, "m1-before")
		if next != nil {
			next(ctx)
		}
		calls = append(calls, "m1-after")
	})
	m2 := AdaptMiddleware(func(ctx Context, next Handler) {
		calls = append(calls, "m2-before")
		if next != nil {
			next(ctx)
		}
		calls = append(calls, "m2-after")
	})

	handler := Chain(m1, m2)(func(ctx Context) {
		calls = append(calls, "handler")
	})

	handler(nil)

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

	m1 := AdaptMiddleware(func(ctx Context, next Handler) {
		calls = append(calls, "m1")
		if next != nil {
			next(ctx)
		}
	})
	m2 := AdaptMiddleware(func(ctx Context, next Handler) {
		calls = append(calls, "m2")
		if next != nil {
			next(ctx)
		}
	})

	handler := Chain(m1, nil, m2)(func(ctx Context) {
		calls = append(calls, "handler")
	})

	handler(nil)

	require.Equal(t, []string{"m1", "m2", "handler"}, calls)
}

func TestChainWithNoMiddlewareFallsBackToNext(t *testing.T) {
	var called bool

	handler := Chain()(func(ctx Context) {
		called = true
	})

	handler(nil)

	require.True(t, called)
}

func TestDefaultServiceGovernanceDefaultsMatchesMainlineTransportPreset(t *testing.T) {
	got := DefaultServiceGovernanceDefaults()
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

func TestDefaultServiceGovernanceSetMatchesMainlineCardinality(t *testing.T) {
	got := DefaultServiceGovernanceSet(nil, DefaultServiceGovernanceOptions{})
	expected := httpmiddleware.DefaultHTTPServiceGovernanceSet(nil, httpmiddleware.DefaultHTTPServiceGovernanceOptions{})
	require.Len(t, got, len(expected))
}

func TestUseDefaultServiceGovernanceForwardsMiddlewareSet(t *testing.T) {
	router := &captureRouter{}
	opts := DefaultServiceGovernanceOptions{
		MaxConcurrent: 8,
	}

	UseDefaultServiceGovernance(router, nil, opts)

	expected := DefaultServiceGovernanceSet(nil, opts)
	require.Len(t, router.used, len(expected))
}
