package consumers

import (
	"context"

	"GoLoad/internal/dataaccess/mq/producer"
	"GoLoad/internal/logic"
	"GoLoad/internal/utils"

	"go.uber.org/zap"
)

type DownloadTaskCreated interface {
	Handle(ctx context.Context, event producer.DownloadTaskCreated) error
}
type downloadTaskCreated struct {
	downloadTaskLogic logic.DownloadTask
	logger            *zap.Logger
}

func NewDownloadTaskCreated(downloadTaskLogic logic.DownloadTask, logger *zap.Logger) DownloadTaskCreated {
	return &downloadTaskCreated{
		downloadTaskLogic: downloadTaskLogic,
		logger:            logger,
	}
}
func (d downloadTaskCreated) Handle(ctx context.Context, event producer.DownloadTaskCreated) error {
	logger := utils.LoggerWithContext(ctx, d.logger).With(zap.Any("event", event))
	logger.Info("download task created event received")

	if err := d.downloadTaskLogic.ExecuteDownloadTask(ctx, event.ID); err != nil {
		logger.With(zap.Error(err)).Error("failed to handle download task created event")
		return err
	}
	return nil
}
