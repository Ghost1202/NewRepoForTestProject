package resend

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	BaseURL      string        `envconfig:"RESEND_BASE_URL" default:"https://api.resend.com"`
	APIKey       string        `envconfig:"RESEND_API_KEY" required:"true"`
	From         string        `envconfig:"RESEND_FROM" required:"true"`
	Timeout      time.Duration `envconfig:"RESEND_TIMEOUT" default:"10s"`
	MaxBodyBytes int64         `envconfig:"RESEND_MAX_BODY_BYTES" default:"1048576"`
}

func (c Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.BaseURL, validation.Required),
		validation.Field(&c.APIKey, validation.Required),
		validation.Field(&c.From, validation.Required),
		validation.Field(&c.Timeout, validation.Required),
	)
}
