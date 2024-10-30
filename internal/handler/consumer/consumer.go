package consumers

import (
	"GoLoad/internal/dataaccess/mq/consumer"
	"GoLoad/internal/dataaccess/mq/producer"
	"context"
	"encoding/json"
	"fmt"
	"log"
)

type Root interface {
	Start(ctx context.Context) error
}
type root struct {
	downloadTaskCreatedHandler DownloadTaskCreated
	mqConsumer                 consumer.Consumer
}

func NewRoot(downloadTaskCreatedHandler DownloadTaskCreated, mqConsumer consumer.Consumer) Root {
	return &root{
		downloadTaskCreatedHandler: downloadTaskCreatedHandler,
		mqConsumer:                 mqConsumer,
	}
}
func (r root) Start(ctx context.Context) error {
	if err := r.mqConsumer.RegisterHandler(
		producer.MessageQueueDownloadTaskCreated,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.DownloadTaskCreated
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}
			return r.downloadTaskCreatedHandler.Handle(ctx, event)
		}); err != nil {
		log.Printf("failed to register download task created handler")
		return fmt.Errorf("failed to register download task created handler: %w", err)
	}
	return r.mqConsumer.Start(ctx)
}
