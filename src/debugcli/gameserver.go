package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sambdavidson/community-chess/src/lib/auth"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

func init() {
	commands["get_game"] = command{
		helpText: "gets a game, updates default active game too.\n\tParams: `gameID`, `playerID`",
		action:   get,
	}
}

func get(args actionArgs) {
	if args.gameID == "" || args.playerID == "" {
		fmt.Printf("game and player ID required\n")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pCtx, err := auth.AppendPlayerIDToOutgoingContext(ctx, args.playerID)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	out, err := gsCli.Game(pCtx, &gs.GameRequest{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("Response: %v\n", out)
	addGame(out.GetGame())
}

func join(args actionArgs) {
	if args.gameID == "" || args.playerID == "" {
		fmt.Printf("game and player ID required\n")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pCtx, err := auth.AppendPlayerIDToOutgoingContext(ctx, args.playerID)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	out, err := gsCli.Join(pCtx, &gs.JoinRequest{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("Response: %v\n", out)
}

func leave(args actionArgs) {
	if args.gameID == "" || args.playerID == "" {
		fmt.Printf("game and player ID required\n")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	pCtx, err := auth.AppendPlayerIDToOutgoingContext(ctx, args.playerID)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	out, err := gsCli.Leave(pCtx, &gs.LeaveRequest{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("Response: %v\n", out)
}
