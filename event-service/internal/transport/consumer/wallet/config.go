package wallet

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string `envconfig:"WALLET_CHARGE_KAFKA_HOST" default:"kafka"`
	Port     int    `envconfig:"WALLET_CHARGE_KAFKA_PORT" default:"29092"`
	Topic    string `envconfig:"WALLET_CHARGE_TOPIC" default:"wallet.charge.completed"`
	GroupID  string `envconfig:"WALLET_CHARGE_GROUP" default:"event_service_wallet_group"`
	MinBytes int    `envconfig:"WALLET_CHARGE_MIN_BYTES" default:"10240"`
	MaxBytes int    `envconfig:"WALLET_CHARGE_MAX_BYTES" default:"10485760"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.Topic, validation.Required),
		validation.Field(&cfg.GroupID, validation.Required),
	)
}
