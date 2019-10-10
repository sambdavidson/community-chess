package chess

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

/* SERVICE STUBS UNIMPLEMENTED BY THIS GAME IMPLEMENTATION */

// Game is not implemented, handled by surrounding gameslave/gamemaster
func (i *Implementation) Game(ctx context.Context, in *pb.GameRequest) (*pb.GameResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Join is not implemented, ultimately handled by master's AddPlayers()
func (i *Implementation) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Leave is not implemented, ultimately handled by master's RemovePlayers()
func (i *Implementation) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	log.Println("Leave", in)
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// AddSlave is not implemented, handled by surrounding gamemaster
func (i *Implementation) AddSlave(ctx context.Context, in *pb.AddSlaveRequest) (*pb.AddSlaveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Status returns the status of this game (and/or the underlying server). TODO: figure out what to do here.
func (i *Implementation) Status(ctx context.Context, in *pb.StatusRequest) (*pb.StatusResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// StopGame is called by an authorized user and shuts down this game.
func (i *Implementation) StopGame(ctx context.Context, in *pb.StopGameRequest) (*pb.StopGameResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
