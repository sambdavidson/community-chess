package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/sambdavidson/community-chess/src/debugwebserver/gamemaster"
	"github.com/sambdavidson/community-chess/src/debugwebserver/gameserver"
	"github.com/sambdavidson/community-chess/src/debugwebserver/players"
)

var (
	port         = flag.Int("port", 8080, "port to serve debug webserver")
	staticDir    = flag.String("static_dir", "src/debugwebserver/static", "directory of static content")
	caBundlePath = flag.String("ca_bundle_path", "./devsecrets/certs/ca_cert.pem", "path to CA bundle for validating TLS connections")
	certPath     = flag.String("tls_cert_path", "./devsecrets/certs/debugadmin/debug_cert.pem", "path to the TLS certificate")
	privPath     = flag.String("tls_private_key_path", "./devsecrets/certs/debugadmin/debug_pk.pem", "path to the TLS private key")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	tlscfg, err := tlsConfig()
	if err != nil {
		log.Fatalln(err)
	}

	http.Handle("/", http.FileServer(http.Dir(*staticDir)))
	http.Handle("/players/", http.StripPrefix("/players/", &players.Handler{TLS: tlscfg}))
	http.Handle("/games/", http.StripPrefix("/games/", &gameserver.Handler{TLS: tlscfg}))
	http.Handle("/gamemaster/", http.StripPrefix("/gamemaster/", &gamemaster.Handler{TLS: tlscfg}))

	/* DONE! */
	log.Printf("Starting HTTP Server on Port: 0.0.0.0:%d\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port), nil))
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
		RootCAs:      caPool,
	}, nil
}
