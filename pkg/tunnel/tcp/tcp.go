package tcp

import (
	"net"
	"time"

	"github.com/xtaci/smux"

	"github.com/jellor/qtunnel/pkg/tunnel"
)

type TcpListener struct {
	listener net.Listener
}

func Listen(address string) (tunnel.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TcpListener{listener: listener}, nil
}

func (l *TcpListener) Accept() (tunnel.Connection, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	session, err := smux.Server(conn, nil)
	if err != nil {
		return nil, err
	}
	return &TcpConnection{conn: conn, session: session}, nil
}

func (l *TcpListener) Close() error {
	return l.listener.Close()
}

func (l *TcpListener) Addr() net.Addr {
	return l.listener.Addr()
}

type TcpDialer struct {
}

func (d *TcpDialer) Dial(address string) (tunnel.Connection, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	session, err := smux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	return &TcpConnection{conn: conn, session: session}, nil
}

type TcpConnection struct {
	conn    net.Conn
	session *smux.Session
}

func (c *TcpConnection) OpenStream() (tunnel.Stream, error) {
	stream, err := c.session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &KcpStream{stream: stream}, nil
}

func (c *TcpConnection) AcceptStream() (tunnel.Stream, error) {
	stream, err := c.session.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &KcpStream{stream: stream}, nil
}

func (c *TcpConnection) Close() error {
	return c.session.Close()
}

type KcpStream struct {
	stream *smux.Stream
}

func (s *KcpStream) Read(b []byte) (n int, err error) {
	return s.stream.Read(b)
}

func (s *KcpStream) Write(b []byte) (n int, err error) {
	return s.stream.Write(b)
}

func (s *KcpStream) Close() error {
	return s.stream.Close()
}

func (s *KcpStream) LocalAddr() net.Addr {
	return s.stream.LocalAddr()
}

func (s *KcpStream) RemoteAddr() net.Addr {
	return s.stream.RemoteAddr()
}

func (s *KcpStream) SetDeadline(t time.Time) error {
	return s.stream.SetDeadline(t)
}

func (s *KcpStream) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}

func (s *KcpStream) SetWriteDeadline(t time.Time) error {
	return s.stream.SetWriteDeadline(t)
}
