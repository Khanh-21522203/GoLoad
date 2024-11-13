package jobs

import (
	"GoLoad/internal/logic"
	"context"
)

type UpdateDownloadingAndFailedDownloadTaskStatusToPending interface {
	Run(context.Context) error
}
type updateDownloadingAndFailedDownloadTaskStatusToPending struct {
	downloadTaskLogic logic.DownloadTask
}

func NewUpdateDownloadingAndFailedDownloadTaskStatusToPending(downloadTaskLogic logic.DownloadTask) UpdateDownloadingAndFailedDownloadTaskStatusToPending {
	return &updateDownloadingAndFailedDownloadTaskStatusToPending{
		downloadTaskLogic: downloadTaskLogic,
	}
}
func (u updateDownloadingAndFailedDownloadTaskStatusToPending) Run(ctx context.Context) error {
	return u.downloadTaskLogic.UpdateDownloadingAndFailedDownloadTaskStatusToPending(ctx)
}
