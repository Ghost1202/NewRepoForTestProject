package handler

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Port              int64         `envconfig:"HTTP_PORT" default:"8080"`
	Host              string        `envconfig:"HTTP_HOST" default:"0.0.0.0"`
	FrontendURL       string        `envconfig:"FRONTEND_URL" default:""`
	ReadTimeout       time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"5s"`
	ReadHeaderTimeout time.Duration `envconfig:"HTTP_READ_HEADER_TIMEOUT" default:"2s"`
	WriteTimeout      time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"10s"`
	IdleTimeout       time.Duration `envconfig:"HTTP_IDLE_TIMEOUT" default:"120s"`
	MaxHeaderBytes    int           `envconfig:"HTTP_MAX_HEADER_BYTES" default:"1048576"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Port, validation.Required),
		validation.Field(&cfg.Host, validation.Required),
		validation.Field(&cfg.FrontendURL),
		validation.Field(&cfg.ReadTimeout, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&cfg.ReadHeaderTimeout, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&cfg.WriteTimeout, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&cfg.IdleTimeout, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&cfg.MaxHeaderBytes, validation.Required),
	)
}
