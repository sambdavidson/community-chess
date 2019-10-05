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

	"github.com/google/uuid"
	chess "github.com/sambdavidson/community-chess/src/game_server/chess"
	gs "github.com/sambdavidson/community-chess/src/proto/services/game_server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	port               = flag.Int("port", 50051, "port the Game Server is accepts connections")
	maxConcurrentGames = flag.Int("max_game_instances", 10, "maximum number of concurrent games on this server")

	playerRegistrarURI  = flag.String("player_registar_uri", "localhost", "URI of the Player Registrar")
	playerRegistrarPort = flag.Int("player_registrar_port", 50052, "Port of the Player Registrar")

	playerRegistrarCli  pr.PlayerRegistrarClient
	playerRegistrarConn *grpc.ClientConn

	gameBuilderMap = map[string]func(opts Opts) (gs.GameServerServer, error){
		"chess": func(opts Opts) (gs.GameServerServer, error) {
			return chess.NewServer(opts.id, opts.playerRegistrarCli)
		},
	}
)

// Opts contains the common options for starting a new server
type Opts struct {
	id                 string
	playerRegistrarCli pr.PlayerRegistrarClient
}

// GameServer implemements gs.GameServerServer
type GameServer struct {
	mux sync.Mutex

	gameInstances map[string]gs.GameServerServer
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

	gameServer := &GameServer{
		gameInstances: map[string]gs.GameServerServer{},
	}

	gs.RegisterGameServerServer(grpcServer, gameServer)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// StartGame starts the game defined in the request
func (s *GameServer) StartGame(ctx context.Context, in *gs.StartGameRequest) (*gs.StartGameResponse, error) {
	gameBuilder, ok := gameBuilderMap[in.GetGameType()]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown game %q", in.GetGameType())
	}
	id := uuid.New().String()
	game, err := gameBuilder(Opts{
		id:                 id,
		playerRegistrarCli: playerRegistrarCli,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to build game: %v", err)
	}

	out, err := game.StartGame(ctx, in)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start created game: %v", err)
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	s.gameInstances[id] = game
	return out, nil
}

// GetGame gets the game details given a GetGameRequest
func (s *GameServer) GetGame(ctx context.Context, in *gs.GetGameRequest) (*gs.GetGameResponse, error) {
	game, ok := s.gameInstances[in.GetGameId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "unknown game with id: %s", in.GetGameId().GetId())
	}
	return game.GetGame(ctx, in)
}

// AddPlayer adds a player to the existing game
func (s *GameServer) AddPlayer(ctx context.Context, in *gs.AddPlayerRequest) (*gs.AddPlayerResponse, error) {
	game, ok := s.gameInstances[in.GetGameId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "unknown game with id: %s", in.GetGameId().GetId())
	}
	return game.AddPlayer(ctx, in)
}

// RemovePlayer removes a player from the current game
func (s *GameServer) RemovePlayer(ctx context.Context, in *gs.RemovePlayerRequest) (*gs.RemovePlayerResponse, error) {
	game, ok := s.gameInstances[in.GetGameId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "unknown game with id: %s", in.GetGameId().GetId())
	}
	return game.RemovePlayer(ctx, in)
}

// PostVotes posts 1+ votes to the current game
func (s *GameServer) PostVotes(ctx context.Context, in *gs.PostVotesRequest) (*gs.PostVotesResponse, error) {
	game, ok := s.gameInstances[in.GetGameId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "unknown game with id: %s", in.GetGameId().GetId())
	}
	return game.PostVotes(ctx, in)
}

// StopGame starts the game defined in the request
func (s *GameServer) StopGame(ctx context.Context, in *gs.StopGameRequest) (*gs.StopGameResponse, error) {
	game, ok := s.gameInstances[in.GetGameId()]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "unknown game with id: %s", in.GetGameId().GetId())
	}
	return game.StopGame(ctx, in)
}

// ListGames starts the game defined in the request
func (s *GameServer) ListGames(ctx context.Context, in *gs.ListGamesRequest) (*gs.ListGamesResponse, error) {
	var games []*messages.Game
	for id, v := range s.gameInstances {
		out, err := v.GetGame(ctx, &gs.GetGameRequest{
			GameId: id,
		})
		if err != nil {
			return nil, err
		}
		games = append(games, out.GetGame())
	}

	return &gs.ListGamesResponse{
		Game: games,
	}, nil
}
