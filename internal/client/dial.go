package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"

	"github.com/KashifKhn/kassie/internal/shared/config"
)

func DefaultCallOptions() []grpc.CallOption {
	return []grpc.CallOption{
		grpc.MaxCallRecvMsgSize(config.MaxMessageSize),
		grpc.MaxCallSendMsgSize(config.MaxMessageSize),
		grpc.UseCompressor(gzip.Name),
	}
}

func DialOptions(interceptor grpc.UnaryClientInterceptor) []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(DefaultCallOptions()...),
	}
	if interceptor != nil {
		opts = append(opts, grpc.WithUnaryInterceptor(interceptor))
	}
	return opts
}
