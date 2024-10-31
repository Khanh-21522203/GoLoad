package http

import (
	"context"
	"log"
	"net/http"
	"time"

	"GoLoad/internal/configs"
	"GoLoad/internal/generated/grpc/go_load"

	handlerGRPC "GoLoad/internal/handler/grpc"
	"GoLoad/internal/handler/http/servemuxoptions"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	//nolint:gosec // This is just to specify the cookie name
	AuthTokenCookieName = "GOLOAD_AUTH"
)

type Server interface {
	Start(ctx context.Context) error
}
type server struct {
	grpcConfig configs.GRPC
	httpConfig configs.HTTP
	authConfig configs.Auth
}

func NewServer(grpcConfig configs.GRPC, httpConfig configs.HTTP, authConfig configs.Auth) Server {
	return &server{
		grpcConfig: grpcConfig,
		httpConfig: httpConfig,
		authConfig: authConfig,
	}
}
func (s server) getGRPCGatewayHandler(ctx context.Context) (http.Handler, error) {
	tokenExpiresInDuration, err := s.authConfig.Token.GetExpiresInDuration()
	if err != nil {
		return nil, err
	}
	grpcMux := runtime.NewServeMux(
		servemuxoptions.WithAuthCookieToAuthMetadata(AuthTokenCookieName, handlerGRPC.AuthTokenMetadataName),
		servemuxoptions.WithAuthMetadataToAuthCookie(
			handlerGRPC.AuthTokenMetadataName, AuthTokenCookieName, tokenExpiresInDuration),
		servemuxoptions.WithRemoveGoAuthMetadata(handlerGRPC.AuthTokenMetadataName),
	)
	err = go_load.RegisterGoLoadServiceHandlerFromEndpoint(
		ctx,
		grpcMux,
		s.grpcConfig.Address,
		[]grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		})
	if err != nil {
		return nil, err
	}
	return grpcMux, nil
}
func (s server) Start(ctx context.Context) error {
	grpcGatewayHandler, err := s.getGRPCGatewayHandler(ctx)
	if err != nil {
		return err
	}
	httpServer := http.Server{
		Addr:              s.httpConfig.Address,
		ReadHeaderTimeout: time.Minute,
		Handler:           grpcGatewayHandler,
	}
	log.Printf("starting http server")
	return httpServer.ListenAndServe()
}
