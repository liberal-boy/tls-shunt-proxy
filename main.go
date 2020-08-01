package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/liberal-boy/tls-shunt-proxy/handler"
	"github.com/liberal-boy/tls-shunt-proxy/sniffer"
	"github.com/stevenjohnstone/sni"
	"log"
	"net"
	"strings"
)

const version = "0.5.3"

var conf config

func main() {
	fmt.Println("tls-shunt-proxy version", version)

	config := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	var err error
	conf, err = readConfig(*config)
	if err != nil {
		log.Fatalf("failed to read config %s: %v", *config, err)
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
		log.Printf("fail to obtain server name: %v\n", err)
		handler.NewPlainTextHandler(handler.SentHttpToHttps).Handle(conn)
		return
	}

	handleWithServerName(sniConn, serverName)
}

func handleWithServerName(conn net.Conn, serverName string) {
	vh, has := conf.vHosts[strings.ToLower(serverName)]
	if !has {
		log.Printf("no available vhost for %s\n", serverName)
		handler.NewPlainTextHandler(handler.NoCertificateAvailable).Handle(conn)
		return
	}

	if vh.TlsConfig != nil {
		conn = tlsOffloading(conn, vh.TlsConfig)
		sniffConn := sniffer.NewPeekPreDataConn(conn)
		conn = sniffConn

		if isHttp := sniffConn.Sniff(); isHttp {
			if handleHttp(sniffConn, vh) {
				return
			}
		}
	}
	vh.Default.Handle(conn)
}

func handleHttp(conn *sniffer.HttpSniffConn, vh vHost) bool {
	for _, p := range vh.PathHandlers {
		if strings.HasPrefix(conn.GetPath(), p.path) {
			conn.SetPath(strings.TrimPrefix(conn.GetPath(), p.trimPrefix))
			p.handler.Handle(conn)
			return true
		}
	}

	if vh.Http != handler.NoopHandler {
		vh.Http.Handle(conn)
		return true
	}

	return false
}

func tlsOffloading(conn net.Conn, tlsConfig *tls.Config) *tls.Conn {
	return tls.Server(conn, tlsConfig)
}
