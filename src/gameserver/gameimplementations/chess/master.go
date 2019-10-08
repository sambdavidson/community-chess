package chess

import (
	"context"
	"log"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Initialize initializes this server to run the game defined in InitializeRequest.
func (i *Implementation) Initialize(ctx context.Context, in *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	log.Println("Initialize Chess", in)

	return &pb.InitializeResponse{}, nil
}

// StopGame is called by an authorized user and shuts down this game.
func (i *Implementation) StopGame(ctx context.Context, in *pb.StopGameRequest) (*pb.StopGameResponse, error) {
	log.Println("StopGame", in)
	return &pb.StopGameResponse{}, nil
}
