package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"GoLoad/internal/configs"
	"GoLoad/internal/generated/grpc/go_load"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server interface {
	Start(ctx context.Context) error
}
type server struct {
	grpcConfig configs.GRPC
	httpConfig configs.HTTP
}

func NewServer(grpcConfig configs.GRPC, httpConfig configs.HTTP) Server {
	return &server{
		grpcConfig: grpcConfig,
		httpConfig: httpConfig,
	}
}
func (s *server) Start(ctx context.Context) error {
	mux := runtime.NewServeMux()
	if err := go_load.RegisterGoLoadServiceHandlerFromEndpoint(
		ctx,
		mux,
		s.grpcConfig.Address,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}); err != nil {
		return err
	}
	httpServer := http.Server{
		Addr:              s.httpConfig.Address,
		ReadHeaderTimeout: time.Minute,
		Handler:           mux,
	}
	log.Printf("starting http server")
	return httpServer.ListenAndServe()
}
