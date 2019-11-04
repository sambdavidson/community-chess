// Package auth provided helpers for various auth/metadata fields used throughout the app.
package auth

import (
	"context"

	"github.com/google/uuid"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/metadata"
)

const (
	playerIDKey = "x-player-id"
)

var (
	errIDAlreadyPresent = status.Errorf(codes.InvalidArgument, "player ID already present")
	errMissingMetadata  = status.Errorf(codes.InvalidArgument, "missing metadata")
	errMissingPlayer    = status.Errorf(codes.InvalidArgument, "missing player ID in metadata")
	errBadPlayerID      = status.Errorf(codes.InvalidArgument, "player ID is incorrectly formatted")
)

// AppendPlayerIDToOutgoingContext appends the player ID id to the outgoing metadata in ctx.
// Returns an updated if successful; if the player ID is invalid or already exists an error is returned.
// The returned error is a valid GRPC error.
func AppendPlayerIDToOutgoingContext(ctx context.Context, id string) (context.Context, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return nil, errBadPlayerID
	}
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		_, ok := md[playerIDKey]
		if ok {
			return nil, errIDAlreadyPresent
		}
	}
	return metadata.AppendToOutgoingContext(ctx, playerIDKey, id), nil
}

// PlayerIDFromIncomingContext returns the player ID from the incoming metadata in ctx if it exists.
// If it is missing or badly formatted an error is returns. The returned error is a valid GRPC error.
func PlayerIDFromIncomingContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingMetadata
	}
	vals, ok := md[playerIDKey]
	if !ok || len(vals) == 0 {
		return "", errMissingPlayer
	}
	_, err := uuid.Parse(vals[0])
	if len(vals) >= 2 || err != nil {
		return "", errBadPlayerID
	}
	return vals[0], nil
}
