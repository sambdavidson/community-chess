// Package debug contains useful debug utilities.
package debug

import (
	"context"
	"flag"
	"log"

	"google.golang.org/grpc"
)

var (
	debug = flag.Bool("debug", false, "Enable debug mode, turning on various logging features such as interceptors.")
)

// UnaryServerInterceptor will log all grpc unary I/O if --debug is set.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if !(*debug) {
		return handler(ctx, req)
	}
	res, err := handler(ctx, req)
	log.Printf("Unary RPC: %s\n\tRequest: %v\n\tResponse: %v\n\tError: %v\n", info.FullMethod, req, res, err)
	return res, err
}
