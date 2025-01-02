package handler

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type ProxyPassHandler struct {
	target        string
	proxyProtocol bool
}

var inboundBufferPool, outboundBufferPool *sync.Pool

func InitBufferPools(inboundBufferSize, outboundBufferSize int) {
	inboundBufferPool = newBufferPool(inboundBufferSize)
	outboundBufferPool = newBufferPool(outboundBufferSize)
}

func newBufferPool(size int) *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return make([]byte, size)
	}}
}

func NewProxyPassHandler(args string) *ProxyPassHandler {
	handler := ProxyPassHandler{}
	parts := strings.Split(args, ";")
	handler.target = parts[0]
	for _, arg := range parts {
		arg = strings.TrimSpace(arg)
		if arg == "proxyProtocol" {
			handler.proxyProtocol = true
		}
	}
	return &handler
}

func (h *ProxyPassHandler) Handle(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	var err error

	var dstConn net.Conn
	if strings.HasPrefix(h.target, "unix:") {
		dstConn, err = net.Dial("unix", h.target[5:])
	} else {
		dstConn, err = net.Dial("tcp", h.target)
	}
	if err != nil {
		log.Printf("fail to connect to %s :%v\n", h.target, err)
		return
	}
	defer func() { _ = dstConn.Close() }()

	if h.proxyProtocol {
		remoteAddr, remotePort, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			log.Printf("fail to send proxy %s :%v\n", h.target, err)
			return
		}
		localAddr, localPort, err := net.SplitHostPort(conn.LocalAddr().String())
		if err != nil {
			log.Printf("fail to send proxy %s :%v\n", h.target, err)
			return
		}

		var ipVer string
		if strings.Contains(remoteAddr, ":") {
			ipVer = "6"
		} else {
			ipVer = "4"
		}
		_, err = fmt.Fprintf(dstConn, "PROXY TCP%s %s %s %s %s\r\n", ipVer, remoteAddr, localAddr, remotePort, localPort)
		if err != nil {
			log.Printf("fail to send proxy %s :%v\n", h.target, err)
			return
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		if tcpConn, ok := dstConn.(*net.TCPConn); ok {
			tcpConn.ReadFrom(conn)
		} else {
			doCopy(dstConn, conn, inboundBufferPool)
		}
		wg.Done()
	}()
	go func() {
		if tcpConn, ok := dstConn.(*net.TCPConn); ok {
			tcpConn.WriteTo(conn)
		} else {
			doCopy(conn, dstConn, outboundBufferPool)
		}
		wg.Done()
	}()

	wg.Wait()
}

func doCopy(dst io.Writer, src io.Reader, bufferPool *sync.Pool) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	_, err := io.CopyBuffer(dst, src, buf)
	if err != nil && err != io.EOF {
		log.Printf("failed to proxy pass: %v\n", err)
	}
}
