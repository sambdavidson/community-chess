package server

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sambdavidson/community-chess/src/lib/auth/grpcplayertokens"
	"github.com/sambdavidson/community-chess/src/playerregistrar/server/playertoken"
	"github.com/sambdavidson/community-chess/src/proto/messages"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

// Opts contains initialization options for a the player registrar server.
type Opts struct {
}

// Server implements an in memory Player Registrar
type Server struct {
	mux sync.RWMutex

	players             map[string]*messages.Player
	usernameCounts      map[string]int32
	usernameNumbersToID map[string]map[int32]string

	tokenIssuer *playertoken.Issuer
}

// New returns a new server that implements a player registrar.
func New(opts *Opts) (*Server, error) {
	// TODO Query DB
	iss, err := playertoken.NewTokenIssuer(nil)
	if err != nil {
		return nil, err
	}
	return &Server{
		players:             map[string]*messages.Player{},
		usernameCounts:      map[string]int32{},
		usernameNumbersToID: map[string]map[int32]string{},
		tokenIssuer:         iss,
	}, nil
}

// RegisterPlayer registers a new player
func (s *Server) RegisterPlayer(ctx context.Context, in *pb.RegisterPlayerRequest) (*pb.RegisterPlayerResponse, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if len(in.GetUsername()) < 2 {
		return nil, status.Error(codes.InvalidArgument, "Username too short.")
	}
	count := s.usernameCounts[in.GetUsername()] + 1
	if count > 9999 {
		return nil, status.Error(codes.ResourceExhausted, "Username all used up.")
	}
	s.usernameCounts[in.GetUsername()] = count

	pid := uuid.New().String()
	out := &pb.RegisterPlayerResponse{
		Player: &messages.Player{
			Id:           pid,
			CreationTime: time.Now().UnixNano(),
			NumberSuffix: count,
			Username:     in.GetUsername(),
		},
	}

	if countToID, ok := s.usernameNumbersToID[in.GetUsername()]; ok {
		countToID[count] = pid
	} else {
		s.usernameNumbersToID[in.GetUsername()] = map[int32]string{count: pid}
	}

	s.players[pid] = out.Player
	return out, nil
}

// GetPlayer gets an existing player's details
func (s *Server) GetPlayer(ctx context.Context, in *pb.GetPlayerRequest) (*pb.GetPlayerReponse, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	player, ok := s.players[in.GetPlayerId()]
	if !ok {
		return nil, status.Error(codes.NotFound, "Unknown player")
	}
	return &pb.GetPlayerReponse{
		Player: player,
	}, nil
}

// Login validates the login credentials and returns a short lived player ID token if successful.
func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	if countToID, ok := s.usernameNumbersToID[in.GetUsername()]; ok {
		if pid, ok := countToID[in.GetNumberSuffix()]; ok {
			if player, ok := s.players[pid]; ok {
				token, err := s.tokenIssuer.TokenForPlayer(player)
				if err != nil {
					return nil, status.Error(codes.Internal, err.Error())
				}
				return &pb.LoginResponse{
					Token: token,
				}, nil
			}
			return nil, status.Error(codes.Internal, "Missing player data.")
		}
	}
	return nil, status.Error(codes.PermissionDenied, "Bad login.")
}

// RefreshToken exchanges the current player token for a fresh one.
func (s *Server) RefreshToken(ctx context.Context, in *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	pid, err := grpcplayertokens.ValidatedPlayerIDFromIncomingContext(ctx)
	if err != nil {
		return nil, err
	}
	pRes, err := s.GetPlayer(ctx, &pb.GetPlayerRequest{
		PlayerId: pid,
	})
	if err != nil {
		return nil, err
	}

	token, err := s.tokenIssuer.TokenForPlayer(pRes.GetPlayer())

	return &pb.RefreshTokenResponse{
		Token: token,
	}, err
}

// TokenPublicKeys returns the current set of all TokenPublicKeys
func (s *Server) TokenPublicKeys(ctx context.Context, in *pb.TokenPublicKeysRequest) (*pb.TokenPublicKeysResponse, error) {
	return &pb.TokenPublicKeysResponse{
		History: s.tokenIssuer.PublicKeys(),
	}, nil
}
