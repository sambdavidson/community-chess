package auth

import (
	"context"
	"fmt"

	"github.com/sambdavidson/community-chess/src/lib/tlsca"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

// AdminAuthorizerClientInterceptor is an authorizor and ensure the request uses Admin auth.
func AdminAuthorizerClientInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	
	fmt.Println("adminInterceptor", info.FullMethod)
	if info.FullMethod != "f" {
		return handler(ctx, req)
	}
	x509Cert, err := X509CertificateFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "AdminAuthorizerInterceptor %v", err)
	}

	for _, san := range x509Cert.DNSNames {
		if san == tlsca.Admin.String() {
			return handler(ctx, req)
		}
	}

	return nil, status.Errorf(codes.PermissionDenied, "missing admin certificate")
}
