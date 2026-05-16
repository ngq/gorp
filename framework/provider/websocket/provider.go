// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements WebSocket server using gws library.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件使用 gws 库实现 WebSocket 服务器。
package websocket

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/lxzan/gws"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	// TextMessage indicates a text message frame.
	// TextMessage 表示文本消息帧。
	TextMessage = 1

	// BinaryMessage indicates a binary message frame.
	// BinaryMessage 表示二进制消息帧。
	BinaryMessage = 2

	// CloseMessage indicates a close frame.
	// CloseMessage 表示关闭帧。
	CloseMessage = 8
)

// Provider provides WebSocket capability.
//
// Provider 提供 WebSocket 能力。
type Provider struct {
	config *transportcontract.WebSocketConfig
}

// NewProvider creates a new WebSocket provider with default config.
//
// NewProvider 使用默认配置创建新的 WebSocket provider。
func NewProvider() *Provider {
	return NewProviderWithConfig(nil)
}

// NewProviderWithConfig creates a new WebSocket provider with custom config.
//
// NewProviderWithConfig 使用自定义配置创建新的 WebSocket provider。
func NewProviderWithConfig(config *transportcontract.WebSocketConfig) *Provider {
	if config == nil {
		config = &transportcontract.WebSocketConfig{
			ParallelEnabled: true,
		}
	}
	return &Provider{config: config}
}

// Name returns the provider name.
//
// Name 返回 provider 名称。
func (p *Provider) Name() string {
	return "websocket"
}

// Register registers WebSocket service into container.
// Also registers a closer to gracefully shutdown connections on container destroy.
//
// Register 将 WebSocket 服务注册到容器。
// 同时注册 closer，在容器销毁时优雅关闭连接。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.WebSocketKey, func(c runtimecontract.Container) (any, error) {
		return NewServerWithConfig(p.config), nil
	}, true)
	c.RegisterCloser(transportcontract.WebSocketKey, &websocketCloser{c: c})
	return nil
}

// Boot initializes the WebSocket service.
//
// Boot 初始化 WebSocket 服务。
func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// IsDefer reports whether the provider should be deferred.
//
// IsDefer 返回该 provider 是否应延迟加载。
func (p *Provider) IsDefer() bool {
	return false
}

// Provides returns the keys that this provider provides.
//
// Provides 返回该 provider 提供的 key 列表。
func (p *Provider) Provides() []string {
	return []string{string(transportcontract.WebSocketKey)}
}

// DependsOn returns the keys that this provider depends on.
//
// DependsOn 返回该 provider 依赖的 key 列表。
func (p *Provider) DependsOn() []string {
	return nil
}

// Server implements WebSocketServer interface using gws.
//
// Server 使用 gws 实现 WebSocketServer 接口。
type Server struct {
	mu           sync.RWMutex
	conns        map[*gws.Conn]context.Context
	config       *transportcontract.WebSocketConfig
	stopCh       chan struct{}           // 信号通道，用于停止心跳检测 goroutine
	lastActivity map[*gws.Conn]time.Time // 每个连接的最后活跃时间
}

// NewServer creates a new WebSocket server with default config.
//
// NewServer 使用默认配置创建新的 WebSocket 服务器。
func NewServer() *Server {
	return NewServerWithConfig(nil)
}

// NewServerWithConfig creates a new WebSocket server with custom config.
//
// NewServerWithConfig 使用自定义配置创建新的 WebSocket 服务器。
func NewServerWithConfig(config *transportcontract.WebSocketConfig) *Server {
	if config == nil {
		config = &transportcontract.WebSocketConfig{
			ParallelEnabled: true,
		}
	}
	s := &Server{
		conns:        make(map[*gws.Conn]context.Context),
		config:       config,
		stopCh:       make(chan struct{}),
		lastActivity: make(map[*gws.Conn]time.Time),
	}
	// 启动心跳检测 goroutine
	if config.ReadTimeout > 0 {
		go s.healthCheck()
	}
	return s
}

// buildServerOption builds gws.ServerOption from WebSocketConfig.
//
// buildServerOption 从 WebSocketConfig 构建 gws.ServerOption。
func (s *Server) buildServerOption() *gws.ServerOption {
	opt := &gws.ServerOption{
		ParallelEnabled: s.config.ParallelEnabled,
		Recovery:        gws.Recovery,
	}

	if s.config.ParallelGolimit > 0 {
		opt.ParallelGolimit = s.config.ParallelGolimit
	}

	if s.config.ReadBufferSize > 0 {
		opt.ReadBufferSize = s.config.ReadBufferSize
	}

	if s.config.WriteBufferSize > 0 {
		opt.WriteBufferSize = s.config.WriteBufferSize
	}

	if s.config.MaxMessageSize > 0 {
		opt.ReadMaxPayloadSize = int(s.config.MaxMessageSize)
	}

	if s.config.EnableCompression {
		level := s.config.CompressionLevel
		if level < 1 || level > 9 {
			level = 6 // default compression level
		}
		opt.PermessageDeflate = gws.PermessageDeflate{
			Enabled: true,
			Level:   level,
		}
	}

	return opt
}

// Upgrade upgrades an HTTP connection to WebSocket.
//
// Upgrade 将 HTTP 连接升级为 WebSocket。
func (s *Server) Upgrade(w http.ResponseWriter, r *http.Request, handler transportcontract.WebSocketHandler) (transportcontract.WebSocketConn, error) {
	// Create adapter for this specific upgrade
	// 为这次升级创建适配器
	adapter := &eventHandlerAdapter{
		server:  s,
		handler: handler,
	}

	// Build server option from config
	// 从配置构建服务器选项
	opt := s.buildServerOption()

	// Create new upgrader with the handler
	// 使用 handler 创建新的 upgrader
	upgrader := gws.NewUpgrader(adapter, opt)

	socket, err := upgrader.Upgrade(w, r)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.conns[socket] = r.Context()
	s.lastActivity[socket] = time.Now()
	s.mu.Unlock()

	// Start read loop in a separate goroutine
	// 在单独的 goroutine 中启动读循环
	go socket.ReadLoop()

	return &connAdapter{socket: socket, server: s}, nil
}

// NewBroadcaster creates a new broadcaster for batch message sending.
//
// NewBroadcaster 创建新的广播器用于批量消息发送。
func (s *Server) NewBroadcaster() transportcontract.WebSocketBroadcaster {
	return &broadcasterAdapter{server: s}
}

// Connections returns all active connections.
//
// Connections 返回所有活跃连接。
func (s *Server) Connections() []transportcontract.WebSocketConn {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conns := make([]transportcontract.WebSocketConn, 0, len(s.conns))
	for gwsConn := range s.conns {
		conns = append(conns, &connAdapter{socket: gwsConn, server: s})
	}
	return conns
}

// Count returns the number of active connections.
//
// Count 返回活跃连接数量。
func (s *Server) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns)
}

// Shutdown gracefully closes all connections.
// Respects context cancellation/timeout and does not hold lock during I/O.
//
// Shutdown 优雅关闭所有连接。
// 尊重上下文取消/超时，不在 I/O 期间持锁。
func (s *Server) Shutdown(ctx context.Context) error {
	// 停止心跳检测 goroutine
	select {
	case <-s.stopCh:
		// 已关闭
	default:
		close(s.stopCh)
	}

	// 持锁收集连接，然后释放锁
	s.mu.Lock()
	conns := make([]*gws.Conn, 0, len(s.conns))
	for socket := range s.conns {
		conns = append(conns, socket)
	}
	s.conns = make(map[*gws.Conn]context.Context)
	s.lastActivity = make(map[*gws.Conn]time.Time)
	s.mu.Unlock()

	// 不持锁发送关闭帧
	for _, socket := range conns {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			socket.WriteClose(1000, nil)
		}
	}

	return nil
}

// removeConn removes a connection from the server.
//
// removeConn 从服务器移除连接。
func (s *Server) removeConn(socket *gws.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, socket)
	delete(s.lastActivity, socket)
}

// updateActivity updates the last activity time for a connection.
//
// updateActivity 更新连接的最后活跃时间。
func (s *Server) updateActivity(socket *gws.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.lastActivity[socket]; ok {
		s.lastActivity[socket] = time.Now()
	}
}

// healthCheck periodically checks for stale connections and closes them.
// Runs as a goroutine when ReadTimeout is configured.
//
// healthCheck 定期检查过期连接并关闭它们。
// 当配置了 ReadTimeout 时以 goroutine 方式运行。
func (s *Server) healthCheck() {
	ticker := time.NewTicker(s.config.ReadTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.closeStaleConnections()
		}
	}
}

// closeStaleConnections closes connections that have exceeded ReadTimeout.
//
// closeStaleConnections 关闭超过 ReadTimeout 的连接。
func (s *Server) closeStaleConnections() {
	now := time.Now()
	var staleConns []*gws.Conn

	s.mu.RLock()
	for socket, lastTime := range s.lastActivity {
		if now.Sub(lastTime) > s.config.ReadTimeout {
			staleConns = append(staleConns, socket)
		}
	}
	s.mu.RUnlock()

	for _, socket := range staleConns {
		socket.WriteClose(1000, nil)
	}
}

// connAdapter adapts gws.Conn to WebSocketConn interface.
//
// connAdapter 将 gws.Conn 适配为 WebSocketConn 接口。
type connAdapter struct {
	socket *gws.Conn
	server *Server
}

// WriteString sends a text message to the client.
//
// WriteString 发送文本消息到客户端。
func (c *connAdapter) WriteString(message string) error {
	return c.socket.WriteMessage(gws.OpcodeText, []byte(message))
}

// WriteBinary sends a binary message to the client.
//
// WriteBinary 发送二进制消息到客户端。
func (c *connAdapter) WriteBinary(data []byte) error {
	return c.socket.WriteMessage(gws.OpcodeBinary, data)
}

// Close closes the connection with a close code and reason.
// Does not call removeConn here; OnClose callback handles removal to avoid double-delete.
//
// Close 关闭连接，可指定关闭码和原因。
// 不在此处调用 removeConn，由 OnClose 回调统一负责删除，避免双重删除。
func (c *connAdapter) Close(code int, reason string) error {
	return c.socket.WriteClose(uint16(code), []byte(reason))
}

// Context returns the context associated with the connection.
//
// Context 返回与连接关联的上下文。
func (c *connAdapter) Context() context.Context {
	c.server.mu.RLock()
	defer c.server.mu.RUnlock()
	ctx, ok := c.server.conns[c.socket]
	if !ok {
		return context.Background()
	}
	return ctx
}

// SetContext sets the context for the connection.
//
// SetContext 设置连接的上下文。
func (c *connAdapter) SetContext(ctx context.Context) {
	c.server.mu.Lock()
	defer c.server.mu.Unlock()
	c.server.conns[c.socket] = ctx
}

// RemoteAddr returns the remote address of the connection.
//
// RemoteAddr 返回连接的远程地址。
func (c *connAdapter) RemoteAddr() string {
	return c.socket.RemoteAddr().String()
}

// LocalAddr returns the local address of the connection.
//
// LocalAddr 返回连接的本地地址。
func (c *connAdapter) LocalAddr() string {
	return c.socket.LocalAddr().String()
}

// RawConn returns the underlying gws.Conn for advanced usage.
// The returned value can be type-asserted to *gws.Conn.
//
// RawConn 返回底层 gws.Conn 供高级使用。
// 返回值可类型断言为 *gws.Conn。
//
// Example:
//
//	if raw, ok := conn.RawConn().(*gws.Conn); ok {
//	    raw.WriteMessage(gws.OpcodeText, []byte("direct message"))
//	}
func (c *connAdapter) RawConn() any {
	return c.socket
}

// broadcasterAdapter adapts gws broadcaster to WebSocketBroadcaster interface.
//
// broadcasterAdapter 将 gws 广播器适配为 WebSocketBroadcaster 接口。
type broadcasterAdapter struct {
	server *Server
}

// BroadcastString sends a text message to all connections.
//
// BroadcastString 向所有连接发送文本消息。
func (b *broadcasterAdapter) BroadcastString(message string) error {
	b.server.mu.RLock()
	defer b.server.mu.RUnlock()

	payload := []byte(message)
	broadcaster := gws.NewBroadcaster(gws.OpcodeText, payload)
	defer broadcaster.Close()

	total := len(b.server.conns)
	var errs []error
	for socket := range b.server.conns {
		if err := broadcaster.Broadcast(socket); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("websocket: %d of %d connections failed: %w", len(errs), total, errors.Join(errs...))
	}
	return nil
}

// BroadcastBinary sends a binary message to all connections.
//
// BroadcastBinary 向所有连接发送二进制消息。
func (b *broadcasterAdapter) BroadcastBinary(data []byte) error {
	b.server.mu.RLock()
	defer b.server.mu.RUnlock()

	broadcaster := gws.NewBroadcaster(gws.OpcodeBinary, data)
	defer broadcaster.Close()

	total := len(b.server.conns)
	var errs []error
	for socket := range b.server.conns {
		if err := broadcaster.Broadcast(socket); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("websocket: %d of %d connections failed: %w", len(errs), total, errors.Join(errs...))
	}
	return nil
}

// BroadcastStringExcept sends a text message to all connections except the specified one.
//
// BroadcastStringExcept 向除指定连接外的所有连接发送文本消息。
func (b *broadcasterAdapter) BroadcastStringExcept(message string, excludeConn transportcontract.WebSocketConn) error {
	b.server.mu.RLock()
	defer b.server.mu.RUnlock()

	adapter, ok := excludeConn.(*connAdapter)
	if !ok {
		return fmt.Errorf("websocket: exclude conn is not a server connection")
	}
	excludeSocket := adapter.socket
	broadcaster := gws.NewBroadcaster(gws.OpcodeText, []byte(message))

	total := len(b.server.conns) - 1
	var errs []error
	for socket := range b.server.conns {
		if socket != excludeSocket {
			if err := broadcaster.Broadcast(socket); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("websocket: %d of %d connections failed: %w", len(errs), total, errors.Join(errs...))
	}
	return nil
}

// BroadcastBinaryExcept sends a binary message to all connections except the specified one.
//
// BroadcastBinaryExcept 向除指定连接外的所有连接发送二进制消息。
func (b *broadcasterAdapter) BroadcastBinaryExcept(data []byte, excludeConn transportcontract.WebSocketConn) error {
	b.server.mu.RLock()
	defer b.server.mu.RUnlock()

	adapter, ok := excludeConn.(*connAdapter)
	if !ok {
		return fmt.Errorf("websocket: exclude conn is not a server connection")
	}
	excludeSocket := adapter.socket
	broadcaster := gws.NewBroadcaster(gws.OpcodeBinary, data)

	total := len(b.server.conns) - 1
	var errs []error
	for socket := range b.server.conns {
		if socket != excludeSocket {
			if err := broadcaster.Broadcast(socket); err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("websocket: %d of %d connections failed: %w", len(errs), total, errors.Join(errs...))
	}
	return nil
}

// eventHandlerAdapter adapts WebSocketHandler to gws.Event interface.
//
// eventHandlerAdapter 将 WebSocketHandler 适配为 gws.Event 接口。
type eventHandlerAdapter struct {
	server  *Server
	handler transportcontract.WebSocketHandler
}

// OnOpen is called when a new connection is established.
//
// OnOpen 在新连接建立时调用。
func (a *eventHandlerAdapter) OnOpen(socket *gws.Conn) {
	conn := &connAdapter{socket: socket, server: a.server}
	a.handler.OnOpen(conn)
}

// OnClose is called when a connection is closed.
//
// OnClose 在连接关闭时调用。
func (a *eventHandlerAdapter) OnClose(socket *gws.Conn, err error) {
	a.server.removeConn(socket)
	conn := &connAdapter{socket: socket, server: a.server}
	a.handler.OnClose(conn, err)
}

// OnPing is called when a ping frame is received.
//
// OnPing 在收到 ping 帧时调用。
func (a *eventHandlerAdapter) OnPing(socket *gws.Conn, payload []byte) {
	a.server.updateActivity(socket)
	socket.WritePong(payload)
}

// OnPong is called when a pong frame is received.
//
// OnPong 在收到 pong 帧时调用。
func (a *eventHandlerAdapter) OnPong(socket *gws.Conn, payload []byte) {
	a.server.updateActivity(socket)
}

// OnMessage is called when a message is received.
//
// OnMessage 在收到消息时调用。
func (a *eventHandlerAdapter) OnMessage(socket *gws.Conn, message *gws.Message) {
	a.server.updateActivity(socket)
	conn := &connAdapter{socket: socket, server: a.server}

	// Convert gws opcode to standard message type
	// 将 gws opcode 转换为标准消息类型
	var messageType int
	switch message.Opcode {
	case gws.OpcodeText:
		messageType = TextMessage
	case gws.OpcodeBinary:
		messageType = BinaryMessage
	default:
		messageType = BinaryMessage
	}

	a.handler.OnMessage(conn, messageType, message.Bytes())

	// Release message buffer
	// 释放消息缓冲区
	message.Close()
}

// ============================================================
// WebSocket Client
// ============================================================

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

// connWrapper methods

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

// DefaultParallelGolimit returns the default parallel goroutine limit.
//
// DefaultParallelGolimit 返回默认并行 goroutine 限制。
func DefaultParallelGolimit() int {
	return runtime.NumCPU()
}

// websocketCloser implements io.Closer to gracefully shutdown WebSocket server on container destroy.
//
// websocketCloser 实现 io.Closer，在容器销毁时优雅关闭 WebSocket 服务器。
type websocketCloser struct {
	c runtimecontract.Container
}

func (wc *websocketCloser) Close() error {
	serverAny, err := wc.c.Make(transportcontract.WebSocketKey)
	if err != nil {
		return nil
	}
	server, ok := serverAny.(*Server)
	if !ok {
		return nil
	}
	return server.Shutdown(context.Background())
}
