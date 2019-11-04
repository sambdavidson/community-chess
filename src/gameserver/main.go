// Package main is the entry point for both kinds of GameServers: GameServerMasters and GameServerSlaves.
package main

/*
MASTER
go run .\src\gameserver --slave=false --game_port=8080 --master_port=8090 --game_id=88888888-4444-2222-1111-000000000000 --debug

SLAVE
go run .\src\gameserver --slave --game_port=8070 --slave_port=8071 --game_id=88888888-4444-2222-1111-000000000000 --debug

*/

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"

	"github.com/google/uuid"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sambdavidson/community-chess/src/gameserver/gamemaster"
	"github.com/sambdavidson/community-chess/src/gameserver/gameslave"
	"github.com/sambdavidson/community-chess/src/lib/debug"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Flags
var (
	gamePort   = flag.Int("game_port", freePort(), "port the GameServer service accepts connections")
	masterPort = flag.Int("master_port", freePort(), "port the GameServerMaster service accepts connections, if enabled")
	slavePort  = flag.Int("slave_port", freePort(), "port the GameServerSlave service accepts connections, if enabled")

	slave                  = flag.Bool("slave", false, "whether or not this server is a GameServerSlave")
	masterAddress          = flag.String("master_address", "localhost:8090", "addres of GameServerMaster; must be set if --slave is also set")
	playerRegistrarAddress = flag.String("player_registar_address", "localhost:8090", "address of the Player Registrar")
	gameID                 = flag.String("game_id", uuid.New().String(), "game_id to use, TODO for now is a UUID random generated at startup")
)

var (
	// Cleaned up in init()
	temporaryListeners []net.Listener
)

// State
var (
	instanceUUID     = uuid.New()
	slaveController  *gameslave.Controller
	masterController *gamemaster.Controller

	gameServer   *grpc.Server
	masterServer *grpc.Server
	slaveServer  *grpc.Server

	serverWG sync.WaitGroup
)

func main() {
	flag.Parse()
	log.Printf("Instance UUID: %s", instanceUUID.String())

	gameUUID, err := uuid.Parse(*gameID)
	if err != nil {
		log.Fatalf("gameID is not a valid UUID: %s", *gameID)
	}

	go handleSIGINT()

	if *slave { // Slave
		slaveTLS, err := gameSlaveTLSConfig(instanceUUID, gameUUID)
		if err != nil {
			log.Fatalf("failed to build TLS config: %v", err)
		}
		slaveController, err = gameslave.NewGameSlaveController(gameslave.Opts{
			GameID:                 *gameID,
			PlayerRegistrarAddress: *playerRegistrarAddress,
			SlaveAddress:           fmt.Sprintf("localhost:%d", *slavePort), // TODO figure this out
			MasterAddress:          *masterAddress,
			SlaveTLSConfig:         slaveTLS,
		})
		if err != nil {
			log.Fatalf("failed to build GameSlaveController: %v", err)
		}
		creds := grpc.Creds(credentials.NewTLS(slaveTLS))

		gameServer = grpc.NewServer(
			creds,
			grpc.UnaryInterceptor(
				middleware.ChainUnaryServer(
					debug.UnaryServerInterceptor,
				),
			),
		)
		gs.RegisterGameServerServer(gameServer, slaveController.GameServerInstance())

		slaveServer = grpc.NewServer(
			creds,
			grpc.UnaryInterceptor(
				middleware.ChainUnaryServer(
					debug.UnaryServerInterceptor,
					gameslave.ValidateMasterUnaryServerInterceptor,
				),
			),
		)
		gs.RegisterGameServerSlaveServer(slaveServer, slaveController.GameServerSlaveInstance())

		asyncServe("GameServer", gameServer, *gamePort)
		asyncServe("SlaveServer", slaveServer, *slavePort)

	} else { // Master
		masterTLS, err := gameMasterTLSConfig(instanceUUID, gameUUID)
		if err != nil {
			log.Fatalf("failed to build TLS config: %v", err)
		}
		masterController, err = gamemaster.NewGameMasterController(gamemaster.Opts{
			GameID:                 *gameID,
			PlayerRegistrarAddress: *playerRegistrarAddress,
			MasterTLSConfig:        masterTLS,
		})
		if err != nil {
			log.Fatalf("failed to build GameMasterController: %v", err)
		}
		creds := grpc.Creds(credentials.NewTLS(masterTLS))
		gameServer = grpc.NewServer(
			creds,
			grpc.UnaryInterceptor(
				middleware.ChainUnaryServer(
					debug.UnaryServerInterceptor,
				),
			),
		)
		gs.RegisterGameServerServer(gameServer, masterController.GameServerInstance())

		masterServer = grpc.NewServer(
			creds,
			grpc.UnaryInterceptor(
				middleware.ChainUnaryServer(
					debug.UnaryServerInterceptor,
					gamemaster.MasterAuthUnaryServerInterceptor,
				),
			),
		)
		gs.RegisterGameServerMasterServer(masterServer, masterController.GameServerMasterInstance())

		asyncServe("GameServer", gameServer, *gamePort)
		asyncServe("MasterServer", masterServer, *masterPort)

	}

	serverWG.Wait()

	closeConnections()
}

func asyncServe(name string, server *grpc.Server, port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Printf("%s failed to listen: %v\n", name, err)
		closeConnections()
	}

	log.Printf("%s serving on port %d\n", name, lis.Addr().(*net.TCPAddr).Port)

	serverWG.Add(1)
	go func() {
		defer serverWG.Done()
		if err := server.Serve(lis); err != nil {
			log.Printf("ERROR: %s failed to serve: %v\n", name, err)
			closeConnections()
		}
	}()
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
	if gameServer != nil {
		log.Println("Gracefully stopping GameServer...")
		gameServer.GracefulStop()
	}
	if slaveServer != nil {
		log.Println("Gracefully stopping SlaveServer...")
		slaveServer.GracefulStop()
	}
	if masterServer != nil {
		log.Println("Gracefully stopping MasterServer...")
		masterServer.GracefulStop()
	}

}

func init() {
	for _, l := range temporaryListeners {
		l.Close()
	}
}

func freePort() int {
	ln, err := net.Listen("tcp", "[::]:0")
	if err != nil {
		log.Fatalf("failed to get free port: %v", err)
	}
	temporaryListeners = append(temporaryListeners, ln)
	return ln.Addr().(*net.TCPAddr).Port
}
