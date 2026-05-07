// Application scenarios:
// - Verify default logger behavior and field helper output.
// - Protect context fallback and derived logger behavior from regressions.
// - Document expected framework/log utility semantics through focused tests.
//
// 适用场景：
// - 验证默认 logger 行为和字段 helper 输出。
// - 防止 context 回退和派生 logger 行为回归。
// - 通过聚焦型测试固化 framework/log 工具层的预期语义。
package log

import (
	"context"
	"testing"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	"github.com/stretchr/testify/require"
)

type loggerStub struct {
	fields []observabilitycontract.Field
}

func (l *loggerStub) Debug(string, ...observabilitycontract.Field) {}
func (l *loggerStub) Info(string, ...observabilitycontract.Field)  {}
func (l *loggerStub) Warn(string, ...observabilitycontract.Field)  {}
func (l *loggerStub) Error(string, ...observabilitycontract.Field) {}
func (l *loggerStub) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	copied := make([]observabilitycontract.Field, len(fields))
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
		observabilitycontract.Field{Key: "trace_id", Value: "trace-1"},
		observabilitycontract.Field{Key: "request_id", Value: "req-1"},
	).(*loggerStub)
	require.True(t, ok)
	require.Equal(t, []observabilitycontract.Field{
		{Key: "trace_id", Value: "trace-1"},
		{Key: "request_id", Value: "req-1"},
	}, withLogger.fields)
}

func TestFieldHelpers(t *testing.T) {
	require.Equal(t, observabilitycontract.Field{Key: "name", Value: "alice"}, String("name", "alice"))
	require.Equal(t, observabilitycontract.Field{Key: "age", Value: 18}, Int("age", 18))
	require.Equal(t, observabilitycontract.Field{Key: "id", Value: int64(42)}, Int64("id", 42))
	require.Equal(t, observabilitycontract.Field{Key: "ok", Value: true}, Bool("ok", true))
	require.Equal(t, observabilitycontract.Field{Key: "payload", Value: map[string]int{"a": 1}}, Any("payload", map[string]int{"a": 1}))

	err := context.Canceled
	require.Equal(t, observabilitycontract.Field{Key: "err", Value: err}, Err(err))
}
