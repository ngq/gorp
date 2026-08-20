package observability_test

import (
	"testing"

	"github.com/ngq/gorp/framework/contract/observability"
	"github.com/stretchr/testify/require"
)

func TestAcquireAndReleaseFields(t *testing.T) {
	fields := observability.AcquireFields(4)
	require.Equal(t, 0, len(fields))
	require.GreaterOrEqual(t, cap(fields), 8)

	fields = append(fields, observability.FieldOf("k1", "v1"))
	fields = append(fields, observability.FieldOf("k2", 123))
	require.Equal(t, 2, len(fields))

	observability.ReleaseFields(fields)

	// Acquire again from pool
	fields2 := observability.AcquireFields(6)
	require.Equal(t, 0, len(fields2))
	require.GreaterOrEqual(t, cap(fields2), 8)
	observability.ReleaseFields(fields2)
}

func BenchmarkFieldsPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fields := observability.AcquireFields(4)
		fields = append(fields, observability.FieldOf("user_id", "12345"))
		fields = append(fields, observability.FieldOf("trace_id", "abc-xyz-123"))
		observability.ReleaseFields(fields)
	}
}

func BenchmarkFieldsHeapAlloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fields := make([]observability.Field, 0, 4)
		fields = append(fields, observability.FieldOf("user_id", "12345"))
		fields = append(fields, observability.FieldOf("trace_id", "abc-xyz-123"))
		_ = fields
	}
}
