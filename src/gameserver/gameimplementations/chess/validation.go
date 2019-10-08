package chess

import (
	"github.com/sambdavidson/community-chess/src/proto/messages"
	"github.com/sambdavidson/community-chess/src/proto/messages/games"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Returns a grpc error if something is encountered.
func validateGenericMetadata(m *messages.Game_Metadata) error {
	if m.GetTitle() == "" {
		return status.Errorf(codes.InvalidArgument, "missing title")
	}
	r := m.GetRules()
	if r.GetVoteAppliedAfterTally() != nil {
		if r.GetVoteAppliedAfterTally().GetTimeoutSeconds() < 3 {
			return status.Errorf(codes.InvalidArgument, "timeout too short, must be 3 or more seconds")
		}
	} else if r.GetVoteAppliedImmediately() == nil {
		return status.Errorf(codes.InvalidArgument, "missing vote application oneof")
	}

	return nil
}

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
