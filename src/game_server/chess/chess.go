package chess

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"google.golang.org/grpc/codes"

	"github.com/notnil/chess"
	"github.com/sambdavidson/community-chess/src/proto/messages"
	gs "github.com/sambdavidson/community-chess/src/proto/services/game_server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc/status"
)

var (
	notation    = chess.AlgebraicNotation{}
	startingFEN = flag.String("chess_starting_fen", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", "Starting postion of chess pieces in Forsyth-Edwards notation.")
)

const (
	gameType = "chess"
)

// Server implement the GameServer service.
type Server struct {
	mux sync.Mutex

	// Pre-start state
	id                 string
	playerRegistrarCli pr.PlayerRegistrarClient

	// Runtime game state
	game                  *messages.Game
	chessGame             *chess.Game
	voteAppliedAfterTally bool
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
	var err error
	if err = validateGenericMetadata(in.GetMetadata()); err != nil {
		return nil, err
	}
	if err = validateChessRules(in.GetMetadata().GetRules().GetChessRules()); err != nil {
		return nil, err
	}
	fen, err := chess.FEN(*startingFEN)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "bad starting FEN %q; err %v", *startingFEN, err)
	}

	startTime := time.Now().UnixNano()
	s.chessGame = chess.NewGame(fen)
	s.game = &messages.Game{
		Id:        s.id,
		StartTime: startTime,
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
				PlayerIdToTeamWhiteIsTrue: map[string]bool{},
				Rounds: []*games.ChessState_ChessRound{
					&games.ChessState_ChessRound{
						StartingBoardFen: s.chessGame.FEN(),
						PlayerToMove:     map[string]string{},
						MoveToCount:      map[string]int64{},
						StartTime:        startTime,
					},
				},
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
	if ok, err := s.validateGameID(in.GetGameId()); !ok {
		return nil, err
	}
	return &gs.GetGameResponse{
		Game: s.game,
	}, nil
}

// AddPlayer adds a player to the existing game
func (s *Server) AddPlayer(ctx context.Context, in *gs.AddPlayerRequest) (*gs.AddPlayerResponse, error) {
	fmt.Printf("AddPlayer %v", in)
	if ok, err := s.validateGameID(in.GetGameId()); !ok {
		return nil, err
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	desiredTeam := in.GetChessData().GetWhiteTeam()
	currentTeam, ok := s.game.GetChessState().GetPlayerIdToTeamWhiteIsTrue()[in.GetPlayerId()]
	var whiteModifier, blackModifier int64
	if ok {
		if currentTeam == desiredTeam {
			// Player already a part of this team.
			return &gs.AddPlayerResponse{}, nil
		} else if !s.game.GetMetadata().GetRules().GetChessRules().GetTeamSwitching() {
			// Team switching not allowed.
			return nil, status.Errorf(codes.InvalidArgument, "cannot join %s team, already on %s team and team switching is now allowed", teamStr(desiredTeam), teamStr(currentTeam))
		}
		if currentTeam {
			whiteModifier = -1
		} else {
			blackModifier = -1
		}
	}

	if !canJoinTeam(
		desiredTeam,
		s.game.GetChessState().GetWhiteTeamCount()+whiteModifier,
		s.game.GetChessState().GetBlackTeamCount()+blackModifier,
		s.game.GetMetadata().GetRules().GetChessRules().GetTolerateDifference(),
		float64(s.game.GetMetadata().GetRules().GetChessRules().GetToleratePercent()),
	) {
		return nil, status.Errorf(codes.InvalidArgument, "cannot join %s team, disallowed by team balance settings", teamStr(desiredTeam))
	}
	s.game.GetChessState().GetPlayerIdToTeamWhiteIsTrue()[in.GetPlayerId()] = desiredTeam
	if desiredTeam {
		s.game.GetChessState().WhiteTeamCount += (1 + whiteModifier)
	} else {
		s.game.GetChessState().BlackTeamCount += (1 + whiteModifier)
	}

	return &gs.AddPlayerResponse{}, nil

}

// RemovePlayer removes a player from the current game
func (s *Server) RemovePlayer(ctx context.Context, in *gs.RemovePlayerRequest) (*gs.RemovePlayerResponse, error) {
	fmt.Printf("RemovePlayer %v", in)
	if ok, err := s.validateGameID(in.GetGameId()); !ok {
		return nil, err
	}

	return nil, status.Error(codes.Unimplemented, "todo implement RemovePlayer")
}

// PostVotes posts 1+ votes to the current game
func (s *Server) PostVotes(ctx context.Context, in *gs.PostVotesRequest) (*gs.PostVotesResponse, error) {
	fmt.Printf("PostVotes %v", in)
	if ok, err := s.validateGameID(in.GetGameId()); !ok {
		return nil, err
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	if in.GetRoundIndex() != s.currentRoundIndex() {
		return nil, status.Errorf(codes.InvalidArgument, "bad round index %d, the current round is %d", in.GetRoundIndex(), s.currentRoundIndex())
	}
	// Validate the votes.
	for _, vote := range in.GetVotes() {
		playerWhite, ok := s.game.GetChessState().GetPlayerIdToTeamWhiteIsTrue()[vote.GetPlayerId()]
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "player %q not part of game %q", vote.GetPlayerId(), in.GetGameId())
		}
		if playerWhite != (s.chessGame.Position().Turn() == chess.White) {
			return nil, status.Errorf(codes.InvalidArgument, "%s player %q cannot cast vote on %s's turn", teamStr(playerWhite), vote.GetPlayerId(), teamStr(!playerWhite))
		}
		_, err := notation.Decode(s.chessGame.Position(), vote.GetChessVote().GetMove())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "bad move %q for player %q", vote.GetChessVote().GetMove(), vote.GetPlayerId())
		}
	}

	round := s.currentRound()
	//Cast the votes
	for _, vote := range in.GetVotes() {
		oldMove, ok := round.GetPlayerToMove()[vote.GetPlayerId()]
		if ok {
			round.GetMoveToCount()[oldMove]--
		}
		round.GetPlayerToMove()[vote.GetPlayerId()] = vote.GetChessVote().GetMove()
		round.GetMoveToCount()[vote.GetChessVote().GetMove()]++
	}

	return &gs.PostVotesResponse{}, nil
}

// StopGame starts the game defined in the request
func (s *Server) StopGame(ctx context.Context, in *gs.StopGameRequest) (*gs.StopGameResponse, error) {
	fmt.Printf("StopGame %v", in)
	if ok, err := s.validateGameID(in.GetGameId()); !ok {
		return nil, err
	}

	return nil, status.Error(codes.Unimplemented, "todo implement StopGame")
}

// ListGames starts the game defined in the request
func (s *Server) ListGames(ctx context.Context, in *gs.ListGamesRequest) (*gs.ListGamesResponse, error) {
	fmt.Printf("ListGames %v", in)

	return nil, status.Error(codes.Unimplemented, "todo implement ListGames")
}

/************************ HELPERS ************************/

// validateGameID returns a grpc error if the gameId doesn't match this game's id.
func (s *Server) validateGameID(gameID string) (bool, error) {
	if gameID != s.id {
		return false, status.Errorf(codes.NotFound, "request game id %q does not match this game's id %q", gameId, s.id)
	}
	return true, nil
}

func (s *Server) currentRoundIndex() int {
	return len(s.game.GetChessState().GetRounds()) - 1
}

func (s *Server) currentRound() *games.ChessState_ChessRound {
	return s.game.GetChessState().GetRounds()[s.currentRoundIndex()]
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

func canJoinTeam(team bool, whiteCount, blackCount, tolerateDifference int64, toleratePercent float64) bool {
	var whiteMod, blackMod int64
	if team {
		if whiteCount == 0 {
			return true
		}
		whiteMod = 1
	} else {
		if blackCount == 0 {
			return true
		}
		blackMod = 1
	}
	newWhite := whiteCount + whiteMod
	newBlack := blackCount + blackMod
	diff := newWhite - newBlack
	if tolerateDifference >= 1 {
		if diff < 0 {
			diff = diff * -1
		}
		if diff > tolerateDifference {
			return false
		}
	} else {
		bigger := newWhite
		smaller := newBlack
		if smaller > bigger {
			smaller = newWhite
			bigger = newBlack
		}
		if math.Ceil(float64(smaller)*(1.0+toleratePercent)) < float64(bigger) {
			return false
		}
	}
	return true
}

func teamStr(t bool) string {
	if t {
		return "white"
	}
	return "false"
}
