// Package log_test provides unit tests for context-based logger storage and retrieval.
//
// 适用场景：
// - 验证基于 context 的 logger 存取行为。
// - 防止 nil 处理和 context 字段派生语义回归。
// - 通过聚焦型测试固化 framework/log context helper 的预期行为。
package log

import (
	"context"
	"testing"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	"github.com/stretchr/testify/require"
)

// TestWithContextHandlesNilInputs verifies that WithContext handles nil context and nil logger gracefully.
//
// TestWithContextHandlesNilInputs 验证 WithContext 能正确处理 nil context 和 nil logger。
func TestWithContextHandlesNilInputs(t *testing.T) {
	ctx := WithContext(nil, nil)
	require.NotNil(t, ctx)
	_, ok := FromContext(ctx)
	require.True(t, ok)
}

// TestFromContextReturnsFalseWhenMissing verifies that FromContext returns false when no logger is stored in context.
//
// TestFromContextReturnsFalseWhenMissing 验证 FromContext 在 context 中没有 logger 时返回 false。
func TestFromContextReturnsFalseWhenMissing(t *testing.T) {
	_, ok := FromContext(context.Background())
	require.False(t, ok)
}

// TestWithContextFieldsUsesContextLogger verifies that WithContextFields uses logger from context when adding fields.
//
// TestWithContextFieldsUsesContextLogger 验证 WithContextFields 在添加字段时使用 context 中的 logger。
func TestWithContextFieldsUsesContextLogger(t *testing.T) {
	defaultStub := &loggerStub{}
	requestStub := &loggerStub{}
	SetDefault(defaultStub)
	ctx := WithContext(context.Background(), requestStub)

	withLogger, ok := WithContextFields(ctx, observabilitycontract.Field{Key: "trace_id", Value: "trace-1"}).(*loggerStub)
	require.True(t, ok)
	require.Equal(t, []observabilitycontract.Field{{Key: "trace_id", Value: "trace-1"}}, withLogger.fields)
}
