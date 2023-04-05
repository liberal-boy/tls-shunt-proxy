package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/caddyserver/certmagic"
	"github.com/go-acme/lego/v3/challenge/tlsalpn01"
	"log"
	"strings"
)

const (
	tlsDefaultMin = tls.VersionTLS12
	tlsDefaultMax = tls.VersionTLS13
)

func init() {
	certmagic.DefaultACME.Agreed = true
}

func getTlsConfig(managedCert bool, serverName string, handleWWW bool, cert, key, keyType, alpn, protocols string) (*tls.Config, error) {
	certificateFunc, err := getCertificateFunc(managedCert, serverName, handleWWW, cert, key, keyType)
	if err != nil {
		return nil, err
	}

	var min, max uint16
	min = tlsDefaultMin
	max = tlsDefaultMax
	if protocols != "" {
		ps := strings.Split(protocols, ",")
		min = getTlsVersion(ps[0])
		if len(ps) > 1 {
			max = getTlsVersion(ps[1])
		} else {
			max = getTlsVersion(ps[0])
		}
	}

	tlsConfig := &tls.Config{
		GetCertificate: certificateFunc,
		NextProtos:     append(strings.Split(alpn, ","), tlsalpn01.ACMETLS1Protocol),
		MinVersion:     min,
		MaxVersion:     max,
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

func getCertificateFunc(managedCert bool, serverName string, handleWWW bool, cert, key, keyType string) (func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error), error) {
	var keyGenerator = certmagic.DefaultKeyGenerator
	if keyType != "" {
		keyGenerator = certmagic.StandardKeyGenerator{KeyType: certmagic.KeyType(keyType)}
	}

	certMagicConfig := certmagic.Config{
		Storage:   &certmagic.FileStorage{Path: "./"},
		KeySource: keyGenerator,
	}

	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(certificate certmagic.Certificate) (c *certmagic.Config, err error) {
			return &certMagicConfig, nil
		},
	})

	magic := certmagic.New(cache, certMagicConfig)

	if managedCert {
		domainNames := []string{serverName}
		if handleWWW {
			domainNames = append(domainNames, fmt.Sprintf("%s%s", WWWPrefix, serverName))
		}
		err := magic.ManageAsync(context.Background(), domainNames)
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

func getTlsVersion(ver string) uint16 {
	switch ver {
	case "tls12":
		return tls.VersionTLS12
	case "tls13":
		return tls.VersionTLS13
	default:
		log.Fatalf("unsupported TLS version")
	}
	return 0
}
