package auth

import (
	"context"

	"github.com/sambdavidson/community-chess/src/lib/tlsca"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

// AdminAuthorizerClientInterceptor is an authorizor and ensure the request uses Admin auth.
func AdminAuthorizerClientInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(adminAuthorizerClientInterceptorImplementation)
}

func adminAuthorizerClientInterceptorImplementation(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...CallOption) error {
	x509Cert, err := X509CertificateFromContext(ctx)
	if err != nil {
		return status.Errorf(codes.PermissionDenied, "AdminAuthorizerInterceptor %v", err)
	}

	for _, san := range x509Cert.DNSNames {
		if san == tlsca.Admin.String() {
			return nil
		}
	}

	return status.Errorf(codes.PermissionDenied, "missing admin certificate")
}
