/* BUILD and RUN
go run .\src\debugcli
*/

// Package main implements a simple front_end client for interacting with the various services through a CLI.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/sambdavidson/community-chess/src/debugcli/certificate"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

/* FLAGS */
var (
	gameServerURI  = flag.String("game_server_uri", "localhost", "uri of the game server")
	gameServerPort = flag.Int("game_server_port", 8080, "port of the game server")

	gameMasterURI  = flag.String("game_master_uri", "localhost", "uri of the game master")
	gameMasterPort = flag.Int("game_master_port", 8090, "port of the game master")

	playerRegistrarURI  = flag.String("player_registar_uri", "localhost", "URI of the Player Registrar")
	playerRegistrarPort = flag.Int("player_registrar_port", 9000, "Port of the Player Registrar")
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
	gmConn, gmCli = getGameMaster()
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
		"leave": command{
			helpText: "player leaves the game, either the active player or a defined one\n\tE.g. 'leave' or 'leave 123456 987654'",
			action:   leaveAction,
		},
		"post_vote": command{
			helpText: "posts a vote to a game",
			action:   postVotesAction,
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
		grpc.WithTransportCredentials(credentials.NewTLS(certificate.InternalTLSConfig())),
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
	gmConn.Close()
	gsConn.Close()
	prConn.Close()
}
