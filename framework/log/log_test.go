package log

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type loggerStub struct {
	fields []contract.Field
}

func (l *loggerStub) Debug(string, ...contract.Field) {}
func (l *loggerStub) Info(string, ...contract.Field)  {}
func (l *loggerStub) Warn(string, ...contract.Field)  {}
func (l *loggerStub) Error(string, ...contract.Field) {}
func (l *loggerStub) With(fields ...contract.Field) contract.Logger {
	copied := make([]contract.Field, len(fields))
	copy(copied, fields)
	return &loggerStub{fields: copied}
}

func TestDefaultReturnsNoopWhenUnset(t *testing.T) {
	SetDefault(nil)
	require.NotNil(t, Default())
}

func TestSetDefaultStoresLogger(t *testing.T) {
	stub := &loggerStub{}
	SetDefault(stub)
	require.Same(t, stub, Default())
}

func TestCtxFallsBackToDefault(t *testing.T) {
	stub := &loggerStub{}
	SetDefault(stub)
	require.Same(t, stub, Ctx(context.Background()))
}

func TestCtxReturnsContextLogger(t *testing.T) {
	defaultStub := &loggerStub{}
	requestStub := &loggerStub{}
	SetDefault(defaultStub)
	ctx := WithContext(context.Background(), requestStub)
	require.Same(t, requestStub, Ctx(ctx))
}

func TestWithUsesDefaultLogger(t *testing.T) {
	stub := &loggerStub{}
	SetDefault(stub)
	withLogger, ok := With(
		contract.Field{Key: "trace_id", Value: "trace-1"},
		contract.Field{Key: "request_id", Value: "req-1"},
	).(*loggerStub)
	require.True(t, ok)
	require.Equal(t, []contract.Field{
		{Key: "trace_id", Value: "trace-1"},
		{Key: "request_id", Value: "req-1"},
	}, withLogger.fields)
}

func TestFieldHelpers(t *testing.T) {
	require.Equal(t, contract.Field{Key: "name", Value: "alice"}, String("name", "alice"))
	require.Equal(t, contract.Field{Key: "age", Value: 18}, Int("age", 18))
	require.Equal(t, contract.Field{Key: "id", Value: int64(42)}, Int64("id", 42))
	require.Equal(t, contract.Field{Key: "ok", Value: true}, Bool("ok", true))
	require.Equal(t, contract.Field{Key: "payload", Value: map[string]int{"a": 1}}, Any("payload", map[string]int{"a": 1}))

	err := context.Canceled
	require.Equal(t, contract.Field{Key: "err", Value: err}, Err(err))
}
