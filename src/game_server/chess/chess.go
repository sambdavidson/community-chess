package chess

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	gs "github.com/sambdavidson/community-chess/src/proto/services/game_server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc/status"
)

// Server implement the GameServer service.
type Server struct {
	playerRegistrarCli pr.PlayerRegistrarClient

	mux sync.Mutex
}

// Opts contains the options for building a chess server
type Opts struct {
	playerRegistrarCli pr.PlayerRegistrarClient
	gameMetadata       *messages.Game_Metadata
}

// NewServer builds a new Server object
func NewServer(o Opts) (*Server, error) {
	s := &Server{
		playerRegistrarCli: o.playerRegistrarCli,
	}

	return s, nil
}

// StartGame starts the game defined in the request
func (s *Server) StartGame(ctx context.Context, in *gs.StartGameRequest) (*gs.StartGameResponse, error) {
	fmt.Printf("GetGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement StartGame")
}

// GetGame gets the game details given a GetGameRequest
func (s *Server) GetGame(ctx context.Context, in *gs.GetGameRequest) (*gs.GetGameResponse, error) {
	fmt.Printf("GetGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement GetGame")
}

// AddPlayer adds a player to the existing game
func (s *Server) AddPlayer(ctx context.Context, in *gs.AddPlayerRequest) (*gs.AddPlayerResponse, error) {
	fmt.Printf("AddPlayer %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement AddPlayer")
}

// RemovePlayer removes a player from the current game
func (s *Server) RemovePlayer(ctx context.Context, in *gs.RemovePlayerRequest) (*gs.RemovePlayerResponse, error) {
	fmt.Printf("RemovePlayer %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement RemovePlayer")
}

// PostVotes posts 1+ votes to the current game
func (s *Server) PostVotes(ctx context.Context, in *gs.PostVotesRequest) (*gs.PostVotesResponse, error) {
	fmt.Printf("PostVotes %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement PostVotes")
}

// StopGame starts the game defined in the request
func (s *GameServer) StopGame(ctx context.Context, in *gs.StopGameRequest) (*gs.StopGameResponse, error) {
	fmt.Printf("StopGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement StopGame")
}
