package main

import (
	"bufio"
	"io"
	"net"
	"strings"
)

type bufferedConn struct {
	r    *bufio.Reader
	rout io.Reader
	net.Conn
}

var httpMethods = [...]string{"GET", "POST", "HEAD", "PUT", "DELETE", "OPTIONS", "CONNECT", "PRI"}

func newBufferedConn(c net.Conn) bufferedConn {
	return bufferedConn{bufio.NewReader(c), nil, c}
}

func (b bufferedConn) Peek(n int) ([]byte, error) {
	return b.r.Peek(n)
}

func (b bufferedConn) Read(p []byte) (int, error) {
	if b.rout != nil {
		return b.rout.Read(p)
	}
	return b.r.Read(p)
}

func sniffHttpFromConn(conn net.Conn) (bool, net.Conn) {
	bufConn := newBufferedConn(conn)
	b, err := bufConn.Peek(8)
	if err != nil {
		return false, bufConn
	}
	header := string(b)
	for _, m := range httpMethods {
		if strings.HasPrefix(header, m+" ") {
			return true, bufConn
		}
	}
	return false, bufConn
}
