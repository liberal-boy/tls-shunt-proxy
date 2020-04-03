package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/caddyserver/certmagic"
	"github.com/go-acme/lego/v3/challenge/tlsalpn01"
)

func init() {
	certmagic.DefaultACME.Agreed = true
}

func getTlsConfig(managedCert bool, serverName, cert, key string) (*tls.Config, error) {
	certificateFunc, err := getCertificateFunc(managedCert, serverName, cert, key)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		GetCertificate: certificateFunc,
		NextProtos:     []string{"http/1.1", tlsalpn01.ACMETLS1Protocol},
		MinVersion:     tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_AES_128_GCM_SHA256,
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,

			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
	}
	return tlsConfig, nil
}

func getCertificateFunc(managedCert bool, serverName, cert, key string) (func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error), error) {
	config := certmagic.Config{
		Storage: &certmagic.FileStorage{Path: "./"},
	}

	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(certificate certmagic.Certificate) (c *certmagic.Config, err error) {
			return &config, nil
		},
	})

	magic := certmagic.New(cache, config)

	if managedCert {
		err := magic.ManageAsync(context.Background(), []string{serverName})
		if err != nil {
			return nil, err
		}
	} else {
		err := magic.CacheUnmanagedCertificatePEMFile(cert, key, nil)
		if err != nil {
			err = fmt.Errorf("fail to load tls key pair for %s: %v", serverName, err)
			return nil, err
		}
	}

	return magic.GetCertificate, nil
}
