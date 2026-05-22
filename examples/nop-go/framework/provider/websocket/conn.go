// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements connection adapter for WebSocket server.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件实现 WebSocket 服务器的连接适配器。
package websocket

import (
	"context"

	"github.com/lxzan/gws"
)

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
	err := c.socket.WriteMessage(gws.OpcodeText, []byte(message))
	if err != nil {
		c.server.metrics.OnError("write")
	} else {
		c.server.metrics.OnMessageSent("text")
	}
	return err
}

// WriteBinary sends a binary message to the client.
//
// WriteBinary 发送二进制消息到客户端。
func (c *connAdapter) WriteBinary(data []byte) error {
	err := c.socket.WriteMessage(gws.OpcodeBinary, data)
	if err != nil {
		c.server.metrics.OnError("write")
	} else {
		c.server.metrics.OnMessageSent("binary")
	}
	return err
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
