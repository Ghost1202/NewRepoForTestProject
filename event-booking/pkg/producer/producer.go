package producer

import (
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

func New(cfg Config, logger *zap.Logger) *Producer {
	addr := net.JoinHostPort(cfg.Host, strconv.FormatInt(cfg.Port, 10))
	return &Producer{
		logger: logger,
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(addr),
			Topic:                  cfg.TopicName,
			AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
			RequiredAcks:           kafka.RequiredAcks(cfg.RequiredAcks),
			MaxAttempts:            cfg.MaxRetries,
			WriteTimeout:           cfg.WriteTimeout,
			BatchSize:              cfg.BatchSize,
			BatchTimeout:           cfg.BatchTimeout,
		},
	}
}
