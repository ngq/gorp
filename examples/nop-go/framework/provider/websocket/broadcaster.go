// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements broadcaster adapter for WebSocket server.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件实现 WebSocket 服务器的广播适配器。
package websocket

import (
	"errors"
	"fmt"

	"github.com/lxzan/gws"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

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

	// Record broadcast metric
	// 记录广播指标
	b.server.metrics.OnBroadcast("text")

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

	// Record broadcast metric
	// 记录广播指标
	b.server.metrics.OnBroadcast("binary")

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

	// Record broadcast metric
	// 记录广播指标
	b.server.metrics.OnBroadcast("text")

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

	// Record broadcast metric
	// 记录广播指标
	b.server.metrics.OnBroadcast("binary")

	if len(errs) > 0 {
		return fmt.Errorf("websocket: %d of %d connections failed: %w", len(errs), total, errors.Join(errs...))
	}
	return nil
}
