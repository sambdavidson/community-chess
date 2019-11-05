// Package gameslave implements the GameServerSlave service
package gameslave

import (
	"context"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

// GameServerSlave implements the GameServerSlave service.
type GameServerSlave struct {
	masterID            string
	masterCli           pb.GameServerMasterClient
	playersRegistrarCli pr.PlayersRegistrarClient
}

// ChangeAcceptingVotes is called by GameServerMasters to set this GameServerSlave to no longer accept votes. Typically done at end of a voting round.
func (s *GameServerSlave) ChangeAcceptingVotes(ctx context.Context, in *pb.ChangeAcceptingVotesRequest) (*pb.ChangeAcceptingVotesResponse, error) {
	return gameImplementation.ChangeAcceptingVotes(ctx, in)
}

// GetVotes is called by GameServerMasters get all votes received by this GameServerSlave for the current round.
func (s *GameServerSlave) GetVotes(ctx context.Context, in *pb.GetVotesRequest) (*pb.GetVotesResponse, error) {
	return gameImplementation.GetVotes(ctx, in)
}

// UpdateMetadata is called by GameServerMasters to update this slave's metadata.
func (s *GameServerSlave) UpdateMetadata(ctx context.Context, in *pb.UpdateMetadataRequest) (*pb.UpdateMetadataResponse, error) {
	return gameImplementation.UpdateMetadata(ctx, in)
}

// UpdateState is called by GameServerMasters to update this slave's state of the game.
func (s *GameServerSlave) UpdateState(ctx context.Context, in *pb.UpdateStateRequest) (*pb.UpdateStateResponse, error) {
	return gameImplementation.UpdateState(ctx, in)
}
