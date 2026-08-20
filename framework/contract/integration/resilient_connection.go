// Package integration provides integration capability contracts and helpers for the gorp framework.
package integration

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	// ErrResilientConnectionClosed is returned when operations are performed on a closed connection.
	ErrResilientConnectionClosed = errors.New("resilient connection: closed")
	// ErrNoDialerConfigured is returned when the Dialer function is nil.
	ErrNoDialerConfigured = errors.New("resilient connection: dialer is required")
)

// ResilientOptions defines configuration and hooks for ResilientConnection.
type ResilientOptions[T any] struct {
	// Dialer is the mandatory factory function to establish a new connection.
	Dialer func(ctx context.Context) (T, error)
	// Checker tests whether the existing connection is healthy. If nil, conn != nil is assumed alive.
	Checker func(conn T) bool
	// OnReconnect is invoked immediately after a successful reconnection to restore state.
	OnReconnect func(ctx context.Context, conn T) error
	// Closer releases the underlying connection resource upon Close or replacement.
	Closer func(conn T) error

	// InitialBackoff is the starting delay for reconnect attempts (default: 100ms).
	InitialBackoff time.Duration
	// MaxBackoff is the upper bound for reconnect delay (default: 5s).
	MaxBackoff time.Duration
}

// ResilientConnection manages the lifecycle of a remote broker/service connection,
// providing automatic thread-safe reconnection, health checking, and post-reconnect recovery hooks.
type ResilientConnection[T any] struct {
	opts ResilientOptions[T]

	mu        sync.RWMutex
	conn      T
	hasConn   bool
	unhealthy bool
	closed    bool

	reconnectMu sync.Mutex
}

// NewResilientConnection creates a ResilientConnection and establishes the initial connection.
func NewResilientConnection[T any](ctx context.Context, opts ResilientOptions[T]) (*ResilientConnection[T], error) {
	if opts.Dialer == nil {
		return nil, ErrNoDialerConfigured
	}
	if opts.InitialBackoff <= 0 {
		opts.InitialBackoff = 100 * time.Millisecond
	}
	if opts.MaxBackoff <= 0 {
		opts.MaxBackoff = 5 * time.Second
	}

	rc := &ResilientConnection[T]{
		opts: opts,
	}

	conn, err := opts.Dialer(ctx)
	if err != nil {
		return nil, fmt.Errorf("initial dial failed: %w", err)
	}
	rc.conn = conn
	rc.hasConn = true

	return rc, nil
}

// Get returns the active connection, automatically detecting broken connections and
// reconnecting with exponential backoff and jitter if needed.
func (rc *ResilientConnection[T]) Get(ctx context.Context) (T, error) {
	rc.mu.RLock()
	if rc.closed {
		rc.mu.RUnlock()
		var zero T
		return zero, ErrResilientConnectionClosed
	}

	if rc.hasConn && !rc.unhealthy {
		if rc.opts.Checker == nil || rc.opts.Checker(rc.conn) {
			conn := rc.conn
			rc.mu.RUnlock()
			return conn, nil
		}
	}
	rc.mu.RUnlock()

	return rc.reconnect(ctx)
}

// MarkUnhealthy signals that the current connection has encountered a transport error,
// forcing the next Get call to re-establish the connection.
func (rc *ResilientConnection[T]) MarkUnhealthy() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.unhealthy = true
}

// Close gracefully closes the underlying connection and stops future reconnects.
func (rc *ResilientConnection[T]) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.closed {
		return nil
	}
	rc.closed = true

	var err error
	if rc.hasConn && rc.opts.Closer != nil {
		err = rc.opts.Closer(rc.conn)
	}
	var zero T
	rc.conn = zero
	rc.hasConn = false
	return err
}

func (rc *ResilientConnection[T]) reconnect(ctx context.Context) (T, error) {
	rc.reconnectMu.Lock()
	defer rc.reconnectMu.Unlock()

	// Double-check under write lock in case another goroutine already reconnected
	rc.mu.RLock()
	if rc.closed {
		rc.mu.RUnlock()
		var zero T
		return zero, ErrResilientConnectionClosed
	}
	if rc.hasConn && !rc.unhealthy {
		if rc.opts.Checker == nil || rc.opts.Checker(rc.conn) {
			conn := rc.conn
			rc.mu.RUnlock()
			return conn, nil
		}
	}
	rc.mu.RUnlock()

	// Close old broken connection if closer is present
	rc.mu.Lock()
	if rc.hasConn && rc.opts.Closer != nil {
		_ = rc.opts.Closer(rc.conn)
	}
	var zero T
	rc.conn = zero
	rc.hasConn = false
	rc.mu.Unlock()

	backoff := rc.opts.InitialBackoff
	for {
		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		default:
		}

		conn, err := rc.opts.Dialer(ctx)
		if err == nil {
			if rc.opts.OnReconnect != nil {
				if hookErr := rc.opts.OnReconnect(ctx, conn); hookErr != nil {
					if rc.opts.Closer != nil {
						_ = rc.opts.Closer(conn)
					}
					err = fmt.Errorf("on_reconnect hook failed: %w", hookErr)
				}
			}
		}

		if err == nil {
			rc.mu.Lock()
			if rc.closed {
				rc.mu.Unlock()
				if rc.opts.Closer != nil {
					_ = rc.opts.Closer(conn)
				}
				return zero, ErrResilientConnectionClosed
			}
			rc.conn = conn
			rc.hasConn = true
			rc.unhealthy = false
			rc.mu.Unlock()
			return conn, nil
		}

		// Calculate backoff with jitter
		jitter := time.Duration(rand.Float64() * float64(backoff) * 0.2)
		sleepDuration := backoff + jitter

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(sleepDuration):
		}

		backoff *= 2
		if backoff > rc.opts.MaxBackoff {
			backoff = rc.opts.MaxBackoff
		}
	}
}
