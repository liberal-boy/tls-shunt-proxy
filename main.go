package main

import (
	"crypto/tls"
	"flag"
	"github.com/liberal-boy/tls-shunt-proxy/sniffer"
	"github.com/stevenjohnstone/sni"
	"log"
	"net"
)

var conf config

func main() {
	config := flag.String("config", "./config.yaml", "Path to config file")
	flag.Parse()

	var err error
	conf, err = readConfig(*config)
	if err != nil {
		log.Fatalf("failed to read config %s: %v", *config, err)
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
	serverName, conn, err := sni.ServerNameFromConn(conn)
	if err != nil {
		log.Printf("fail to obtain server name: %v\n", err)
		return
	}

	handleWithServerName(conn, serverName)
}

func handleWithServerName(conn net.Conn, serverName string) {
	vh, has := conf.vHosts[serverName]
	if !has {
		log.Printf("no available vhost for %s\n", serverName)
		return
	}

	isHttp := false

	if vh.TlsConfig != nil {
		conn = tlsOffloading(conn, vh.TlsConfig)
		isHttp, conn = sniffer.SniffHttpFromConn(conn)
	}

	if isHttp && vh.Http != nil {
		vh.Http.Handle(conn)
		return
	}
	vh.Default.Handle(conn)
}

func tlsOffloading(conn net.Conn, tlsConfig *tls.Config) *tls.Conn {
	return tls.Server(conn, tlsConfig)
}
