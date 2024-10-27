package kcp

import (
	"net"
	"time"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"

	"github.com/jellor/qtunnel/pkg/tunnel"
)

type KcpListener struct {
	listener net.Listener
}

func Listen(address string) (tunnel.Listener, error) {
	listener, err := kcp.Listen(address)
	if err != nil {
		return nil, err
	}
	return &KcpListener{listener: listener}, nil
}

func (l *KcpListener) Accept() (tunnel.Connection, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}
	session, err := smux.Server(conn, nil)
	if err != nil {
		return nil, err
	}
	return &KcpConnection{conn: conn, session: session}, nil
}

func (l *KcpListener) Close() error {
	return l.listener.Close()
}

func (l *KcpListener) Addr() net.Addr {
	return l.listener.Addr()
}

type KcpDialer struct {
}

func (d *KcpDialer) Dial(address string) (tunnel.Connection, error) {
	conn, err := kcp.Dial(address)
	if err != nil {
		return nil, err
	}

	session, err := smux.Client(conn, nil)
	if err != nil {
		return nil, err
	}

	return &KcpConnection{conn: conn, session: session}, nil
}

type KcpConnection struct {
	conn    net.Conn
	session *smux.Session
}

func (c *KcpConnection) OpenStream() (tunnel.Stream, error) {
	stream, err := c.session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &KcpStream{stream: stream}, nil
}

func (c *KcpConnection) AcceptStream() (tunnel.Stream, error) {
	stream, err := c.session.AcceptStream()
	if err != nil {
		return nil, err
	}
	return &KcpStream{stream: stream}, nil
}

func (c *KcpConnection) Close() error {
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
