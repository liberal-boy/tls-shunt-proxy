package handler

import (
	"errors"
	"net"
	"sync"
)

var ErrNetClosing = errors.New("use of closed network connection")

type ConnListener struct {
	isClosed bool
	connChan chan net.Conn
	exitChan chan struct{}
	mu       sync.Mutex
	once     sync.Once
}

func NewConnListener() *ConnListener {
	return &ConnListener{
		connChan: make(chan net.Conn),
	}
}

func (c *ConnListener) Accept() (net.Conn, error) {
	if c.isClosed {
		return nil, ErrNetClosing
	}

	select {
	case <-c.exitChan:
		return nil, ErrNetClosing
	case conn := <-c.connChan:
		return conn, nil
	}
}

func (c *ConnListener) Close() error {
	c.once.Do(func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		if c.isClosed {
			return
		}
		c.isClosed = true
		close(c.exitChan)
		close(c.connChan)
	})
	return nil
}

func (c *ConnListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4zero, Port: 0}
}

func (c *ConnListener) HandleConn(conn net.Conn) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return ErrNetClosing
	}
	c.connChan <- conn
	return nil
}
