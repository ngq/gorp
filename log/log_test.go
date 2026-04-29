package log

import (
	"context"
	"testing"

	frameworklog "github.com/ngq/gorp/framework/log"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

type exportLoggerStub struct {
	fields []contract.Field
}

func (l *exportLoggerStub) Debug(string, ...contract.Field) {}
func (l *exportLoggerStub) Info(string, ...contract.Field)  {}
func (l *exportLoggerStub) Warn(string, ...contract.Field)  {}
func (l *exportLoggerStub) Error(string, ...contract.Field) {}
func (l *exportLoggerStub) With(fields ...contract.Field) contract.Logger {
	copied := make([]contract.Field, len(fields))
	copy(copied, fields)
	return &exportLoggerStub{fields: copied}
}

func TestExportedCtxAndWith(t *testing.T) {
	stub := &exportLoggerStub{}
	frameworklog.SetDefault(stub)

	ctx := frameworklog.WithContext(context.Background(), stub)
	require.Same(t, stub, Ctx(ctx))

	withLogger, ok := With(contract.Field{Key: "trace_id", Value: "trace-1"}).(*exportLoggerStub)
	require.True(t, ok)
	require.Equal(t, []contract.Field{{Key: "trace_id", Value: "trace-1"}}, withLogger.fields)
}

func TestExportedHelpers(t *testing.T) {
	require.Equal(t, contract.Field{Key: "name", Value: "alice"}, String("name", "alice"))
	require.Equal(t, contract.Field{Key: "count", Value: 2}, Int("count", 2))
	require.Equal(t, contract.Field{Key: "id", Value: int64(9)}, Int64("id", 9))
	require.Equal(t, contract.Field{Key: "ok", Value: true}, Bool("ok", true))
	require.Equal(t, contract.Field{Key: "value", Value: 1.5}, Any("value", 1.5))

	err := context.DeadlineExceeded
	require.Equal(t, contract.Field{Key: "err", Value: err}, Err(err))

	stub := &exportLoggerStub{}
	frameworklog.SetDefault(stub)
	ctx := frameworklog.WithContext(context.Background(), stub)
	withLogger, ok := WithContextFields(ctx, String("trace_id", "trace-1")).(*exportLoggerStub)
	require.True(t, ok)
	require.Equal(t, []contract.Field{{Key: "trace_id", Value: "trace-1"}}, withLogger.fields)
}
