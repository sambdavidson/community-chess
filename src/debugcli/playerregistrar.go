package main

import (
	"context"
	"fmt"
	"time"

	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

// ARGS
const (
	USERNAME = "username"
)

func createPlayerAction(args actionArgs) {
	username, ok := args.otherUserInput[USERNAME]
	if !ok {
		fmt.Printf("missing 'username' arg\n")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := prCli.RegisterPlayer(ctx, &pr.RegisterPlayerRequest{
		Username: username,
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("Response: %v\n", out)
	addPlayer(out.GetPlayer())
}

func getPlayerAction(args actionArgs) {
	if args.playerID == "" {
		fmt.Printf("player ID required\n")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := prCli.GetPlayer(ctx, &pr.GetPlayerRequest{
		PlayerId: args.playerID,
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	fmt.Printf("Response: %v\n", out)
	addPlayer(out.GetPlayer())
}
