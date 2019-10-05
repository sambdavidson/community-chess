// Package gamemaster implements the GameServerMaster service.
package gamemaster

import (
	"context"

	pb "github.com/sambdavidson/community-chess/src/proto/services/game_server"
)

// Server implements the GameServerMaster service.
type Server struct{}

// AddSlave is called by a GameServerSlave to request to be accepted as a valid slave for this game.
func (s *Server) AddSlave(ctx context.Context, in *pb.AddSlaveRequest) (*pb.AddSlaveResponse, error) {
	return nil, nil
}

// AddPlayers is called by a GameServerSlave to request 1+ player(s) be added to this game.
func (s *Server) AddPlayers(ctx context.Context, in *pb.AddPlayersRequest) (*pb.AddPlayersResponse, error) {
	return nil, nil
}

// RemovePlayers is called by a GameServerSlave to request 1+ player(s) be removed from this game.
func (s *Server) RemovePlayers(ctx context.Context, in *pb.RemovePlayerRequest) (*pb.RemovePlayerResponse, error) {
	return nil, nil
}

// StopGame is called by an authorized user and shuts down this game.
func (s *Server) StopGame(ctx context.Context, in *pb.StopGameRequest) (*pb.StopGameResponse, error) {
	return nil, nil
}
