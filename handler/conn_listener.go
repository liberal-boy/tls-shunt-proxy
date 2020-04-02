package handler

import (
	"errors"
	"net"
)

var ErrNetClosing = errors.New("use of closed network connection")

type ConnListener struct {
	isClosed bool
	connChan chan net.Conn
	exitChan chan struct{}
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
	if c.isClosed {
		return ErrNetClosing
	}
	c.isClosed = true
	c.exitChan <- struct{}{}
	close(c.exitChan)
	close(c.connChan)
	return nil
}

func (c *ConnListener) Addr() net.Addr {
	return nil
}

func (c *ConnListener) HandleConn(conn net.Conn) error {
	if c.isClosed {
		return ErrNetClosing
	}
	c.connChan <- conn
	return nil
}
