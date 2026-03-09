package transfer

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string `envconfig:"TRANSFER_KAFKA_HOST" default:"kafka"`
	Port     int    `envconfig:"TRANSFER_KAFKA_PORT" default:"29092"`
	Topic    string `envconfig:"TRANSFER_KAFKA_TOPIC" default:"ticket_transfer_requested"`
	GroupID  string `envconfig:"TRANSFER_KAFKA_CONSUMER_GROUP" default:"event_service_transfer_group"`
	MinBytes int    `envconfig:"TRANSFER_KAFKA_MIN_BYTES" default:"10240"`
	MaxBytes int    `envconfig:"TRANSFER_KAFKA_MAX_BYTES" default:"10485760"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.Topic, validation.Required),
		validation.Field(&cfg.GroupID, validation.Required),
	)
}
