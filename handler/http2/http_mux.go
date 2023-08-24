package http2

import (
	"github.com/liberal-boy/tls-shunt-proxy/config/raw"
	"github.com/liberal-boy/tls-shunt-proxy/handler"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net"
	"net/http"
)

type HttpMuxHandler struct {
	connListener *handler.ConnListener
}

func NewHttpMuxHandler(paths []raw.RawPathHandler) *HttpMuxHandler {
	ln := handler.NewConnListener()
	go func() {
		h2s := &http2.Server{}
		mux := http.NewServeMux()

		for _, path := range paths {
			mux.Handle(path.Path, newHandler(path.Handler, path.Args))
		}

		withGzip := handler.DefaultGzipHandler().WrapHandler(mux)

		err := http.Serve(ln, h2c.NewHandler(withGzip, h2s))
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return &HttpMuxHandler{connListener: ln}
}

func (h *HttpMuxHandler) Handle(conn net.Conn) {
	err := h.connListener.HandleConn(conn)
	if err != nil {
		log.Println(err)
	}
}

func newHandler(name, args string) http.Handler {
	switch name {
	case "proxyPass":
		return NewProxyPassHandler(args)
	case "fileServer":
		return http.FileServer(http.Dir(args))
	default:
		log.Fatalf("handler %s not supported\n", name)
	}
	return nil
}
