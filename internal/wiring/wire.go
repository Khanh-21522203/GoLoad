//go:build wireinject
// +build wireinject

// go:generate go run github.com/google/wire/cmd/wire
package wiring

import (
	"GoLoad/internal/app"
	"GoLoad/internal/configs"
	"GoLoad/internal/dataaccess"
	"GoLoad/internal/handler"
	"GoLoad/internal/logic"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	configs.WireSet,
	dataaccess.WireSet,
	logic.WireSet,
	handler.WireSet,
	app.WireSet,
	// cache.WireSet,
)

func InitializeServer(configFilePath configs.ConfigFilePath) (*app.Server, func(), error) {
	wire.Build(WireSet)
	return nil, nil, nil
}
