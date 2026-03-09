package stripe

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	APIKey        string `envconfig:"STRIPE_API_KEY"`
	WebhookSecret string `envconfig:"STRIPE_WEBHOOK_SECRET"`
	SuccessURL    string `envconfig:"STRIPE_SUCCESS_URL"`
	CancelURL     string `envconfig:"STRIPE_CANCEL_URL"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.APIKey, validation.Required),
		validation.Field(&cfg.WebhookSecret, validation.Required),
		validation.Field(&cfg.SuccessURL, validation.Required),
		validation.Field(&cfg.CancelURL, validation.Required),
	)
}
