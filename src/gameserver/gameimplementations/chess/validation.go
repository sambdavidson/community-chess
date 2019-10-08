package chess

import (
	"github.com/sambdavidson/community-chess/src/proto/messages"
	"github.com/sambdavidson/community-chess/src/proto/messages/games"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validateChessRules(r *games.ChessRules) error {
	if r == nil {
		return status.Errorf(codes.InvalidArgument, "missing chess specific rules")
	}
	if r.GetTolerateDifference() != 0 {
		if r.GetTolerateDifference() < 1 {
			return status.Errorf(codes.InvalidArgument, "balance enforcement tolerate difference cannot be less than 1")
		}
		return nil
	} else if r.GetToleratePercent() != 0.0 {
		if r.GetToleratePercent() <= 0.0 {
			return status.Errorf(codes.InvalidArgument, "balance enforcement tolerate percent cannot be less than or equal to 0")
		}
		return nil
	}
	return status.Errorf(codes.InvalidArgument, "balance enforcement undefined")
}
