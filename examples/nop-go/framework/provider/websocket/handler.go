// Package websocket provides WebSocket capability provider for gorp framework.
// This file implements event handler adapter for WebSocket server.
//
// WebSocket 包提供 gorp 框架的 WebSocket 能力 provider。
// 本文件实现 WebSocket 服务器的事件处理适配器。
package websocket

import (
	"time"

	"github.com/lxzan/gws"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

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
	startTime := time.Now()
	a.server.updateActivity(socket)
	conn := &connAdapter{socket: socket, server: a.server}

	// Convert gws opcode to standard message type
	// 将 gws opcode 转换为标准消息类型
	var messageType int
	var msgType string
	switch message.Opcode {
	case gws.OpcodeText:
		messageType = TextMessage
		msgType = "text"
	case gws.OpcodeBinary:
		messageType = BinaryMessage
		msgType = "binary"
	default:
		messageType = BinaryMessage
		msgType = "binary"
	}

	a.handler.OnMessage(conn, messageType, message.Bytes())

	// Record message received metric with latency
	// 记录消息接收指标和延迟
	latency := time.Since(startTime)
	a.server.metrics.OnMessageReceived(msgType, latency)

	// Release message buffer
	// 释放消息缓冲区
	message.Close()
}
