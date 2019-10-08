package chess

import (
	"context"
	"log"
	"sync"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	ch "github.com/notnil/chess"
	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Implementation is an implementation of chess for use by both the master and slave
type Implementation struct {
	gameMux sync.Mutex
	game    *ch.Game

	playersMux sync.Mutex
	// player ID to is_white_team
	players map[string]bool

	// Game proto stuff, the state is built dynamically.
	metadata *messages.Game_Metadata
	history  *messages.Game_History
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
