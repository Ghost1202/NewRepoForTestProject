package payment

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string `envconfig:"KAFKA_HOST" default:"kafka"`
	Port     int    `envconfig:"KAFKA_PORT" default:"29092"`
	Topic    string `envconfig:"KAFKA_TOPIC" default:"payment_events"`
	GroupID  string `envconfig:"KAFKA_CONSUMER_GROUP" default:"event_service_group"`
	MinBytes int    `envconfig:"KAFKA_MIN_BYTES" default:"10240"`
	MaxBytes int    `envconfig:"KAFKA_MAX_BYTES" default:"10485760"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.Topic, validation.Required),
		validation.Field(&cfg.GroupID, validation.Required),
	)
}
