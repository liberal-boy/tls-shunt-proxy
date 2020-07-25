package main

import (
	"crypto/tls"
	"github.com/liberal-boy/tls-shunt-proxy/handler"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

type (
	rawConfig struct {
		Listen                                string
		InboundBufferSize, OutboundBufferSize int
		VHosts                                []rawVHost
	}
	rawVHost struct {
		Name          string
		TlsOffloading bool
		ManagedCert   bool
		Cert          string
		Key           string
		KeyType       string
		Alpn          string
		Protocols     string
		Http          rawHttpHandler
		Default       rawHandler
	}
	rawHandler struct {
		Handler string
		Args    string
	}
	rawHttpHandler struct {
		Paths   []rawPathHandler
		Handler string
		Args    string
	}
	rawPathHandler struct {
		Path       string
		Handler    string
		Args       string
		TrimPrefix string
	}
)

type (
	config struct {
		Listen string
		vHosts map[string]vHost
	}
	vHost struct {
		TlsConfig    *tls.Config
		Http         handler.Handler
		PathHandlers []pathHandler
		Default      handler.Handler
	}
	pathHandler struct {
		path, trimPrefix string
		handler          handler.Handler
	}
)

func readRawConfig(path string) (conf rawConfig, err error) {
	conf = rawConfig{InboundBufferSize: 4, OutboundBufferSize: 32}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return
	}
	return
}

func readConfig(path string) (conf config, err error) {
	rawConf, err := readRawConfig(path)
	if err != nil {
		return
	}

	handler.InitBufferPools(rawConf.InboundBufferSize*1024, rawConf.OutboundBufferSize*1024)

	conf.Listen = rawConf.Listen
	conf.vHosts = make(map[string]vHost, len(rawConf.VHosts))

	for _, vh := range rawConf.VHosts {
		var tlsConfig *tls.Config

		if vh.TlsOffloading {
			tlsConfig, err = getTlsConfig(vh.ManagedCert, vh.Name, vh.Cert, vh.Key, vh.KeyType, vh.Alpn, vh.Protocols)
		}

		pathHandlers := make([]pathHandler, len(vh.Http.Paths))

		for i, p := range vh.Http.Paths {
			pathHandlers[i] = pathHandler{
				path:       p.Path,
				trimPrefix: p.TrimPrefix,
				handler:    newHandler(p.Handler, p.Args),
			}
		}

		conf.vHosts[strings.ToLower(vh.Name)] = vHost{
			TlsConfig:    tlsConfig,
			Http:         newHandler(vh.Http.Handler, vh.Http.Args),
			PathHandlers: pathHandlers,
			Default:      newHandler(vh.Default.Handler, vh.Default.Args),
		}
	}
	return
}

func newHandler(name, args string) handler.Handler {
	switch name {
	case "":
		return handler.NoopHandler
	case "proxyPass":
		return handler.NewProxyPassHandler(args)
	case "fileServer":
		return handler.NewFileServerHandler(args)
	default:
		log.Fatalf("handler %s not supported\n", name)
	}
	return nil
}
