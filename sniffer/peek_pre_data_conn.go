package sniffer

import (
	"bytes"
	"errors"
	"io"
	"net"
)

type peekPreDataConn struct {
	net.Conn
	rout         io.Reader
	peeked, read bool
}

func newPeekPreDataConn(c net.Conn) *peekPreDataConn {
	return &peekPreDataConn{Conn: c, rout: c}
}

func (c *peekPreDataConn) PeekPreData(n int) ([]byte, error) {
	if c.read {
		return nil, errors.New("pre-data must be peek before read")
	}
	if c.peeked {
		return nil, errors.New("can only peek once")
	}
	c.peeked = true
	preDate := make([]byte, n)
	n, err := c.Conn.Read(preDate)
	c.rout = io.MultiReader(bytes.NewReader(preDate[:n]), c.Conn)
	return preDate[:n], err
}

func (c *peekPreDataConn) Read(p []byte) (int, error) {
	c.read = true
	return c.rout.Read(p)
}
