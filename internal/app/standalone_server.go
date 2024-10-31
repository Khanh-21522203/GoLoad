package app

import (
	"GoLoad/internal/configs"
	consumers "GoLoad/internal/handler/consumer"
	"GoLoad/internal/handler/grpc"
	"GoLoad/internal/handler/http"
	"GoLoad/internal/handler/jobs"
	"context"
	"log"

	"github.com/go-co-op/gocron/v2"
)

type StandaloneServer struct {
	grpcServer                                               grpc.Server
	httpServer                                               http.Server
	rootConsumer                                             consumers.Root
	executeAllPendingDownloadTaskJob                         jobs.ExecuteAllPendingDownloadTask
	updateDownloadingAndFailedDownloadTaskStatusToPendingJob jobs.UpdateDownloadingAndFailedDownloadTaskStatusToPending
	cronConfig                                               configs.Cron
}

func NewStandaloneServer(
	grpcServer grpc.Server,
	httpServer http.Server,
	rootConsumer consumers.Root,
	executeAllPendingDownloadTaskJob jobs.ExecuteAllPendingDownloadTask,
	updateDownloadingAndFailedDownloadTaskStatusToPendingJob jobs.UpdateDownloadingAndFailedDownloadTaskStatusToPending,
	cronConfig configs.Cron,
) *StandaloneServer {
	return &StandaloneServer{
		grpcServer:                       grpcServer,
		httpServer:                       httpServer,
		rootConsumer:                     rootConsumer,
		executeAllPendingDownloadTaskJob: executeAllPendingDownloadTaskJob,
		updateDownloadingAndFailedDownloadTaskStatusToPendingJob: updateDownloadingAndFailedDownloadTaskStatusToPendingJob,
		cronConfig: cronConfig,
	}
}
func (s StandaloneServer) scheduleCronJobs(scheduler gocron.Scheduler) error {
	if _, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.ExecuteAllPendingDownloadTask.Schedule, true),
		gocron.NewTask(func() {
			if err := s.executeAllPendingDownloadTaskJob.Run(context.Background()); err != nil {
				log.Printf("failed to run execute all pending download task job")
			}
		}),
	); err != nil {
		log.Printf("failed to schedule execute all pending download task job")
		return err
	}
	if _, err := scheduler.NewJob(
		gocron.CronJob(s.cronConfig.UpdateDownloadingAndFailedDownloadTaskStatusToPending.Schedule, true),
		gocron.NewTask(func() {
			if err := s.executeAllPendingDownloadTaskJob.Run(context.Background()); err != nil {
				log.Printf("failed to run update downloading and failed download task status to pending job")
			}
		}),
	); err != nil {
		log.Printf("failed to schedule update downloading and failed download task status to pending job")
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
		log.Printf("failed to initialize scheduler")
		return err
	}
	defer func() {
		if shutdownErr := scheduler.Shutdown(); shutdownErr != nil {
			log.Printf("failed to shutdown scheduler")
		}
	}()
	err = s.scheduleCronJobs(scheduler)
	if err != nil {
		return err
	}
	go func() {
		s.grpcServer.Start(context.Background())
		log.Printf("grpc server stopped")
	}()
	go func() {
		s.httpServer.Start(context.Background())
		log.Printf("http server stopped")
	}()
	go func() {
		s.rootConsumer.Start(context.Background())
		log.Printf("message queue consumer stopped")
	}()
	return nil
}
