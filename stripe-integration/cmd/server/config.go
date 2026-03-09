package main

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kelseyhightower/envconfig"

	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"
	"github.com/turtlepavlo/stripe_integration/internal/service"
	"github.com/turtlepavlo/stripe_integration/internal/storage/database/postgres"
	"github.com/turtlepavlo/stripe_integration/internal/transport/grpc"
	http "github.com/turtlepavlo/stripe_integration/internal/transport/http"
	"github.com/turtlepavlo/stripe_integration/pkg/producer"
	"github.com/turtlepavlo/stripe_integration/pkg/provider/resend"
	stripe "github.com/turtlepavlo/stripe_integration/pkg/stripe"
)

type Config struct {
	Logger        logger.Config
	Telemetry     telemetry.Config
	Postgres      postgres.Config
	GRPC          grpc.Config
	Stripe        stripe.Config
	Service       service.Config
	Producer      producer.Config
	HTTP          http.Config
	EmeilProvider resend.Config
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Logger),
		validation.Field(&cfg.Telemetry),
		validation.Field(&cfg.Postgres),
		validation.Field(&cfg.GRPC),
		validation.Field(&cfg.Stripe),
		validation.Field(&cfg.Service),
		validation.Field(&cfg.Producer),
		validation.Field(&cfg.HTTP),
		validation.Field(&cfg.EmeilProvider),
	)
}

func Load(ctx context.Context) (Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.ValidateWithContext(ctx); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
