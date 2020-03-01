package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sambdavidson/community-chess/src/debugwebserver/players"
)

var (
	port         = flag.Int("port", 8080, "port to serve debug webserver")
	caBundlePath = flag.String("ca_bundle_path", "", "path to CA bundle for validating TLS connections")
	certPath     = flag.String("tls_cert_path", "", "path to the TLS certificate")
	privPath     = flag.String("tls_private_key_path", "", "path to the TLS private key")
)

func main() {
	flag.Parse()
	tlscfg, err := tlsConfig()
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", http.FileServer(http.Dir("src/debugwebserver/static")))
	http.Handle("/players/", http.StripPrefix("/players/", &players.Handler{TLS: tlscfg}))

	/* DONE! */
	log.Printf("Starting HTTP Server on Port: %d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func tlsConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(*certPath, *privPath)
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
