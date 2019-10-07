// Package main is the entry point for both kinds of GameServers: GameServerMasters and GameServerSlaves.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/sambdavidson/community-chess/src/gameserver/gamemaster"
	"github.com/sambdavidson/community-chess/src/gameserver/gameslave"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
)

// Flags
var (
	port = flag.Int("port", 0, "port the GameServer accepts connections")

	playerRegistrarAddress = flag.String("player_registar_address", "localhost:50052", "address of the Player Registrar")

	slave         = flag.Bool("slave", false, "whether or not this server is a GameServerSlave")
	masterAddress = flag.String("master_address", "", "addres of GameServerMaster; must be set if --slave is also set")
)

// State
var (
	playerRegistrarCli  pr.PlayersRegistrarClient
	playerRegistrarConn *grpc.ClientConn

	slaveController  *gameslave.Controller
	masterController *gamemaster.Controller
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	playerRegistrarConn, err = grpc.Dial(*playerRegistrarAddress, grpc.WithInsecure() /* Figure out Auth story */)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	playerRegistrarCli = pr.NewPlayersRegistrarClient(playerRegistrarConn)

	if *slave {
		slaveController, err = gameslave.NewGameSlaveController(gameslave.Opts{
			PlayerRegistrarCli: playerRegistrarCli,
			ReturnAddress:      lis.Addr().String(),
			MasterAddress:      *masterAddress,
		})
		if err != nil {
			playerRegistrarConn.Close()
			log.Fatalf("failed to build GameSlaveController: %v", err)
		}
		gs.RegisterGameServerServer(grpcServer, slaveController.GameServerInstance())
		gs.RegisterGameServerSlaveServer(grpcServer, slaveController.GameServerSlaveInstance())
	} else {
		masterController, err = gamemaster.NewGameMasterController(gamemaster.Opts{
			PlayerRegistrarCli: playerRegistrarCli,
		})
		if err != nil {
			playerRegistrarConn.Close()
			log.Fatalf("failed to build GameMasterController: %v", err)
		}
		gs.RegisterGameServerServer(grpcServer, masterController.GameServerInstance())
		gs.RegisterGameServerMasterServer(grpcServer, masterController.GameServerMasterInstance())
	}

	go handleSIGINT()

	// START LISTENING
	log.Printf("Listening at %s\n", lis.Addr())
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
	if playerRegistrarConn != nil {
		playerRegistrarConn.Close()
	}
	if slaveController != nil {
		slaveController.Close()
	}
	if masterController != nil {
		masterController.Close()
	}

}
