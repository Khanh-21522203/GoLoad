// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package wiring

import (
	"GoLoad/internal/app"
	"GoLoad/internal/configs"
	"GoLoad/internal/dataaccess"
	"GoLoad/internal/dataaccess/cache"
	"GoLoad/internal/dataaccess/database"
	"GoLoad/internal/dataaccess/mq/consumer"
	"GoLoad/internal/dataaccess/mq/producer"
	"GoLoad/internal/handler"
	"GoLoad/internal/handler/consumer"
	"GoLoad/internal/handler/grpc"
	"GoLoad/internal/handler/http"
	"GoLoad/internal/logic"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitializeServer(configFilePath configs.ConfigFilePath) (*app.Server, func(), error) {
	config, err := configs.NewConfig(configFilePath)
	if err != nil {
		return nil, nil, err
	}
	configsDatabase := config.Database
	db, cleanup, err := database.InitializeAndMigrateUpDB(configsDatabase)
	if err != nil {
		return nil, nil, err
	}
	goquDatabase := database.InitializeGoquDB(db)
	configsCache := config.Cache
	client := cache.NewRedisClient(configsCache)
	takenAccountName := cache.NewTakenAccountName(client)
	accountDataAccessor := database.NewAccountDataAccessor(goquDatabase)
	accountPasswordDataAccessor := database.NewAccountPasswordDataAccessor(goquDatabase)
	auth := config.Auth
	hash := logic.NewHash(auth)
	tokenPublicKey := cache.NewTokenPublicKey(client)
	tokenPublicKeyDataAccessor := database.NewTokenPublicKeyDataAccessor(goquDatabase)
	token, err := logic.NewToken(accountDataAccessor, tokenPublicKey, tokenPublicKeyDataAccessor, auth)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	account := logic.NewAccount(goquDatabase, takenAccountName, accountDataAccessor, accountPasswordDataAccessor, hash, token)
	downloadTaskDataAccessor := database.NewDownloadTaskDataAccessor(goquDatabase)
	mq := config.MQ
	producerClient, err := producer.NewClient(mq)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	downloadTaskCreatedProducer := producer.NewDownloadTaskCreatedProducer(producerClient)
	downloadTask := logic.NewDownloadTask(token, downloadTaskDataAccessor, downloadTaskCreatedProducer, goquDatabase)
	goLoadServiceServer := grpc.NewHandler(account, downloadTask)
	configsGRPC := config.GRPC
	server := grpc.NewServer(goLoadServiceServer, configsGRPC)
	configsHTTP := config.HTTP
	httpServer := http.NewServer(configsGRPC, configsHTTP)
	downloadTaskCreated := consumers.NewDownloadTaskCreated()
	consumerConsumer, err := consumer.NewConsumer(mq)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	root := consumers.NewRoot(downloadTaskCreated, consumerConsumer)
	appServer := app.NewServer(server, httpServer, root)
	return appServer, func() {
		cleanup()
	}, nil
}

// wire.go:

var WireSet = wire.NewSet(configs.WireSet, dataaccess.WireSet, logic.WireSet, handler.WireSet, app.WireSet)
