package telemetry

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Enabled        bool   `envconfig:"OTEL_ENABLED" default:"true"`
	Endpoint       string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
	Insecure       bool   `envconfig:"OTEL_EXPORTER_OTLP_INSECURE" default:"true"`
	ServiceName    string `envconfig:"OTEL_SERVICE_NAME"`
	Environment    string `envconfig:"OTEL_ENVIRONMENT" default:"local"`
	ServiceVersion string `envconfig:"OTEL_SERVICE_VERSION"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	if !cfg.Enabled {
		return nil
	}

	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Endpoint, validation.Required),
		validation.Field(&cfg.ServiceName, validation.Required),
		validation.Field(&cfg.Environment, validation.Required),
	)
}
