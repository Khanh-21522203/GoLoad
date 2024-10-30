package grpc

import (
	"context"
	"log"
	"net"

	"GoLoad/internal/configs"
	"GoLoad/internal/generated/grpc/go_load"

	"google.golang.org/grpc"
)

type Server interface {
	Start(ctx context.Context) error
}
type server struct {
	handler    go_load.GoLoadServiceServer
	grpcConfig configs.GRPC
}

func NewServer(handler go_load.GoLoadServiceServer, grpcConfig configs.GRPC) Server {
	return &server{
		handler: handler,
	}
}
func (s *server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.grpcConfig.Address)
	if err != nil {
		log.Printf("failed to open tcp listener")
		return err
	}
	defer listener.Close()
	server := grpc.NewServer()
	go_load.RegisterGoLoadServiceServer(server, s.handler)
	log.Printf("starting grpc server")
	return server.Serve(listener)
}
