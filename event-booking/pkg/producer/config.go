package producer

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host                   string        `envconfig:"TRANSFER_KAFKA_HOST" default:"kafka"`
	Port                   int64         `envconfig:"TRANSFER_KAFKA_PORT" default:"9092"`
	TopicName              string        `envconfig:"TRANSFER_KAFKA_TOPIC" default:"ticket_transfer_requested"`
	AllowAutoTopicCreation bool          `envconfig:"TRANSFER_KAFKA_AUTO_CREATE_TOPIC" default:"false"`
	RequiredAcks           int64         `envconfig:"TRANSFER_KAFKA_REQUIRED_ACKS" default:"1"`
	MaxRetries             int           `envconfig:"TRANSFER_KAFKA_MAX_RETRIES" default:"3"`
	WriteTimeout           time.Duration `envconfig:"TRANSFER_KAFKA_WRITE_TIMEOUT" default:"10s"`
	BatchSize              int           `envconfig:"TRANSFER_KAFKA_BATCH_SIZE" default:"100"`
	BatchTimeout           time.Duration `envconfig:"TRANSFER_KAFKA_BATCH_TIMEOUT" default:"1s"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.TopicName, validation.Required),
		validation.Field(&cfg.RequiredAcks, validation.Required),
		validation.Field(&cfg.MaxRetries, validation.Required),
		validation.Field(&cfg.WriteTimeout, validation.Required),
		validation.Field(&cfg.BatchSize, validation.Required),
		validation.Field(&cfg.BatchTimeout, validation.Required),
	)
}
