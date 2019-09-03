// Package main implements a server for the Game Server service.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	sv "github.com/samdamana/community-chess/src/game_server/server"
	pb "github.com/samdamana/community-chess/src/proto/services/game_server"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "port the Game Server is accepts connections")
)

func main() {
	server, err := sv.NewServer(sv.Opts{})
	if err != nil {
		log.Fatalf("failed to get server: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGameServerServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
