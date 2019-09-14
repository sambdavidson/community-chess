/* BUILD and RUN
go run .\src\game_server
*/

// Package main implements a server for the Game Server service.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	chess "github.com/sambdavidson/community-chess/src/game_server/chess"
	gs "github.com/sambdavidson/community-chess/src/proto/services/game_server"
	"google.golang.org/grpc"
)

var (
	game = flag.String("game", "chess", "game type to run the server as, currently the only supported string is \"chess\"")
	port = flag.Int("port", 50051, "port the Game Server is accepts connections")

	playerRegistrarURI  = flag.String("player_registar_uri", "localhost", "URI of the Player Registrar")
	playerRegistrarPort = flag.Int("player_registrar_port", 50052, "Port of the Player Registrar")

	gameMap = map[string]func(*grpc.Server) (gs.GameServerServer, error){
		"chess": buildChess,
	}
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	builder, ok := gameMap[*game]
	if !ok {
		log.Fatalf("unknown game flag %q", *game)
	}
	server, err := builder(s)
	if err != nil {
		log.Fatalf("failed to get server: %v", err)
	}
	gs.RegisterGameServerServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func buildChess(sv *grpc.Server) (gs.GameServerServer, error) {
	// TODO: Parse relevant flags.
	server, err := chess.NewServer(chess.Opts{
		Server:                 sv,
		PlayerRegistrarAddress: fmt.Sprintf("%s:%d", *playerRegistrarURI, *playerRegistrarPort),
	})
	return server, err
}
