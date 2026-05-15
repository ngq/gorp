// Package gws provides a WebSocket service implementation based on the lxzan/gws library.
//
// gws 包提供基于 lxzan/gws 库的 WebSocket 服务实现。
// 支持通过 Upgrader 将 HTTP 连接升级为 WebSocket，以及创建 WebSocket 客户端。
//
// 使用示例：
//
//	ws, err := rt.Container.Make(transportcontract.WebSocketKey)
//	wsService := ws.(transportcontract.WebSocket)
//	conn, err := wsService.Upgrade(c.Writer, c.Request, &MyHandler{})
//
// 配置路径：websocket.*
package gws

import (
	"errors"
	"net/http"

	gwslib "github.com/lxzan/gws"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider registers WebSocket service backed by gws.
//
// Provider 注册基于 gws 的 WebSocket 服务。
type Provider struct{}

// NewProvider creates a new gws WebSocket provider.
func NewProvider() *Provider { return &Provider{} }

// Name returns provider name for identification.
func (p *Provider) Name() string { return "websocket.gws" }

// IsDefer indicates this provider should defer loading until needed.
func (p *Provider) IsDefer() bool { return true }

// Provides returns the capability keys this provider exposes.
func (p *Provider) Provides() []string { return []string{transportcontract.WebSocketKey} }

// DependsOn returns the keys this provider depends on.
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Register binds the WebSocket factory to the container.
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.WebSocketKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getWebSocketConfig(c)
		if err != nil {
			return nil, err
		}
		return newService(cfg), nil
	}, true)
	return nil
}

// Boot initializes the provider. No additional startup logic required.
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// config holds the WebSocket configuration read from the container.
type config struct {
	ReadBufferSize     int
	WriteBufferSize    int
	ParallelEnabled    bool
	CompressionEnabled bool
}

func getWebSocketConfig(c runtimecontract.Container) (*config, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("websocket.gws: invalid config service")
	}
	rc := &config{}
	if v := cfg.GetInt("websocket.read_buffer_size"); v > 0 {
		rc.ReadBufferSize = v
	}
	if v := cfg.GetInt("websocket.write_buffer_size"); v > 0 {
		rc.WriteBufferSize = v
	}
	rc.ParallelEnabled = cfg.GetBool("websocket.parallel_enabled")
	rc.CompressionEnabled = cfg.GetBool("websocket.compression_enabled")
	return rc, nil
}

// service implements transportcontract.WebSocket using gws.
type service struct {
	cfg *config
}

func newService(cfg *config) *service {
	return &service{cfg: cfg}
}

// buildServerOption creates a gws.ServerOption from our config.
func (s *service) buildServerOption() *gwslib.ServerOption {
	opts := &gwslib.ServerOption{
		ReadBufferSize:  s.cfg.ReadBufferSize,
		ParallelEnabled: s.cfg.ParallelEnabled,
		Recovery:        gwslib.Recovery,
	}
	if s.cfg.CompressionEnabled {
		opts.PermessageDeflate = gwslib.PermessageDeflate{Enabled: true}
	}
	return opts
}

// buildClientOption creates a gws.ClientOption from our config and the given options.
func (s *service) buildClientOption(opts *transportcontract.WebSocketClientOptions) *gwslib.ClientOption {
	co := &gwslib.ClientOption{
		ReadBufferSize:  s.cfg.ReadBufferSize,
		ParallelEnabled: s.cfg.ParallelEnabled,
		Recovery:        gwslib.Recovery,
	}
	if opts != nil {
		co.Addr = opts.Addr
		if opts.TLSInsecureSkipVerify {
			co.TlsConfig = newTLSInsecureConfig()
		}
		if opts.CompressionEnabled {
			co.PermessageDeflate = gwslib.PermessageDeflate{Enabled: true}
		}
	}
	return co
}

// Upgrade upgrades an HTTP connection to WebSocket with the given handler.
func (s *service) Upgrade(w http.ResponseWriter, r *http.Request, handler transportcontract.WebSocketHandler) (transportcontract.WebSocketConn, error) {
	adapter := &eventAdapter{handler: handler}
	upgrader := gwslib.NewUpgrader(adapter, s.buildServerOption())
	conn, err := upgrader.Upgrade(w, r)
	if err != nil {
		return nil, err
	}
	wsConn := &connWrapper{conn: conn}
	adapter.conn = wsConn
	go conn.ReadLoop()
	return wsConn, nil
}

// NewClient connects to a WebSocket server with the given handler and options.
func (s *service) NewClient(handler transportcontract.WebSocketHandler, opts *transportcontract.WebSocketClientOptions) (transportcontract.WebSocketConn, error) {
	adapter := &eventAdapter{handler: handler}
	conn, _, err := gwslib.NewClient(adapter, s.buildClientOption(opts))
	if err != nil {
		return nil, err
	}
	wsConn := &connWrapper{conn: conn}
	adapter.conn = wsConn
	go conn.ReadLoop()
	return wsConn, nil
}
