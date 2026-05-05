package gorp

import (
	"context"
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

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
