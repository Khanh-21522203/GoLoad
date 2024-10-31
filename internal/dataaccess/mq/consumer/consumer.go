package consumer

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"GoLoad/internal/configs"

	"github.com/IBM/sarama"
)

type HandlerFunc func(ctx context.Context, queueName string, payload []byte) error
type Consumer interface {
	RegisterHandler(queueName string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
}

type consumer struct {
	saramaConsumer            sarama.Consumer
	queueNameToHandlerFuncMap map[string]HandlerFunc
}

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}
func NewConsumer(mqConfig configs.MQ) (Consumer, error) {
	saramaConsumer, err := sarama.NewConsumer(mqConfig.Addresses, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}
	return &consumer{
		saramaConsumer:            saramaConsumer,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}
func (c *consumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
}
func (c consumer) consume(queueName string, handlerFunc HandlerFunc, exitSignalChannel chan os.Signal) error {
	partitionConsumer, err := c.saramaConsumer.ConsumePartition(queueName, 0, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("failed to create sarama partition consumer: %w", err)
	}
	for {
		select {
		case message := <-partitionConsumer.Messages():
			err = handlerFunc(context.Background(), queueName, message.Value)
			if err != nil {
				log.Printf("failed to handle message")
			}
		case <-exitSignalChannel:
			break
		}
	}
}
func (c consumer) Start(ctx context.Context) error {
	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, os.Interrupt)
	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		go func(queueName string, handlerFunc HandlerFunc) {
			if err := c.consume(queueName, handlerFunc, exitSignalChannel); err != nil {
				log.Printf("failed to consume message from queue")
			}
		}(queueName, handlerFunc)
	}
	<-exitSignalChannel
	return nil
}
