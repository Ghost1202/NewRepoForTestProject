package main

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kelseyhightower/envconfig"

	"github.com/turtlepavlo/event-booking/internal/lib/jwt"
	"github.com/turtlepavlo/event-booking/internal/service"
	"github.com/turtlepavlo/event-booking/internal/storage/database/postgres"
	"github.com/turtlepavlo/event-booking/internal/storage/database/redis"
	paymentclient "github.com/turtlepavlo/event-booking/internal/transport/client/payment"
	"github.com/turtlepavlo/event-booking/internal/transport/client/wallet"
	transport "github.com/turtlepavlo/event-booking/internal/transport/http"
	"github.com/turtlepavlo/event-booking/pkg/producer"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

type Config struct {
	Logger        logger.Config
	Telemetry     telemetry.Config
	Postgres      postgres.Config
	Redis         redis.Config
	Service       service.Config
	Transport     transport.Config
	JWT           jwt.Config
	PaymentClient paymentclient.Config
	WalletClient  wallet.Config
	Producer      producer.Config
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Logger),
		validation.Field(&cfg.Telemetry),
		validation.Field(&cfg.Postgres),
		validation.Field(&cfg.Redis),
		validation.Field(&cfg.Service),
		validation.Field(&cfg.Transport),
		validation.Field(&cfg.JWT),
		validation.Field(&cfg.PaymentClient),
		validation.Field(&cfg.WalletClient),
		validation.Field(&cfg.Producer),
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
