package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/sambdavidson/community-chess/src/debugcli/certificate"
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

/* Clients */
var (
	gsConn *grpc.ClientConn
	gsCli  gs.GameServerClient
	gmConn *grpc.ClientConn
	gmCli  gs.GameServerMasterClient
	prConn *grpc.ClientConn
	prCli  pr.PlayersRegistrarClient
)

func init() {
	gsConn, gsCli = getGameServer()
	gmConn, gmCli = getGameMaster()
	prConn, prCli = getPlayerRegistrar()
}

func clientConnectionCleanup() {
	gmConn.Close()
	gsConn.Close()
	prConn.Close()
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
