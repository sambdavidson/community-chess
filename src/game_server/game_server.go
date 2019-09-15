/* BUILD and RUN
go run .\src\game_server
*/

// Package main implements a server for the Game Server service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	chess "github.com/sambdavidson/community-chess/src/game_server/chess"
	gs "github.com/sambdavidson/community-chess/src/proto/services/game_server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	port     = flag.Int("port", 50051, "port the Game Server is accepts connections")
	maxGames = flag.Int("max_concurrent_games", 10, "maximum number of concurrent games that can be started on this server")

	playerRegistrarURI  = flag.String("player_registar_uri", "localhost", "URI of the Player Registrar")
	playerRegistrarPort = flag.Int("player_registrar_port", 50052, "Port of the Player Registrar")

	playerRegistrarCli  pr.PlayerRegistrarClient
	playerRegistrarConn *grpc.ClientConn

	gameMap = map[string]func(opts Opts) (gs.GameServerServer, error){
		"chess": chess.NewServer,
	}
)

// Opts contains the common options for starting a new server
type Opts struct {
	playerRegistrarCli pr.PlayerRegistrarClient
	gameMetadata       *messages.Game_Metadata
}

// GameServer implemements gs.GameServerServer
type GameServer struct {
	mux sync.Mutex

	activeGames map[string]gs.GameServerServer
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	playerRegistrarConn, err = grpc.Dial(fmt.Sprintf("%s:%d", *playerRegistrarURI, *playerRegistrarPort), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	playerRegistrarCli = pr.NewPlayerRegistrarClient(playerRegistrarConn)

	gameServer := GameServer{}

	gs.RegisterGameServerServer(s, gameServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// StartGame starts the game defined in the request
func (s *GameServer) StartGame(ctx context.Context, in *gs.StartGameRequest) (*gs.StartGameResponse, error) {
	fmt.Printf("StartGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement StartGame")
}

// GetGame gets the game details given a GetGameRequest
func (s *GameServer) GetGame(ctx context.Context, in *gs.GetGameRequest) (*gs.GetGameResponse, error) {
	fmt.Printf("GetGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement GetGame")
}

// AddPlayer adds a player to the existing game
func (s *GameServer) AddPlayer(ctx context.Context, in *gs.AddPlayerRequest) (*gs.AddPlayerResponse, error) {
	fmt.Printf("AddPlayer %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement AddPlayer")
}

// RemovePlayer removes a player from the current game
func (s *GameServer) RemovePlayer(ctx context.Context, in *gs.RemovePlayerRequest) (*gs.RemovePlayerResponse, error) {
	fmt.Printf("RemovePlayer %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement RemovePlayer")
}

// PostVotes posts 1+ votes to the current game
func (s *GameServer) PostVotes(ctx context.Context, in *gs.PostVotesRequest) (*gs.PostVotesResponse, error) {
	fmt.Printf("PostVotes %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement PostVotes")
}
x
// StopGame starts the game defined in the request
func (s *GameServer) StopGame(ctx context.Context, in *gs.StopGameRequest) (*gs.StopGameResponse, error) {
	fmt.Printf("StopGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement StopGame")
}
