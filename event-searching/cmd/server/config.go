package main

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kelseyhightower/envconfig"

	"github.com/turtlepavlo/event-searching/internal/service"
	"github.com/turtlepavlo/event-searching/internal/storage/database/elastic"
	"github.com/turtlepavlo/event-searching/internal/storage/database/postgres"
	"github.com/turtlepavlo/event-searching/internal/storage/database/redis"
	"github.com/turtlepavlo/event-searching/internal/transport/http"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

type Config struct {
	HTTP      http.Config
	Logger    logger.Config
	Telemetry telemetry.Config
	Elastic   elastic.Config
	Redis     redis.Config
	Postgres  postgres.Config
	Service   service.Config
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.HTTP),
		validation.Field(&cfg.Logger),
		validation.Field(&cfg.Telemetry),
		validation.Field(&cfg.Elastic),
		validation.Field(&cfg.Redis),
		validation.Field(&cfg.Postgres),
		validation.Field(&cfg.Service),
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
