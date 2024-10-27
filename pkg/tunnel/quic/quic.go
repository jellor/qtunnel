package quic

import (
	"context"
	"net"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/jellor/qtunnel/pkg/tunnel"
)

type QuicListener struct {
	listener *quic.Listener
}

func Listen(address string) (tunnel.Listener, error) {
	listener, err := quic.ListenAddr(address, nil, nil)
	if err != nil {
		return nil, err
	}
	return &QuicListener{listener: listener}, nil
}

func (l *QuicListener) Accept() (tunnel.Connection, error) {
	conn, err := l.listener.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	return &QuicConnection{conn: conn}, nil
}

func (l *QuicListener) Close() error {
	return l.listener.Close()
}

func (l *QuicListener) Addr() net.Addr {
	return l.listener.Addr()
}

type QuicDialer struct {
}

func (d *QuicDialer) Dial(address string) (tunnel.Connection, error) {
	conn, err := quic.DialAddr(context.Background(), address, nil, nil)
	if err != nil {
		return nil, err
	}

	return &QuicConnection{conn: conn}, nil
}

type QuicConnection struct {
	conn quic.Connection
}

func (c *QuicConnection) OpenStream() (tunnel.Stream, error) {
	stream, err := c.conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}
	return &QuicStream{stream: stream, localAddr: c.conn.LocalAddr(), remoteAddr: c.conn.RemoteAddr()}, nil
}

func (c *QuicConnection) AcceptStream() (tunnel.Stream, error) {
	stream, err := c.conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}
	return &QuicStream{stream: stream, localAddr: c.conn.LocalAddr(), remoteAddr: c.conn.RemoteAddr()}, nil
}

func (c *QuicConnection) Close() error {
	return c.conn.CloseWithError(0, "")
}

type QuicStream struct {
	stream     quic.Stream
	localAddr  net.Addr
	remoteAddr net.Addr
}

func (s *QuicStream) Read(b []byte) (n int, err error) {
	return s.stream.Read(b)
}

func (s *QuicStream) Write(b []byte) (n int, err error) {
	return s.stream.Write(b)
}

func (s *QuicStream) Close() error {
	return s.stream.Close()
}

func (s *QuicStream) LocalAddr() net.Addr {
	return s.localAddr
}

func (s *QuicStream) RemoteAddr() net.Addr {
	return s.remoteAddr
}

func (s *QuicStream) SetDeadline(t time.Time) error {
	return s.stream.SetDeadline(t)
}

func (s *QuicStream) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}

func (s *QuicStream) SetWriteDeadline(t time.Time) error {
	return s.stream.SetWriteDeadline(t)
}
