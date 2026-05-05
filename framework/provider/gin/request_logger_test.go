package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ginpkg "github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/container"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
	"github.com/stretchr/testify/require"
)

type requestLoggerStub struct {
	fields      []observabilitycontract.Field
	infoMsg     string
	infoFields  []observabilitycontract.Field
	errorMsg    string
	errorFields []observabilitycontract.Field
}

func (l *requestLoggerStub) Debug(string, ...observabilitycontract.Field) {}
func (l *requestLoggerStub) Info(msg string, fields ...observabilitycontract.Field) {
	l.infoMsg = msg
	l.infoFields = append([]observabilitycontract.Field(nil), fields...)
}
func (l *requestLoggerStub) Warn(string, ...observabilitycontract.Field) {}
func (l *requestLoggerStub) Error(msg string, fields ...observabilitycontract.Field) {
	l.errorMsg = msg
	l.errorFields = append([]observabilitycontract.Field(nil), fields...)
}
func (l *requestLoggerStub) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	copied := make([]observabilitycontract.Field, len(fields))
	copy(copied, fields)
	return &requestLoggerStub{fields: copied}
}

type requestLoggerContainerStub struct {
	logger observabilitycontract.Logger
}

func (s *requestLoggerContainerStub) Bind(string, runtimecontract.Factory, bool) {}
func (s *requestLoggerContainerStub) IsBind(key string) bool {
	return key == observabilitycontract.LogKey
}
func (s *requestLoggerContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (s *requestLoggerContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (s *requestLoggerContainerStub) MustMake(key string) any { v, _ := s.Make(key); return v }
func (s *requestLoggerContainerStub) Make(key string) (any, error) {
	if key == observabilitycontract.LogKey {
		return s.logger, nil
	}
	return nil, http.ErrNoLocation
}

func TestLoggingMiddlewareStoresRequestLoggerInContext(t *testing.T) {
	base := &requestLoggerStub{}
	mw := LoggingMiddleware(base)

	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/users/42", nil)
	ctx.Params = ginpkg.Params{{Key: "id", Value: "42"}}
	ctx.Request = ctx.Request.WithContext(supportcontract.NewTraceIDContext(ctx.Request.Context(), "trace-1"))
	ctx.Request = ctx.Request.WithContext(supportcontract.NewRequestIDContext(ctx.Request.Context(), "req-1"))
	httpCtx := newHTTPContext(ctx)
	wrapped := mw(func(c transportcontract.HTTPContext) {
		if gc, ok := unwrapGinContext(c); ok {
			gc.FullPath()
		}
		c.Status(http.StatusNoContent)
	})
	require.NotNil(t, wrapped)
	wrapped(httpCtx)

	requestLogger := frameworkbizlog.Ctx(httpCtx.Context())
	stub, ok := requestLogger.(*requestLoggerStub)
	require.True(t, ok)
	require.Equal(t, []observabilitycontract.Field{
		{Key: "trace_id", Value: "trace-1"},
		{Key: "request_id", Value: "req-1"},
	}, stub.fields)
	require.Equal(t, "http request", stub.infoMsg)
	require.Contains(t, stub.infoFields, observabilitycontract.Field{Key: "path", Value: "/users/42"})
	require.Contains(t, stub.infoFields, observabilitycontract.Field{Key: "route", Value: ""})
	require.Contains(t, stub.infoFields, observabilitycontract.Field{Key: "status", Value: http.StatusNoContent})
	require.Contains(t, stub.infoFields, observabilitycontract.Field{Key: "request_id", Value: "req-1"})
	require.Contains(t, stub.infoFields, observabilitycontract.Field{Key: "trace_id", Value: "trace-1"})
}

func TestRecoveryMiddlewareRecoversPanic(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	httpCtx := newHTTPContext(ctx)
	httpCtx.SetContext(frameworkbizlog.WithContext(httpCtx.Context(), &requestLoggerStub{}))

	wrapped := RecoveryMiddleware()(func(transportcontract.HTTPContext) {
		panic("boom")
	})
	require.NotNil(t, wrapped)
	wrapped(httpCtx)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "internal server error")
}

type recoveryResponderStub struct {
	message string
}

func (r *recoveryResponderStub) Success(transportcontract.HTTPContext, any)                    {}
func (r *recoveryResponderStub) SuccessWithMessage(transportcontract.HTTPContext, string, any) {}
func (r *recoveryResponderStub) SuccessWithStatus(transportcontract.HTTPContext, int, any)     {}
func (r *recoveryResponderStub) Error(transportcontract.HTTPContext, error)                    {}
func (r *recoveryResponderStub) BadRequest(transportcontract.HTTPContext, string)              {}
func (r *recoveryResponderStub) InternalError(c transportcontract.HTTPContext, message string) {
	r.message = message
	c.Status(http.StatusTeapot)
}

func TestRecoveryMiddlewareUsesCustomResponder(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	c := container.New()
	custom := &recoveryResponderStub{}
	c.Bind(transportcontract.HTTPResponderKey, func(runtimecontract.Container) (any, error) {
		return custom, nil
	}, true)

	httpCtx := newHTTPContext(ctx)
	httpCtx.SetContext(supportcontract.NewContainerContext(httpCtx.Context(), c))
	httpCtx.SetContext(frameworkbizlog.WithContext(httpCtx.Context(), &requestLoggerStub{}))

	wrapped := RecoveryMiddleware()(func(transportcontract.HTTPContext) {
		panic("boom")
	})
	require.NotNil(t, wrapped)
	wrapped(httpCtx)

	require.Equal(t, "internal server error", custom.message)
	require.Equal(t, http.StatusTeapot, httpCtx.ResponseStatus())
}
