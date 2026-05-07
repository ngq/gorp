// Application scenarios:
// - Verify the top-level log package export helpers and context logger behavior.
// - Protect field builders and default-logger wrappers from regressions.
// - Document expected use through focused export tests.
//
// 适用场景：
// - 验证顶层 log 包的导出 helper 和 context logger 行为。
// - 防止字段构造器和默认 logger 包装器回归。
// - 通过聚焦型导出测试固化预期用法。
package log

import (
	"context"
	"testing"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	frameworklog "github.com/ngq/gorp/framework/log"
	"github.com/stretchr/testify/require"
)

type exportLoggerStub struct {
	fields []observabilitycontract.Field
}

func (l *exportLoggerStub) Debug(string, ...observabilitycontract.Field) {}
func (l *exportLoggerStub) Info(string, ...observabilitycontract.Field)  {}
func (l *exportLoggerStub) Warn(string, ...observabilitycontract.Field)  {}
func (l *exportLoggerStub) Error(string, ...observabilitycontract.Field) {}
func (l *exportLoggerStub) With(fields ...observabilitycontract.Field) observabilitycontract.Logger {
	copied := make([]observabilitycontract.Field, len(fields))
	copy(copied, fields)
	return &exportLoggerStub{fields: copied}
}

func TestExportedCtxAndWith(t *testing.T) {
	stub := &exportLoggerStub{}
	frameworklog.SetDefault(stub)

	ctx := frameworklog.WithContext(context.Background(), stub)
	require.Same(t, stub, Ctx(ctx))

	withLogger, ok := With(observabilitycontract.Field{Key: "trace_id", Value: "trace-1"}).(*exportLoggerStub)
	require.True(t, ok)
	require.Equal(t, []observabilitycontract.Field{{Key: "trace_id", Value: "trace-1"}}, withLogger.fields)
}

func TestExportedHelpers(t *testing.T) {
	require.Equal(t, observabilitycontract.Field{Key: "name", Value: "alice"}, String("name", "alice"))
	require.Equal(t, observabilitycontract.Field{Key: "count", Value: 2}, Int("count", 2))
	require.Equal(t, observabilitycontract.Field{Key: "id", Value: int64(9)}, Int64("id", 9))
	require.Equal(t, observabilitycontract.Field{Key: "ok", Value: true}, Bool("ok", true))
	require.Equal(t, observabilitycontract.Field{Key: "value", Value: 1.5}, Any("value", 1.5))

	err := context.DeadlineExceeded
	require.Equal(t, observabilitycontract.Field{Key: "err", Value: err}, Err(err))

	stub := &exportLoggerStub{}
	frameworklog.SetDefault(stub)
	ctx := frameworklog.WithContext(context.Background(), stub)
	withLogger, ok := WithContextFields(ctx, String("trace_id", "trace-1")).(*exportLoggerStub)
	require.True(t, ok)
	require.Equal(t, []observabilitycontract.Field{{Key: "trace_id", Value: "trace-1"}}, withLogger.fields)
}
