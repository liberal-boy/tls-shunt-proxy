package main

import (
	"crypto/tls"
	"github.com/liberal-boy/tls-shunt-proxy/handler"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type (
	rawConfig struct {
		Listen string
		VHosts []rawVHost
	}
	rawVHost struct {
		Name          string
		TlsOffloading bool
		ManagedCert   bool
		Cert          string
		Key           string
		Http          rawHandler
		Default       rawHandler
	}
	rawHandler struct {
		Handler string
		Args    string
	}
)

type (
	config struct {
		Listen string
		vHosts map[string]vHost
	}
	vHost struct {
		TlsConfig *tls.Config
		Http      handler.Handler
		Default   handler.Handler
	}
)

func readRawConfig(path string) (conf rawConfig, err error) {
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

	conf.Listen = rawConf.Listen
	conf.vHosts = make(map[string]vHost, len(rawConf.VHosts))

	for _, vh := range rawConf.VHosts {
		var tlsConfig *tls.Config

		if vh.TlsOffloading {
			tlsConfig, err = getTlsConfig(vh.ManagedCert, vh.Name, vh.Cert, vh.Key)
		}

		conf.vHosts[vh.Name] = vHost{
			TlsConfig: tlsConfig,
			Http:      newHandler(vh.Http.Handler, vh.Http.Args),
			Default:   newHandler(vh.Default.Handler, vh.Default.Args),
		}
	}
	return
}

func newHandler(name, args string) handler.Handler {
	switch name {
	case "":
		return nil
	case "proxyPass":
		return handler.NewProxyPassHandler(args)
	case "fileServer":
		return handler.NewFileServerHandler(args)
	default:
		log.Fatalf("handler %s not supported\n", name)
	}
	return nil
}
