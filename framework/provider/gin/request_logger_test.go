package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ginpkg "github.com/gin-gonic/gin"
	frameworkbizlog "github.com/ngq/gorp/framework/log"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type requestLoggerStub struct {
	fields []contract.Field
}

func (l *requestLoggerStub) Debug(string, ...contract.Field) {}
func (l *requestLoggerStub) Info(string, ...contract.Field)  {}
func (l *requestLoggerStub) Warn(string, ...contract.Field)  {}
func (l *requestLoggerStub) Error(string, ...contract.Field) {}
func (l *requestLoggerStub) With(fields ...contract.Field) contract.Logger {
	copied := make([]contract.Field, len(fields))
	copy(copied, fields)
	return &requestLoggerStub{fields: copied}
}

type requestLoggerContainerStub struct {
	logger contract.Logger
}

func (s *requestLoggerContainerStub) Bind(string, contract.Factory, bool)                     {}
func (s *requestLoggerContainerStub) IsBind(key string) bool                                  { return key == contract.LogKey }
func (s *requestLoggerContainerStub) RegisterProvider(contract.ServiceProvider) error         { return nil }
func (s *requestLoggerContainerStub) RegisterProviders(...contract.ServiceProvider) error     { return nil }
func (s *requestLoggerContainerStub) MustMake(key string) any                                 { v, _ := s.Make(key); return v }
func (s *requestLoggerContainerStub) Make(key string) (any, error) {
	if key == contract.LogKey {
		return s.logger, nil
	}
	return nil, http.ErrNoLocation
}

func TestInjectRequestLoggerStoresRequestLoggerInContext(t *testing.T) {
	base := &requestLoggerStub{}
	mw := injectRequestLogger(&requestLoggerContainerStub{logger: base})

	rec := httptest.NewRecorder()
	ctx, _ := ginpkg.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set("trace_id", "trace-1")
	ctx.Set("request_id", "req-1")

	called := false
	ctx.Set("__test_next", true)
	mw(ctx)
	called = true

	require.True(t, called)
	requestLogger := frameworkbizlog.Ctx(ctx.Request.Context())
	stub, ok := requestLogger.(*requestLoggerStub)
	require.True(t, ok)
	require.Equal(t, []contract.Field{
		{Key: "trace_id", Value: "trace-1"},
		{Key: "request_id", Value: "req-1"},
	}, stub.fields)
}
