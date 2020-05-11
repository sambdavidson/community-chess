package server

import (
	"context"
	"fmt"

	"github.com/sambdavidson/community-chess/src/playerregistrar/database"

	"github.com/sambdavidson/community-chess/src/lib/auth/grpcplayertokens"
	"github.com/sambdavidson/community-chess/src/playerregistrar/server/playertoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

// Opts contains initialization options for a the player registrar server.
type Opts struct {
	DB database.Database
}

// Server implements an in memory Player Registrar
type Server struct {
	db database.Database

	tokenIssuer *playertoken.Issuer
}

// New returns a new server that implements a player registrar.
func New(opts *Opts) (*Server, error) {
	iss, err := playertoken.NewTokenIssuer(&playertoken.IssuerOpts{
		DB: opts.DB,
	})
	if err != nil {
		return nil, err
	}

	if opts.DB == nil {
		return nil, fmt.Errorf("database in options cannot be nil")
	}
	return &Server{opts.DB, iss}, nil
}

// RegisterPlayer registers a new player
func (s *Server) RegisterPlayer(ctx context.Context, in *pb.RegisterPlayerRequest) (*pb.RegisterPlayerResponse, error) {
	if len(in.GetUsername()) < 2 {
		return nil, status.Error(codes.InvalidArgument, "Username too short.")
	}
	player, err := s.db.RegisterPlayer(in.GetUsername())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.RegisterPlayerResponse{
		Player: player,
	}, nil
}

// GetPlayer gets an existing player's details
func (s *Server) GetPlayer(ctx context.Context, in *pb.GetPlayerRequest) (*pb.GetPlayerReponse, error) {
	player, err := s.db.GetPlayerByID(in.GetPlayerId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetPlayerReponse{
		Player: player,
	}, nil
}

// Login validates the login credentials and returns a short lived player ID token if successful.
func (s *Server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	player, err := s.db.GetPlayerByUsername(in.GetUsername(), in.GetNumberSuffix())
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	token, err := s.tokenIssuer.TokenForPlayer(player)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.LoginResponse{
		Token: token,
	}, nil
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
