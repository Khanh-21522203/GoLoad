package producer

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MessageQueueDownloadTaskCreated = "download_task_created"
)

type DownloadTaskCreated struct {
	ID uint64 `json:"id"`
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
		return status.Error(codes.Internal, "failed to marshal download task created event")
	}
	err = d.client.Produce(ctx, MessageQueueDownloadTaskCreated, eventBytes)
	if err != nil {
		log.Printf("failed to produce download task created event")
		return status.Error(codes.Internal, "failed to marshal download task created event")
	}
	return nil
}
