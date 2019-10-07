package chess

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Initialize initializes this server to run the game defined in InitializeRequest.
func (i *Implementation) Initialize(ctx context.Context, in *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	log.Println("Initialize", in)
	return &pb.InitializeResponse{}, nil
}

// AddSlave is called by a GameServerSlave to request to be accepted as a valid slave for this game.
func (i *Implementation) AddSlave(ctx context.Context, in *pb.AddSlaveRequest) (*pb.AddSlaveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// AddPlayers is called by a GameServerSlave to request 1+ player(s) be added to this game.
func (i *Implementation) AddPlayers(ctx context.Context, in *pb.AddPlayersRequest) (*pb.AddPlayersResponse, error) {
	log.Println("AddPlayers", in)
	return &pb.AddPlayersResponse{}, nil
}

// RemovePlayers is called by a GameServerSlave to request 1+ player(s) be removed from this game.
func (i *Implementation) RemovePlayers(ctx context.Context, in *pb.RemovePlayersRequest) (*pb.RemovePlayersResponse, error) {
	log.Println("RemovePlayers", in)
	return &pb.RemovePlayersResponse{}, nil
}

// StopGame is called by an authorized user and shuts down this game.
func (i *Implementation) StopGame(ctx context.Context, in *pb.StopGameRequest) (*pb.StopGameResponse, error) {
	log.Println("StopGame", in)
	return &pb.StopGameResponse{}, nil
}
