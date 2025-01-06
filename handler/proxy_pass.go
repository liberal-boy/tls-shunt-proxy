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
	if conn == nil {
		return
	}

	defer conn.Close()

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
	defer dstConn.Close()

	if h.proxyProtocol {
		if err := h.sendProxyProtocol(conn, dstConn); err != nil {
			log.Printf("fail to send proxy %s :%v\n", h.target, err)
			return
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go doCopy(dstConn, conn, inboundBufferPool, &wg)
	go doCopy(conn, dstConn, outboundBufferPool, &wg)

	wg.Wait()
}

func (h *ProxyPassHandler) sendProxyProtocol(srcConn, dstConn net.Conn) error {
	remoteAddr, remotePort, err := net.SplitHostPort(srcConn.RemoteAddr().String())
	if err != nil {
		return err
	}
	localAddr, localPort, err := net.SplitHostPort(srcConn.LocalAddr().String())
	if err != nil {
		return err
	}

	ipVer := "4"
	if strings.Contains(remoteAddr, ":") {
		ipVer = "6"
	}
	_, err = fmt.Fprintf(dstConn, "PROXY TCP%s %s %s %s %s\r\n", ipVer, remoteAddr, localAddr, remotePort, localPort)
	return err
}

func doCopy(dst io.Writer, src io.Reader, bufferPool *sync.Pool, wg *sync.WaitGroup) {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	_, err := io.CopyBuffer(dst, src, buf)
	if err != nil && err != io.EOF {
		log.Printf("failed to proxy pass: %v\n", err)
	}
	wg.Done()
}
