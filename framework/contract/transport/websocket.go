// Application scenarios:
// - Define transport-layer WebSocket service contract shared by providers and bootstrap logic.
// - Keep upgrade, connection, and message semantics stable and provider-neutral.
// - Let application code depend on one WebSocket abstraction instead of a concrete library.
//
// 适用场景：
// - 定义 provider 和 bootstrap 共同使用的 transport 层 WebSocket 服务契约。
// - 稳定维护 upgrade、连接和消息语义。
// - 让应用代码依赖统一 WebSocket 抽象，而不是具体库实现。
package transport

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// WebSocketKey is the container key for the WebSocket service capability.
//
// WebSocketKey 是 WebSocket 服务能力的容器键。
const WebSocketKey = "framework.websocket"

// WebSocket defines the transport-layer WebSocket service abstraction.
//
// WebSocket 定义 transport 层 WebSocket 服务抽象。
type WebSocket interface {
	// Upgrade upgrades an HTTP connection to WebSocket with the given handler.
	// The returned WebSocketConn is already running its read loop in a background goroutine.
	//
	// Upgrade 将 HTTP 连接升级为 WebSocket，并使用给定的 handler 处理事件。
	// 返回的 WebSocketConn 已在后台 goroutine 中运行读循环。
	Upgrade(w http.ResponseWriter, r *http.Request, handler WebSocketHandler) (WebSocketConn, error)

	// NewClient connects to a WebSocket server with the given handler and options.
	//
	// NewClient 使用给定的 handler 和选项连接到 WebSocket 服务端。
	NewClient(handler WebSocketHandler, opts *WebSocketClientOptions) (WebSocketConn, error)
}

// WebSocketHandler handles WebSocket events.
// OnPing and OnPong have default no-op implementations via WebSocketHandlerAdapter.
//
// WebSocketHandler 处理 WebSocket 事件。
// OnPing 和 OnPong 可通过 WebSocketHandlerAdapter 获得默认空实现。
type WebSocketHandler interface {
	// OnOpen is called when a WebSocket connection is established.
	//
	// OnOpen 在 WebSocket 连接建立时调用。
	OnOpen(conn WebSocketConn)

	// OnClose is called when a WebSocket connection is closed.
	//
	// OnClose 在 WebSocket 连接关闭时调用。
	OnClose(conn WebSocketConn, err error)

	// OnMessage is called when a WebSocket message is received.
	//
	// OnMessage 在收到 WebSocket 消息时调用。
	OnMessage(conn WebSocketConn, msg WebSocketMessage)
}

// WebSocketHandlerAdapter provides default no-op implementations for WebSocketHandler.
// Embed this struct to only implement the methods you need.
//
// WebSocketHandlerAdapter 提供 WebSocketHandler 的默认空实现。
// 嵌入此结构体即可只实现需要的方法。
type WebSocketHandlerAdapter struct{}

func (WebSocketHandlerAdapter) OnOpen(WebSocketConn)          {}
func (WebSocketHandlerAdapter) OnClose(WebSocketConn, error)  {}
func (WebSocketHandlerAdapter) OnMessage(WebSocketConn, WebSocketMessage) {}

// WebSocketOpcode represents the type of a WebSocket message.
//
// WebSocketOpcode 表示 WebSocket 消息类型。
type WebSocketOpcode int

const (
	// OpcodeText represents a text message.
	OpcodeText WebSocketOpcode = iota + 1
	// OpcodeBinary represents a binary message.
	OpcodeBinary
)

// WebSocketMessage represents a received WebSocket message.
//
// WebSocketMessage 表示收到的 WebSocket 消息。
type WebSocketMessage interface {
	// Opcode returns the message type (text or binary).
	Opcode() WebSocketOpcode

	// Bytes returns the raw message payload.
	Bytes() []byte

	// String returns the message payload as a string.
	// For binary messages, this returns the string representation of the byte slice.
	String() string
}

// WebSocketConn represents a WebSocket connection.
//
// WebSocketConn 表示一个 WebSocket 连接。
type WebSocketConn interface {
	// WriteText sends a text message.
	WriteText(data []byte) error

	// WriteBinary sends a binary message.
	WriteBinary(data []byte) error

	// WriteClose sends a close frame with the given status code and reason.
	WriteClose(code int, reason string) error

	// SetDeadline sets the read and write deadlines for the connection.
	SetDeadline(t time.Time) error

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr

	// Session returns per-connection session storage for user-defined key-value pairs.
	Session() WebSocketSession

	// Close closes the connection.
	Close() error
}

// WebSocketSession provides per-connection key-value storage.
//
// WebSocketSession 提供每个连接的键值对存储。
type WebSocketSession interface {
	Load(key any) (any, bool)
	Store(key, value any)
	Delete(key any)
}

// WebSocketServerOptions describes WebSocket server-side configuration.
//
// WebSocketServerOptions 描述 WebSocket 服务端配置。
type WebSocketServerOptions struct {
	// ReadBufferSize sets the read buffer size in bytes (0 = default).
	ReadBufferSize int `mapstructure:"read_buffer_size"`
	// WriteBufferSize sets the write buffer size in bytes (0 = default).
	WriteBufferSize int `mapstructure:"write_buffer_size"`
	// ParallelEnabled enables parallel message handling.
	ParallelEnabled bool `mapstructure:"parallel_enabled"`
	// CompressionEnabled enables permessage-deflate compression.
	CompressionEnabled bool `mapstructure:"compression_enabled"`
}

// WebSocketClientOptions describes WebSocket client-side configuration.
//
// WebSocketClientOptions 描述 WebSocket 客户端配置。
type WebSocketClientOptions struct {
	// Addr is the WebSocket server address to connect to (e.g., "ws://127.0.0.1:8080/ws").
	Addr string `mapstructure:"addr"`
	// TLSInsecureSkipVerify skips TLS certificate verification.
	TLSInsecureSkipVerify bool `mapstructure:"tls_insecure_skip_verify"`
	// CompressionEnabled enables permessage-deflate compression.
	CompressionEnabled bool `mapstructure:"compression_enabled"`
}

// MapSession is an in-memory implementation of WebSocketSession backed by sync.Map.
//
// MapSession 是基于 sync.Map 的内存 WebSocketSession 实现。
type MapSession struct {
	store sync.Map
}

// NewMapSession creates a new empty MapSession.
func NewMapSession() *MapSession {
	return &MapSession{}
}

func (s *MapSession) Load(key any) (any, bool)   { return s.store.Load(key) }
func (s *MapSession) Store(key, value any)        { s.store.Store(key, value) }
func (s *MapSession) Delete(key any)              { s.store.Delete(key) }
