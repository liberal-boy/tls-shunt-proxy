package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/liberal-boy/tls-shunt-proxy/config"
	"github.com/liberal-boy/tls-shunt-proxy/handler"
	"github.com/liberal-boy/tls-shunt-proxy/sniffer"
	"github.com/stevenjohnstone/sni"
	"log"
	"net"
	"strings"
)

const version = "0.9.2"

var conf config.Config

func main() {
	fmt.Println("tls-shunt-proxy version", version)

	configPath := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	var err error
	conf, err = config.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to read config %s: %v", *configPath, err)
	}

	if conf.RedirectHttps != "" {
		handler.ServeRedirectHttps(conf.RedirectHttps)
	}

	listenAndServe()
}

func listenAndServe() {
	ln, err := net.Listen("tcp", conf.Listen)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", conf.Listen, err)
	}
	defer func() { _ = ln.Close() }()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("fail to establish conn: %v\n", err)
			continue
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	serverName, sniConn, err := sni.ServerNameFromConn(conn)
	if err != nil {
		if conf.Fallback != handler.NoopHandler {
			conf.Fallback.Handle(sniConn)
		} else {
			log.Printf("fail to obtain server name: %v\n", err)
			handler.NewPlainTextHandler(handler.SentHttpToHttps).Handle(conn)
		}
		return
	}

	handleWithServerName(sniConn, serverName)
}

func handleWithServerName(conn net.Conn, serverName string) {
	vh, has := conf.VHosts[strings.ToLower(serverName)]
	if !has {
		if conf.Fallback != handler.NoopHandler {
			conf.Fallback.Handle(conn)
		} else {
			log.Printf("no available vhost for %s\n", serverName)
			handler.NewPlainTextHandler(handler.NoCertificateAvailable).Handle(conn)
		}
		return
	}

	if vh.TlsConfig != nil {
		conn = tlsOffloading(conn, vh.TlsConfig)
		sniffConn := sniffer.NewPeekPreDataConn(conn)
		conn = sniffConn

		switch sniffConn.Type {
		case sniffer.TypeHttp:
			if handleHttp(sniffConn, vh) {
				return
			}
		case sniffer.TypeHttp2:
			if handleHttp2(sniffConn, vh) {
				return
			}
		case sniffer.TypeTrojan:
			if handleTrojan(sniffConn, vh) {
				return
			}
		}
	}
	vh.Default.Handle(conn)
}

func handleHttp(conn *sniffer.SniffConn, vh config.VHost) bool {
	for _, p := range vh.PathHandlers {
		if strings.HasPrefix(conn.GetPath(), p.Path) {
			conn.SetPath(strings.TrimPrefix(conn.GetPath(), p.TrimPrefix))
			p.Handler.Handle(conn)
			return true
		}
	}

	if vh.Http != handler.NoopHandler {
		vh.Http.Handle(conn)
		return true
	}

	return false
}

func handleHttp2(conn *sniffer.SniffConn, vh config.VHost) bool {
	if vh.Http2 != handler.NoopHandler {
		vh.Http2.Handle(conn)
		return true
	}
	return handleHttp(conn, vh)
}

func handleTrojan(conn *sniffer.SniffConn, vh config.VHost) bool {
	if vh.Trojan != handler.NoopHandler {
		vh.Trojan.Handle(conn)
		return true
	}
	return false
}

func tlsOffloading(conn net.Conn, tlsConfig *tls.Config) *tls.Conn {
	return tls.Server(conn, tlsConfig)
}
