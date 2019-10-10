package chess

import (
	"context"
	"log"

	"github.com/sambdavidson/community-chess/src/proto/messages"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// Metadata gets this game's metadata.
func (i *Implementation) Metadata(ctx context.Context, in *pb.MetadataRequest) (*pb.MetadataResponse, error) {
	log.Println("GetGameMetadata", in)
	return &pb.MetadataResponse{
		Metadata: i.metadata,
	}, nil
}

// State gets this game's state.
func (i *Implementation) State(ctx context.Context, in *pb.StateRequest) (*pb.StateResponse, error) {
	log.Println("GetGameState", in)
	var details *games.ChessState_Details
	if in.GetDetailed() {
		details = &games.ChessState_Details{
			PlayerIdToTeam: i.playerToTeam,
			PlayerToMove:   i.playerToMove,
		}
	}
	return &pb.StateResponse{
		State: &messages.Game_State{
			Game: &messages.Game_State_ChessState{
				ChessState: &games.ChessState{
					WhiteTeamCount: i.teamToCount[true],
					BlackTeamCount: i.teamToCount[false],
					BoardFen:       i.game.FEN(),
					MoveToCount:    i.moveToCount,
					RoundStartTime: i.startTime.UnixNano(),
					RoundEndTime:   i.endTime.UnixNano(),
					Details:        details,
					RoundIndex:     i.roundIndex,
				},
			},
		},
	}, nil
}

// History gets this game's history.
func (i *Implementation) History(ctx context.Context, in *pb.HistoryRequest) (*pb.HistoryResponse, error) {
	log.Println("GetGameHistory", in)
	return &pb.HistoryResponse{
		History: i.history,
	}, nil
}
