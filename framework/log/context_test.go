package log

import (
	"context"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestWithContextHandlesNilInputs(t *testing.T) {
	ctx := WithContext(nil, nil)
	require.NotNil(t, ctx)
	_, ok := FromContext(ctx)
	require.True(t, ok)
}

func TestFromContextReturnsFalseWhenMissing(t *testing.T) {
	_, ok := FromContext(context.Background())
	require.False(t, ok)
}

func TestWithContextFieldsUsesContextLogger(t *testing.T) {
	defaultStub := &loggerStub{}
	requestStub := &loggerStub{}
	SetDefault(defaultStub)
	ctx := WithContext(context.Background(), requestStub)

	withLogger, ok := WithContextFields(ctx, contract.Field{Key: "trace_id", Value: "trace-1"}).(*loggerStub)
	require.True(t, ok)
	require.Equal(t, []contract.Field{{Key: "trace_id", Value: "trace-1"}}, withLogger.fields)
}
