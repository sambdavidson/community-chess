package gamemaster

import (
	"context"

	"github.com/sambdavidson/community-chess/src/lib/auth"
	"github.com/sambdavidson/community-chess/src/lib/tlsca"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MasterAuthUnaryServerInterceptor is an authorizor and ensure the request uses Admin auth.
func MasterAuthUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	x509Cert, err := auth.X509CertificateFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "could not get peer certificate from context: %v", err)
	}

	for _, san := range x509Cert.DNSNames {
		if san == tlsca.Internal.String() {
			return handler(ctx, req)
		}
	}

	return nil, status.Errorf(codes.PermissionDenied, "certificate does not grant authz for master")
}
