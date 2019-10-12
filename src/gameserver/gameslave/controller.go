package gameslave

import (
	"context"
	"fmt"
	"log"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	chess "github.com/sambdavidson/community-chess/src/gameserver/gameimplementations/chess"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Opts contains intialization options and variables for a new GameServerSlave.
type Opts struct {
	PlayerRegistrarCli pr.PlayersRegistrarClient
	ReturnAddress      string
	MasterAddress      string
}

// Controller owns both the GameServer and GameServerSlave and manages their game data.
type Controller struct {
	server      *GameServer
	serverSlave *GameServerSlave

	masterCli  gs.GameServerMasterClient
	masterConn *grpc.ClientConn
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

	game GameImplementation
	// Missing state, history, and game-specific metadata.
	partialGameProto *messages.Game
	controller       *Controller
)

// Returns GRPC error if the slave is not yet ready to receieve RPCS.
// TODO SAM NEXT: Have this be called for every RPC in a clean way.
func ready() error {
	if partialGameProto == nil || game == nil {
		return status.Errof(codes.Unavailable, "not yet available")
	}
}

// NewGameSlaveController todo
func NewGameSlaveController(opts Opts) (*Controller, error) {
	var err error
	if controller != nil {
		return nil, fmt.Errorf("GameSlave Controller already initialized")
	}

	controller = &Controller{
		server: &GameServer{
			playersRegistrarCli: opts.PlayerRegistrarCli,
		},
		serverSlave: &GameServerSlave{
			playersRegistrarCli: opts.PlayerRegistrarCli,
		},
	}

	controller.masterConn, err = grpc.Dial(opts.MasterAddress, grpc.WithInsecure() /* Figure out Auth story */)
	if err != nil {
		return nil, fmt.Errorf("failed to dial master: %v", err)
	}
	controller.masterCli = gs.NewGameServerMasterClient(controller.masterConn)

	log.Printf("Connecting to master and adding self as slave...")
	res, err := controller.masterCli.AddSlave(context.Background(), &gs.AddSlaveRequest{
		ReturnAddress: opts.ReturnAddress,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add self as slave to master: %v", err)
	}

	log.Printf("Success!\n%v", res)

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
