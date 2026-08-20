package transport_test

import (
	"testing"

	"github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestMetadataPool(t *testing.T) {
	md := transport.AcquireMetadata()
	require.NotNil(t, md)

	md.Set("x-trace-id", "123456")
	md.Set("Authorization", "Bearer token")

	require.Equal(t, "123456", md.Get("X-Trace-Id"))
	require.Equal(t, "Bearer token", md.Get("authorization"))

	transport.ReleaseMetadata(md)

	// Acquire again should yield clean metadata
	md2 := transport.AcquireMetadata()
	require.Equal(t, "", md2.Get("x-trace-id"))
	require.Equal(t, 0, len(md2.ToMap()))
	transport.ReleaseMetadata(md2)
}

func BenchmarkMetadataPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		md := transport.AcquireMetadata()
		md.Set("x-trace-id", "abc-123")
		md.Set("x-request-id", "req-456")
		_ = md.Get("x-trace-id")
		transport.ReleaseMetadata(md)
	}
}

func BenchmarkMetadataNew(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		md := transport.NewMetadata()
		md.Set("x-trace-id", "abc-123")
		md.Set("x-request-id", "req-456")
		_ = md.Get("x-trace-id")
	}
}
