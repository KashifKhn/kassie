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

func DialOptions(unary grpc.UnaryClientInterceptor, stream grpc.StreamClientInterceptor) []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(DefaultCallOptions()...),
	}
	if unary != nil {
		opts = append(opts, grpc.WithUnaryInterceptor(unary))
	}
	if stream != nil {
		opts = append(opts, grpc.WithStreamInterceptor(stream))
	}
	return opts
}
