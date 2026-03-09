package main

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kelseyhightower/envconfig"

	"github.com/turtlepavlo/event-service/internal/storage/database"
	"github.com/turtlepavlo/event-service/internal/transport/consumer/payment"
	"github.com/turtlepavlo/event-service/internal/transport/consumer/transfer"
	"github.com/turtlepavlo/event-service/internal/transport/consumer/wallet"
	"github.com/turtlepavlo/event-service/internal/transport/http"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

type Config struct {
	Logger           logger.Config
	Telemetry        telemetry.Config
	DB               database.Config
	Transport        http.Config
	PaymentConsumer  payment.Config
	WalletConsumer   wallet.Config
	TransferConsumer transfer.Config
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.Logger),
		validation.Field(&cfg.Telemetry),
		validation.Field(&cfg.PaymentConsumer),
		validation.Field(&cfg.WalletConsumer),
		validation.Field(&cfg.DB),
		validation.Field(&cfg.Transport),
		validation.Field(&cfg.TransferConsumer),
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
