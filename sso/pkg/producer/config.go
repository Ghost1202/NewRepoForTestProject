package producer

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host                   string        `envconfig:"ANALYTICS_KAFKA_HOST" default:"kafka"`
	Port                   int64         `envconfig:"ANALYTICS_KAFKA_PORT" default:"29092"`
	TopicName              string        `envconfig:"ANALYTICS_KAFKA_TOPIC" default:"analytics_events"`
	AllowAutoTopicCreation bool          `envconfig:"ALLOW_AUTO_TOPIC_CREATION" default:"false"`
	RequiredAcks           int64         `envconfig:"ANALYTICS_KAFKA_REQUIRED_ACKS" default:"1"`
	MaxRetries             int           `envconfig:"ANALYTICS_KAFKA_MAX_RETRIES" default:"3"`
	WriteTimeout           time.Duration `envconfig:"ANALYTICS_KAFKA_WRITE_TIMEOUT" default:"10s"`
	BatchSize              int           `envconfig:"ANALYTICS_KAFKA_BATCH_SIZE" default:"100"`
	BatchTimeout           time.Duration `envconfig:"ANALYTICS_KAFKA_BATCH_TIMEOUT" default:"1s"`
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
