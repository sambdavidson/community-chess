package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sambdavidson/community-chess/src/lib/auth"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

func getGameAction(cmdParts []string) {
	// get_game <game_id>
	var id string
	if len(cmdParts) < 2 {
		if activeGame == nil {
			fmt.Println("missing game_id, either set an game_id or define one.")
			return
		}
		id = activeGame.GetId()
	} else {
		id = cmdParts[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pCtx, err := auth.AppendPlayerIDToOutgoingContext(ctx, id)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	out, err := gsCli.Game(pCtx, &gs.GameRequest{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	log.Println("ok", out, err)
}

func joinAction(cmdParts []string) {
	// add_player <player_id>
	var pid string
	if len(cmdParts) < 2 {
		if activePlayer == nil {
			fmt.Println("missing player_id, either set an active player or define one.")
			return
		}
		pid = activePlayer.GetId()
	} else {
		pid = cmdParts[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pCtx, err := auth.AppendPlayerIDToOutgoingContext(ctx, pid)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	out, err := gsCli.Join(pCtx, &gs.JoinRequest{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out)
}

func leaveAction(cmdParts []string) {
	// leave <player_id> <game_id>
	var pid string
	if len(cmdParts) < 3 {
		if activePlayer == nil {
			fmt.Println("missing player_id, either set an active player or define one.")
			return
		}
		if activeGame == nil {
			fmt.Println("missing game_id, either set an active player or define one.")
			return
		}
		pid = activePlayer.GetId()
	} else {
		pid = cmdParts[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pCtx, err := auth.AppendPlayerIDToOutgoingContext(ctx, pid)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	out, err := gsCli.Leave(pCtx, &gs.LeaveRequest{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out)
}
