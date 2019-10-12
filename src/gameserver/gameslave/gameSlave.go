// Package gameslave implements the GameServerSlave service
package gameslave

import (
	"context"
	"log"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

// GameServerSlave implements the GameServerSlave service.
type GameServerSlave struct {
	playersRegistrarCli pr.PlayersRegistrarClient

}

// ChangeAcceptingVotes is called by GameServerMasters to set this GameServerSlave to no longer accept votes. Typically done at end of a voting round.
func (s *GameServerSlave) ChangeAcceptingVotes(ctx context.Context, in *pb.ChangeAcceptingVotesRequest) (*pb.ChangeAcceptingVotesResponse, error) {
	log.Println("ChangeAcceptingVotes", in)
	return &pb.ChangeAcceptingVotesResponse{}, nil
}

// GetVotes is called by GameServerMasters get all votes received by this GameServerSlave for the current round.
func (s *GameServerSlave) GetVotes(ctx context.Context, in *pb.GetVotesRequest) (*pb.GetVotesResponse, error) {
	log.Println("GetVotes", in)
	return &pb.GetVotesResponse{}, nil
}

// UpdateMetadata is called by GameServerMasters to update this slave's metadata.
func (s *GameServerSlave) UpdateMetadata(ctx context.Context, in *pb.UpdateMetadataRequest) (*pb.UpdateMetadataResponse, error) {
	log.Println("UpdateMetadata", in)
	return &pb.UpdateMetadataResponse{}, nil
}

// UpdateState is called by GameServerMasters to update this slave's state of the game.
func (s *GameServerSlave) UpdateState(ctx context.Context, in *pb.UpdateStateRequest) (*pb.UpdateStateResponse, error) {
	log.Println("UpdateState", in)
	return &pb.UpdateStateResponse{}, nil
}
