package app

import (
	consumers "GoLoad/internal/handler/consumer"
	"GoLoad/internal/handler/grpc"
	"GoLoad/internal/handler/http"
	"context"
	"log"
)

type Server struct {
	grpcServer   grpc.Server
	httpServer   http.Server
	rootConsumer consumers.Root
}

func NewServer(grpcServer grpc.Server, httpServer http.Server, rootConsumer consumers.Root) *Server {
	return &Server{
		grpcServer:   grpcServer,
		httpServer:   httpServer,
		rootConsumer: rootConsumer,
	}
}
func (s Server) Start() error {
	go func() {
		s.grpcServer.Start(context.Background())
		log.Printf("grpc server stopped")
	}()
	go func() {
		s.httpServer.Start(context.Background())
		log.Printf("http server stopped")
	}()
	go func() {
		s.rootConsumer.Start(context.Background())
		log.Printf("message queue consumer stopped")
	}()
	return nil
}
