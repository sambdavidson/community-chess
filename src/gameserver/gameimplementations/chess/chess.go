package chess

import (
	"context"
	"log"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Implementation is an implementation of chess for use by both the master and slave
type Implementation struct{}

func (i *Implementation) Enable() {

}

// Game gets this game.
func (i *Implementation) Game(ctx context.Context, in *pb.GameRequest) (*pb.GameResponse, error) {
	log.Println("GetGame", in)
	return &pb.GameResponse{}, nil
}

// Metadata gets this game's metadata.
func (i *Implementation) Metadata(ctx context.Context, in *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	log.Println("GetGameMetadata", in)
	return &pb.MetadataResponse{}, nil
}

// State gets this game's state.
func (i *Implementation) State(ctx context.Context, in *pb.StateRequest) (*pb.StateResponse, error) {
	log.Println("GetGameState", in)
	return &pb.StateResponse{}, nil
}

// History gets this game's history.
func (i *Implementation) History(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	log.Println("GetGameHistory", in)
	return &pb.HistoryResponse{}, nil
}

// Join joins this game.
func (i *Implementation) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {
	log.Println("Join", in)
	return &pb.JoinResponse{}, nil
}

// Leave leaves this game.
func (i *Implementation) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	log.Println("Leave", in)
	return &pb.LeaveResponse{}, nil
}

// PostVote posts a vote to this game.
func (i *Implementation) PostVote(ctx context.Context, in *pb.PostVoteRequest) (*pb.PostVoteResponse, error) {
	log.Println("PostVote", in)
	return &pb.PostVoteResponse{}, nil
}

// Status returns the status of this game (and/or the underlying server).
func (i *Implementation) Status(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	log.Println("Status", in)
	return &pb.StatusResponse{}, nil
}
