package gamemaster

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"github.com/sambdavidson/community-chess/src/gameserver/gameimplementations/chess"
	"github.com/sambdavidson/community-chess/src/proto/messages"
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

// GameImplementation joins a GameServerServer and GameServerSlaveServer.
type GameImplementation interface {
	gs.GameServerServer
	gs.GameServerMasterServer
}

var (
	gameImplementations = map[messages.Game_Type]GameImplementation{
		messages.Game_CHESS: &chess.Implementation{},
	}

	instanceID      string
	gameID          string
	game            GameImplementation
	controller      *Controller
	masterTLSConfig *tls.Config
)

// NewGameMasterController todo
func NewGameMasterController(opts Opts) (*Controller, error) {
	if controller != nil {
		return nil, fmt.Errorf("GameMaster Controller already initialized")
	}
	instanceID = opts.InstanceID
	gameID = opts.GameID
	masterTLSConfig = opts.MasterTLSConfig

	playerRegistrarConn, err := grpc.Dial(opts.PlayerRegistrarAddress, grpc.WithTransportCredentials(credentials.NewTLS(masterTLSConfig)))
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
	_, err = controller.gameServerMaster.Initialize(context.Background(), &gs.InitializeRequest{
		Game: &messages.Game{
			Id:   opts.GameID,
			Type: messages.Game_CHESS,
			Metadata: &messages.Game_Metadata{
				Title: "foo",
				Rules: &messages.Game_Metadata_Rules{
					GameSpecific: &messages.Game_Metadata_Rules_ChessRules{
						ChessRules: &games.ChessRules{
							BalanceEnforcement: &games.ChessRules_TolerateDifference{
								TolerateDifference: 10,
							},
						},
					},
				},
			},
			State: &messages.Game_State{
				Game: &messages.Game_State_ChessState{
					ChessState: &games.ChessState{
						BoardFen:       "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
						RoundIndex:     1,
						RoundStartTime: time.Now().UnixNano(),
						RoundEndTime:   time.Now().Add(time.Minute * 10).UnixNano(),
						Details: &games.ChessState_Details{
							PlayerIdToTeam: map[string]bool{},
							PlayerToMove:   map[string]string{},
						},
					},
				},
			},
		},
	})
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
