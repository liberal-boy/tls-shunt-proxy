package config

import (
	"crypto/tls"
	"github.com/liberal-boy/tls-shunt-proxy/config/raw"
	"github.com/liberal-boy/tls-shunt-proxy/handler"
	"github.com/liberal-boy/tls-shunt-proxy/handler/http2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strings"
)

type (
	Config struct {
		Listen        string
		RedirectHttps string
		VHosts        map[string]VHost
	}
	VHost struct {
		TlsConfig    *tls.Config
		Http         handler.Handler
		Http2        handler.Handler
		PathHandlers []PathHandler
		Trojan       handler.Handler
		Default      handler.Handler
	}
	PathHandler struct {
		Path, TrimPrefix string
		Handler          handler.Handler
	}
)

func readRawConfig(path string) (conf raw.RawConfig, err error) {
	conf = raw.RawConfig{InboundBufferSize: 4, OutboundBufferSize: 32}
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

func ReadConfig(path string) (conf Config, err error) {
	rawConf, err := readRawConfig(path)
	if err != nil {
		return
	}

	handler.InitBufferPools(rawConf.InboundBufferSize*1024, rawConf.OutboundBufferSize*1024)

	conf.Listen = rawConf.Listen
	conf.RedirectHttps = rawConf.RedirectHttps
	conf.VHosts = make(map[string]VHost, len(rawConf.VHosts))

	for _, vh := range rawConf.VHosts {
		var tlsConfig *tls.Config

		if vh.TlsOffloading {
			tlsConfig, err = getTlsConfig(vh.ManagedCert, vh.Name, vh.Cert, vh.Key, vh.KeyType, vh.Alpn, vh.Protocols)
		}

		pathHandlers := make([]PathHandler, len(vh.Http.Paths))

		for i, p := range vh.Http.Paths {
			pathHandlers[i] = PathHandler{
				Path:       p.Path,
				TrimPrefix: p.TrimPrefix,
				Handler:    newHandler(p.Handler, p.Args),
			}
		}

		var http2Handler handler.Handler
		if len(vh.Http2) != 0 {
			http2Handler = http2.NewHttpMuxHandler(vh.Http2)
		} else {
			http2Handler = handler.NoopHandler
		}

		conf.VHosts[strings.ToLower(vh.Name)] = VHost{
			TlsConfig:    tlsConfig,
			Http:         newHandler(vh.Http.Handler, vh.Http.Args),
			PathHandlers: pathHandlers,
			Http2:        http2Handler,
			Trojan:       newHandler(vh.Trojan.Handler, vh.Trojan.Args),
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
	case "dohServer":
		return handler.NewDohServer(args)
	default:
		log.Fatalf("handler %s not supported\n", name)
	}
	return nil
}
