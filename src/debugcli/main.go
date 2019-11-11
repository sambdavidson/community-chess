/* BUILD and RUN
go run .\src\debugcli
*/

// Package main implements a simple front_end client for interacting with the various services through a CLI.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/sambdavidson/community-chess/src/proto/messages"
)

/* State */
var (
	commands = map[string]command{}

	// IDs to objects
	knownPlayers = map[string]*messages.Player{}
	knownGames   = map[string]*messages.Game{}
	// IDs using IDConsts
	activeIDs = map[string]string{}
)

// IDConsts
const (
	ACTIVEPLAYERID = "ACTIVEPLAYERID"
	ACTIVEGAMEID   = "ACTIVEGAMEID"
)

func init() {
	flag.Parse()
}

type command struct {
	helpText string
	action   func(actionArgs)
}

type actionArgs struct {
	gameID         string
	playerID       string
	otherUserInput map[string]string
}

func defaultArgs() actionArgs {
	return actionArgs{
		gameID:         activeIDs[ACTIVEGAMEID],
		playerID:       activeIDs[ACTIVEPLAYERID],
		otherUserInput: map[string]string{},
	}
}

func addPlayer(p *messages.Player) {
	id := p.GetId()
	knownPlayers[id] = p
	activeIDs[ACTIVEPLAYERID] = id
	fmt.Printf("Updated active player ID to: %s\n", id)
}
func addGame(g *messages.Game) {
	id := g.GetId()
	knownGames[id] = g
	activeIDs[ACTIVEGAMEID] = id
	fmt.Printf("Updated active game ID to: %s\n", id)
}

func listKnownPlayersAction(cmdParts []string) {
	for k, p := range knownPlayers {
		fmt.Printf("%s : %s %d\n", k, p.GetUsername(), p.GetNumberSuffix())
	}
}

func main() {
	defer clientConnectionCleanup()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Front End Client\nEnter 'help' for commands, or 'quit' to exit.\n\n")
	run := true
	for run {
		fmt.Print("> ")
		read, _ := reader.ReadString('\n')
		text := strings.Trim(read, " \n\r")
		verb, args := parseInput(text)
		if verb == "" {
			continue
		}
		if verb == "quit" || verb == "exit" {
			break
		}

		cmd, ok := commands[verb]
		if !ok {
			fmt.Printf("unknown command: %s\n", verb)
			continue
		}
		cmd.action(args)
	}
}

func parseInput(in string) (string, actionArgs) {
	args := defaultArgs()
	parts := strings.Split(in, " ")
	if len(parts) == 0 {
		return "", args
	}
	verb := parts[0]
	kvMap := map[string]string{}
	for _, v := range parts[1:] {
		kv := strings.Split(v, "=")
		if len(kv) != 2 {
			fmt.Printf("bad input, should be key value pair: %s\n", v)
			return "", args
		}
		if _, ok := kvMap[kv[0]]; ok {
			fmt.Printf("key '%s' defined twice\n", kv[0])
		}
		kvMap[kv[0]] = kv[1]
	}

	for k, v := range kvMap {
		switch k {
		case "game":
			args.gameID = v
		case "player":
			args.playerID = v
		default:
			args.otherUserInput[k] = v
		}
	}

	return verb, args
}
