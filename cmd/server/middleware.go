package main

import (
	"fmt"
	"runtime/debug"

	zlog "github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	grpcmdlw "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
)

func StdUnaryMiddleware(interceptors ...grpc.UnaryServerInterceptor) grpc.ServerOption {
	arr := []grpc.UnaryServerInterceptor{
		grpcctxtags.UnaryServerInterceptor(),
		grpcprom.UnaryServerInterceptor,
		grpcrecovery.UnaryServerInterceptor(grpcrecovery.WithRecoveryHandler(func(i interface{}) error {
			zlog.Error().Str("stack", string(debug.Stack())).Msg("gRPC handler panicked")

			return fmt.Errorf("%#v", i)
		})),
	}
	arr = append(arr, interceptors...)

	return grpc.UnaryInterceptor(
		grpcmdlw.ChainUnaryServer(arr...),
	)
}

func StdStreamMiddleware(interceptors ...grpc.StreamServerInterceptor) grpc.ServerOption {
	arr := []grpc.StreamServerInterceptor{
		grpcctxtags.StreamServerInterceptor(),
		grpcprom.StreamServerInterceptor,
		grpcrecovery.StreamServerInterceptor(),
	}
	arr = append(arr, interceptors...)

	return grpc.StreamInterceptor(
		grpcmdlw.ChainStreamServer(arr...),
	)
}
