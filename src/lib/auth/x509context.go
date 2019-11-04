package auth

import (
	"context"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// X509CertificateFromContext returns the verified x509 leaf certificate for a request.
func X509CertificateFromContext(ctx context.Context) (*x509.Certificate, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no peer within context")
	}

	tlsAuth, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, fmt.Errorf("peer AuthInfo is not a TLSInfo")
	}
	return tlsAuth.State.VerifiedChains[0][0], nil
}
