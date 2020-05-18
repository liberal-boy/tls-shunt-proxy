package handler

import "net"

var NoopHandler = new(noopHandler)

type noopHandler struct{}

func (h *noopHandler) Handle(conn net.Conn) {
	_ = conn.Close()
}
