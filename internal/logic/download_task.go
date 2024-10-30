package logic

import (
	"GoLoad/internal/dataaccess/database"
	"GoLoad/internal/dataaccess/mq/producer"
	"GoLoad/internal/generated/grpc/go_load"
	"context"

	"github.com/doug-martin/goqu/v9"
)

type CreateDownloadTaskParams struct {
	Token        string
	DownloadType go_load.DownloadType
	URL          string
}
type CreateDownloadTaskOutput struct {
	DownloadTask go_load.DownloadTask
}
type GetDownloadTaskListParams struct {
	Token  string
	Offset uint64
	Limit  uint64
}
type GetDownloadTaskListOutput struct {
	DownloadTask           go_load.DownloadTask
	TotalDownloadTaskCount uint64
}
type UpdateDownloadTaskParams struct {
	Token          string
	DownloadTaskID uint64
	URL            string
}
type UpdateDownloadTaskOutput struct {
	DownloadTask go_load.DownloadTask
}
type DeleteDownloadTaskParams struct {
	Token          string
	DownloadTaskID uint64
}
type DownloadTask interface {
	CreateDownloadTask(context.Context, CreateDownloadTaskParams) (CreateDownloadTaskOutput, error)
	GetDownloadTaskList(context.Context, GetDownloadTaskListParams) (GetDownloadTaskListOutput, error)
	UpdateDownloadTask(context.Context, UpdateDownloadTaskParams) (UpdateDownloadTaskOutput, error)
	DeleteDownloadTask(context.Context, DeleteDownloadTaskParams) error
}
type downloadTask struct {
	tokenLogic                  Token
	downloadTaskDataAccessor    database.DownloadTaskDataAccessor
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer
	goquDatabase                *goqu.Database
}

func NewDownloadTask(tokenLogic Token, downloadTaskDataAccessor database.DownloadTaskDataAccessor,
	downloadTaskCreatedProducer producer.DownloadTaskCreatedProducer, goquDatabase *goqu.Database) DownloadTask {
	return &downloadTask{
		tokenLogic:                  tokenLogic,
		downloadTaskDataAccessor:    downloadTaskDataAccessor,
		downloadTaskCreatedProducer: downloadTaskCreatedProducer,
		goquDatabase:                goquDatabase,
	}
}
func (d downloadTask) CreateDownloadTask(ctx context.Context, params CreateDownloadTaskParams) (CreateDownloadTaskOutput, error) {
	accountID, _, err := d.tokenLogic.GetAccountIDAndExpireTime(ctx, params.Token)
	if err != nil {
		return CreateDownloadTaskOutput{}, err
	}
	databaseDownloadTask := database.DownloadTask{
		OfAccountID:    accountID,
		DownloadType:   params.DownloadType,
		URL:            params.URL,
		DownloadStatus: go_load.DownloadStatus_Pending,
		Metadata:       "{}",
	}
	txErr := d.goquDatabase.WithTx(func(td *goqu.TxDatabase) error {
		downloadTaskID, createDownloadTaskErr := d.downloadTaskDataAccessor.
			WithDatabase(td).
			CreateDownloadTask(ctx, databaseDownloadTask)
		if createDownloadTaskErr != nil {
			return createDownloadTaskErr
		}
		databaseDownloadTask.ID = downloadTaskID
		produceErr := d.downloadTaskCreatedProducer.Produce(ctx, producer.DownloadTaskCreated{
			DownloadTask: databaseDownloadTask,
		})
		if produceErr != nil {
			return produceErr
		}
		return nil
	})
	if txErr != nil {
		return CreateDownloadTaskOutput{}, txErr
	}
	return CreateDownloadTaskOutput{
		DownloadTask: go_load.DownloadTask{
			Id:             databaseDownloadTask.ID,
			OfAccount:      nil,
			DownloadType:   params.DownloadType,
			Url:            params.URL,
			DownloadStatus: go_load.DownloadStatus_Pending,
		},
	}, nil
}
func (d downloadTask) GetDownloadTaskList(context.Context, GetDownloadTaskListParams) (GetDownloadTaskListOutput, error) {
	panic("Not implemented")
}
func (d downloadTask) UpdateDownloadTask(context.Context, UpdateDownloadTaskParams) (UpdateDownloadTaskOutput, error) {
	panic("Not implemented")
}
func (d downloadTask) DeleteDownloadTask(context.Context, DeleteDownloadTaskParams) error {
	panic("Not implemented")
}
