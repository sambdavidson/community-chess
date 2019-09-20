package chess

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"google.golang.org/grpc/codes"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	gs "github.com/sambdavidson/community-chess/src/proto/services/game_server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc/status"
)

const (
	gameType = "chess"
)

// Server implement the GameServer service.
type Server struct {
	mux sync.Mutex

	// Pre-start state
	id                 string
	startTime          int64
	playerRegistrarCli pr.PlayerRegistrarClient

	// Runtime game state
	game                  *messages.Game
	voteAppliedAfterTally bool
}

// Opts contains the options for building a chess server
type Opts struct {
	playerRegistrarCli pr.PlayerRegistrarClient
	gameMetadata       *messages.Game_Metadata
}

// NewServer builds a new Server object
func NewServer(id string, prCli pr.PlayerRegistrarClient) (*Server, error) {
	s := &Server{
		id:                 id,
		playerRegistrarCli: prCli,
	}

	return s, nil
}

func (s *Server) infof(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("[%s]: ", s.id)+format, v)
}

// StartGame starts the game defined in the request
func (s *Server) StartGame(ctx context.Context, in *gs.StartGameRequest) (*gs.StartGameResponse, error) {
	s.infof("StartGame(%v)", in)
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.game != nil {
		return nil, status.Errorf(codes.Unavailable, "game already started")
	}
	if in.GetGameType() != gameType {
		return nil, status.Errorf(codes.InvalidArgument, "game type is not %s: %q", gameType, in.GetGameType())
	}

	err := validateGenericMetadata(in.GetMetadata())
	if err != nil {
		return nil, err
	}
	err = validateChessRules(in.GetMetadata().GetRules().GetChessRules())
	if err != nil {
		return nil, err
	}

	s.game = &messages.Game{
		Id: &messages.Game_Id{
			Id: s.id,
		},
		StartTime: time.Now().UnixNano(),
		Metadata: &messages.Game_Metadata{
			Title:      in.GetMetadata().GetTitle(),
			Visibility: in.GetMetadata().GetVisibility(),
			Rules: &messages.Game_Metadata_Rules{
				VoteApplication: in.GetMetadata().GetRules().GetVoteApplication(),
				GameSpecificRules: &messages.Game_Metadata_Rules_ChessRules{
					ChessRules: &games.ChessRules{
						BalancedTeams:      in.GetMetadata().GetRules().GetChessRules().GetBalancedTeams(),
						BalanceEnforcement: in.GetMetadata().GetRules().GetChessRules().GetBalanceEnforcement(),
					},
				},
			},
		},
		State: &messages.Game_ChessState{
			ChessState: &games.ChessState{
				// TODO: figure out precide round timings
			},
		},
	}

	return &gs.StartGameResponse{
		Game: s.game,
	}, nil
}

// GetGame gets the game details given a GetGameRequest
func (s *Server) GetGame(ctx context.Context, in *gs.GetGameRequest) (*gs.GetGameResponse, error) {
	fmt.Printf("GetGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement GetGame")
}

// AddPlayer adds a player to the existing game
func (s *Server) AddPlayer(ctx context.Context, in *gs.AddPlayerRequest) (*gs.AddPlayerResponse, error) {
	fmt.Printf("AddPlayer %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement AddPlayer")
}

// RemovePlayer removes a player from the current game
func (s *Server) RemovePlayer(ctx context.Context, in *gs.RemovePlayerRequest) (*gs.RemovePlayerResponse, error) {
	fmt.Printf("RemovePlayer %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement RemovePlayer")
}

// PostVotes posts 1+ votes to the current game
func (s *Server) PostVotes(ctx context.Context, in *gs.PostVotesRequest) (*gs.PostVotesResponse, error) {
	fmt.Printf("PostVotes %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement PostVotes")
}

// StopGame starts the game defined in the request
func (s *Server) StopGame(ctx context.Context, in *gs.StopGameRequest) (*gs.StopGameResponse, error) {
	fmt.Printf("StopGame %v", in)
	return nil, status.Error(codes.Unimplemented, "TODO implement StopGame")
}

// ListGames starts the game defined in the request
func (s *Server) ListGames(ctx context.Context, in *gs.ListGamesRequest) (*gs.ListGamesResponse, error) {
	fmt.Printf("ListGames %v", in)

	return nil, status.Error(codes.Unimplemented, "TODO implement ListGames")
}

// Returns a grpc error if something is encountered.
func validateGenericMetadata(m *messages.Game_Metadata) error {
	if m.GetTitle() == "" {
		return status.Errorf(codes.InvalidArgument, "missing title")
	}
	r := m.GetRules()
	if r.GetVoteAppliedAfterTally() != nil {
		if r.GetVoteAppliedAfterTally().GetTimeoutSeconds() < 3 {
			return status.Errorf(codes.InvalidArgument, "timeout too short, must be 3 or more seconds")
		}
	} else if r.GetVoteAppliedImmediately() == nil {
		return status.Errorf(codes.InvalidArgument, "missing vote application oneof")
	}
	return nil
}

func validateChessRules(r *games.ChessRules) error {
	if r == nil {
		return status.Errorf(codes.InvalidArgument, "missing chess specific rules")
	}
	if r.GetTolerateDifference() != 0 {
		if r.GetTolerateDifference() < 1 {
			return status.Errorf(codes.InvalidArgument, "balance enforcement tolerate difference cannot be less than 1")
		}
		return nil
	} else if r.GetToleratePercent() != 0.0 {
		if r.GetToleratePercent() <= 0.0 {
			return status.Errorf(codes.InvalidArgument, "balance enforcement tolerate percent cannot be less than or equal to 0")
		}
		return nil
	}
	return status.Errorf(codes.InvalidArgument, "balance enforcement undefined")
}
