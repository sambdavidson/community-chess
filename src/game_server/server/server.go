package server

import (
	"context"

	pb "github.com/samdamana/community-chess/src/proto/services/game_server"
)

// Server implement the GameServer service.
type Server struct {
}

// Opts is the options for setting up a GameServer.
type Opts struct {
}

// NewServer builds a new Server object
func NewServer(o Opts) (*Server, error) {
	return &Server{}, nil
}

// GetGame gets the game details given a GetGameRequest
func (s *Server) GetGame(ctx context.Context, in *pb.GetGameRequest) (*pb.GetGameResponse, error) {
	return nil, nil
}

// AddPlayer adds a player to the existing game
func (s *Server) AddPlayer(ctx context.Context, in *pb.AddPlayerRequest) (*pb.AddPlayerResponse, error) {
	return nil, nil
}

// RemovePlayer removes a player from the current game
func (s *Server) RemovePlayer(ctx context.Context, in *pb.RemovePlayerRequest) (*pb.RemovePlayerResponse, error) {
	return nil, nil
}

// PostVotes posts 1+ votes to the current game
func (s *Server) PostVotes(ctx context.Context, in *pb.PostVotesRequest) (*pb.PostVotesResponse, error) {
	return nil, nil
}
