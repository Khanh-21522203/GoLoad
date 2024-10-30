package dataaccess

import (
	"GoLoad/internal/dataaccess/cache"
	"GoLoad/internal/dataaccess/database"
	"GoLoad/internal/dataaccess/mq"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	cache.WireSet,
	database.WireSet,
	mq.WireSet,
)
