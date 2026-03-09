package wallet

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Host        string        `envconfig:"WALLET_SERVICE_HOST" default:"host.docker.internal"`
	Port        int           `envconfig:"WALLET_SERVICE_PORT" default:"2437"`
	DialTimeout time.Duration `envconfig:"WALLET_SERVICE_DIAL_TIMEOUT" default:"5s"`
	CallTimeout time.Duration `envconfig:"WALLET_SERVICE_CALL_TIMEOUT" default:"5s"`
}

func (cfg Config) Validate() error {
	return validation.ValidateStruct(&cfg,
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.DialTimeout, validation.Required),
		validation.Field(&cfg.CallTimeout, validation.Required),
	)
}
