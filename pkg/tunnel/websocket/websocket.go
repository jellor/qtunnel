package websocket

import (
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xtaci/smux"

	"github.com/jellor/qtunnel/pkg/tunnel"
)

var upgrader = websocket.Upgrader{}

type WebSocketListener struct {
	httpServer *http.Server
	connCh     chan *websocket.Conn
	dieErr     error
	die        chan struct{}
	dieOnce    sync.Once
}

func Listen(address string) (tunnel.Listener, error) {
	server := &http.Server{
		Addr: address,
	}
	websocketListener := &WebSocketListener{
		httpServer: server,
		connCh:     make(chan *websocket.Conn, 4096),
		die:        make(chan struct{}),
	}
	server.Handler = handleHTTPHandler(websocketListener)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			return
		}
	}()
	return websocketListener, nil
}

func handleHTTPHandler(l *WebSocketListener) http.Handler {
	return http.HandlerFunc(l.handleHTTP)
}

func (l *WebSocketListener) handleHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()

	l.connCh <- c

	<-l.die
}

func (l *WebSocketListener) Accept() (tunnel.Connection, error) {
	select {
	case <-l.die:
		return nil, l.dieErr
	case conn, ok := <-l.connCh:
		if !ok {
			return nil, l.dieErr
		}
		wrapConn := newWrapConn(conn)
		session, err := smux.Server(wrapConn, nil)
		if err != nil {
			return nil, err
		}
		return &WebSocketConnection{conn: conn, session: session}, nil
	}
}

func (l *WebSocketListener) Close() error {
	l.dieOnce.Do(func() {
		close(l.die)
	})
	return l.httpServer.Close()
}

func (l *WebSocketListener) Addr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", l.httpServer.Addr)
	if err != nil {
		panic(err)
	}
	return addr
}

type WebSocketDialer struct {
}

func (d *WebSocketDialer) Dial(address string) (tunnel.Connection, error) {
	u := url.URL{Scheme: "ws", Host: address, Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	wrapConn := newWrapConn(conn)
	session, err := smux.Server(wrapConn, nil)
	if err != nil {
		return nil, err
	}
	return &WebSocketConnection{conn: conn, session: session}, nil
}

type WebSocketConnection struct {
	conn    *websocket.Conn
	session *smux.Session
}

func (c *WebSocketConnection) OpenStream() (tunnel.Stream, error) {
	stream, err := c.session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &WebSocketStream{stream: stream}, nil
}

func (c *WebSocketConnection) AcceptStream() (tunnel.Stream, error) {
	stream, err := c.session.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &WebSocketStream{stream: stream}, nil
}

func (c *WebSocketConnection) Close() error {
	return c.session.Close()
}

type WebSocketStream struct {
	stream *smux.Stream
}

func (s *WebSocketStream) Read(b []byte) (n int, err error) {
	return s.stream.Read(b)
}

func (s *WebSocketStream) Write(b []byte) (n int, err error) {
	return s.stream.Write(b)
}

func (s *WebSocketStream) Close() error {
	return s.stream.Close()
}

func (s *WebSocketStream) LocalAddr() net.Addr {
	return s.stream.LocalAddr()
}

func (s *WebSocketStream) RemoteAddr() net.Addr {
	return s.stream.RemoteAddr()
}

func (s *WebSocketStream) SetDeadline(t time.Time) error {
	return s.stream.SetDeadline(t)
}

func (s *WebSocketStream) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}

func (s *WebSocketStream) SetWriteDeadline(t time.Time) error {
	return s.stream.SetWriteDeadline(t)
}
