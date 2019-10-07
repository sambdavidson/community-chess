package gamemaster

import (
	"fmt"

	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
)

// Opts contains intialization options and variables for a new GameServerMaster
type Opts struct {
	PlayerRegistrarCli pr.PlayersRegistrarClient
}

// Controller owns both the GameServer and GameServerMaster and manages their game data.
type Controller struct {
	gameServer       *GameServer
	gameServerMaster *GameServerMaster

	slaveClis  []gs.GameServerSlaveClient
	slaveConns []*grpc.ClientConn
}

// GameImplementation joins a GameServerServer and GameServerSlaveServer.
type GameImplementation interface {
	gs.GameServerServer
	gs.GameServerMasterServer
}

var (
	controller *Controller
)

// NewGameMasterController todo
func NewGameMasterController(opts Opts) (*Controller, error) {
	if controller != nil {
		return nil, fmt.Errorf("GameMaster Controller already initialized")
	}

	controller = &Controller{
		gameServer: &GameServer{
			playersRegistrarCli: opts.PlayerRegistrarCli,
		},
		gameServerMaster: &GameServerMaster{
			playersRegistrarCli: opts.PlayerRegistrarCli,
		},
	}
	return controller, nil
}

// GameServerInstance todo
func (c *Controller) GameServerInstance() *GameServer {
	return c.gameServer
}

// GameServerMasterInstance todo
func (c *Controller) GameServerMasterInstance() *GameServerMaster {
	return c.gameServerMaster
}

// Close all open connections
func (c *Controller) Close() {
	for _, conn := range c.slaveConns {
		conn.Close()
	}
}
