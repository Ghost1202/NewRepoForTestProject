package client

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host     string        `envconfig:"PAYMENT_SERVICE_HOST" default:"host.docker.internal"`
	Port     int           `envconfig:"PAYMENT_SERVICE_PORT" default:"9094"`
	TokenTTL time.Duration `envconfig:"PAYMENT_TOKEN_TTL" default:"30m"`
}

func (cfg Config) Validate() error {
	return validation.ValidateStruct(&cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.TokenTTL, validation.Required),
	)
}
