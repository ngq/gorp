// Package cache_test provides unit tests for the cache provider.
//
// 适用场景：
// - 验证 Cache provider 的 Get / Set / Delete / Exists / TTL 等核心行为。
package cache

import (
	"context"
	"os"
	"testing"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	testinghelper "github.com/ngq/gorp/framework/testing"

	"github.com/stretchr/testify/require"
)

func TestCache_Memory_TTL_Del_MGet_Remember(t *testing.T) {
	c, cleanup := testinghelper.NewTestContainer(t)
	defer cleanup()

	// use memory driver
	_ = os.Setenv("CACHE_DRIVER", "memory")
	require.NoError(t, c.RegisterProvider(NewProvider()))

	anySvc, err := c.Make(datacontract.CacheKey)
	require.NoError(t, err)
	svc := anySvc.(datacontract.Cache)

	ctx := context.Background()

	// Set/Get + TTL
	require.NoError(t, svc.Set(ctx, "k1", "v1", 50*time.Millisecond))
	v, err := svc.Get(ctx, "k1")
	require.NoError(t, err)
	require.Equal(t, "v1", v)

	time.Sleep(70 * time.Millisecond)
	_, err = svc.Get(ctx, "k1")
	require.ErrorIs(t, err, datacontract.ErrCacheMiss)

	// Del
	require.NoError(t, svc.Set(ctx, "k2", "v2", 0))
	require.NoError(t, svc.Del(ctx, "k2"))
	_, err = svc.Get(ctx, "k2")
	require.ErrorIs(t, err, datacontract.ErrCacheMiss)

	// MGet
	require.NoError(t, svc.Set(ctx, "a", "1", 0))
	require.NoError(t, svc.Set(ctx, "b", "2", 0))
	m, err := svc.MGet(ctx, "a", "b", "c")
	require.NoError(t, err)
	require.Equal(t, map[string]string{"a": "1", "b": "2"}, m)

	// Remember
	calls := 0
	val, err := svc.Remember(ctx, "rem", 0, func(ctx context.Context) (string, error) {
		calls++
		return "computed", nil
	})
	require.NoError(t, err)
	require.Equal(t, "computed", val)
	val, err = svc.Remember(ctx, "rem", 0, func(ctx context.Context) (string, error) {
		calls++
		return "computed2", nil
	})
	require.NoError(t, err)
	require.Equal(t, "computed", val)
	require.Equal(t, 1, calls)
}
