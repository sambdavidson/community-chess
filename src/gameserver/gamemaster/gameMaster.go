// Package gamemaster implements the GameServerMaster service.
package gamemaster

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	"github.com/sambdavidson/community-chess/src/lib/auth"
	"github.com/sambdavidson/community-chess/src/lib/tlsca"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
)

// GameServerMaster implements the GameServerMaster service.
type GameServerMaster struct {
	mux                 sync.Mutex
	playersRegistrarCli pr.PlayersRegistrarClient
	slaves              map[string]pb.GameServerSlaveClient
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
	slaveID, err := validateSlave(ctx)
	if err != nil {
		return nil, err
	}
	s.mux.Lock()
	defer s.mux.Unlock()

	_, ok := s.slaves[slaveID]
	if ok {
		return nil, status.Errorf(codes.FailedPrecondition, "slave %s already added to this master", slaveID)
	}
	slaveConn, err := grpc.Dial(in.GetReturnAddress(), grpc.WithTransportCredentials(credentials.NewTLS(masterTLSConfig)))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to dial return address")
	}
	s.slaves[slaveID] = pb.NewGameServerSlaveClient(slaveConn)
	controller.slaveConns = append(controller.slaveConns, slaveConn)
	return &pb.AddSlaveResponse{}, nil
}

// AddPlayers is called by a GameServerSlave to request 1+ player(s) be added to this game.
func (s *GameServerMaster) AddPlayers(ctx context.Context, in *pb.AddPlayersRequest) (*pb.AddPlayersResponse, error) {
	log.Println("AddPlayers", in)
	slaveID, err := validateSlave(ctx)
	if err != nil {
		return nil, err
	}
	res, err := game.AddPlayers(ctx, in)
	if err == nil {
		s.otherSlavesUpdateState(slaveID, res.GetState())
	}

	return res, err
}

// RemovePlayers is called by a GameServerSlave to request 1+ player(s) be removed from this game.
func (s *GameServerMaster) RemovePlayers(ctx context.Context, in *pb.RemovePlayersRequest) (*pb.RemovePlayersResponse, error) {
	slaveID, err := validateSlave(ctx)
	if err != nil {
		return nil, err
	}
	res, err := game.RemovePlayers(ctx, in)
	if err == nil {
		s.otherSlavesUpdateState(slaveID, res.GetState())
	}
	return res, nil
}

// StopGame is called by an authorized user and shuts down this game.
func (s *GameServerMaster) StopGame(ctx context.Context, in *pb.StopGameRequest) (*pb.StopGameResponse, error) {
	log.Println("StopGame", in)
	// TODO
	return &pb.StopGameResponse{}, nil
}

// otherSlavesUpdateState updates the state of all slaves except skipSlave.
func (s *GameServerMaster) otherSlavesUpdateState(skipSlave string, state *messages.Game_State) {
	// TODO: Consider some sort of watcher thread instead.
	for id, slaveCli := range s.slaves { // TODO: Consider some sort of watcher thread instead.
		if id == skipSlave {
			continue
		}

		_, err := slaveCli.UpdateState(context.Background(), &pb.UpdateStateRequest{
			State: state,
		})
		if err != nil {
			fmt.Println("TODO: DO SOMETHING, unable to update slave state", err)

		}
	}
}

// validateSlave returns its unique InstanceID. If anything goes wrong returns a GRPC status error.
func validateSlave(ctx context.Context) (string, error) {
	x509Cert, err := auth.X509CertificateFromContext(ctx)
	if err != nil {
		return "", status.Errorf(codes.Unauthenticated, err)
	}
	if !contains(x509Cert.DNSNames, string(tlsca.GameSlave)) {
		return "", status.Error(codes.Unauthenticated, "peer is not a slave")
	}
	if !contains(x509Cert.DNSNames, gameID) {
		return "", status.Error(codes.Unauthenticated, "peer is not a slave")
	}

	if len(x509Cert.Subject.CommonName) == 0 {
		return "", status.Error(codes.Unauthenticated, "slave common name empty")
	}
	return x509Cert.Subject.CommonName, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
