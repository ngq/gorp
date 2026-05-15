// Package noop provides a no-op WebSocket service implementation for gorp.
//
// noop 包提供 gorp 的 WebSocket 服务的空实现。
// 所有方法返回错误或零值，用于不需要 WebSocket 能力的场景。
package noop

import (
	"errors"
	"net"
	"net/http"
	"time"

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
		return &noopWebSocket{}, nil
	}, true)
	return nil
}

// Boot initializes the provider. No additional startup logic required.
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// noopWebSocket implements transportcontract.WebSocket with no-op methods.
type noopWebSocket struct{}

var _ transportcontract.WebSocket = (*noopWebSocket)(nil)

func (n *noopWebSocket) Upgrade(http.ResponseWriter, *http.Request, transportcontract.WebSocketHandler) (transportcontract.WebSocketConn, error) {
	return nil, errors.New("websocket: not configured (noop provider)")
}

func (n *noopWebSocket) NewClient(transportcontract.WebSocketHandler, *transportcontract.WebSocketClientOptions) (transportcontract.WebSocketConn, error) {
	return nil, errors.New("websocket: not configured (noop provider)")
}

// noopConn implements transportcontract.WebSocketConn with no-op methods.
type noopConn struct{}

var _ transportcontract.WebSocketConn = (*noopConn)(nil)

func (c *noopConn) WriteText([]byte) error       { return errors.New("websocket: noop connection") }
func (c *noopConn) WriteBinary([]byte) error      { return errors.New("websocket: noop connection") }
func (c *noopConn) WriteClose(int, string) error  { return errors.New("websocket: noop connection") }
func (c *noopConn) SetDeadline(time.Time) error    { return errors.New("websocket: noop connection") }
func (c *noopConn) RemoteAddr() net.Addr          { return nil }
func (c *noopConn) Session() transportcontract.WebSocketSession { return &noopSession{} }
func (c *noopConn) Close() error                   { return nil }

// noopSession implements transportcontract.WebSocketSession with no-op methods.
type noopSession struct{}

var _ transportcontract.WebSocketSession = (*noopSession)(nil)

func (s *noopSession) Load(any) (any, bool) { return nil, false }
func (s *noopSession) Store(any, any)       {}
func (s *noopSession) Delete(any)           {}
