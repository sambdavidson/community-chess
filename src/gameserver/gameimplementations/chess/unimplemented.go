package chess

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

/* SERVICE STUBS UNIMPLEMENTED BY THIS GAME IMPLEMENTATION */

// Game is not implemented
func (i *Implementation) Game(ctx context.Context, in *pb.GameRequest) (*pb.GameResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// AddSlave is not implemented
func (i *Implementation) AddSlave(ctx context.Context, in *pb.AddSlaveRequest) (*pb.AddSlaveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
