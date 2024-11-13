package jobs

import (
	"GoLoad/internal/logic"
	"context"
)

type ExecuteAllPendingDownloadTask interface {
	Run(context.Context) error
}
type executeAllPendingDownloadTask struct {
	downloadTaskLogic logic.DownloadTask
}

func NewExecuteAllPendingDownloadTask(downloadTaskLogic logic.DownloadTask) ExecuteAllPendingDownloadTask {
	return &executeAllPendingDownloadTask{
		downloadTaskLogic: downloadTaskLogic,
	}
}
func (e executeAllPendingDownloadTask) Run(ctx context.Context) error {
	return e.downloadTaskLogic.ExecuteAllPendingDownloadTask(ctx)
}
