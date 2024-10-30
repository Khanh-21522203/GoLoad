package database

import (
	"GoLoad/internal/generated/grpc/go_load"
	"context"

	"github.com/doug-martin/goqu/v9"
)

type DownloadTaskDataAccessor interface {
	CreateDownloadTask(ctx context.Context, task DownloadTask) (uint64, error)
	GetDownloadTaskListOfUser(ctx context.Context, userID, offset, limit uint64) ([]DownloadTask, error)
	GetDownloadTaskCountOfUser(ctx context.Context, userID uint64) (uint64, error)
	UpdateDownloadTask(ctx context.Context, task DownloadTask) error
	DeleteDownloadTask(ctx context.Context, id uint64) error
	WithDatabase(database Database) DownloadTaskDataAccessor
}

type DownloadTask struct {
	ID             uint64                 `db:"id"`
	OfAccountID    uint64                 `db:"of_account_id"`
	DownloadType   go_load.DownloadType   `db:"download_type"`
	URL            string                 `db:"url"`
	DownloadStatus go_load.DownloadStatus `db:"download_status"`
	Metadata       string                 `db:"metadata"`
}

type downloadTaskDataAccessor struct {
	database Database
}

func NewDownloadTaskDataAccessor(database *goqu.Database) DownloadTaskDataAccessor {
	return &downloadTaskDataAccessor{
		database: database,
	}
}
func (d downloadTaskDataAccessor) CreateDownloadTask(ctx context.Context, task DownloadTask) (uint64, error) {
	return 1, nil
}
func (d downloadTaskDataAccessor) DeleteDownloadTask(ctx context.Context, id uint64) error {
	panic("unimplemented")
}
func (d downloadTaskDataAccessor) GetDownloadTaskCountOfUser(ctx context.Context, userID uint64) (uint64, error) {
	panic("unimplemented")
}
func (d downloadTaskDataAccessor) GetDownloadTaskListOfUser(ctx context.Context, userID uint64, offset uint64, limit uint64) ([]DownloadTask, error) {
	panic("unimplemented")
}
func (d downloadTaskDataAccessor) UpdateDownloadTask(ctx context.Context, task DownloadTask) error {
	panic("unimplemented")
}
func (d downloadTaskDataAccessor) WithDatabase(database Database) DownloadTaskDataAccessor {
	panic("unimplemented")
}
