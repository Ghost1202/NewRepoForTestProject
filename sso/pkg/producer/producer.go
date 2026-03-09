package producer

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	eventv1 "github.com/turtlepavlo/proto-contract/gen/go/analytics/event/v1"
	"google.golang.org/protobuf/proto"
)

const (
	eventTypeLogin    = "login"
	eventTypeRegister = "register"
	eventTypeKey      = "eventType"
)

type Producer struct {
	kafkaWriter *kafka.Writer
}

func New(config Config) *Producer {
	brokerAddress := fmt.Sprintf("%s:%d", config.Host, config.Port)

	return &Producer{
		kafkaWriter: &kafka.Writer{
			Addr:                   kafka.TCP(brokerAddress),
			Topic:                  config.TopicName,
			AllowAutoTopicCreation: config.AllowAutoTopicCreation,

			RequiredAcks: kafka.RequiredAcks(config.RequiredAcks),
			MaxAttempts:  config.MaxRetries,
			WriteTimeout: config.WriteTimeout,
			BatchSize:    config.BatchSize,
			BatchTimeout: config.BatchTimeout,
		},
	}
}

func (producer *Producer) PublishLoginEvent(
	ctx context.Context,
	event *eventv1.LoginEvent,
) error {
	return producer.sendToQueue(ctx, event, eventTypeLogin)
}

func (producer *Producer) PublishRegisterEvent(
	ctx context.Context,
	event *eventv1.RegisterEvent,
) error {
	return producer.sendToQueue(ctx, event, eventTypeRegister)
}

func (producer *Producer) sendToQueue(ctx context.Context, message proto.Message, eventTypeVal string) error {
	messageBytes, err := proto.Marshal(message)

	if err != nil {
		return fmt.Errorf("producer: failed to marshal proto event: %w", err)
	}

	err = producer.kafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   nil,
		Value: messageBytes,
		Headers: []kafka.Header{
			{
				Key:   eventTypeKey,
				Value: []byte(eventTypeVal),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("producer: failed to write to kafka: %w", err)
	}

	return nil
}

func (producer *Producer) Close() error {
	return producer.kafkaWriter.Close()
}
