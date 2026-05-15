package gws

import (
	"crypto/tls"
	"net"
	"time"

	gwslib "github.com/lxzan/gws"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// connWrapper wraps a gws.Conn to implement transportcontract.WebSocketConn.
type connWrapper struct {
	conn *gwslib.Conn
}

var _ transportcontract.WebSocketConn = (*connWrapper)(nil)

func (c *connWrapper) WriteText(data []byte) error {
	return c.conn.WriteMessage(gwslib.OpcodeText, data)
}

func (c *connWrapper) WriteBinary(data []byte) error {
	return c.conn.WriteMessage(gwslib.OpcodeBinary, data)
}

func (c *connWrapper) WriteClose(code int, reason string) error {
	return c.conn.WriteClose(uint16(code), []byte(reason))
}

func (c *connWrapper) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *connWrapper) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *connWrapper) Session() transportcontract.WebSocketSession {
	return &sessionAdapter{storage: c.conn.Session()}
}

func (c *connWrapper) Close() error {
	return c.conn.WriteClose(1000, []byte(""))
}

// sessionAdapter adapts gws.SessionStorage to transportcontract.WebSocketSession.
type sessionAdapter struct {
	storage gwslib.SessionStorage
}

var _ transportcontract.WebSocketSession = (*sessionAdapter)(nil)

func (s *sessionAdapter) Load(key any) (any, bool) {
	return s.storage.Load(key.(string))
}

func (s *sessionAdapter) Store(key, value any) {
	s.storage.Store(key.(string), value)
}

func (s *sessionAdapter) Delete(key any) {
	s.storage.Delete(key.(string))
}

// eventAdapter adapts transportcontract.WebSocketHandler to gws.Event.
type eventAdapter struct {
	handler transportcontract.WebSocketHandler
	conn    transportcontract.WebSocketConn
}

var _ gwslib.Event = (*eventAdapter)(nil)

func (a *eventAdapter) OnOpen(socket *gwslib.Conn) {
	if a.conn == nil {
		a.conn = &connWrapper{conn: socket}
	}
	a.handler.OnOpen(a.conn)
}

func (a *eventAdapter) OnClose(socket *gwslib.Conn, err error) {
	if a.conn == nil {
		a.conn = &connWrapper{conn: socket}
	}
	a.handler.OnClose(a.conn, err)
}

func (a *eventAdapter) OnPing(socket *gwslib.Conn, payload []byte) {
	_ = socket.WritePong(payload)
}

func (a *eventAdapter) OnPong(socket *gwslib.Conn, payload []byte) {
	// Default: no-op. Override by embedding WebSocketHandlerAdapter.
}

func (a *eventAdapter) OnMessage(socket *gwslib.Conn, message *gwslib.Message) {
	defer message.Close()
	if a.conn == nil {
		a.conn = &connWrapper{conn: socket}
	}
	var opcode transportcontract.WebSocketOpcode
	switch message.Opcode {
	case gwslib.OpcodeText:
		opcode = transportcontract.OpcodeText
	case gwslib.OpcodeBinary:
		opcode = transportcontract.OpcodeBinary
	default:
		return
	}
	a.handler.OnMessage(a.conn, &messageAdapter{opcode: opcode, data: message.Bytes()})
}

// messageAdapter wraps gws message data to implement transportcontract.WebSocketMessage.
type messageAdapter struct {
	opcode transportcontract.WebSocketOpcode
	data   []byte
}

var _ transportcontract.WebSocketMessage = (*messageAdapter)(nil)

func (m *messageAdapter) Opcode() transportcontract.WebSocketOpcode { return m.opcode }
func (m *messageAdapter) Bytes() []byte                            { return m.data }
func (m *messageAdapter) String() string                           { return string(m.data) }

// newTLSInsecureConfig creates a TLS config that skips certificate verification.
func newTLSInsecureConfig() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true}
}
