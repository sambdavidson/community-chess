// Package gameserver provides types for the GameServer service.
package gameserver

import (
	"context"

	pb "github.com/sambdavidson/community-chess/src/proto/services/game_server"
)

// Server implements the GameServer service.
type Server struct{}

// GetGame gets this game.
func (s *Server) GetGame(ctx context.Context, in *pb.GetGameRequest) (*pb.GetGameResponse, error) {
	return nil, nil
}

// GetGameMetadata gets this game's metadata.
func (s *Server) GetGameMetadata(ctx context.Context, in *pb.GetGameMetadataRequest) (*pb.GetGameMetadataResponse, error) {
	return nil, nil
}

// GetGameState gets this game's state.
func (s *Server) GetGameState(ctx context.Context, in *pb.GetGameStateRequest) (*pb.GetGameStateResponse, error) {
	return nil, nil
}

// GetGameHistory gets this game's history.
func (s *Server) GetGameHistory(ctx context.Context, in *pb.GetGameHistoryRequest) (*pb.GetGameHistoryResponse, error) {
	return nil, nil
}

// Join joins this game.
func (s *Server) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {
	return nil, nil
}

// Leave leaves this game.
func (s *Server) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	return nil, nil
}

// PostVote posts a vote to this game.
func (s *Server) PostVote(ctx context.Context, in *pb.GetGameRequest) (*pb.PostVoteResponse, error) {
	return nil, nil
}
