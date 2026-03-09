package service

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	OTPMin int64 `envconfig:"OTP_MIN" default:"1000"`
	OTPMax int64 `envconfig:"OTP_MAX" default:"9999"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.OTPMin, validation.Required),
		validation.Field(&cfg.OTPMax, validation.Required),
	)
}
