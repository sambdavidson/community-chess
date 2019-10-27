// Package main is the entry point for both kinds of GameServers: GameServerMasters and GameServerSlaves.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/google/uuid"

	"github.com/sambdavidson/community-chess/src/gameserver/gamemaster"
	"github.com/sambdavidson/community-chess/src/gameserver/gameslave"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Flags
var (
	port = flag.Int("port", 0, "port the GameServer accepts connections")

	slave                  = flag.Bool("slave", false, "whether or not this server is a GameServerSlave")
	masterAddress          = flag.String("master_address", "", "addres of GameServerMaster; must be set if --slave is also set")
	playerRegistrarAddress = flag.String("player_registar_address", "localhost:50052", "address of the Player Registrar")
	gameID                 = flag.String("game_id", uuid.New().String(), "game_id to use, TODO for now is a UUID random generated at startup")
)

// State
var (
	slaveController  *gameslave.Controller
	masterController *gamemaster.Controller
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	log.Printf("Using address %s\n", lis.Addr())

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var grpcServer *grpc.Server
	if *slave {
		slaveTLS, err := gameSlaveTLSConfig(*gameID)
		if err != nil {
			log.Fatalf("failed to build game slave TLS config: %v", err)
		}
		slaveController, err = gameslave.NewGameSlaveController(gameslave.Opts{
			GameID:                 *gameID,
			PlayerRegistrarAddress: *playerRegistrarAddress,
			ReturnAddress:          lis.Addr().String(),
			MasterAddress:          *masterAddress,
			SlaveTLSConfig:         slaveTLS,
		})
		if err != nil {
			log.Fatalf("failed to build GameSlaveController: %v", err)
		}
		grpcServer = grpc.NewServer(grpc.Creds(credentials.NewTLS(slaveTLS)))
		gs.RegisterGameServerServer(grpcServer, slaveController.GameServerInstance())
		gs.RegisterGameServerSlaveServer(grpcServer, slaveController.GameServerSlaveInstance())
	} else {
		masterTLS, err := gameMasterTLSConfig(*gameID)
		if err != nil {
			log.Fatalf("failed to build game master TLS config: %v", err)
		}
		masterController, err = gamemaster.NewGameMasterController(gamemaster.Opts{
			GameID:                 *gameID,
			PlayerRegistrarAddress: *playerRegistrarAddress,
			MasterTLSConfig:        masterTLS,
		})
		if err != nil {
			log.Fatalf("failed to build GameMasterController: %v", err)
		}
		grpcServer = grpc.NewServer(grpc.Creds(credentials.NewTLS(masterTLS)))
		gs.RegisterGameServerServer(grpcServer, masterController.GameServerInstance())
		gs.RegisterGameServerMasterServer(grpcServer, masterController.GameServerMasterInstance())
	}

	go handleSIGINT()

	// START LISTENING
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	closeConnections()
}

func handleSIGINT() {
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Print("\nReceived an interrupt, stopping services...\n")
		closeConnections()
		close(cleanupDone)
	}()
	<-cleanupDone
	os.Exit(1)
}

func closeConnections() {
	if slaveController != nil {
		slaveController.Close()
	}
	if masterController != nil {
		masterController.Close()
	}

}
