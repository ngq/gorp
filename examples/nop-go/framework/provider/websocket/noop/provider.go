// Package noop provides a no-op WebSocket service implementation for gorp.
//
// noop 包提供 gorp 的 WebSocket 服务的空实现。
// 所有方法返回错误或零值，用于不需要 WebSocket 能力的场景。
package noop

import (
	"context"
	"errors"
	"net/http"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider registers a no-op WebSocket service.
//
// Provider 注册空 WebSocket 服务。
type Provider struct{}

// NewProvider creates a new no-op WebSocket provider.
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
func (p *Provider) Name() string { return "websocket.noop" }

// IsDefer indicates this provider should defer loading.
func (p *Provider) IsDefer() bool { return true }

// Provides returns the capability keys this provider exposes.
func (p *Provider) Provides() []string { return []string{transportcontract.WebSocketKey} }

// DependsOn returns empty — no dependencies.
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op WebSocket factory to the container.
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.WebSocketKey, func(c runtimecontract.Container) (any, error) {
		return &noopWebSocketServer{}, nil
	}, true)
	return nil
}

// Boot initializes the provider. No additional startup logic required.
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopWebSocketServer implements transportcontract.WebSocketServer with no-op methods.
type noopWebSocketServer struct{}

var _ transportcontract.WebSocketServer = (*noopWebSocketServer)(nil)

func (n *noopWebSocketServer) Upgrade(http.ResponseWriter, *http.Request, transportcontract.WebSocketHandler) (transportcontract.WebSocketConn, error) {
	return nil, errors.New("websocket: not configured (noop provider)")
}

func (n *noopWebSocketServer) NewBroadcaster() transportcontract.WebSocketBroadcaster {
	return &noopBroadcaster{}
}

func (n *noopWebSocketServer) Connections() []transportcontract.WebSocketConn {
	return nil
}

func (n *noopWebSocketServer) Count() int { return 0 }

func (n *noopWebSocketServer) Shutdown(context.Context) error { return nil }

// noopConn implements transportcontract.WebSocketConn with no-op methods.
type noopConn struct{}

var _ transportcontract.WebSocketConn = (*noopConn)(nil)

func (c *noopConn) WriteString(string) error   { return errors.New("websocket: noop connection") }
func (c *noopConn) WriteBinary([]byte) error   { return errors.New("websocket: noop connection") }
func (c *noopConn) Close(int, string) error    { return nil }
func (c *noopConn) Context() context.Context   { return context.Background() }
func (c *noopConn) SetContext(context.Context) {}
func (c *noopConn) RemoteAddr() string         { return "" }
func (c *noopConn) LocalAddr() string          { return "" }
func (c *noopConn) RawConn() any               { return nil }

// noopBroadcaster implements transportcontract.WebSocketBroadcaster with no-op methods.
type noopBroadcaster struct{}

var _ transportcontract.WebSocketBroadcaster = (*noopBroadcaster)(nil)

func (b *noopBroadcaster) BroadcastString(string) error { return nil }
func (b *noopBroadcaster) BroadcastBinary([]byte) error { return nil }
func (b *noopBroadcaster) BroadcastStringExcept(string, transportcontract.WebSocketConn) error {
	return nil
}
func (b *noopBroadcaster) BroadcastBinaryExcept([]byte, transportcontract.WebSocketConn) error {
	return nil
}
