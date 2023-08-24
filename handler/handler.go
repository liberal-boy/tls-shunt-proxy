package handler

import (
	"net"

	"github.com/nanmu42/gzip"
)

type Handler interface {
	Handle(conn net.Conn)
}

func DefaultGzipHandler() *gzip.Handler{
	return gzip.NewHandler(gzip.Config{
		CompressionLevel: gzip.BestSpeed,
		MinContentLength: 1 * 1024,
		RequestFilter: []gzip.RequestFilter{
			gzip.NewCommonRequestFilter(),
			gzip.DefaultExtensionFilter(),
		},
		ResponseHeaderFilter: []gzip.ResponseHeaderFilter{
			gzip.NewSkipCompressedFilter(),
			gzip.DefaultContentTypeFilter(),
		},
	})
}
