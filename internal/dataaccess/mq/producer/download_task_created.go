package producer

import (
	"context"
	"encoding/json"
	"log"

	"GoLoad/internal/dataaccess/database"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MessageQueueDownloadTaskCreated = "download_task_created"
)

type DownloadTaskCreated struct {
	DownloadTask database.DownloadTask
}
type DownloadTaskCreatedProducer interface {
	Produce(ctx context.Context, event DownloadTaskCreated) error
}
type downloadTaskCreatedProducer struct {
	client Client
}

func NewDownloadTaskCreatedProducer(
	client Client,
) DownloadTaskCreatedProducer {
	return &downloadTaskCreatedProducer{
		client: client,
	}
}
func (d downloadTaskCreatedProducer) Produce(ctx context.Context, event DownloadTaskCreated) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal download task created event")
		return status.Errorf(codes.Internal, "failed to marshal download task created event: %+v", err)
	}
	err = d.client.Produce(ctx, MessageQueueDownloadTaskCreated, eventBytes)
	if err != nil {
		log.Printf("failed to produce download task created event")
		return status.Errorf(codes.Internal, "failed to produce download task created event: %+v", err)
	}
	return nil
}
