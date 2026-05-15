// Package transport defines transport-layer contracts for gorp framework.
// This file defines the WebSocket abstraction contract.
//
// Transport 包定义 gorp 框架的传输层契约。
// 本文件定义 WebSocket 抽象契约。
package transport

import (
	"context"
	"net/http"
)

// WebSocketKey is the container key for WebSocket service.
//
// WebSocketKey 是 WebSocket 服务的容器键。
const WebSocketKey = "framework.transport.websocket"

// WebSocketConn represents a WebSocket connection.
//
// WebSocketConn 表示 WebSocket 连接。
type WebSocketConn interface {
	// WriteString sends a text message to the client.
	// WriteString 发送文本消息到客户端。
	WriteString(message string) error

	// WriteBinary sends a binary message to the client.
	// WriteBinary 发送二进制消息到客户端。
	WriteBinary(data []byte) error

	// Close closes the connection with a close code and reason.
	// Close 关闭连接，可指定关闭码和原因。
	Close(code int, reason string) error

	// Context returns the context associated with the connection.
	// Context 返回与连接关联的上下文。
	Context() context.Context

	// SetContext sets the context for the connection.
	// SetContext 设置连接的上下文。
	SetContext(ctx context.Context)

	// RemoteAddr returns the remote address of the connection.
	// RemoteAddr 返回连接的远程地址。
	RemoteAddr() string

	// LocalAddr returns the local address of the connection.
	// LocalAddr 返回连接的本地地址。
	LocalAddr() string

	// RawConn returns the underlying raw connection for advanced usage.
	// The returned value is the underlying library's connection type.
	// Use with caution - bypassing the abstraction may break framework guarantees.
	//
	// RawConn 返回底层原生连接供高级使用。
	// 返回值是底层库的连接类型。
	// 请谨慎使用 - 绕过抽象层可能破坏框架保证。
	RawConn() any
}

// WebSocketHandler handles WebSocket events.
//
// WebSocketHandler 处理 WebSocket 事件。
type WebSocketHandler interface {
	// OnOpen is called when a new connection is established.
	// OnOpen 在新连接建立时调用。
	OnOpen(conn WebSocketConn)

	// OnClose is called when a connection is closed.
	// OnClose 在连接关闭时调用。
	OnClose(conn WebSocketConn, err error)

	// OnMessage is called when a message is received.
	// OnMessage 在收到消息时调用。
	OnMessage(conn WebSocketConn, messageType int, data []byte)
}

// WebSocketBroadcaster broadcasts messages to multiple connections.
//
// WebSocketBroadcaster 向多个连接广播消息。
type WebSocketBroadcaster interface {
	// BroadcastString sends a text message to all connections.
	// BroadcastString 向所有连接发送文本消息。
	BroadcastString(message string) error

	// BroadcastBinary sends a binary message to all connections.
	// BroadcastBinary 向所有连接发送二进制消息。
	BroadcastBinary(data []byte) error

	// BroadcastStringExcept sends a text message to all connections except the specified one.
	// BroadcastStringExcept 向除指定连接外的所有连接发送文本消息。
	BroadcastStringExcept(message string, excludeConn WebSocketConn) error

	// BroadcastBinaryExcept sends a binary message to all connections except the specified one.
	// BroadcastBinaryExcept 向除指定连接外的所有连接发送二进制消息。
	BroadcastBinaryExcept(data []byte, excludeConn WebSocketConn) error
}

// WebSocketServer manages WebSocket connections and provides broadcasting capability.
//
// WebSocketServer 管理 WebSocket 连接并提供广播能力。
type WebSocketServer interface {
	// Upgrade upgrades an HTTP connection to WebSocket.
	// Upgrade 将 HTTP 连接升级为 WebSocket。
	Upgrade(w http.ResponseWriter, r *http.Request, handler WebSocketHandler) (WebSocketConn, error)

	// NewBroadcaster creates a new broadcaster for batch message sending.
	// NewBroadcaster 创建新的广播器用于批量消息发送。
	NewBroadcaster() WebSocketBroadcaster

	// Connections returns all active connections.
	// Connections 返回所有活跃连接。
	Connections() []WebSocketConn

	// Count returns the number of active connections.
	// Count 返回活跃连接数量。
	Count() int

	// Shutdown gracefully closes all connections.
	// Shutdown 优雅关闭所有连接。
	Shutdown(ctx context.Context) error
}

// WebSocketClient represents a WebSocket client connection.
//
// WebSocketClient 表示 WebSocket 客户端连接。
type WebSocketClient interface {
	// Conn returns the underlying connection.
	// Conn 返回底层连接。
	Conn() WebSocketConn

	// Close closes the client connection.
	// Close 关闭客户端连接。
	Close() error

	// WriteString sends a text message to the server.
	// WriteString 发送文本消息到服务器。
	WriteString(message string) error

	// WriteBinary sends a binary message to the server.
	// WriteBinary 发送二进制消息到服务器。
	WriteBinary(data []byte) error
}

// WebSocketConfig holds WebSocket server configuration.
//
// WebSocketConfig 保存 WebSocket 服务器配置。
type WebSocketConfig struct {
	// Enable compression (permessage-deflate).
	// 启用压缩（permessage-deflate）。
	EnableCompression bool

	// Compression level (1-9, default 6).
	// 压缩级别（1-9，默认 6）。
	CompressionLevel int

	// Max message size in bytes (default 0 = unlimited).
	// 最大消息大小（字节），默认 0 表示无限制。
	MaxMessageSize int64

	// Read buffer size in bytes (default 4096).
	// 读缓冲区大小（字节），默认 4096。
	ReadBufferSize int

	// Write buffer size in bytes (default 4096).
	// 写缓冲区大小（字节），默认 4096。
	WriteBufferSize int

	// ParallelEnabled enables parallel message processing.
	// 启用并行消息处理。
	ParallelEnabled bool

	// ParallelGolimit is the max goroutines for parallel processing (default runtime.NumCPU).
	// 并行处理的最大 goroutine 数，默认 runtime.NumCPU。
	ParallelGolimit int
}

// WebSocketClientConfig holds WebSocket client configuration.
//
// WebSocketClientConfig 保存 WebSocket 客户端配置。
type WebSocketClientConfig struct {
	// Server URL (ws:// or wss://).
	// 服务器 URL（ws:// 或 wss://）。
	URL string

	// Enable compression (permessage-deflate).
	// 启用压缩（permessage-deflate）。
	EnableCompression bool

	// Compression level (1-9, default 6).
	// 压缩级别（1-9，默认 6）。
	CompressionLevel int

	// Max message size in bytes (default 0 = unlimited).
	// 最大消息大小（字节），默认 0 表示无限制。
	MaxMessageSize int64

	// Read buffer size in bytes (default 4096).
	// 读缓冲区大小（字节），默认 4096。
	ReadBufferSize int

	// Write buffer size in bytes (default 4096).
	// 写缓冲区大小（字节），默认 4096。
	WriteBufferSize int

	// Handshake timeout in seconds (default 10).
	// 握手超时（秒），默认 10。
	HandshakeTimeout int

	// Request headers to send during handshake.
	// 握手时发送的请求头。
	RequestHeader http.Header
}