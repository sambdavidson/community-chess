package gamemaster

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
	metadataRes, err := gameImplementation.Metadata(ctx, &pb.MetadataRequest{})
	if err != nil {
		return nil, err
	}
	stateRes, err := gameImplementation.State(ctx, &pb.StateRequest{Detailed: in.GetDetailed()})
	if err != nil {
		return nil, err
	}
	historyRes, err := gameImplementation.History(ctx, &pb.HistoryRequest{Detailed: in.GetDetailed()})
	if err != nil {
		return nil, err
	}
	f := &pb.GameResponse{
		Game: &messages.Game{
			Type:      gameType,
			Id:        gameID,
			StartTime: initializeTime.UnixNano(),
			Location:  "localhost", // TODO
			Metadata:  metadataRes.GetMetadata(),
			State:     stateRes.GetState(),
			History:   historyRes.GetHistory(),
		},
	}
	log.Println(f)
	return f, nil
}

// Metadata gets this game's metadata.
func (s *GameServer) Metadata(ctx context.Context, in *pb.MetadataRequest) (*pb.MetadataResponse, error) {

	return &pb.MetadataResponse{}, nil
}

// State gets this game's state.
func (s *GameServer) State(ctx context.Context, in *pb.StateRequest) (*pb.StateResponse, error) {

	return &pb.StateResponse{}, nil
}

// History gets this game's history.
func (s *GameServer) History(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {

	return &pb.HistoryResponse{}, nil
}

// Join joins this game.
func (s *GameServer) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {

	return &pb.JoinResponse{}, nil
}

// Leave leaves this game.
func (s *GameServer) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {

	return &pb.LeaveResponse{}, nil
}

// PostVote posts a vote to this game.
func (s *GameServer) PostVote(ctx context.Context, in *pb.PostVoteRequest) (*pb.PostVoteResponse, error) {

	return &pb.PostVoteResponse{}, nil
}

// Status returns the status of this game (and/or the underlying server).
func (s *GameServer) Status(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {

	return &pb.StatusResponse{}, nil
}
