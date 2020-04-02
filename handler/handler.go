package handler

import "net"

type Handler interface {
	Handle(conn net.Conn)
}
