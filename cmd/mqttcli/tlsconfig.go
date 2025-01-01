// tlsconfig.go
package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
)

// NewTLSConfig loads CA, client cert, and key files into a tls.Config.
// If insecure is true, it won't verify the server's certificate.
func NewTLSConfig(caFile, certFile, keyFile string, insecure bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
		MinVersion:         tls.VersionTLS12,
	}

	// If CA file is provided, load it so the client trusts that root CA
	if caFile != "" {
		certs := x509.NewCertPool()
		ca, err := ioutil.ReadFile(caFile)
		if err != nil {
			return nil, err
		}
		if !certs.AppendCertsFromPEM(ca) {
			return nil, errors.New("failed to append CA certificate")
		}
		tlsConfig.RootCAs = certs
	}

	// If client certificate & key are provided, use mutual TLS
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
