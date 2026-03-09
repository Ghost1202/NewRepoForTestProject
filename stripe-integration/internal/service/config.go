package service

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	DefaultTokenTTL time.Duration `envconfig:"PAYMENT_DEFAULT_TOKEN_TTL" default:"30m"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.DefaultTokenTTL, validation.Required, validation.Min(time.Minute)),
	)
}
