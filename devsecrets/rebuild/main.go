// Rebuilds all the dev secrets.
package main

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
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sambdavidson/community-chess/src/lib/tlsconsts"
)

var (
	secretDir   = flag.String("secret_dir", "certs", "directory to write dev secrets. Defaults to empty (the parent dir of devscrets)")
	instanceID  = flag.String("instance_id", uuid.New().String(), "instance ID to be used in the TLS certificate")
	gameID      = flag.String("game_id", "", "game ID to be used in the TLS certificate")
	serviceType = flag.String("service_type", "", "type of client we are generating a cert for, must be one of the consts")
)

/*
	"localhost", // The address of services will need to be figured out and injected here.
	gameID.String(),
	serverType.String(),
	tlsca.GameServer.String(),
	tlsca.Internal.String(),
*/

const (
	caPrefix   = "ca"
	certSuffix = "cert.pem"
	pkSuffix   = "pk.pem"
)

var (
	privateKey *rsa.PrivateKey
	ca         *x509.Certificate
)

func main() {
	flag.Parse()

	var cert *x509.Certificate
	var prefix string
	switch strings.ToLower(*serviceType) {
	case "gameserver/master":
		prefix = "master"
		cert = certForGameserver(false)
	case "gameserver/slave":
		prefix = "slave"
		cert = certForGameserver(true)
	case "playerregistrar":
		prefix = "pr"
		cert = certForPlayerregistrar()
	case "debugadmin":
		prefix = "debug"
		cert = certForDebugAdmin()
	default:
		log.Fatalf("invalid service_type selection: %s\n", *serviceType)
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}
	certPEMBytes, err := SignCertificate(cert, &certPrivKey.PublicKey)
	if err != nil {
		log.Fatal(err)
	}

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	err = writeToDisk(*serviceType, prefix, certPEMBytes, certPrivKeyPEM.Bytes())
	if err != nil {
		log.Fatal(err)
	}
}

func certForGameserver(slave bool) *x509.Certificate {
	if len(*gameID) == 0 {
		log.Fatal("Missing game_id flag for gameserver")
	}
	serverType := tlsconsts.GameMaster
	dockerName := "community-chess_gameserver_master_1"
	if slave {
		serverType = tlsconsts.GameSlave
		dockerName = "community-chess_gameserver_slave_1"
	}

	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: *instanceID,
		},
		SerialNumber: big.NewInt(time.Now().Unix()),
		DNSNames: []string{
			"localhost", // The address of services will need to be figured out and injected here.
			dockerName,
			*gameID,
			serverType.String(),
			tlsconsts.GameServer.String(),
			tlsconsts.Internal.String(),
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
}

func certForPlayerregistrar() *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: *instanceID,
		},
		SerialNumber: big.NewInt(time.Now().Unix()),
		DNSNames: []string{
			"localhost", // The address of services will need to be figured out and injected here.
			tlsconsts.PlayerRegistrar.String(),
			tlsconsts.Internal.String(),
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
}

func certForDebugAdmin() *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "admin",
		},
		SerialNumber: big.NewInt(time.Now().Unix()),
		DNSNames: []string{
			"localhost", // The address of services will need to be figured out and injected here.
			tlsconsts.Admin.String(),
			tlsconsts.Internal.String(),
		},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
}

// Initializes the CA if it doesn't already exist.
func init() {
	log.Println("Initializing TLS CA library...")
	if err := loadCAFiles(*secretDir); err != nil {
		log.Printf("Error loading existing CA files: %v\n", err)
		if _, err = RekeyCA(); err != nil {
			log.Fatalf("Unable rekey CA: %v\n", err)
		}

	}
}

// CAPool returns a CertPool containing all CAs to recognize for RPCs. Returns an error if something goes wrong.
func CAPool() (*x509.CertPool, error) {
	caCertBytesPEM, err := ioutil.ReadFile(filepath.Join(*secretDir, caPrefix+"_"+certSuffix))
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
			Organization:  tlsconsts.Organization(),
			Country:       tlsconsts.Country(),
			Province:      tlsconsts.Province(),
			Locality:      tlsconsts.Locality(),
			StreetAddress: tlsconsts.StreetAddress(),
			PostalCode:    tlsconsts.PostalCode(),
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
	if err = writeToDisk("", caPrefix, caPEM.Bytes(), caPrivKeyPEM.Bytes()); err != nil {
		return nil, err
	}
	ca = newCa
	privateKey = caPrivKey
	return caPEM.Bytes(), nil
}

func loadCAFiles(dir string) error {
	caCertBytesPEM, err := ioutil.ReadFile(filepath.Join(dir, caPrefix+"_"+certSuffix))
	if err != nil {
		return err
	}
	p, _ := pem.Decode(caCertBytesPEM)
	parsedCA, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return err
	}

	caPrivateKeyPEM, err := ioutil.ReadFile(filepath.Join(dir, caPrefix+"_"+pkSuffix))
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

func writeToDisk(dir, prefix string, cert, pk []byte) error {
	_ = os.MkdirAll(filepath.Join(*secretDir, dir), 0777)
	fi, err := os.Stat(*secretDir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("path is not a directory: %s", *secretDir)
	}
	log.Printf("Writing CA cert and private key to directory: %s\n", *secretDir)
	if err = ioutil.WriteFile(filepath.Join(*secretDir, dir, prefix+"_"+certSuffix), cert, 0777); err != nil {
		return err
	}
	if err = ioutil.WriteFile(filepath.Join(*secretDir, dir, prefix+"_"+pkSuffix), pk, 0777); err != nil {
		return err
	}

	return nil
}
