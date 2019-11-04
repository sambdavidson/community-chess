package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	"github.com/sambdavidson/community-chess/src/proto/messages/games"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

func initializeAction(cmdParts []string) {
	if len(cmdParts) != 2 {
		fmt.Println("badly formatted request")
		return
	}
	id := cmdParts[1]

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	out, err := gmCli.Initialize(ctx, &gs.InitializeRequest{
		Game: &messages.Game{
			Id:   id,
			Type: messages.Game_CHESS,
			Metadata: &messages.Game_Metadata{
				Title: "foo",
				Rules: &messages.Game_Metadata_Rules{
					GameSpecific: &messages.Game_Metadata_Rules_ChessRules{
						ChessRules: &games.ChessRules{
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
						RoundStartTime: time.Now().UnixNano(),
						RoundEndTime:   time.Now().Add(time.Minute * 10).UnixNano(),
						Details: &games.ChessState_Details{
							PlayerIdToTeam: map[string]bool{},
							PlayerToMove:   map[string]string{},
						},
					},
				},
			},
		},
	})
	log.Println(out, err)
}

func postVotesAction(cmdParts []string) {
	// post_votes <player_id>
	fmt.Println("post_votes not yet implemented on the cli")
	// if len(cmdParts) < 3 {
	// 	fmt.Println("missing player_id")
	// 	return
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// out, err :=
	// if err != nil {
	// 	fmt.Printf("error: %v\n", err)
	// 	return
	// }
	// fmt.Printf("ok: %v\n", out.GetGame())
}
