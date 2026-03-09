package transport

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Port              int           `envconfig:"HTTP_PORT" default:"3434"`
	Host              string        `envconfig:"HTTP_HOST" default:"0.0.0.0"`
	ReadTimeout       time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"5s"`
	ReadHeaderTimeout time.Duration `envconfig:"HTTP_READ_HEADER_TIMEOUT" default:"2s"`
	WriteTimeout      time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout       time.Duration `envconfig:"HTTP_IDLE_TIMEOUT" default:"120s"`
	MaxHeaderBytes    int           `envconfig:"HTTP_MAX_HEADER_BYTES" default:"1048576"`

	CORSEnabled         bool   `envconfig:"MW_CORS_ENABLED" default:"true"`
	CORSAllowOrigins    string `envconfig:"MW_CORS_ALLOW_ORIGINS" default:"*"`
	CORSAllowMethods    string `envconfig:"MW_CORS_ALLOW_METHODS" default:"GET,POST,PUT,DELETE,OPTIONS"`
	CORSAllowHeaders    string `envconfig:"MW_CORS_ALLOW_HEADERS" default:"Content-Type, Authorization"`
	LoggerEnabled       bool   `envconfig:"MW_LOGGER_ENABLED" default:"true"`
	RecoverEnabled      bool   `envconfig:"MW_RECOVER_ENABLED" default:"true"`
	RecoverStack        bool   `envconfig:"MW_RECOVER_STACK" default:"true"`
	RecoverMessage      string `envconfig:"MW_RECOVER_MESSAGE" default:"internal server error"`
	StripeWebhookSecret string `envconfig:"STRIPE_WEBHOOK_SECRET" default:""`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.ReadTimeout, validation.Required),
		validation.Field(&cfg.ReadHeaderTimeout, validation.Required),
		validation.Field(&cfg.WriteTimeout, validation.Required),
		validation.Field(&cfg.IdleTimeout, validation.Required),
		validation.Field(&cfg.MaxHeaderBytes, validation.Required),

		validation.Field(&cfg.CORSAllowOrigins, validation.Required),
		validation.Field(&cfg.CORSAllowMethods, validation.Required),
		validation.Field(&cfg.CORSAllowHeaders, validation.Required),
		validation.Field(&cfg.RecoverMessage, validation.Required),
	)
}
