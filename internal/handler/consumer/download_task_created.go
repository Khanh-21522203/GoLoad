package consumers

import (
	"context"
	"log"

	"GoLoad/internal/dataaccess/mq/producer"
)

type DownloadTaskCreated interface {
	Handle(ctx context.Context, event producer.DownloadTaskCreated) error
}
type downloadTaskCreated struct {
}

func NewDownloadTaskCreated() DownloadTaskCreated {
	return &downloadTaskCreated{}
}
func (d downloadTaskCreated) Handle(ctx context.Context, event producer.DownloadTaskCreated) error {
	log.Printf("download task created event received")
	return nil
}
