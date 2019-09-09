/* BUILD and RUN
go run .\src\player_registrar
*/

// Package main implements a server for the Player Registrar
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/samdamana/community-chess/src/proto/messages"

	pb "github.com/samdamana/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50052, "port the Game Server is accepts connections")
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPlayerRegistrarServer(s, &Server{})

	log.Printf("Starting listen of Player Registrar on port %v\n", *port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Server implements an in memory Player Registrar
type Server struct {
	mux sync.Mutex

	players        map[string]*messages.Player
	usernameCounts map[string]int32
}

// RegisterPlayer registers a new player
func (s *Server) RegisterPlayer(ctx context.Context, in *pb.RegisterPlayerRequest) (*pb.RegisterPlayerResponse, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	log.Printf("RegisterPlayer: %v\n", in)

	if len(in.GetUsername()) < 2 {
		return nil, status.Error(codes.InvalidArgument, "Username too short.")
	}
	count := s.usernameCounts[in.GetUsername()]
	if count > 9999 {
		return nil, status.Error(codes.ResourceExhausted, "Usename all used up.")
	}
	s.usernameCounts[in.GetUsername()] = count + 1

	id := uuid.New().String()
	out := &pb.RegisterPlayerResponse{
		Player: &messages.Player{
			Id: &messages.Player_Id{
				Id: id,
			},
			CreationTime: time.Now().UnixNano(),
			NumberSuffix: count + 1,
			Username:     in.GetUsername(),
		},
	}

	s.players[id] = out.Player
	return out, nil
}

// GetPlayer gets an existing player's details
func (s *Server) GetPlayer(ctx context.Context, in *pb.GetPlayerRequest) (*pb.GetPlayerReponse, error) {
	log.Printf("GetPlayer: %v\n", in)
	player, ok := s.players[in.GetPlayerId().Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "Unknown player")
	}
	return &pb.GetPlayerReponse{
		Player: player,
	}, nil
}
