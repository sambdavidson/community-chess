// Package gameslave implements the GameServerSlave service
package gameslave

import (
	"context"
	"log"

	pb "github.com/sambdavidson/community-chess/src/proto/services/game_server"
)

// Server implements the GameServerSlave service.
type Server struct{}

// ChangeAcceptingVotes is called by GameServerMasters to set this GameServerSlave to no longer accept votes. Typically done at end of a voting round.
func (s *Server) ChangeAcceptingVotes(ctx context.Context, in *pb.ChangeAcceptingVotesRequest) (*pb.ChangeAcceptingVotesResponse, error) {
	log.Println("ChangeAcceptingVotes", in)
	return nil, nil
}

// GetVotes is called by GameServerMasters get all votes received by this GameServerSlave for the current round.
func (s *Server) GetVotes(ctx context.Context, in *pb.GetVotesRequest) (*pb.GetVotesResponse, error) {
	log.Println("GetVotes", in)
	return nil, nil
}

// UpdateMetadata is called by GameServerMasters to update this slave's metadata.
func (s *Server) UpdateMetadata(ctx context.Context, in *pb.UpdateMetadataRequest) (*pb.UpdateStateResponse, error) {
	log.Println("UpdateMetadata", in)
	return nil, nil
}

// UpdateState is called by GameServerMasters to update this slave's state of the game.
func (s *Server) UpdateState(ctx context.Context, in *pb.UpdateStateRequest) (*pb.UpdateStateResponse, error) {
	log.Println("UpdateState", in)
	return nil, nil
}
