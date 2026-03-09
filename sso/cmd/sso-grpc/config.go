package main

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kelseyhightower/envconfig"

	"github.com/turtlepavlo/sso/internal/lib/jwt"
	"github.com/turtlepavlo/sso/internal/service"
	"github.com/turtlepavlo/sso/internal/storage/database/postgres"
	"github.com/turtlepavlo/sso/internal/storage/database/redis"
	"github.com/turtlepavlo/sso/internal/transport/grpc/interceptor"
	"github.com/turtlepavlo/sso/internal/transport/grpc/server"
	"github.com/turtlepavlo/sso/pkg/logger"
	"github.com/turtlepavlo/sso/pkg/producer"
	"github.com/turtlepavlo/sso/pkg/provider/google"
	"github.com/turtlepavlo/sso/pkg/provider/infobip"
	"github.com/turtlepavlo/sso/pkg/provider/resend"
	"github.com/turtlepavlo/sso/pkg/telemetry"
)

type Config struct {
	DB              postgres.Config
	Logger          logger.Config
	JWT             jwt.Config
	Telemetry       telemetry.Config
	GRPCServer      server.Config
	GRPCInterceptor interceptor.Config
	Producer        producer.Config
	Google          google.Config
	SMSProvider     infobip.Config
	Cache           redis.Config
	Service         service.Config
	EmeilProvider   resend.Config
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.DB),
		validation.Field(&cfg.Logger),
		validation.Field(&cfg.JWT),
		validation.Field(&cfg.Telemetry),
		validation.Field(&cfg.GRPCServer),
		validation.Field(&cfg.GRPCInterceptor),
		validation.Field(&cfg.Producer),
		validation.Field(&cfg.SMSProvider),
		validation.Field(&cfg.Cache),
		validation.Field(&cfg.Service),
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
