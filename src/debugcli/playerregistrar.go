package main

import (
	"context"
	"fmt"
	"time"

	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
)

func createPlayerAction(cmdParts []string) {
	if len(cmdParts) < 2 {
		fmt.Println("missing player username")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := prCli.RegisterPlayer(ctx, &pr.RegisterPlayerRequest{
		Username: cmdParts[1],
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	knownPlayers[out.GetPlayer().GetId()] = out.GetPlayer()
	activePlayer = out.GetPlayer()
	fmt.Printf("created, updated active player: %v\n", out)
}

func getPlayerAction(cmdParts []string) {
	var id string
	if len(cmdParts) < 2 {
		if activePlayer == nil {
			fmt.Println("missing player_id, either set an active player or define one.")
			return
		}
		id = activePlayer.GetId()
	} else {
		id = cmdParts[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := prCli.GetPlayer(ctx, &pr.GetPlayerRequest{
		PlayerId: id,
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	knownPlayers[out.GetPlayer().GetId()] = out.GetPlayer()
	activePlayer = out.GetPlayer()
	fmt.Printf("got, updated active player: %v\n", out)
}
