package handler

import (
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type ProxyPassHandler struct {
	target string
}

var bufPool = sync.Pool{New: func() interface{} {
	return make([]byte, 32*1024)
}}

func NewProxyPassHandler(target string) *ProxyPassHandler {
	return &ProxyPassHandler{target: target}
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

	var wg sync.WaitGroup
	wg.Add(2)

	go doCopy(dstConn, conn, &wg)
	go doCopy(conn, dstConn, &wg)

	wg.Wait()
}

func doCopy(dst io.Writer, src io.Reader, wg *sync.WaitGroup) {
	buf := bufPool.Get().([]byte)
	defer bufPool.Put(buf)
	_, err := io.CopyBuffer(dst, src, buf)
	if err != nil && err != io.EOF {
		log.Printf("failed to proxy pass: %v\n", err)
	}
	wg.Done()
}
