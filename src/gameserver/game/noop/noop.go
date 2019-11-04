// Package noop defines a no-op GameImplementation that just returns "game uninitialized" for everything.
package noop

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Implementation defines a GameImplementation that returns FailedPrecondition for everything.
type Implementation struct{}

var (
	err = status.Errorf(codes.FailedPrecondition, "game not initialized")
)

// Initialize returns FailedPrecondition for everything.
func (i *Implementation) Initialize(ctx context.Context, in *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	return nil, err
}

// UpdateMetadata returns FailedPrecondition for everything.
func (i *Implementation) UpdateMetadata(ctx context.Context, in *pb.UpdateMetadataRequest) (*pb.UpdateMetadataResponse, error) {
	return nil, err
}

// UpdateState returns FailedPrecondition for everything.
func (i *Implementation) UpdateState(ctx context.Context, in *pb.UpdateStateRequest) (*pb.UpdateStateResponse, error) {
	return nil, err
}

// Metadata returns FailedPrecondition for everything.
func (i *Implementation) Metadata(ctx context.Context, in *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	return nil, err
}

// State returns FailedPrecondition for everything.
func (i *Implementation) State(ctx context.Context, in *pb.StateRequest) (*pb.StateResponse, error) {
	return nil, err
}

// History returns FailedPrecondition for everything.
func (i *Implementation) History(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	return nil, err
}

// AddPlayers returns FailedPrecondition for everything.
func (i *Implementation) AddPlayers(ctx context.Context, in *pb.AddPlayersRequest) (*pb.AddPlayersResponse, error) {
	return nil, err
}

// RemovePlayers returns FailedPrecondition for everything.
func (i *Implementation) RemovePlayers(ctx context.Context, in *pb.RemovePlayersRequest) (*pb.RemovePlayersResponse, error) {
	return nil, err
}

// Game returns FailedPrecondition for everything.
func (i *Implementation) Game(ctx context.Context, in *pb.GameRequest) (*pb.GameResponse, error) {
	return nil, err
}

// Join returns FailedPrecondition for everything.
func (i *Implementation) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {
	return nil, err
}

// Leave returns FailedPrecondition for everything.
func (i *Implementation) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	return nil, err
}

// AddSlave returns FailedPrecondition for everything.
func (i *Implementation) AddSlave(ctx context.Context, in *pb.AddSlaveRequest) (*pb.AddSlaveResponse, error) {
	return nil, err
}

// Status returns FailedPrecondition for everything.
func (i *Implementation) Status(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	return nil, err
}

// StopGame returns FailedPrecondition for everything.
func (i *Implementation) StopGame(ctx context.Context, in *pb.StopGameRequest) (*pb.StopGameResponse, error) {
	return nil, err
}

// ChangeAcceptingVotes returns FailedPrecondition for everything.
func (i *Implementation) ChangeAcceptingVotes(ctx context.Context, in *pb.ChangeAcceptingVotesRequest) (*pb.ChangeAcceptingVotesResponse, error) {
	return nil, err
}

// GetVotes returns FailedPrecondition for everything.
func (i *Implementation) GetVotes(ctx context.Context, in *pb.GetVotesRequest) (*pb.GetVotesResponse, error) {
	return nil, err
}

// PostVote returns FailedPrecondition for everything.
func (i *Implementation) PostVote(ctx context.Context, in *pb.PostVoteRequest) (*pb.PostVoteResponse, error) {
	return nil, err
}
