// Package game defines the implementation interface
package game

import (
	"github.com/sambdavidson/community-chess/src/gameserver/game/chess"
	"github.com/sambdavidson/community-chess/src/gameserver/game/noop"
	"github.com/sambdavidson/community-chess/src/proto/messages"
	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Implementation joins the interfaces of a GameServerServer, GameServerMasterServer, and GameServerSlaveServer.
type Implementation interface {
	pb.GameServerServer
	pb.GameServerMasterServer
	pb.GameServerSlaveServer
}

var (
	// Noop is an instanciated no-op game implementation.
	Noop Implementation = &noop.Implementation{}

	// ImplementationMap maps game types to their implementations.
	ImplementationMap = map[messages.Game_Type]Implementation{
		messages.Game_CHESS: &chess.Implementation{},
	}
)
