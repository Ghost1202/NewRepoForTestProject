package jwt

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	JWTSecret []byte        `envconfig:"JWT_SECRET" required:"true"`
	TokenTTL  time.Duration `envconfig:"JWT_TTL" default:"24h"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.JWTSecret, validation.Required),
		validation.Field(&cfg.TokenTTL, validation.Required),
	)
}
