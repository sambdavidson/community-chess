// Package validation provides validation feature for generic Game proto bits.
// Implementation specific validation can be found in the thier implementation packages.
package validation

import (
	"github.com/google/uuid"
	"github.com/sambdavidson/community-chess/src/proto/messages"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Game validates the generic game fields as well as calls sub-validators such as GameMetadata().
func Game(g *messages.Game) error {
	if _, err := uuid.Parse(g.GetId()); err != nil {
		return status.Errorf(codes.InvalidArgument, "game id is not a UUID: %q; %v", g.GetId(), err)
	}
	if g.GetStartTime() == 0 {
		return status.Errorf(codes.InvalidArgument, "start time not set")
	}
	if g.GetLocation() == "" {
		return status.Errorf(codes.InvalidArgument, "location is empty")
	}
	return nil
}

// GameMetadata validates generic game metadata and returns grpc error if anything is wrong.
func GameMetadata(m *messages.Game_Metadata) error {
	if m.GetTitle() == "" {
		return status.Errorf(codes.InvalidArgument, "missing title")
	}
	r := m.GetRules()
	if r == nil {
		return status.Errorf(codes.InvalidArgument, "missing rules")
	}

	if r.GetVoteAppliedAfterTally() != nil {
		if r.GetVoteAppliedAfterTally().GetTimeoutSeconds() < 3 {
			return status.Errorf(codes.InvalidArgument, "timeout too short, must be 3 or more seconds")
		}
	} else if r.GetVoteAppliedImmediately() == nil {
		return status.Errorf(codes.InvalidArgument, "missing vote application oneof")
	}

	if r.GetGameSpecific() == nil {
		return status.Errorf(codes.InvalidArgument, "missing game specific rules")
	}

	return nil
}
