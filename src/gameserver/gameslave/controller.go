package gameslave

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	chess "github.com/sambdavidson/community-chess/src/gameserver/gameimplementations/chess"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"

	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Opts contains intialization options and variables for a new GameServerSlave.
type Opts struct {
	InstanceID             string
	GameID                 string
	PlayerRegistrarAddress string
	SlaveAddress           string
	MasterAddress          string
	ServerTLSConfig        *tls.Config
	SlaveTLSConfig         *tls.Config
}

// Controller owns both the GameServer and GameServerSlave and manages their game data.
type Controller struct {
	server      *GameServer
	serverSlave *GameServerSlave

	masterCli  gs.GameServerMasterClient
	masterConn *grpc.ClientConn

	playerRegistarCli   pr.PlayersRegistrarClient
	playerRegistrarConn *grpc.ClientConn
}

// GameImplementation joins a GameServerServer and GameServerSlaveServer.
type GameImplementation interface {
	gs.GameServerServer
	gs.GameServerSlaveServer
}

var (
	gameImplementations = map[messages.Game_Type]GameImplementation{
		messages.Game_CHESS: &chess.Implementation{},
	}

	instanceID string
	gameID     string
	game       GameImplementation
	// Missing state, history, and game-specific metadata.
	partialGameProto *messages.Game
	controller       *Controller
	slaveTLSConfig   *tls.Config
)

// Returns GRPC error if the slave is not yet ready to receieve RPCS.
// TODO SAM NEXT: Have this be called for every RPC in a clean way.
func ready() error {
	if partialGameProto == nil || game == nil {
		return status.Errorf(codes.Unavailable, "not yet available")
	}
	return nil
}

// NewGameSlaveController builts a new slave and registers itself to the master.
func NewGameSlaveController(opts Opts) (*Controller, error) {
	var err error
	if controller != nil {
		return nil, fmt.Errorf("GameSlave Controller already initialized")
	}
	instanceID = opts.InstanceID
	gameID = opts.GameID
	slaveTLSConfig = opts.SlaveTLSConfig

	playerRegistrarConn, err := grpc.Dial(opts.PlayerRegistrarAddress, grpc.WithTransportCredentials(credentials.NewTLS(slaveTLSConfig)))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	playerRegistrarCli := pr.NewPlayersRegistrarClient(playerRegistrarConn)

	masterConn, err := grpc.Dial(opts.MasterAddress, grpc.WithTransportCredentials(credentials.NewTLS(slaveTLSConfig)))
	if err != nil {
		return nil, fmt.Errorf("failed to dial master: %v", err)
	}
	masterCli := gs.NewGameServerMasterClient(masterConn)

	log.Printf("Connecting to master and adding self as slave...")
	res, err := masterCli.AddSlave(context.Background(), &gs.AddSlaveRequest{
		ReturnAddress: opts.SlaveAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add self as slave to master: %v", err)
	}

	log.Printf("Added self as slave to master: %s!\n%v", opts.MasterAddress, res)

	controller = &Controller{
		server: &GameServer{
			masterCli:           masterCli,
			playersRegistrarCli: playerRegistrarCli,
		},
		serverSlave: &GameServerSlave{
			masterID:            res.GetMasterId(),
			masterCli:           masterCli,
			playersRegistrarCli: playerRegistrarCli,
		},
		masterConn:          masterConn,
		playerRegistrarConn: playerRegistrarConn,
	}
	return controller, nil
}

// GameServerInstance todo
func (c *Controller) GameServerInstance() *GameServer {
	return c.server
}

// GameServerSlaveInstance todo
func (c *Controller) GameServerSlaveInstance() *GameServerSlave {
	return c.serverSlave
}

// Close all open connections
func (c *Controller) Close() {
	if c.playerRegistrarConn != nil {
		c.playerRegistrarConn.Close()
	}
	if c.masterConn != nil {
		c.masterConn.Close()
	}
}
