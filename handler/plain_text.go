package handler

import (
	"log"
	"net"
)

var (
	SentHttpToHttps = []byte("HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.")

	// TLSv1.2 Record Layer: Alert (Level: Fatal, Description: Internal Error)
	NoCertificateAvailable = []byte{0x15, 0x03, 0x03, 0x00, 0x02, 0x02, 0x50}
)

type PlainTextHandler struct {
	text []byte
}

func NewPlainTextHandler(text []byte) *PlainTextHandler {
	return &PlainTextHandler{text: text}
}

func (h *PlainTextHandler) Handle(conn net.Conn) {
	_, err := conn.Write(h.text)
	if err != nil {
		log.Println("fail to serve plain text: ", err)
	}
	_ = conn.Close()
}
