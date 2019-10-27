// Package gamemaster implements the GameServerMaster service.
package gamemaster

import (
	"context"
	"log"

	"github.com/sambdavidson/community-chess/src/lib/tlsca"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
)

// GameServerMaster implements the GameServerMaster service.
type GameServerMaster struct {
	playersRegistrarCli pr.PlayersRegistrarClient
}

// Initialize initializes this server to run the game defined in InitializeRequest.
func (s *GameServerMaster) Initialize(ctx context.Context, in *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	log.Println("Initialize", in)
	if game == nil {
		var ok bool
		if game, ok = gameImplementations[in.GetGame().GetType()]; !ok {
			return nil, status.Errorf(codes.InvalidArgument, "unknown game type: %v", in.GetGame().GetType())
		}
	} else {
		return nil, status.Error(codes.FailedPrecondition, "this master is already initialized")
	}
	return game.Initialize(ctx, in)
}

// AddSlave is called by a GameServerSlave to request to be accepted as a valid slave for this game.
func (s *GameServerMaster) AddSlave(ctx context.Context, in *pb.AddSlaveRequest) (*pb.AddSlaveResponse, error) {
	log.Println("AddSlave", in)
	if err := validateSlave(ctx); err != nil {
		return nil, err
	}
	slaveConn, err := grpc.Dial(in.GetReturnAddress(), grpc.WithTransportCredentials(credentials.NewTLS(masterTLSConfig)))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to dial return address")
	}
	slaveCli := pb.NewGameServerSlaveClient(slaveConn)
	controller.slaveConns = append(controller.slaveConns, slaveConn)
	controller.slaveClis = append(controller.slaveClis, slaveCli)
	return &pb.AddSlaveResponse{}, nil
}

// AddPlayers is called by a GameServerSlave to request 1+ player(s) be added to this game.
func (s *GameServerMaster) AddPlayers(ctx context.Context, in *pb.AddPlayersRequest) (*pb.AddPlayersResponse, error) {
	log.Println("AddPlayers", in)
	return &pb.AddPlayersResponse{}, nil
}

// RemovePlayers is called by a GameServerSlave to request 1+ player(s) be removed from this game.
func (s *GameServerMaster) RemovePlayers(ctx context.Context, in *pb.RemovePlayersRequest) (*pb.RemovePlayersResponse, error) {
	log.Println("RemovePlayers", in)
	return &pb.RemovePlayersResponse{}, nil
}

// StopGame is called by an authorized user and shuts down this game.
func (s *GameServerMaster) StopGame(ctx context.Context, in *pb.StopGameRequest) (*pb.StopGameResponse, error) {
	log.Println("StopGame", in)
	return &pb.StopGameResponse{}, nil
}

// validateSlave returns an error if the peer in ctx is not a our slave
func validateSlave(ctx context.Context) error {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no peer found")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return status.Error(codes.Unauthenticated, "unexpected peer transport credentials")
	}
	if !contains(tlsAuth.State.VerifiedChains[0][0].DNSNames, string(tlsca.GameSlave)) {
		return status.Error(codes.Unauthenticated, "peer is not a slave")
	}
	if !contains(tlsAuth.State.VerifiedChains[0][0].DNSNames, id) {
		return status.Error(codes.Unauthenticated, "peer is not a slave")
	}
	return nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
