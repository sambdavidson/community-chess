/* BUILD and RUN
go run .\src\front_end
*/

// Package main implements a simple front_end client for interacting with the various services through a CLI.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	gs "github.com/sambdavidson/community-chess/src/proto/services/game_server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/player_registrar"
	"google.golang.org/grpc"
)

/* FLAGS */
var (
	gameServerURI  = flag.String("game_server_uri", "localhost", "uri of the game server")
	gameServerPort = flag.Int("game_server_port", 50051, "port of the game server")

	playerRegistrarURI  = flag.String("player_registar_uri", "localhost", "URI of the Player Registrar")
	playerRegistrarPort = flag.Int("player_registrar_port", 50052, "Port of the Player Registrar")
)

/* Clients which are defined in init() */
var (
	gsConn       *grpc.ClientConn
	gsCli        gs.GameServerClient
	prConn       *grpc.ClientConn
	prCli        pr.PlayerRegistrarClient
	commands     map[string]command
	knownPlayers = map[string]*messages.Player{}
	activePlayer *messages.Player
	activeGame   *messages.Game
)

func init() {
	flag.Parse()
	gsConn, gsCli = getGameServer()
	prConn, prCli = getPlayerRegistrar()
	commands = map[string]command{
		"create_player": command{
			helpText: "creates a new player with username, updates active player with created player\n\tE.g. 'create_player sam'",
			action:   createPlayerAction,
		},
		"get_player": command{
			helpText: "gets a player, updates active player if found\n\tE.g. 'get_player 123456'",
			action:   getPlayerAction,
		},
		"list_known_players": command{
			helpText: "lists all locally known players",
			action:   listKnownPlayersAction,
		},
		"start_game": command{
			helpText: "starts a game\n\tE.g. 'start_game chess ComeChessItUp'",
			action:   startGameAction,
		},
		"get_game": command{
			helpText: "gets a game\n\tE.g. 'get_game 123456'",
			action:   getGameAction,
		},
		"add_player": command{
			helpText: "adds a player to the game, either the active player or a defined one\n\tE.g. 'add_player' or 'add_player 123456 987654'",
			action:   addPlayerAction,
		},
		"remove_player": command{
			helpText: "removes a player from the game, either the active player or a defined one\n\tE.g. 'remove_player' or 'remove_player 123456 987654'",
			action:   removePlayerAction,
		},
		"post_vote": command{
			helpText: "posts a vote to a game",
			action:   postVotesAction,
		},
		"stop_game": command{
			helpText: "stops a game, either the active one or one defined\n\tE.g. 'stop_game' or 'stop_game 987654'",
			action:   stopGameAction,
		},
		"list_games": command{
			helpText: "list all games by the game server",
			action:   listGamesAction,
		},
		"help": command{
			helpText: "displays all commands and they help text",
			action:   helpAction,
		},
	}
}

type command struct {
	helpText string
	action   func(cmdParts []string)
}

func getGameServer() (*grpc.ClientConn, gs.GameServerClient) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *gameServerURI, *gameServerPort), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	return conn, gs.NewGameServerClient(conn)
}

func getPlayerRegistrar() (*grpc.ClientConn, pr.PlayerRegistrarClient) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *playerRegistrarURI, *playerRegistrarPort), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	return conn, pr.NewPlayerRegistrarClient(conn)
}

func helpAction(cmdParts []string) {
	fmt.Println("Commands: ")
	var keys []string
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cmd := commands[k]
		fmt.Printf("- '%s': %s\n", k, cmd.helpText)
	}
}

func listKnownPlayersAction(cmdParts []string) {
	for k, p := range knownPlayers {
		fmt.Printf("%s : %s %d\n", k, p.GetUsername(), p.GetNumberSuffix())
	}
}

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
	knownPlayers[out.GetPlayer().GetId().GetId()] = out.GetPlayer()
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
		id = activePlayer.GetId().GetId()
	} else {
		id = cmdParts[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := prCli.GetPlayer(ctx, &pr.GetPlayerRequest{
		PlayerId: &messages.Player_Id{
			Id: id,
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	knownPlayers[out.GetPlayer().GetId().GetId()] = out.GetPlayer()
	activePlayer = out.GetPlayer()
	fmt.Printf("got, updated active player: %v\n", out)
}

func startGameAction(cmdParts []string) {
	// start_game <game_type> <title>
	if len(cmdParts) < 3 {
		fmt.Println("missing game_type and title")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := gsCli.StartGame(ctx, &gs.StartGameRequest{
		GameType: cmdParts[1],
		Metadata: &messages.Game_Metadata{
			Title: cmdParts[2],
			Rules: &messages.Game_Metadata_Rules{
				VoteApplication: &messages.Game_Metadata_Rules_VoteAppliedAfterTally_{
					VoteAppliedAfterTally: &messages.Game_Metadata_Rules_VoteAppliedAfterTally{
						TimeoutSeconds:  10,
						WaitFullTimeout: false,
					},
				},
				GameSpecificRules: &messages.Game_Metadata_Rules_ChessRules{
					ChessRules: &games.ChessRules{
						BalancedTeams: true,
						BalanceEnforcement: &games.ChessRules_TolerateDifference{
							TolerateDifference: 1,
						},
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	activeGame = out.GetGame()
	fmt.Printf("ok: %v\n", out)
}

func getGameAction(cmdParts []string) {
	// get_game <game_id>
	var id string
	if len(cmdParts) < 2 {
		if activeGame == nil {
			fmt.Println("missing game_id, either set an game_id or define one.")
			return
		}
		id = activeGame.GetId().GetId()
	} else {
		id = cmdParts[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := gsCli.GetGame(ctx, &gs.GetGameRequest{
		GameId: &messages.Game_Id{
			Id: id,
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out)
}

func addPlayerAction(cmdParts []string) {
	// add_player <player_id> <game_id>
	var pid, gid string
	if len(cmdParts) < 3 {
		if activePlayer == nil {
			fmt.Println("missing player_id, either set an active player or define one.")
			return
		}
		if activeGame == nil {
			fmt.Println("missing game_id, either set an active player or define one.")
			return
		}
		pid = activePlayer.GetId().GetId()
		gid = activeGame.GetId().GetId()
	} else {
		pid = cmdParts[1]
		gid = cmdParts[2]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := gsCli.AddPlayer(ctx, &gs.AddPlayerRequest{
		GameId: &messages.Game_Id{
			Id: gid,
		},
		PlayerId: &messages.Player_Id{
			Id: pid,
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out)
}

func removePlayerAction(cmdParts []string) {
	// remove_player <player_id> <game_id>
	var pid, gid string
	if len(cmdParts) < 3 {
		if activePlayer == nil {
			fmt.Println("missing player_id, either set an active player or define one.")
			return
		}
		if activeGame == nil {
			fmt.Println("missing game_id, either set an active player or define one.")
			return
		}
		pid = activePlayer.GetId().GetId()
		gid = activeGame.GetId().GetId()
	} else {
		pid = cmdParts[1]
		gid = cmdParts[2]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := gsCli.RemovePlayer(ctx, &gs.RemovePlayerRequest{
		GameId: &messages.Game_Id{
			Id: gid,
		},
		PlayerId: &messages.Player_Id{
			Id: pid,
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out)
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

func listGamesAction(cmdParts []string) {
	// list_games
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := gsCli.ListGames(ctx, &gs.ListGamesRequest{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out)
}

func stopGameAction(cmdParts []string) {
	// stop_game <player_id>
	var id string
	if len(cmdParts) < 2 {
		if activeGame == nil {
			fmt.Println("missing game_id, either set an game_id or define one.")
			return
		}
		id = activeGame.GetId().GetId()
	} else {
		id = cmdParts[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := gsCli.StopGame(ctx, &gs.StopGameRequest{
		GameId: &messages.Game_Id{
			Id: id,
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Front End Client\nEnter 'help' for commands\n\n")
	run := true
	for run {
		fmt.Print("> ")
		read, _ := reader.ReadString('\n')
		text := strings.Trim(read, " \n\r")
		cmdParts := strings.Split(text, " ")
		verb := strings.ToLower(cmdParts[0])
		if verb == "quit" || verb == "exit" {
			break
		}

		cmd, ok := commands[verb]
		if !ok {
			fmt.Printf("unknown command: %s\n", verb)
			continue
		}
		cmd.action(cmdParts)
	}
	gsConn.Close()
	prConn.Close()
}
