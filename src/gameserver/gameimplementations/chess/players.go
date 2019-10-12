package chess

import (
	"context"
	"log"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// AddPlayers is called by a GameServerSlave to request 1+ player(s) be added to this game.
func (i *Implementation) AddPlayers(ctx context.Context, in *pb.AddPlayersRequest) (*pb.AddPlayersResponse, error) {
	i.teamsMux.Lock()
	defer i.teamsMux.Unlock()

	// Calculate if these new players will break balance enforcement
	deltas := [2]int64{0, 0}
	for _, newPlayer := range in.GetPlayers() {
		isWhite, onTeam := i.playerToTeam[newPlayer.GetPlayerId()]
		if onTeam {
			i := 0
			if !isWhite {
				i = 1
			}
			deltas[i]--
		}
		i := 0
		if !newPlayer.GetRequest().GetFields().GetChessFields().GetWhiteTeam() {
			i = 1
		}
		deltas[i]++
	}
	newWhite := i.teamToCount[true] + deltas[0]
	newBlack := i.teamToCount[false] + deltas[1]
	if err := validateNewTeamSizes(newWhite, newBlack, i.metadata.GetRules().GetChessRules()); err != nil {
		return nil, err
	}

	// New sizes check out, lets apply them.
	for _, newPlayer := range in.GetPlayers() {
		if isWhite, onTeam := i.playerToTeam[newPlayer.GetPlayerId()]; onTeam {
			i.teamToCount[isWhite]--
		}
		i.teamToCount[newPlayer.GetRequest().GetFields().GetChessFields().GetWhiteTeam()]++
		i.playerToTeam[newPlayer.GetPlayerId()] = newPlayer.GetRequest().GetFields().GetChessFields().GetWhiteTeam()

	}

	return &pb.AddPlayersResponse{}, nil
}

// RemovePlayers is called by a GameServerSlave to request 1+ player(s) be removed from this game.
func (i *Implementation) RemovePlayers(ctx context.Context, in *pb.RemovePlayersRequest) (*pb.RemovePlayersResponse, error) {
	i.teamsMux.Lock()
	defer i.teamsMux.Unlock()

	for _, playerID := range in.GetPlayerIds() {
		if t, ok := i.playerToTeam[playerID]; ok {
			i.teamToCount[t]--
			delete(i.playerToTeam, playerID)
		} else {
			// TODO: metrics and logging
			log.Printf("Removing already removed player %s\n", playerID)
		}
	}
	return &pb.RemovePlayersResponse{}, nil
}

func validateNewTeamSizes(white, black int64, rules *games.ChessRules) error {
	if rules.GetTolerateDifference() >= 1 {
		return validateNewTeamSizesByDiff(white, black, rules.GetTolerateDifference())
	}
	if rules.GetToleratePercent() > 0 {
		return validateNewTeamSizesByPercent(white, black, rules.GetToleratePercent())
	}
	return status.Errorf(codes.Internal, "incorrectly setup chess balance enforcement")

}

func validateNewTeamSizesByDiff(white, black, diffTolerance int64) error {
	diff := white - black
	// abs
	if diff < 0 {
		diff = diff * -1
	}

	if diff > diffTolerance {
		return status.Errorf(codes.FailedPrecondition, "new team sizes (white: %d; black: %d; diff: %d) would exceed maximum team size difference of %d", white, black, diff, diffTolerance)
	}
	return nil
}

func validateNewTeamSizesByPercent(white, black int64, percentTolerance float32) error {
	bigger := white
	smaller := black
	if smaller > bigger {
		smaller = white
		bigger = black
	}
	diff := (float32(bigger) / float32(smaller)) - 1.0
	if diff > percentTolerance {
		return status.Errorf(codes.FailedPrecondition, "new team sizes (white: %d; black: %d; diff: %.2f%%) would exceed maximum team size difference percange of %.2f%%", white, black, diff, percentTolerance)
	}
	return nil
}
