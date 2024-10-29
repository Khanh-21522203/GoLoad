package http

import (
	"context"
	"net/http"

	"GoLoad/internal/generated/grpc/go_load"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server interface {
	Start(ctx context.Context) error
}
type server struct {
}

func NewServer() Server {
	return &server{}
}
func (s *server) Start(ctx context.Context) error {
	mux := runtime.NewServeMux()
	if err := go_load.RegisterGoLoadServiceHandlerFromEndpoint(
		ctx,
		mux,
		"0.0.0.0:8080",
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}); err != nil {
		return err
	}
	return http.ListenAndServe(":8081", mux)
}
