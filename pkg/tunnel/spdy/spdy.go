package spdy

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/moby/spdystream"

	"github.com/jellor/qtunnel/pkg/tunnel"
)

type SpdyListener struct {
	listener net.Listener
}

func Listen(address string) (tunnel.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	return &SpdyListener{listener: listener}, nil
}

func (l *SpdyListener) Accept() (tunnel.Connection, error) {
	conn, err := l.listener.Accept()
	if err != nil {
		return nil, err
	}

	spdyConn, err := spdystream.NewConnection(conn, true)
	if err != nil {
		return nil, err
	}

	spdyConnection, err := newSpdyConnection(spdyConn)
	if err != nil {
		return nil, err
	}
	return spdyConnection, nil
}

func (l *SpdyListener) Close() error {
	return l.listener.Close()
}

func (l *SpdyListener) Addr() net.Addr {
	return l.listener.Addr()
}

type SpdyDialer struct {
}

func (d *SpdyDialer) Dial(address string) (tunnel.Connection, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	spdyConn, err := spdystream.NewConnection(conn, false)
	if err != nil {
		return nil, err
	}
	spdyConnection, err := newSpdyConnection(spdyConn)
	if err != nil {
		return nil, err
	}
	return spdyConnection, nil
}

type SpdyConnection struct {
	conn          *spdystream.Connection
	newStreamChan chan *spdystream.Stream
	dieErr        error
	die           chan struct{}
	dieOnce       sync.Once
}

func newSpdyConnection(conn *spdystream.Connection) (*SpdyConnection, error) {
	spdyConn := &SpdyConnection{
		conn:          conn,
		newStreamChan: make(chan *spdystream.Stream, 4096),
		die:           make(chan struct{}),
	}

	go spdyConn.conn.Serve(spdyConn.newStreamHandler)

	return spdyConn, nil
}

func (c *SpdyConnection) newStreamHandler(stream *spdystream.Stream) {
	select {
	case <-c.die:
		return
	case c.newStreamChan <- stream:
	}
}

func (c *SpdyConnection) OpenStream() (tunnel.Stream, error) {
	stream, err := c.conn.CreateStream(http.Header{}, nil, false)
	if err != nil {
		return nil, err
	}
	err = stream.Wait()
	if err != nil {
		return nil, err
	}

	return &SpdyStream{stream: stream}, nil
}

func (c *SpdyConnection) AcceptStream() (tunnel.Stream, error) {
	select {
	case <-c.die:
		return nil, c.dieErr
	case stream, ok := <-c.newStreamChan:
		if ok {
			return &SpdyStream{stream: stream}, nil
		} else {
			return nil, c.dieErr
		}
	}
}

func (c *SpdyConnection) Close() error {
	var err error
	c.dieOnce.Do(func() {
		close(c.die)
		err = c.conn.Close()
	})
	return err
}

type SpdyStream struct {
	stream *spdystream.Stream
}

func (s *SpdyStream) Read(b []byte) (n int, err error) {
	return s.stream.Read(b)
}

func (s *SpdyStream) Write(b []byte) (n int, err error) {
	return s.stream.Write(b)
}

func (s *SpdyStream) Close() error {
	return s.stream.Close()
}

func (s *SpdyStream) LocalAddr() net.Addr {
	return s.stream.LocalAddr()
}

func (s *SpdyStream) RemoteAddr() net.Addr {
	return s.stream.RemoteAddr()
}

func (s *SpdyStream) SetDeadline(t time.Time) error {
	return s.stream.SetDeadline(t)
}

func (s *SpdyStream) SetReadDeadline(t time.Time) error {
	return s.stream.SetReadDeadline(t)
}

func (s *SpdyStream) SetWriteDeadline(t time.Time) error {
	return s.stream.SetWriteDeadline(t)
}
