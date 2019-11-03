package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// ServerNameFromContext returns the TLS ServerName from the context's TLSAuth. Returns an error if a problem occurs.
func ServerNameFromContext(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("no peer found in context")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "", fmt.Errorf("unexpected peer transport credentials, could not cast to TLSInfo")
	}

	return tlsAuth.State.ServerName, nil
}
