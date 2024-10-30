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
	RegisterHandler(queueName string, handlerFunc HandlerFunc) error
	Start(ctx context.Context) error
}
type partitionConsumerAndHandlerFunc struct {
	queueName         string
	partitionConsumer sarama.PartitionConsumer
	handlerFunc       HandlerFunc
}
type consumer struct {
	saramaConsumer                      sarama.Consumer
	partitionConsumerAndHandlerFuncList []partitionConsumerAndHandlerFunc
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
		saramaConsumer: saramaConsumer,
	}, nil
}
func (c *consumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) error {
	partitionConsumer, err := c.saramaConsumer.ConsumePartition(queueName, 0, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("failed to create sarama partition consumer: %w", err)
	}
	c.partitionConsumerAndHandlerFuncList = append(
		c.partitionConsumerAndHandlerFuncList,
		partitionConsumerAndHandlerFunc{
			queueName:         queueName,
			partitionConsumer: partitionConsumer,
			handlerFunc:       handlerFunc,
		})
	return nil
}
func (c consumer) Start(_ context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for i := range c.partitionConsumerAndHandlerFuncList {
		go func(i int) {
			queueName := c.partitionConsumerAndHandlerFuncList[i].queueName
			partitionConsumer := c.partitionConsumerAndHandlerFuncList[i].partitionConsumer
			handlerFunc := c.partitionConsumerAndHandlerFuncList[i].handlerFunc
			for {
				select {
				case message := <-partitionConsumer.Messages():
					if err := handlerFunc(context.Background(), queueName, message.Value); err != nil {
						log.Printf("failed to handle message")
					}
				case <-signals:
					break
				}
			}
		}(i)
	}
	<-signals
	return nil
}
