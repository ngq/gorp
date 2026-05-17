// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements WebSocket server using gws library.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件使用 gws 库实现 WebSocket 服务器。
package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/lxzan/gws"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Server implements WebSocketServer interface using gws.
//
// Server 使用 gws 实现 WebSocketServer 接口。
type Server struct {
	mu           sync.RWMutex
	conns        map[*gws.Conn]context.Context
	config       *transportcontract.WebSocketConfig
	stopCh       chan struct{}           // 信号通道，用于停止心跳检测 goroutine
	lastActivity map[*gws.Conn]time.Time // 每个连接的最后活跃时间
	metrics      *WSMetricsRecorder      // Prometheus 指标记录器
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
		metrics:      NewWSMetricsRecorder(),
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
		s.metrics.OnError("upgrade")
		return nil, err
	}

	s.mu.Lock()
	s.conns[socket] = r.Context()
	s.lastActivity[socket] = time.Now()
	s.mu.Unlock()

	// Record connection metric
	// 记录连接指标
	s.metrics.OnConnect()

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
	if _, ok := s.conns[socket]; ok {
		s.metrics.OnDisconnect()
	}
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