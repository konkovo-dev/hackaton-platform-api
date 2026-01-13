package grpc

import (
	commongrpc "github.com/belikoooova/hackaton-platform-api/pkg/grpc"
	"google.golang.org/grpc"
)

func NewGRPCServer() *grpc.Server {
	return commongrpc.NewServer()
}
