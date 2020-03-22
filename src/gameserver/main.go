// Package main is the entry point for both kinds of GameServers: GameServerMasters and GameServerSlaves.
package main

/*
MASTER
go run .\src\gameserver --slave=false --game_port=8080 --master_port=8090 --game_id=88888888-4444-2222-1111-000000000000 --debug

SLAVE
go run .\src\gameserver --slave --game_port=8070 --slave_port=8071 --game_id=88888888-4444-2222-1111-000000000000 --debug

*/

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"text/tabwriter"
	"time"

	"google.golang.org/grpc/connectivity"

	"github.com/sambdavidson/community-chess/src/lib/auth/grpcplayertokens"

	"github.com/google/uuid"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/sambdavidson/community-chess/src/gameserver/gamemaster"
	"github.com/sambdavidson/community-chess/src/gameserver/gameslave"
	"github.com/sambdavidson/community-chess/src/lib/debug"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Flags
var (
	gamePort   = flag.Int("game_port", freePort(), "port the GameServer service accepts connections")
	masterPort = flag.Int("master_port", freePort(), "port the GameServerMaster service accepts connections, if enabled")
	slavePort  = flag.Int("slave_port", freePort(), "port the GameServerSlave service accepts connections, if enabled")

	slave                  = flag.Bool("slave", false, "whether or not this server is a GameServerSlave")
	masterAddress          = flag.String("master_address", "", "addres of GameServerMaster; must be set if --slave is also set")
	playerRegistrarAddress = flag.String("player_registar_address", "playerregistrar:443", "address of the Player Registrar")
	gameID                 = flag.String("game_id", "", "game_id to use, TODO for now is a UUID random generated at startup")
	instanceID             = flag.String("instance_id", uuid.New().String(), "instance_id which uniquely identifies this running gameserver instance")
)

var (
	// Cleaned up in init()
	temporaryListeners []net.Listener
)

// State
var (
	slaveController  *gameslave.Controller
	masterController *gamemaster.Controller

	gameServer   *grpc.Server
	masterServer *grpc.Server
	slaveServer  *grpc.Server

	serverWG sync.WaitGroup
)

// Dependent Services
var (
	playersRegistrarClient pr.PlayersRegistrarClient
	playersRegistrarConn   *grpc.ClientConn
)

func main() {
	flag.Parse()
	printConfig()
	defer closeConnections()

	go handleSIGINT()

	if *slave { // Slave
		slaveTLS, err := gameSlaveTLSConfig()
		if err != nil {
			log.Fatalf("failed to build slave TLS config: %v", err)
		}
		playersRegistrarClient, playersRegistrarConn, err = setupPlayerRegistrar(*playerRegistrarAddress, slaveTLS)
		if err != nil {
			log.Fatalf("failed to connect to playerristrar service as slave: %v", err)
		}
		slaveController, err = gameslave.NewGameSlaveController(gameslave.Opts{
			GameID:              *gameID,
			SlaveAddress:        slaveAddress(),
			MasterAddress:       *masterAddress,
			SlaveTLSConfig:      slaveTLS,
			PlayersRegistrarCli: playersRegistrarClient,
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
					grpcplayertokens.NewPlayerAuthIngress(grpcplayertokens.PlayerAuthIngressArgs{
						PlayersRegistrarClient: playersRegistrarClient,
					}).GetUnaryServerInterceptor(grpcplayertokens.Reject),
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
		masterTLS, err := gameMasterTLSConfig()
		if err != nil {
			log.Fatalf("failed to build master TLS config: %v", err)
		}
		playersRegistrarClient, playersRegistrarConn, err = setupPlayerRegistrar(*playerRegistrarAddress, masterTLS)
		if err != nil {
			log.Fatalf("failed to connect to playerristrar service as master: %v", err)
		}
		masterController, err = gamemaster.NewGameMasterController(gamemaster.Opts{
			GameID:              *gameID,
			MasterTLSConfig:     masterTLS,
			PlayersRegistrarCli: playersRegistrarClient,
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
					grpcplayertokens.NewPlayerAuthIngress(grpcplayertokens.PlayerAuthIngressArgs{
						PlayersRegistrarClient: playersRegistrarClient,
					}).GetUnaryServerInterceptor(grpcplayertokens.Reject),
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
}

func printConfig() {
	slaveOrMaster := "Master"
	slaveOrMasterPort := *masterPort
	if *slave {
		slaveOrMaster = "Slave"
		slaveOrMasterPort = *slavePort
	}

	log.Println("Starting Game Server with configuration")
	w := &tabwriter.Writer{}
	w.Init(log.Writer(), 0, 8, 1, '\t', 0)
	fmt.Fprintf(w, "Slave/Master\tMain Port\t%s Port\tInstance UUID\tGame UUID\tMaster Address\tPR Address\n", slaveOrMaster)
	fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%s\t%s\n", slaveOrMaster, *gamePort, slaveOrMasterPort, *instanceID, *gameID, *masterAddress, *playerRegistrarAddress)
	fmt.Fprintln(w)
	w.Flush()
}

func asyncServe(name string, server *grpc.Server, port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Printf("%s failed to listen: %v\n", name, err)
		closeConnections()
		return
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
	if playersRegistrarConn != nil {
		playersRegistrarConn.Close()
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

// Returns the host address + slave port and tries to use the environement
// variable HOSTNAME, if not defaults to localhost.
func slaveAddress() string {
	hostname, ok := os.LookupEnv("HOSTNAME")
	if !ok {
		hostname = "localhost"
	}
	return fmt.Sprintf("%s:%d", hostname, *slavePort)
}

func setupPlayerRegistrar(addr string, tlsConf *tls.Config) (pr.PlayersRegistrarClient, *grpc.ClientConn, error) {
	log.Printf("Connecting to playerregistrar at address: %s...", addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
	if err != nil {
		return nil, nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	ok := conn.WaitForStateChange(ctx, connectivity.Connecting)
	if !ok || conn.GetState() == connectivity.TransientFailure || conn.GetState() == connectivity.Shutdown {
		return nil, nil, fmt.Errorf("failed to connect to playerregistar, conn state: %v", conn.GetState())
	}
	fmt.Print("done\n")
	cli := pr.NewPlayersRegistrarClient(conn)
	return cli, conn, nil
}
