package producer

import (
	"errors"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer       *kafka.Writer
	statusWriter *kafka.Writer
	logger       *zap.Logger
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
		statusWriter: &kafka.Writer{
			Addr:                   kafka.TCP(addr),
			Topic:                  cfg.StatusTopicName,
			AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
			RequiredAcks:           kafka.RequiredAcks(cfg.RequiredAcks),
			MaxAttempts:            cfg.MaxRetries,
			WriteTimeout:           cfg.WriteTimeout,
			BatchSize:              cfg.BatchSize,
			BatchTimeout:           cfg.BatchTimeout,
		},
	}
}

func (p *Producer) Close() error {
	var err1, err2 error
	if p.statusWriter != nil {
		err1 = p.statusWriter.Close()
	}
	if p.writer != nil {
		err2 = p.writer.Close()
	}
	return errors.Join(err1, err2)
}
