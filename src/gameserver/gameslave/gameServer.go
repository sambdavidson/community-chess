package gameslave

import (
	"context"
	"log"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

// GameServer implements the GameServer service.
type GameServer struct {
	playersRegistrarCli pr.PlayersRegistrarClient
}

// Game gets this game.
func (s *GameServer) Game(ctx context.Context, in *pb.GameRequest) (*pb.GameResponse, error) {
	log.Println("GetGame", in)
	return &pb.GameResponse{
		Game: &messages.Game{},
	}, nil
}

// Metadata gets this game's metadata.
func (s *GameServer) Metadata(ctx context.Context, in *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	log.Println("GetGameMetadata", in)
	return &pb.MetadataResponse{}, nil
}

// State gets this game's state.
func (s *GameServer) State(ctx context.Context, in *pb.StateRequest) (*pb.StateResponse, error) {
	log.Println("GetGameState", in)
	return &pb.StateResponse{}, nil
}

// History gets this game's history.
func (s *GameServer) History(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	log.Println("GetGameHistory", in)
	return &pb.HistoryResponse{}, nil
}

// Join joins this game.
func (s *GameServer) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {
	log.Println("Join", in)
	return &pb.JoinResponse{}, nil
}

// Leave leaves this game.
func (s *GameServer) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	log.Println("Leave", in)
	return &pb.LeaveResponse{}, nil
}

// PostVote posts a vote to this game.
func (s *GameServer) PostVote(ctx context.Context, in *pb.PostVoteRequest) (*pb.PostVoteResponse, error) {
	log.Println("PostVote", in)
	return &pb.PostVoteResponse{}, nil
}

// Status returns the status of this game (and/or the underlying server).
func (s *GameServer) Status(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	log.Println("Status", in)
	return &pb.StatusResponse{}, nil
}
