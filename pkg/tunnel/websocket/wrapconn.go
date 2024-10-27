package websocket

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
)

type wrapConn struct {
	conn   *websocket.Conn
	buffer []byte
}

func newWrapConn(conn *websocket.Conn) net.Conn {
	return &wrapConn{
		conn: conn,
	}
}

func (w *wrapConn) Read(b []byte) (int, error) {
	var src []byte
	if l := len(w.buffer); l > 0 {
		src = w.buffer
		w.buffer = nil
	} else {
		_, msg, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}
		src = msg
	}
	var n int
	blen := len(b)
	if len(src) > blen {
		n = copy(b, src[:blen])
		remaining := src[blen:]
		remainingLen := len(remaining)
		w.buffer = make([]byte, remainingLen)
		copy(w.buffer, remaining)
	} else {
		n = copy(b, src)
	}
	return n, nil
}

func (w *wrapConn) Write(b []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *wrapConn) Close() error {
	return w.conn.Close()
}

func (w *wrapConn) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *wrapConn) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *wrapConn) SetDeadline(t time.Time) error {
	if err := w.conn.SetReadDeadline(t); err != nil {
		return err
	}
	return w.conn.SetWriteDeadline(t)
}

func (w *wrapConn) SetReadDeadline(t time.Time) error {
	return w.conn.SetReadDeadline(t)
}

func (w *wrapConn) SetWriteDeadline(t time.Time) error {
	return w.conn.SetWriteDeadline(t)
}
