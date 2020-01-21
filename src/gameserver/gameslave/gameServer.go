package gameslave

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sambdavidson/community-chess/src/lib/auth/grpcplayertokens"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

// GameServer implements the GameServer service.
type GameServer struct {
	playersRegistrarCli pr.PlayersRegistrarClient
	masterCli           pb.GameServerMasterClient
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
	return &pb.GameResponse{
		Game: &messages.Game{
			Type:      gameType,
			Id:        gameID,
			StartTime: initializeTime.UnixNano(),
			Location:  "localhost", // TODO
			Metadata:  metadataRes.GetMetadata(),
			State:     stateRes.GetState(),
			History:   historyRes.GetHistory(),
		},
	}, nil
}

// Metadata gets this game's metadata.
func (s *GameServer) Metadata(ctx context.Context, in *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	return gameImplementation.Metadata(ctx, in)
}

// State gets this game's state.
func (s *GameServer) State(ctx context.Context, in *pb.StateRequest) (*pb.StateResponse, error) {
	return gameImplementation.State(ctx, in)
}

// History gets this game's history.
func (s *GameServer) History(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	return gameImplementation.History(ctx, in)
}

// Join joins this game.
func (s *GameServer) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {
	pid, err := grpcplayertokens.ValidatedPlayerIDFromIncomingContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "missing player id from incoming context")
	}
	_, err = s.masterCli.AddPlayers(ctx, &pb.AddPlayersRequest{
		Players: []*pb.AddPlayersRequest_NewPlayer{
			&pb.AddPlayersRequest_NewPlayer{
				PlayerId: pid,
				Request: &pb.AddPlayersRequest_NewPlayer_JoinRequest{
					Fields: in.Fields,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	// TODO: update state and stuff.
	return &pb.JoinResponse{}, nil
}

// Leave leaves this game.
func (s *GameServer) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	return gameImplementation.Leave(ctx, in)
}

// PostVote posts a vote to this game.
func (s *GameServer) PostVote(ctx context.Context, in *pb.PostVoteRequest) (*pb.PostVoteResponse, error) {
	return gameImplementation.PostVote(ctx, in)
}

// Status returns the status of this game (and/or the underlying server).
func (s *GameServer) Status(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	return gameImplementation.Status(ctx, in)
}
