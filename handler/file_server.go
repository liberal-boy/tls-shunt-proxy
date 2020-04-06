package handler

import (
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
		h2s := &http2.Server{}
		fileServer := http.FileServer(http.Dir(path))
		err := http.Serve(ln, h2c.NewHandler(fileServer, h2s))
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
