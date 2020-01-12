// Package main implements a server for the Player Registrar
package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/sambdavidson/community-chess/src/lib/debug"
	"github.com/sambdavidson/community-chess/src/playerregistrar/server"

	pb "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50052, "port the Game Server is accepts connections")
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(
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
