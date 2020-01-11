package chess

import (
	"context"
	"testing"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

var (
	tNow = time.Now()
)

func TestInitialization(t *testing.T) {
	_, _, err := initializedDefaultGame()
	if err != nil {
		t.Fatal(err)
	}
}

func initializedDefaultGame() (*Implementation, *pb.InitializeResponse, error) {
	c := &Implementation{}

	ctx := context.TODO()

	o, err := c.Initialize(ctx, &pb.InitializeRequest{
		Game: &messages.Game{
			Id:        "testID",
			Location:  "testLocation",
			StartTime: tNow.UnixNano(),
			History:   nil,
			Metadata: &messages.Game_Metadata{
				Description: "testDescription",
				Title:       "testTitle",
				Visibility:  messages.Game_Metadata_OPEN,
				Rules: &messages.Game_Metadata_Rules{
					VoteApplication: &messages.Game_Metadata_Rules_VoteAppliedImmediately_{},
					GameSpecific: &messages.Game_Metadata_Rules_ChessRules{
						ChessRules: &games.ChessRules{
							TeamSwitching: true,
							BalancedTeams: true,
							BalanceEnforcement: &games.ChessRules_TolerateDifference{
								TolerateDifference: 10,
							},
						},
					},
				},
			},
			State: &messages.Game_State{
				Game: &messages.Game_State_ChessState{
					ChessState: &games.ChessState{
						BoardFen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
						RoundIndex:     1,
						RoundStartTime: tNow.UnixNano(),
						RoundEndTime:   tNow.Add(time.Minute * 10).UnixNano(),
						Details: &games.ChessState_Details{
							PlayerIdToTeam: map[string]bool{},
							PlayerToMove:   map[string]string{},
						},
					},
				},
			},
		},
	})
	return c, o, err
}
