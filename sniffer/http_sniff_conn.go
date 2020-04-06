package sniffer

import (
	"bytes"
	"errors"
	"io"
	"net"
)

type HttpSniffConn struct {
	net.Conn
	rout         io.Reader
	peeked, read bool
	preDataParts [][]byte
}

var (
	httpMethods = [...][]byte{
		[]byte("GET"),
		[]byte("POST"),
		[]byte("HEAD"),
		[]byte("PUT"),
		[]byte("DELETE"),
		[]byte("OPTIONS"),
		[]byte("CONNECT"),
		[]byte("PRI"),
	}
	sep = []byte(" ")
)

func NewPeekPreDataConn(c net.Conn) *HttpSniffConn {
	return &HttpSniffConn{Conn: c, rout: c}
}

func (c *HttpSniffConn) peekPreData(n int) ([]byte, error) {
	if c.read {
		return nil, errors.New("pre-data must be peek before read")
	}
	if c.peeked {
		return nil, errors.New("can only peek once")
	}
	c.peeked = true
	preDate := make([]byte, n)
	n, err := c.Conn.Read(preDate)
	return preDate[:n], err
}

func (c *HttpSniffConn) Read(p []byte) (int, error) {
	if !c.read {
		c.read = true
		preDate := bytes.Join(c.preDataParts, sep)
		c.preDataParts = nil
		c.rout = io.MultiReader(bytes.NewReader(preDate), c.Conn)
	}
	return c.rout.Read(p)
}

func (c *HttpSniffConn) Sniff() bool {
	preDate, err := c.peekPreData(64)

	c.preDataParts = bytes.Split(preDate, sep)

	if err != nil && err != io.EOF {
		return false
	}

	if len(c.preDataParts) < 2 {
		return false
	}

	for _, m := range httpMethods {
		if bytes.Compare(c.preDataParts[0], m) == 0 {
			return true
		}
	}
	return false
}

func (c *HttpSniffConn) SetPath(path string) {
	c.preDataParts[1] = []byte(path)
}

func (c *HttpSniffConn) GetPath() string {
	return string(c.preDataParts[1])
}
