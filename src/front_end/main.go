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
	"strings"
	"time"

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
)

func init() {
	flag.Parse()
	gsConn, gsCli = getGameServer()
	prConn, prCli = getPlayerRegistrar()
	commands = map[string]command{
		"create_player": command{
			helpText: "creates a new player with username, updates active player with created player.\nE.g. 'create_player sam'",
			action:   createPlayerAction,
		},
		"get_player": command{
			helpText: "gets a player, updates active player if found.\nE.g. 'get_player 123456'",
			action:   getPlayerAction,
		},
		"get_game": command{
			helpText: "gets a game.\nE.g. 'get_game 123456'",
			action:   getGameAction,
		},
		"help": command{
			helpText: "foo",
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
	for k, cmd := range commands {
		fmt.Printf("- '%s': %s\n", k, cmd.helpText)
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
	fmt.Printf("created, updated active player: %v\n", out.GetPlayer())
}

func getPlayerAction(cmdParts []string) {
	if len(cmdParts) < 2 {
		fmt.Println("missing player id")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := prCli.GetPlayer(ctx, &pr.GetPlayerRequest{
		PlayerId: &messages.Player_Id{
			Id: cmdParts[1],
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
	knownPlayers[out.GetPlayer().GetId().GetId()] = out.GetPlayer()
	activePlayer = out.GetPlayer()
	fmt.Printf("got, updated active player: %v\n", out.GetPlayer())
}

func getGameAction(cmdParts []string) {
	if len(cmdParts) != 2 {
		fmt.Println("missing game id")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	out, err := gsCli.GetGame(ctx, &gs.GetGameRequest{
		GameId: &messages.Game_Id{
			Id: cmdParts[1],
		},
	})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("ok: %v\n", out.GetGame())
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
