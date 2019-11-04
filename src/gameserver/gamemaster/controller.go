package gamemaster

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	"github.com/sambdavidson/community-chess/src/gameserver/game"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Opts contains intialization options and variables for a new GameServerMaster
type Opts struct {
	InstanceID             string
	GameID                 string
	PlayerRegistrarAddress string
	ServerTLSConfig        *tls.Config
	MasterTLSConfig        *tls.Config
}

// Controller owns both the GameServer and GameServerMaster and manages their game data.
type Controller struct {
	gameServer       *GameServer
	gameServerMaster *GameServerMaster

	slaveConns []*grpc.ClientConn

	playerRegistrarConn *grpc.ClientConn
}

var (
	instanceID         string
	gameID             string
	gameType           messages.Game_Type
	gameImplementation = game.Noop
	initializeTime     time.Time
	controller         *Controller
	masterTLSConfig    *tls.Config
)

// NewGameMasterController todo
func NewGameMasterController(opts Opts) (*Controller, error) {
	if controller != nil {
		return nil, fmt.Errorf("GameMaster Controller already initialized")
	}
	instanceID = opts.InstanceID
	gameID = opts.GameID
	masterTLSConfig = opts.MasterTLSConfig

	playerRegistrarConn, err := grpc.Dial(opts.PlayerRegistrarAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(masterTLSConfig)),
	)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	playerRegistrarCli := pr.NewPlayersRegistrarClient(playerRegistrarConn)
	controller = &Controller{
		gameServer: &GameServer{
			playersRegistrarCli: playerRegistrarCli,
		},
		gameServerMaster: &GameServerMaster{
			playersRegistrarCli: playerRegistrarCli,
			slaves:              map[string]gs.GameServerSlaveClient{},
		},
		playerRegistrarConn: playerRegistrarConn,
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
	if c.playerRegistrarConn != nil {
		c.playerRegistrarConn.Close()
	}
	for _, conn := range c.slaveConns {
		conn.Close()
	}
}
