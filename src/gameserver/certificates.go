package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/sambdavidson/community-chess/src/lib/tlsca"
)

func gameSlaveTLSConfig(gameID string) (*tls.Config, error) {
	return tlsConfig(tlsca.GameSlave, gameID)
}

func gameMasterTLSConfig(gameID string) (*tls.Config, error) {
	return tlsConfig(tlsca.GameMaster, gameID)
}

func tlsConfig(serverType tlsca.SAN, gameID string) (*tls.Config, error) {
	certTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		DNSNames:     []string{serverType.String(), tlsca.GameServer.String(), gameID},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
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
		ServerName:   gameID,
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		RootCAs:      caPool,
		ClientCAs:    caPool,
	}, nil
}
