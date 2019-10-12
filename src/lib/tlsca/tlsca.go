// Package tlsca provides a simple TLS CA that will be used for basic auth between services for now. Provides packages for programs
// to get their own temporary dev CA signed by a common CA.
package tlsca

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	caCertFileName       = "communityChessCACert.pem"
	caPrivateKeyFileName = "communityChessCAPrivateKey.pem"
)

var (
	caDirectory = flag.String("ca_directory", os.TempDir(), "Optional directory to store CA certificate and private key. Defaults to os.TempDir().")
)

var (
	privateKey *rsa.PrivateKey
	ca         *x509.Certificate
)

// Initializes the CA if it doesn't already exist.
func init() {
	log.Println("Initializing TLS CA library...")
	if err := loadCAFiles(); err != nil {
		log.Printf("Error loading existing CA files: %v\n", err)
		if _, err = RekeyCA(); err != nil {
			log.Fatalf("Unable rekey CA: %v\n", err)
		}

	}
}

// CAPool returns a CertPool containing all CAs to recognize for RPCs. Returns an error if something goes wrong.
func CAPool() (*x509.CertPool, error) {
	caCertBytesPEM, err := ioutil.ReadFile(filepath.Join(*caDirectory, caCertFileName))
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCertBytesPEM); !ok {
		return nil, fmt.Errorf("could not add cert to CA pool")
	}
	return caCertPool, nil
}

// SignCertificate returns a PEM encoded signed certificate of the input certificate.
func SignCertificate(cert *x509.Certificate, pub *rsa.PublicKey) ([]byte, error) {
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, pub, privateKey)
	if err != nil {
		return nil, err
	}
	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	return certPEM.Bytes(), nil

}

// RekeyCA returns the bytes of a new PEM encoded CA certificate. The new cert and private key are also written to disk,
// overwritting any previous version.
func RekeyCA() ([]byte, error) {
	log.Println("Rekeying CA...")
	newCa := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization:  []string{"Community Chess"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Seattle, WA"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, newCa, newCa, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err = writeToDisk(caPEM.Bytes(), caPrivKeyPEM.Bytes()); err != nil {
		return nil, err
	}
	ca = newCa
	privateKey = caPrivKey
	return caPEM.Bytes(), nil
}

func loadCAFiles() error {
	caCertBytesPEM, err := ioutil.ReadFile(filepath.Join(*caDirectory, caCertFileName))
	if err != nil {
		return err
	}
	p, _ := pem.Decode(caCertBytesPEM)
	parsedCA, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return err
	}

	caPrivateKeyPEM, err := ioutil.ReadFile(filepath.Join(*caDirectory, caPrivateKeyFileName))
	if err != nil {
		return err
	}
	p, _ = pem.Decode(caPrivateKeyPEM)
	parsedPrivateKey, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		return err
	}
	ca = parsedCA
	privateKey = parsedPrivateKey
	return nil
}

func writeToDisk(caCert, caPrivateKey []byte) error {
	fi, err := os.Stat(*caDirectory)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("path is not a directory: %s", *caDirectory)
	}
	log.Printf("Writing CA cert and private key to directory: %s\n", *caDirectory)
	if err = ioutil.WriteFile(filepath.Join(*caDirectory, caCertFileName), caCert, 0777); err != nil {
		return err
	}
	if err = ioutil.WriteFile(filepath.Join(*caDirectory, caPrivateKeyFileName), caPrivateKey, 0777); err != nil {
		return err
	}

	return nil
}
