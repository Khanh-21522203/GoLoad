package consumers

import (
	"GoLoad/internal/dataaccess/mq/consumer"
	"GoLoad/internal/dataaccess/mq/producer"
	"context"
	"encoding/json"

	"go.uber.org/zap"
)

type Root interface {
	Start(ctx context.Context) error
}
type root struct {
	downloadTaskCreatedHandler DownloadTaskCreated
	mqConsumer                 consumer.Consumer
	logger                     *zap.Logger
}

func NewRoot(downloadTaskCreatedHandler DownloadTaskCreated, mqConsumer consumer.Consumer, logger *zap.Logger) Root {
	return &root{
		downloadTaskCreatedHandler: downloadTaskCreatedHandler,
		mqConsumer:                 mqConsumer,
		logger:                     logger,
	}
}
func (r root) Start(ctx context.Context) error {
	r.mqConsumer.RegisterHandler(
		producer.MessageQueueDownloadTaskCreated,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.DownloadTaskCreated
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}
			return r.downloadTaskCreatedHandler.Handle(ctx, event)
		},
	)
	return r.mqConsumer.Start(ctx)
}
