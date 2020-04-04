package sniffer

import (
	"io"
	"net"
	"strings"
)

var httpMethods = [...]string{"GET", "POST", "HEAD", "PUT", "DELETE", "OPTIONS", "CONNECT", "PRI"}

func SniffHttpFromConn(conn net.Conn) (bool, string, net.Conn) {
	preConn := newPeekPreDataConn(conn)
	preDate, err := preConn.PeekPreData(64)
	if err != nil && err != io.EOF {
		return false, "", preConn
	}
	isHttp, path := sniff(preDate)
	return isHttp, path, preConn
}

func sniff(b []byte) (isHttp bool, path string) {
	parts := strings.Split(string(b), " ")

	if len(parts) < 2 {
		return
	}

	for _, m := range httpMethods {
		if parts[0] == m {
			isHttp = true
			break
		}
	}

	path = parts[1]

	return
}
