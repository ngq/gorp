package integration_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract/integration"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	id     int64
	alive  atomic.Bool
	closed atomic.Bool
}

func TestResilientConnection_BasicLifecycle(t *testing.T) {
	ctx := context.Background()
	var nextID int64

	opts := integration.ResilientOptions[*mockConn]{
		Dialer: func(ctx context.Context) (*mockConn, error) {
			id := atomic.AddInt64(&nextID, 1)
			c := &mockConn{id: id}
			c.alive.Store(true)
			return c, nil
		},
		Checker: func(conn *mockConn) bool {
			return conn != nil && conn.alive.Load()
		},
		Closer: func(conn *mockConn) error {
			conn.closed.Store(true)
			return nil
		},
	}

	rc, err := integration.NewResilientConnection(ctx, opts)
	require.NoError(t, err)
	require.NotNil(t, rc)

	// First Get returns initial connection
	c1, err := rc.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(1), c1.id)

	// Subsequent Get returns cached connection when healthy
	c1Again, err := rc.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, c1, c1Again)

	// Mark connection dead
	c1.alive.Store(false)

	// Next Get detects dead connection and reconnects
	c2, err := rc.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(2), c2.id)
	require.True(t, c1.closed.Load())

	// Close connection
	require.NoError(t, rc.Close())
	require.True(t, c2.closed.Load())

	// Get after Close fails
	_, err = rc.Get(ctx)
	require.ErrorIs(t, err, integration.ErrResilientConnectionClosed)
}

func TestResilientConnection_ConcurrentGetDuringReconnect(t *testing.T) {
	ctx := context.Background()
	var dialCount int64

	opts := integration.ResilientOptions[*mockConn]{
		Dialer: func(ctx context.Context) (*mockConn, error) {
			count := atomic.AddInt64(&dialCount, 1)
			if count == 2 {
				// simulate slight dial delay
				time.Sleep(10 * time.Millisecond)
			}
			c := &mockConn{id: count}
			c.alive.Store(true)
			return c, nil
		},
		Checker: func(conn *mockConn) bool {
			return conn != nil && conn.alive.Load()
		},
		InitialBackoff: 10 * time.Millisecond,
	}

	rc, err := integration.NewResilientConnection(ctx, opts)
	require.NoError(t, err)

	c1, err := rc.Get(ctx)
	require.NoError(t, err)
	c1.alive.Store(false) // break connection

	// Concurrently call Get from 50 goroutines
	var wg sync.WaitGroup
	results := make([]*mockConn, 50)
	errorsList := make([]error, 50)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errorsList[idx] = rc.Get(ctx)
		}(i)
	}
	wg.Wait()

	for i := 0; i < 50; i++ {
		require.NoError(t, errorsList[i])
		require.NotNil(t, results[i])
		require.Equal(t, int64(2), results[i].id)
	}

	require.Equal(t, int64(2), atomic.LoadInt64(&dialCount))
}

func TestResilientConnection_OnReconnectHook(t *testing.T) {
	ctx := context.Background()
	var hookCalled atomic.Int64

	opts := integration.ResilientOptions[*mockConn]{
		Dialer: func(ctx context.Context) (*mockConn, error) {
			c := &mockConn{id: 1}
			c.alive.Store(true)
			return c, nil
		},
		Checker: func(conn *mockConn) bool {
			return conn != nil && conn.alive.Load()
		},
		OnReconnect: func(ctx context.Context, conn *mockConn) error {
			hookCalled.Add(1)
			return nil
		},
		InitialBackoff: 5 * time.Millisecond,
	}

	rc, err := integration.NewResilientConnection(ctx, opts)
	require.NoError(t, err)

	rc.MarkUnhealthy()
	c, err := rc.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, c)
	require.Equal(t, int64(1), hookCalled.Load())
}

func TestResilientConnection_OnReconnectFailureRetries(t *testing.T) {
	ctx := context.Background()
	var dialAttempts atomic.Int64

	opts := integration.ResilientOptions[*mockConn]{
		Dialer: func(ctx context.Context) (*mockConn, error) {
			c := &mockConn{id: dialAttempts.Add(1)}
			c.alive.Store(true)
			return c, nil
		},
		Checker: func(conn *mockConn) bool {
			return conn != nil && conn.alive.Load()
		},
		OnReconnect: func(ctx context.Context, conn *mockConn) error {
			if conn.id == 2 {
				return errors.New("simulated hook failure")
			}
			return nil
		},
		InitialBackoff: 5 * time.Millisecond,
	}

	rc, err := integration.NewResilientConnection(ctx, opts)
	require.NoError(t, err)

	rc.MarkUnhealthy()
	c, err := rc.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(3), c.id) // id 2 failed hook and retried with id 3
}
