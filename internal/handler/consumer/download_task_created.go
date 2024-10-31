package consumers

import (
	"context"
	"log"

	"GoLoad/internal/dataaccess/mq/producer"
	"GoLoad/internal/logic"
)

type DownloadTaskCreated interface {
	Handle(ctx context.Context, event producer.DownloadTaskCreated) error
}
type downloadTaskCreated struct {
	downloadTaskLogic logic.DownloadTask
}

func NewDownloadTaskCreated(downloadTaskLogic logic.DownloadTask) DownloadTaskCreated {
	return &downloadTaskCreated{
		downloadTaskLogic: downloadTaskLogic,
	}
}
func (d downloadTaskCreated) Handle(ctx context.Context, event producer.DownloadTaskCreated) error {
	log.Printf("download task created event received")
	if err := d.downloadTaskLogic.ExecuteDownloadTask(ctx, event.ID); err != nil {
		log.Printf("failed to handle download task created event")
		return err
	}
	return nil
}
