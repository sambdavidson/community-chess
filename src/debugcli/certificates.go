package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/sambdavidson/community-chess/src/lib/tlsca"
)

func clientTLSConfig() (*tls.Config, error) {
	certTmpl := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "debugcli",
		},
		SerialNumber: big.NewInt(time.Now().Unix()),
		DNSNames: []string{
			"localhost", // The address of services will need to be figured out and injected here.
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	certPEM, err := tlsca.SignCertificate(certTmpl, &certPrivKey.PublicKey)
	if err != nil {
		return nil, err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	cert, err := tls.X509KeyPair(certPEM, certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}
	caPool, err := tlsca.CAPool()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		RootCAs:      caPool,
		ClientCAs:    caPool,
	}, nil
}
