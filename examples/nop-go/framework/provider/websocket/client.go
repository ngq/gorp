// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements WebSocket client using gws library.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件使用 gws 库实现 WebSocket 客户端。
package websocket

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/lxzan/gws"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Client implements WebSocketClient interface using gws.
//
// Client 使用 gws 实现 WebSocketClient 接口。
type Client struct {
	socket *gws.Conn
	conn   *connWrapper
}

// connWrapper wraps gws.Conn to implement WebSocketConn.
//
// connWrapper 包装 gws.Conn 以实现 WebSocketConn。
type connWrapper struct {
	socket *gws.Conn
	ctx    context.Context
}

// NewClient creates a new WebSocket client connection.
//
// NewClient 创建新的 WebSocket 客户端连接。
func NewClient(handler transportcontract.WebSocketHandler, config *transportcontract.WebSocketClientConfig) (transportcontract.WebSocketClient, *http.Response, error) {
	if config == nil {
		config = &transportcontract.WebSocketClientConfig{}
	}

	opt := buildClientOption(config)

	// Create adapter for client events
	// 为客户端事件创建适配器
	adapter := &clientHandlerAdapter{handler: handler}

	socket, resp, err := gws.NewClient(adapter, opt)
	if err != nil {
		return nil, resp, err
	}

	client := &Client{
		socket: socket,
		conn:   &connWrapper{socket: socket},
	}

	adapter.client = client

	// Start read loop in a separate goroutine
	// 在单独的 goroutine 中启动读循环
	go socket.ReadLoop()

	return client, resp, nil
}

// buildClientOption builds gws.ClientOption from WebSocketClientConfig.
//
// buildClientOption 从 WebSocketClientConfig 构建 gws.ClientOption。
func buildClientOption(config *transportcontract.WebSocketClientConfig) *gws.ClientOption {
	opt := &gws.ClientOption{
		Addr:          config.URL,
		RequestHeader: config.RequestHeader,
	}

	if config.HandshakeTimeout > 0 {
		opt.HandshakeTimeout = time.Duration(config.HandshakeTimeout) * time.Second
	}

	if config.ReadBufferSize > 0 {
		opt.ReadBufferSize = config.ReadBufferSize
	}

	if config.MaxMessageSize > 0 {
		opt.ReadMaxPayloadSize = int(config.MaxMessageSize)
	}

	if config.EnableCompression {
		level := config.CompressionLevel
		if level < 1 || level > 9 {
			level = 6
		}
		opt.PermessageDeflate = gws.PermessageDeflate{
			Enabled: true,
			Level:   level,
		}
	}

	return opt
}

// Conn returns the underlying connection.
//
// Conn 返回底层连接。
func (c *Client) Conn() transportcontract.WebSocketConn {
	return c.conn
}

// Close closes the client connection.
//
// Close 关闭客户端连接。
func (c *Client) Close() error {
	return c.socket.WriteClose(1000, nil)
}

// WriteString sends a text message to the server.
//
// WriteString 发送文本消息到服务器。
func (c *Client) WriteString(message string) error {
	return c.socket.WriteMessage(gws.OpcodeText, []byte(message))
}

// WriteBinary sends a binary message to the server.
//
// WriteBinary 发送二进制消息到服务器。
func (c *Client) WriteBinary(data []byte) error {
	return c.socket.WriteMessage(gws.OpcodeBinary, data)
}

// ============================================================
// connWrapper methods
// ============================================================

// WriteString sends a text message.
//
// WriteString 发送文本消息。
func (c *connWrapper) WriteString(message string) error {
	return c.socket.WriteMessage(gws.OpcodeText, []byte(message))
}

// WriteBinary sends a binary message.
//
// WriteBinary 发送二进制消息。
func (c *connWrapper) WriteBinary(data []byte) error {
	return c.socket.WriteMessage(gws.OpcodeBinary, data)
}

// Close closes the connection.
//
// Close 关闭连接。
func (c *connWrapper) Close(code int, reason string) error {
	return c.socket.WriteClose(uint16(code), []byte(reason))
}

// Context returns the connection context.
//
// Context 返回连接上下文。
func (c *connWrapper) Context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// SetContext sets the connection context.
//
// SetContext 设置连接上下文。
func (c *connWrapper) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// RemoteAddr returns the remote address.
//
// RemoteAddr 返回远程地址。
func (c *connWrapper) RemoteAddr() string {
	return c.socket.RemoteAddr().String()
}

// LocalAddr returns the local address.
//
// LocalAddr 返回本地地址。
func (c *connWrapper) LocalAddr() string {
	return c.socket.LocalAddr().String()
}

// RawConn returns the underlying gws.Conn for advanced usage.
//
// RawConn 返回底层 gws.Conn 供高级使用。
func (c *connWrapper) RawConn() any {
	return c.socket
}

// ============================================================
// clientHandlerAdapter
// ============================================================

// clientHandlerAdapter adapts WebSocketHandler for client.
//
// clientHandlerAdapter 为客户端适配 WebSocketHandler。
type clientHandlerAdapter struct {
	handler transportcontract.WebSocketHandler
	client  *Client
}

// OnOpen is called when connection is established.
//
// OnOpen 在连接建立时调用。
func (a *clientHandlerAdapter) OnOpen(socket *gws.Conn) {
	a.handler.OnOpen(a.client.conn)
}

// OnClose is called when connection is closed.
//
// OnClose 在连接关闭时调用。
func (a *clientHandlerAdapter) OnClose(socket *gws.Conn, err error) {
	a.handler.OnClose(a.client.conn, err)
}

// OnPing handles ping frame.
//
// OnPing 处理 ping 帧。
func (a *clientHandlerAdapter) OnPing(socket *gws.Conn, payload []byte) {
	socket.WritePong(payload)
}

// OnPong handles pong frame.
//
// OnPong 处理 pong 帧。
func (a *clientHandlerAdapter) OnPong(socket *gws.Conn, payload []byte) {
	// Default pong handler does nothing
	// 默认 pong 处理器不做任何操作
}

// OnMessage handles incoming message.
//
// OnMessage 处理收到的消息。
func (a *clientHandlerAdapter) OnMessage(socket *gws.Conn, message *gws.Message) {
	var messageType int
	switch message.Opcode {
	case gws.OpcodeText:
		messageType = TextMessage
	case gws.OpcodeBinary:
		messageType = BinaryMessage
	default:
		messageType = BinaryMessage
	}

	a.handler.OnMessage(a.client.conn, messageType, message.Bytes())
	message.Close()
}

// NewClientWithTLS creates a new WebSocket client with TLS config.
//
// NewClientWithTLS 使用 TLS 配置创建新的 WebSocket 客户端。
func NewClientWithTLS(handler transportcontract.WebSocketHandler, config *transportcontract.WebSocketClientConfig, tlsConfig *tls.Config) (transportcontract.WebSocketClient, *http.Response, error) {
	if config == nil {
		config = &transportcontract.WebSocketClientConfig{}
	}

	opt := buildClientOption(config)
	opt.TlsConfig = tlsConfig

	adapter := &clientHandlerAdapter{handler: handler}
	socket, resp, err := gws.NewClient(adapter, opt)
	if err != nil {
		return nil, resp, err
	}

	client := &Client{
		socket: socket,
		conn:   &connWrapper{socket: socket},
	}
	adapter.client = client

	go socket.ReadLoop()

	return client, resp, nil
}
