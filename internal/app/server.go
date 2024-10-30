package app

import (
	"GoLoad/internal/handler/grpc"
	"GoLoad/internal/handler/http"
	"context"
	"log"
)

type Server struct {
	grpcServer grpc.Server
	httpServer http.Server
}

func NewServer(grpcServer grpc.Server, httpServer http.Server) *Server {
	return &Server{
		grpcServer: grpcServer,
		httpServer: httpServer,
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
	return nil
}
