package gameslave

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/sambdavidson/community-chess/src/gameserver/game"
	"github.com/sambdavidson/community-chess/src/proto/messages"

	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"

	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Opts contains intialization options and variables for a new GameServerSlave.
type Opts struct {
	InstanceID          string
	GameID              string
	SlaveAddress        string
	MasterAddress       string
	ServerTLSConfig     *tls.Config
	SlaveTLSConfig      *tls.Config
	PlayersRegistrarCli pr.PlayersRegistrarClient
}

// Controller owns both the GameServer and GameServerSlave and manages their game data.
type Controller struct {
	server      *GameServer
	serverSlave *GameServerSlave

	masterCli  gs.GameServerMasterClient
	masterConn *grpc.ClientConn
}

var (
	instanceID         string
	gameID             string
	gameType           messages.Game_Type
	gameImplementation = game.Noop
	initializeTime     time.Time
	partialGameProto   *messages.Game
	controller         *Controller
	slaveTLSConfig     *tls.Config
)

// NewGameSlaveController builts a new slave and registers itself to the master.
func NewGameSlaveController(opts Opts) (*Controller, error) {
	var err error
	if controller != nil {
		return nil, fmt.Errorf("GameSlave Controller already initialized")
	}
	instanceID = opts.InstanceID
	gameID = opts.GameID
	slaveTLSConfig = opts.SlaveTLSConfig

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
			playersRegistrarCli: opts.PlayersRegistrarCli,
		},
		serverSlave: &GameServerSlave{
			masterID:            res.GetMasterId(),
			masterCli:           masterCli,
			playersRegistrarCli: opts.PlayersRegistrarCli,
		},
		masterConn: masterConn,
	}

	var ok bool
	if gameImplementation, ok = game.ImplementationMap[res.GetGame().GetType()]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown game type: %v", res.GetGame().GetType())
	}
	if _, err = gameImplementation.Initialize(context.Background(), &gs.InitializeRequest{
		Game: res.GetGame(),
	}); err != nil {
		return nil, fmt.Errorf("unable to initialize game implementation: %v", err)
	}
	gameType = res.GetGame().GetType()
	initializeTime = time.Unix(0, res.GetGame().GetStartTime())
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
	if c.masterConn != nil {
		c.masterConn.Close()
	}
}
