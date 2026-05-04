package gorp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ngq/gorp/framework/container"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
	"github.com/stretchr/testify/require"
)

type responseCapture struct {
	status int
	body   any
	header http.Header
}

type responderSpy struct {
	calls []string
}

type customResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Result  any    `json:"result,omitempty"`
}

type customResponder struct{}

func (s *responderSpy) Success(transportcontract.HTTPContext, any) {
	s.calls = append(s.calls, "success")
}

func (s *responderSpy) SuccessWithMessage(transportcontract.HTTPContext, string, any) {
	s.calls = append(s.calls, "success_with_message")
}

func (s *responderSpy) SuccessWithStatus(transportcontract.HTTPContext, int, any) {
	s.calls = append(s.calls, "success_with_status")
}

func (s *responderSpy) Error(transportcontract.HTTPContext, error) {
	s.calls = append(s.calls, "error")
}

func (s *responderSpy) BadRequest(transportcontract.HTTPContext, string) {
	s.calls = append(s.calls, "bad_request")
}

func (s *responderSpy) InternalError(transportcontract.HTTPContext, string) {
	s.calls = append(s.calls, "internal_error")
}

func (customResponder) Success(c transportcontract.HTTPContext, data any) {
	c.JSON(http.StatusOK, customResponse{
		Success: true,
		Message: "ok",
		Result:  data,
	})
}

func (customResponder) SuccessWithMessage(c transportcontract.HTTPContext, message string, data any) {
	c.JSON(http.StatusOK, customResponse{
		Success: true,
		Message: message,
		Result:  data,
	})
}

func (customResponder) SuccessWithStatus(c transportcontract.HTTPContext, status int, data any) {
	c.JSON(status, customResponse{
		Success: true,
		Message: "ok",
		Result:  data,
	})
}

func (customResponder) Error(c transportcontract.HTTPContext, err error) {
	c.JSON(http.StatusBadRequest, customResponse{
		Success: false,
		Message: err.Error(),
	})
}

func (customResponder) BadRequest(c transportcontract.HTTPContext, message string) {
	c.JSON(http.StatusBadRequest, customResponse{
		Success: false,
		Message: message,
	})
}

func (customResponder) InternalError(c transportcontract.HTTPContext, message string) {
	c.JSON(http.StatusInternalServerError, customResponse{
		Success: false,
		Message: message,
	})
}

func TestSuccessFallsBackToDefaultResponder(t *testing.T) {
	ctx, captured := newResponseTestContext(t, context.Background())

	Success(ctx, map[string]any{"ok": true})

	require.Equal(t, http.StatusOK, captured.status)
	resp, ok := captured.body.(ginprovider.Response)
	require.True(t, ok)
	require.Equal(t, 0, resp.Code)
	require.Equal(t, "success", resp.Message)
	require.Equal(t, map[string]any{"ok": true}, resp.Data)
}

func TestErrorUsesBusinessResponderFromContainer(t *testing.T) {
	c := container.New()
	spy := &responderSpy{}
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return spy, nil
	}, true)

	ctx, _ := newResponseTestContext(t, supportcontract.NewContainerContext(context.Background(), c))

	Error(ctx, errors.New("boom"))

	require.Equal(t, []string{"error"}, spy.calls)
}

func TestSuccessWithMessageUsesBusinessResponderFromContainer(t *testing.T) {
	c := container.New()
	spy := &responderSpy{}
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return spy, nil
	}, true)

	ctx, _ := newResponseTestContext(t, supportcontract.NewContainerContext(context.Background(), c))

	SuccessWithMessage(ctx, "done", map[string]any{"id": 1})

	require.Equal(t, []string{"success_with_message"}, spy.calls)
}

func TestSuccessCanUseCustomBusinessResponseShape(t *testing.T) {
	c := container.New()
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return customResponder{}, nil
	}, true)

	ctx, captured := newResponseTestContext(t, supportcontract.NewContainerContext(context.Background(), c))

	Success(ctx, map[string]any{"id": 7})

	require.Equal(t, http.StatusOK, captured.status)
	resp, ok := captured.body.(customResponse)
	require.True(t, ok)
	require.True(t, resp.Success)
	require.Equal(t, "ok", resp.Message)
	require.Equal(t, map[string]any{"id": 7}, resp.Result)
}

func newResponseTestContext(t *testing.T, ctx context.Context) (HTTPContext, *responseCapture) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
	captured := &responseCapture{header: make(http.Header)}
	httpCtx := transportcontract.NewDefaultHTTPContext(req.Context(), req)
	httpCtx.SetHeaderFuncs(func(key string) string {
		return captured.header.Get(key)
	}, func(key, value string) {
		captured.header.Set(key, value)
	})
	httpCtx.SetResponseFuncs(func(status int, body any) {
		captured.status = status
		captured.body = body
	}, func(status int, body string) {
		captured.status = status
		captured.body = body
	}, func(status int, body any) {
		captured.status = status
		captured.body = body
	}, func(status int, contentType string, body []byte) {
		captured.status = status
		captured.header.Set("Content-Type", contentType)
		captured.body = body
	}, func(status int, location string) {
		captured.status = status
		captured.header.Set("Location", location)
	}, func(status int) {
		captured.status = status
	}, func() int {
		return captured.status
	})
	return httpCtx, captured
}
