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

type consumerHandler struct {
	handlerFunc       HandlerFunc
	exitSignalChannel chan os.Signal
}

func newConsumerHandler(handlerFunc HandlerFunc, exitSignalChannel chan os.Signal) *consumerHandler {
	return &consumerHandler{
		handlerFunc:       handlerFunc,
		exitSignalChannel: exitSignalChannel,
	}
}
func (h consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}
func (h consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
func (h consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				session.Commit()
				return nil
			}
			if err := h.handlerFunc(session.Context(), message.Topic, message.Value); err != nil {
				return err
			}
		case <-h.exitSignalChannel:
			session.Commit()
			break
		}
	}
}

type Consumer interface {
	RegisterHandler(queueName string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
}

type consumer struct {
	saramaConsumer            sarama.ConsumerGroup
	queueNameToHandlerFuncMap map[string]HandlerFunc
}

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}
func NewConsumer(mqConfig configs.MQ) (Consumer, error) {
	saramaConsumer, err := sarama.NewConsumerGroup(mqConfig.Addresses, mqConfig.ClientID, newSaramaConfig(mqConfig))
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

func (c consumer) Start(ctx context.Context) error {
	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, os.Interrupt)
	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		go func(queueName string, handlerFunc HandlerFunc) {
			if err := c.saramaConsumer.Consume(
				context.Background(),
				[]string{queueName},
				newConsumerHandler(handlerFunc, exitSignalChannel),
			); err != nil {
				log.Printf("failed to consume message from queue")
			}
		}(queueName, handlerFunc)
	}
	<-exitSignalChannel
	return nil
}
