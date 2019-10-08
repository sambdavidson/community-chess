package chess

import (
	"context"
	"log"
	"math"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Join joins this game.
func (i *Implementation) Join(ctx context.Context, in *pb.JoinRequest) (*pb.JoinResponse, error) {
	log.Println("Join", in)
	return &pb.JoinResponse{}, nil
}

// Leave leaves this game.
func (i *Implementation) Leave(ctx context.Context, in *pb.LeaveRequest) (*pb.LeaveResponse, error) {
	log.Println("Leave", in)
	return &pb.LeaveResponse{}, nil
}

// AddPlayers is called by a GameServerSlave to request 1+ player(s) be added to this game.
func (i *Implementation) AddPlayers(ctx context.Context, in *pb.AddPlayersRequest) (*pb.AddPlayersResponse, error) {
	log.Println("AddPlayers", in)
	return &pb.AddPlayersResponse{}, nil
}

// RemovePlayers is called by a GameServerSlave to request 1+ player(s) be removed from this game.
func (i *Implementation) RemovePlayers(ctx context.Context, in *pb.RemovePlayersRequest) (*pb.RemovePlayersResponse, error) {
	log.Println("RemovePlayers", in)
	return &pb.RemovePlayersResponse{}, nil
}

func canJoinTeam(team bool, whiteCount, blackCount, tolerateDifference int64, toleratePercent float64) bool {
	var whiteMod, blackMod int64
	if team {
		if whiteCount == 0 {
			return true
		}
		whiteMod = 1
	} else {
		if blackCount == 0 {
			return true
		}
		blackMod = 1
	}
	newWhite := whiteCount + whiteMod
	newBlack := blackCount + blackMod
	diff := newWhite - newBlack
	if tolerateDifference >= 1 {
		if diff < 0 {
			diff = diff * -1
		}
		if diff > tolerateDifference {
			return false
		}
	} else {
		bigger := newWhite
		smaller := newBlack
		if smaller > bigger {
			smaller = newWhite
			bigger = newBlack
		}
		if math.Ceil(float64(smaller)*(1.0+toleratePercent)) < float64(bigger) {
			return false
		}
	}
	return true
}

func teamStr(t bool) string {
	if t {
		return "white"
	}
	return "false"
}
