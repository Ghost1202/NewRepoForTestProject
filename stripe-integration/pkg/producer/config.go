package producer

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host                   string        `envconfig:"STRIPE_KAFKA_HOST" default:"kafka"`
	Port                   int64         `envconfig:"STRIPE_KAFKA_PORT" default:"9092"`
	TopicName              string        `envconfig:"STRIPE_KAFKA_TOPIC" default:"payment_events"`
	StatusTopicName        string        `envconfig:"STRIPE_KAFKA_STATUS_TOPIC" default:"payment.status.changed"`
	AllowAutoTopicCreation bool          `envconfig:"STRIPE_KAFKA_AUTO_CREATE_TOPIC" default:"false"`
	RequiredAcks           int           `envconfig:"STRIPE_KAFKA_REQUIRED_ACKS" default:"1"`
	MaxRetries             int           `envconfig:"STRIPE_KAFKA_MAX_RETRIES" default:"3"`
	WriteTimeout           time.Duration `envconfig:"STRIPE_KAFKA_WRITE_TIMEOUT" default:"10s"`
	BatchSize              int           `envconfig:"STRIPE_KAFKA_BATCH_SIZE" default:"100"`
	BatchTimeout           time.Duration `envconfig:"STRIPE_KAFKA_BATCH_TIMEOUT" default:"1s"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.TopicName, validation.Required),
		validation.Field(&cfg.StatusTopicName, validation.Required),
		validation.Field(&cfg.RequiredAcks, validation.Required),
		validation.Field(&cfg.MaxRetries, validation.Required),
		validation.Field(&cfg.WriteTimeout, validation.Required),
		validation.Field(&cfg.BatchSize, validation.Required),
		validation.Field(&cfg.BatchTimeout, validation.Required),
	)
}
