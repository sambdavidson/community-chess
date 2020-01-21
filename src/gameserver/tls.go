package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
)

// TLS
var (
	caBundlePath         = flag.String("ca_bundle_path", "", "path to CA bundle for validating TLS connections")
	masterCertPath       = flag.String("master_cert_path", "", "path to the master gameserver TLS certificate, if enabled")
	masterPrivateKeyPath = flag.String("master_private_key_path", "", "path to the master gameserver TLS private key, if enabled")
	slaveCertPath        = flag.String("slave_cert_path", "", "path to the slave gameserver TLS certificate, if enabled")
	slavePrivateKeyPath  = flag.String("slave_private_key_path", "", "path to the master gameserver TLS private key, if enabled")
)

func gameSlaveTLSConfig() (*tls.Config, error) {
	return tlsConfig(*slaveCertPath, *slavePrivateKeyPath)
}

func gameMasterTLSConfig() (*tls.Config, error) {
	return tlsConfig(*masterCertPath, *masterPrivateKeyPath)
}

func tlsConfig(certPath, privPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, privPath)
	if err != nil {
		return nil, fmt.Errorf("failed loading X509KeyPair: %v", err)
	}

	caPool := x509.NewCertPool()
	caPEM, err := ioutil.ReadFile(*caBundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading CA bundle file: %v", err)
	}
	if ok := caPool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("appending CA cert to cert pool not ok")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		RootCAs:      caPool,
		ClientCAs:    caPool,
	}, nil
}
