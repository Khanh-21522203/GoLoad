package app

import (
	"GoLoad/internal/configs"
	consumers "GoLoad/internal/handler/consumer"
	"GoLoad/internal/handler/grpc"
	"GoLoad/internal/handler/http"
	"GoLoad/internal/handler/jobs"
	"GoLoad/internal/utils"
	"context"
	"syscall"

	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
)

type StandaloneServer struct {
	grpcServer                                               grpc.Server
	httpServer                                               http.Server
	rootConsumer                                             consumers.Root
	executeAllPendingDownloadTaskJob                         jobs.ExecuteAllPendingDownloadTask
	updateDownloadingAndFailedDownloadTaskStatusToPendingJob jobs.UpdateDownloadingAndFailedDownloadTaskStatusToPending
	cronConfig                                               configs.Cron
	logger                                                   *zap.Logger
}

func NewStandaloneServer(
	grpcServer grpc.Server,
	httpServer http.Server,
	rootConsumer consumers.Root,
	executeAllPendingDownloadTaskJob jobs.ExecuteAllPendingDownloadTask,
	updateDownloadingAndFailedDownloadTaskStatusToPendingJob jobs.UpdateDownloadingAndFailedDownloadTaskStatusToPending,
	cronConfig configs.Cron,
	logger *zap.Logger,
) *StandaloneServer {
	return &StandaloneServer{
		grpcServer:                       grpcServer,
		httpServer:                       httpServer,
		rootConsumer:                     rootConsumer,
		executeAllPendingDownloadTaskJob: executeAllPendingDownloadTaskJob,
		updateDownloadingAndFailedDownloadTaskStatusToPendingJob: updateDownloadingAndFailedDownloadTaskStatusToPendingJob,
		cronConfig: cronConfig,
		logger:     logger,
	}
}
func (s StandaloneServer) scheduleCronJobs(scheduler gocron.Scheduler) error {
	if _, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.ExecuteAllPendingDownloadTask.Schedule, true),
		gocron.NewTask(func() {
			if err := s.executeAllPendingDownloadTaskJob.Run(context.Background()); err != nil {
				s.logger.With(zap.Error(err)).Error("failed to run execute all pending download task job")
			}
		}),
	); err != nil {
		s.logger.With(zap.Error(err)).Error("failed to schedule execute all pending download task job")
		return err
	}
	if _, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.UpdateDownloadingAndFailedDownloadTaskStatusToPending.Schedule, true),
		gocron.NewTask(func() {
			if err := s.executeAllPendingDownloadTaskJob.Run(context.Background()); err != nil {
				s.logger.With(zap.Error(err)).
					Error("failed to run update downloading and failed download task status to pending job")
			}
		}),
	); err != nil {
		s.logger.With(zap.Error(err)).
			Error("failed to schedule update downloading and failed download task status to pending job")
		return err
	}
	return nil
}
func (s StandaloneServer) Start() error {
	if err := s.updateDownloadingAndFailedDownloadTaskStatusToPendingJob.Run(context.Background()); err != nil {
		return err
	}
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		s.logger.With(zap.Error(err)).Error("failed to initialize scheduler")
		return err
	}
	defer func() {
		if shutdownErr := scheduler.Shutdown(); shutdownErr != nil {
			s.logger.With(zap.Error(shutdownErr)).Error("failed to shutdown scheduler")
		}
	}()
	err = s.scheduleCronJobs(scheduler)
	if err != nil {
		return err
	}
	go func() {
		grpcStartErr := s.grpcServer.Start(context.Background())
		s.logger.With(zap.Error(grpcStartErr)).Info("grpc server stopped")
	}()
	go func() {
		httpStartErr := s.httpServer.Start(context.Background())
		s.logger.With(zap.Error(httpStartErr)).Info("http server stopped")
	}()
	go func() {
		consumerStartErr := s.rootConsumer.Start(context.Background())
		s.logger.With(zap.Error(consumerStartErr)).Info("message queue consumer stopped")
	}()

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)
	return nil
}
