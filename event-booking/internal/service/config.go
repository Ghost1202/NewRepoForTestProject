package service

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	TicketLockTTL     time.Duration `envconfig:"BOOKING_TICKET_LOCK_TTL" default:"10m"`
	EmailValidateHost bool          `envconfig:"EMAIL_VALIDATE_HOST" default:"false"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.TicketLockTTL, validation.Required),
	)
}
