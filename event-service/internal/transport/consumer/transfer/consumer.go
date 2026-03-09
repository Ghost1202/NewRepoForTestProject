package transfer

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/turtlepavlo/event-service/internal/domain"
	"go.uber.org/zap"
)

type Service interface {
	TransferTicket(ctx context.Context, transfer domain.TicketTransfer) error
}

type Consumer struct {
	reader *kafka.Reader
	log    *zap.Logger
	svc    Service
	conv   *ToDomainConvert
}

func NewConsumer(brokers []string, cfg Config, log *zap.Logger, svc Service) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       cfg.Topic,
			GroupID:     cfg.GroupID,
			MinBytes:    cfg.MinBytes,
			MaxBytes:    cfg.MaxBytes,
			StartOffset: kafka.FirstOffset,
		}),
		log:  log,
		svc:  svc,
		conv: NewToDomainConvert(),
	}
}
