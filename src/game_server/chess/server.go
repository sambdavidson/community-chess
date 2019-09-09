package chess

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/samdamana/community-chess/src/proto/messages"

	gs "github.com/samdamana/community-chess/src/proto/services/game_server"
	pr "github.com/samdamana/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc"
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
	log.Println(o.PlayerRegistrarAddress)
	go s.commandLineCli()

	return s, nil
}

// GetGame gets the game details given a GetGameRequest
func (s *Server) GetGame(ctx context.Context, in *gs.GetGameRequest) (*gs.GetGameResponse, error) {
	return nil, nil
}

// AddPlayer adds a player to the existing game
func (s *Server) AddPlayer(ctx context.Context, in *gs.AddPlayerRequest) (*gs.AddPlayerResponse, error) {
	return nil, nil
}

// RemovePlayer removes a player from the current game
func (s *Server) RemovePlayer(ctx context.Context, in *gs.RemovePlayerRequest) (*gs.RemovePlayerResponse, error) {
	return nil, nil
}

// PostVotes posts 1+ votes to the current game
func (s *Server) PostVotes(ctx context.Context, in *gs.PostVotesRequest) (*gs.PostVotesResponse, error) {
	return nil, nil
}

func (s *Server) commandLineCli() {
	reader := bufio.NewReader(os.Stdin)
	for true {
		fmt.Print("Enter Command: ")
		text, _ := reader.ReadString('\n')
		switch strings.ToLower(strings.Trim(text, " \n\r")) {
		case "f":
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			out, err := s.playerRegistrar.GetPlayer(ctx, &pr.GetPlayerRequest{
				PlayerId: &messages.Player_Id{
					Id: "foo",
				},
			})
			fmt.Println(out, err)
		case "q", "exit", "quit":
			s.shutdown()
			return
		}
	}
}

func (s *Server) shutdown() {
	s.conn.Close()
	s.server.Stop()
}
