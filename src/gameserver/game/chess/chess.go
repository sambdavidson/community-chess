package chess

import (
	"context"
	"sync"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	ch "github.com/notnil/chess"
	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Implementation is an implementation of chess for use by both the master and slave
type Implementation struct {
	initialized bool

	gameMux    sync.Mutex
	startTime  time.Time
	endTime    time.Time
	game       *ch.Game
	roundIndex int32

	teamsMux sync.Mutex
	// player ID to is_white_team
	playerToTeam map[string]bool
	teamToCount  map[bool]int64

	moveMux sync.Mutex
	// Move in the form of Algebraic Notation
	acceptingVotes bool
	playerToMove   map[string]string
	moveToCount    map[string]int64

	// Game proto stuff, the state is built dynamically.
	metadata *messages.Game_Metadata
	history  *games.ChessHistory
}

// resetWithState resets all the state variables of this chess implementation and updates them to the input state.
// This function is NON-LOCKING so wrap it in a mux if necessary.
func (i *Implementation) resetWithState(s *games.ChessState) {
	i.startTime = time.Unix(0, s.GetRoundStartTime())
	i.endTime = time.Unix(0, s.GetRoundEndTime())
	fen, _ := ch.FEN(s.GetBoardFen())
	i.game = ch.NewGame(fen)
	i.roundIndex = s.GetRoundIndex()
	// TODO figure out if copying inputs is necessary
	i.playerToTeam = s.GetDetails().GetPlayerIdToTeam()
	if i.playerToTeam == nil {
		i.playerToTeam = map[string]bool{}
	}
	i.teamToCount = map[bool]int64{
		true:  s.GetWhiteTeamCount(),
		false: s.GetBlackTeamCount(),
	}
	i.playerToMove = s.GetDetails().GetPlayerToMove()
	if i.playerToMove == nil {
		i.playerToMove = map[string]string{}
	}
	i.moveToCount = s.GetMoveToCount()
	if i.moveToCount == nil {
		i.moveToCount = map[string]int64{}
	}
	i.history = &games.ChessHistory{
		StateHistory: []*games.ChessState{},
	}
}

// Initialize initializes this server to run the game defined in InitializeRequest.
func (i *Implementation) Initialize(ctx context.Context, in *pb.InitializeRequest) (*pb.InitializeResponse, error) {
	if err := validateChessRules(in.GetGame().GetMetadata().GetRules().GetChessRules()); err != nil {
		return nil, err
	}
	if err := validateChessState(in.GetGame().GetState().GetChessState(), true); err != nil {
		return nil, err
	}

	i.resetWithState(in.GetGame().GetState().GetChessState())
	if m := in.GetGame().GetMetadata(); m != nil {
		i.metadata = m
	}
	if h := in.GetGame().GetHistory().GetChessHistory(); h != nil {
		i.history = h
	}
	i.initialized = true
	return &pb.InitializeResponse{}, nil
}

// UpdateMetadata is called by GameServerMasters to update this slave's metadata.
func (i *Implementation) UpdateMetadata(ctx context.Context, in *pb.UpdateMetadataRequest) (*pb.UpdateMetadataResponse, error) {
	i.metadata = in.GetMetadata()
	return nil, status.Error(codes.Unimplemented, "todo")
}

// UpdateState is called by GameServerMasters to update this slave's state of the game.
func (i *Implementation) UpdateState(ctx context.Context, in *pb.UpdateStateRequest) (*pb.UpdateStateResponse, error) {
	validateChessState(in.GetState().GetChessState(), true)

	i.gameMux.Lock()
	i.teamsMux.Lock()
	i.moveMux.Lock()
	i.resetWithState(in.GetState().GetChessState())
	i.moveMux.Unlock()
	i.teamsMux.Unlock()
	i.gameMux.Unlock()

	return &pb.UpdateStateResponse{}, nil
}

// TODO figure out if copying inputs is necessary
// func copyMetadata(in *messages.Game_Metadata) *messages.Game_Metadata {
// 	o := &messages.Game_Metadata{
// 		Description: in.GetDescription(),
// 		Rules: &messages.Game_Metadata_Rules{
// 			GameSpecific: &messages.Game_Metadata_Rules_ChessRules{
// 				ChessRules: &games.ChessRules{
// 					BalancedTeams:      in.GetRules().GetChessRules().GetBalancedTeams(),
// 					BalanceEnforcement: in.GetRules().GetChessRules().GetBalanceEnforcement(),
// 					TeamSwitching:      in.GetRules().GetChessRules().GetTeamSwitching(),
// 				},
// 			},
// 		},
// 		Title:      in.GetTitle(),
// 		Visibility: in.GetVisibility(),
// 	}
// 	if in.GetRules().GetVoteAppliedAfterTally() != nil {
// 		o.Rules.VoteApplication = &messages.Game_Metadata_Rules_VoteAppliedAfterTally_{
// 			VoteAppliedAfterTally: &messages.Game_Metadata_Rules_VoteAppliedAfterTally{
// 				SelectionType:   in.GetRules().GetVoteAppliedAfterTally().GetSelectionType(),
// 				TimeoutSeconds:  in.GetRules().GetVoteAppliedAfterTally().GetTimeoutSeconds(),
// 				WaitFullTimeout: in.GetRules().GetVoteAppliedAfterTally().GetWaitFullTimeout(),
// 			},
// 		}
// 	} else if in.GetRules().GetVoteAppliedImmediately() != nil {
// 		o.GetRules().VoteApplication = &messages.Game_Metadata_Rules_VoteAppliedImmediately_{
// 			VoteAppliedImmediately: &messages.Game_Metadata_Rules_VoteAppliedImmediately{},
// 		}
// 	}

// 	if in.GetRules().GetChessRules().GetTolerateDifference() > 0 {
// 		o.GetRules().GetChessRules().BalanceEnforcement = &games.ChessRules_TolerateDifference{
// 			TolerateDifference: in.GetRules().GetChessRules().GetTolerateDifference(),
// 		}
// 	} else if in.GetRules().GetChessRules().GetToleratePercent() > 0 {
// 		o.GetRules().GetChessRules().BalanceEnforcement = &games.ChessRules_ToleratePercent{
// 			ToleratePercent: in.GetRules().GetChessRules().GetToleratePercent(),
// 		}
// 	}

// 	return o
// }
