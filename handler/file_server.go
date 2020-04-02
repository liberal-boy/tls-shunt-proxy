package handler

import (
	"log"
	"net"
	"net/http"
)

type FileServerHandler struct {
	connListener *ConnListener
}

func NewFileServerHandler(path string) *FileServerHandler {
	ln := NewConnListener()
	go func() {
		err := http.Serve(ln, http.FileServer(http.Dir(path)))
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return &FileServerHandler{connListener: ln}
}

func (h *FileServerHandler) Handle(conn net.Conn) {
	err := h.connListener.HandleConn(conn)
	if err != nil {
		log.Println(err)
	}
}
