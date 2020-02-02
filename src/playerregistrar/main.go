// Package main implements a server for the Player Registrar
package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/sambdavidson/community-chess/src/lib/debug"
	"github.com/sambdavidson/community-chess/src/playerregistrar/server"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	pb "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	port         = flag.Int("port", 8080, "port the Game Server is accepts connections")
	caBundlePath = flag.String("ca_bundle_path", "", "path to CA bundle for validating TLS connections")
	tlsCertPath  = flag.String("tls_cert_path", "", "path to the master gameserver TLS certificate, if enabled")
	tlsPKPath    = flag.String("tls_private_key_path", "", "path to the master gameserver TLS private key, if enabled")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	config, err := tlsConfig()
	if err != nil {
		log.Fatalf("failed to build tls config: %v", err)
	}

	s := grpc.NewServer(
		grpc.Creds(credentials.NewTLS(config)),
		grpc.UnaryInterceptor(
			middleware.ChainUnaryServer(
				debug.UnaryServerInterceptor,
			),
		),
	)
	svr, err := server.New(&server.Opts{})
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterPlayersRegistrarServer(s, svr)

	log.Printf("Starting listen of Player Registrar on port %v\n", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func tlsConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(*tlsCertPath, *tlsPKPath)
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
