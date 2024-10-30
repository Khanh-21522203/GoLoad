package mq

import (
	"GoLoad/internal/dataaccess/mq/consumer"
	"GoLoad/internal/dataaccess/mq/producer"
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	consumer.WireSet,
	producer.WireSet,
)
