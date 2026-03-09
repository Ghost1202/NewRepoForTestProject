package infobip

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	BaseURL  string        `envconfig:"INFOBIP_BASE_URL"`
	APIKey   string        `envconfig:"INFOBIP_API_KEY"`
	SenderID string        `envconfig:"INFOBIP_SENDER_ID" default:"SSO"`
	Timeout  time.Duration `envconfig:"INFOBIP_TIMEOUT" default:"10s"`
}

func (c Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.BaseURL, validation.Required),
		validation.Field(&c.APIKey, validation.Required),
		validation.Field(&c.Timeout, validation.Required))
}
