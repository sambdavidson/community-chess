package chess

import (
	ch "github.com/notnil/chess"
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

func validateChessState(s *games.ChessState, details bool) error {
	if s == nil {
		return status.Errorf(codes.InvalidArgument, "missing chess state")
	}
	_, err := ch.FEN(s.GetBoardFen())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "bad state FEN: %q", s.GetBoardFen())
	}
	if s.GetRoundIndex() < 1 {
		return status.Errorf(codes.InvalidArgument, "round index %d cannot be less than 1", s.GetRoundIndex())
	}
	if details {
		if s.GetDetails() == nil {
			return status.Errorf(codes.InvalidArgument, "missing detailed state")
		}
	}
	return nil
}
