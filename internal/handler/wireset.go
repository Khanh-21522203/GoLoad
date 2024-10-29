package handler

import (
	"GoLoad/internal/handler/grpc"
	"GoLoad/internal/handler/http"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	grpc.WireSet,
	http.WireSet,
)
