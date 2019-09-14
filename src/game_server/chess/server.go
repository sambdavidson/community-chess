package chess

import (
	"context"
	"sync"

	"google.golang.org/grpc/codes"

	gs "github.com/samdamana/community-chess/src/proto/services/game_server"
	pr "github.com/samdamana/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// Server implement the GameServer service.
type Server struct {
	server          *grpc.Server
	conn            *grpc.ClientConn
	playerRegistrar pr.PlayerRegistrarClient

	mux sync.Mutex
}

// Opts is the options for setting up a GameServer.
type Opts struct {
	Server                 *grpc.Server
	PlayerRegistrarAddress string
}

// NewServer builds a new Server object
func NewServer(o Opts) (*Server, error) {
	conn, err := grpc.Dial(o.PlayerRegistrarAddress, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	s := &Server{
		server:          o.Server,
		conn:            conn,
		playerRegistrar: pr.NewPlayerRegistrarClient(conn),
	}

	return s, nil
}

// GetGame gets the game details given a GetGameRequest
func (s *Server) GetGame(ctx context.Context, in *gs.GetGameRequest) (*gs.GetGameResponse, error) {
	return nil, status.Error(codes.Unimplemented, "TODO implement GetGame")
}

// AddPlayer adds a player to the existing game
func (s *Server) AddPlayer(ctx context.Context, in *gs.AddPlayerRequest) (*gs.AddPlayerResponse, error) {
	return nil, status.Error(codes.Unimplemented, "TODO implement AddPlayer")
}

// RemovePlayer removes a player from the current game
func (s *Server) RemovePlayer(ctx context.Context, in *gs.RemovePlayerRequest) (*gs.RemovePlayerResponse, error) {
	return nil, status.Error(codes.Unimplemented, "TODO implement RemovePlayer")
}

// PostVotes posts 1+ votes to the current game
func (s *Server) PostVotes(ctx context.Context, in *gs.PostVotesRequest) (*gs.PostVotesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "TODO implement PostVotes")
}
