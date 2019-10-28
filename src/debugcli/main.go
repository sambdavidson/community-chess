/* BUILD and RUN
go run .\src\debugcli
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

	"github.com/sambdavidson/community-chess/src/debugcli/certificate"
	"github.com/sambdavidson/community-chess/src/lib/auth"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	"github.com/sambdavidson/community-chess/src/proto/messages/games"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

/* FLAGS */
var (
	gameServerURI  = flag.String("game_server_uri", "localhost", "uri of the game server")
	gameServerPort = flag.Int("game_server_port", 8081, "port of the game server")

	gameMasterURI  = flag.String("game_master_uri", "localhost", "uri of the game master")
	gameMasterPort = flag.Int("game_master_port", 8080, "port of the game master")

	playerRegistrarURI  = flag.String("player_registar_uri", "localhost", "URI of the Player Registrar")
	playerRegistrarPort = flag.Int("player_registrar_port", 50052, "Port of the Player Registrar")
)

/* Clients which are defined in init() */
var (
	gsConn       *grpc.ClientConn
	gsCli        gs.GameServerClient
	gmConn       *grpc.ClientConn
	gmCli        gs.GameServerMasterClient
	prConn       *grpc.ClientConn
	prCli        pr.PlayersRegistrarClient
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
		"initialize": command{
			helpText: "initializes the master for a game\n\tE.g. `initialize <game-id>`",
			action:   initializeAction,
		},
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
		"get_game": command{
			helpText: "gets a game\n\tE.g. 'get_game 123456'",
			action:   getGameAction,
		},
		"join": command{
			helpText: "join the game, either the active player or a defined one\n\tE.g. 'join' or 'join <player_id>'",
			action:   joinAction,
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
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", *gameServerURI, *gameServerPort),
		grpc.WithTransportCredentials(credentials.NewTLS(certificate.ClientTLSConfig())),
	)
	if err != nil {
		log.Fatal(err)
	}
	return conn, gs.NewGameServerClient(conn)
}

func getGameMaster() (*grpc.ClientConn, gs.GameServerMasterClient) {
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", *gameMasterURI, *gameMasterPort),
		grpc.WithTransportCredentials(credentials.NewTLS(certificate.AdminTLSConfig())),
	)
	if err != nil {
		log.Fatal(err)
	}
	return conn, gs.NewGameServerMasterClient(conn)
}

func getPlayerRegistrar() (*grpc.ClientConn, pr.PlayersRegistrarClient) {
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", *playerRegistrarURI, *playerRegistrarPort),
		grpc.WithTransportCredentials(credentials.NewTLS(certificate.ClientTLSConfig())),
	)
	if err != nil {
		log.Fatal(err)
	}
	return conn, pr.NewPlayersRegistrarClient(conn)
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
	log.Println("ok", out, err)
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

func removePlayerAction(cmdParts []string) {
	// remove_player <player_id> <game_id>
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

func stopGameAction(cmdParts []string) {
	// stop_game <player_id>
	// var id string
	// if len(cmdParts) < 2 {
	// 	if activeGame == nil {
	// 		fmt.Println("missing game_id, either set an game_id or define one.")
	// 		return
	// 	}
	// 	id = activeGame.GetId()
	// } else {
	// 	id = cmdParts[1]
	// }

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// out, err := gsCli.StopGame(ctx, &gs.StopGameRequest{
	// 	GameId: id,
	// })
	// if err != nil {
	// 	fmt.Printf("error: %v\n", err)
	// 	return
	// }
	// fmt.Printf("ok: %v\n", out)
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
