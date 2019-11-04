package gameslave

import (
	"context"

	"github.com/sambdavidson/community-chess/src/lib/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

// ValidateMasterUnaryServerInterceptor validates the certificate within ctx is the slave's master.
func ValidateMasterUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	slave, ok := info.Server.(*GameServerSlave)
	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to cast server to slave for master validation")
	}
	cert, err := auth.X509CertificateFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "couldn't get master peer certificate: %v", err)
	}
	if cert.Subject.CommonName != slave.masterID {
		return nil, status.Errorf(codes.InvalidArgument, "master certificate subject is not expected got: %s; want: %s", cert.Subject.CommonName, slave.masterID)
	}
	return handler(ctx, req)
}
