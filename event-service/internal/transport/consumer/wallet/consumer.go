package wallet

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Service interface {
	ConfirmFromWallet(ctx context.Context, ticketID, userID, eventID int64) error
}

type Consumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	service Service
}

func NewConsumer(brokers []string, cfg Config, logger *zap.Logger, service Service) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       cfg.Topic,
			GroupID:     cfg.GroupID,
			MinBytes:    cfg.MinBytes,
			MaxBytes:    cfg.MaxBytes,
			StartOffset: kafka.FirstOffset,
		}),
		logger:  logger,
		service: service,
	}
}
