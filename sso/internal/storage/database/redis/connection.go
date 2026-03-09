package redis

import (
	"context"
	"net"
	"strconv"

	"github.com/redis/go-redis/v9"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func New(ctx context.Context, cfg Config, log *zap.Logger) (*redis.Client, error) {
	const op = "storage.redis.New"
	tracer := otel.Tracer("database/redis")

	if log == nil {
		log = zap.NewNop()
	}

	addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.address", addr),
			attribute.Int("db.redis.database_index", cfg.DB),
		),
	)
	defer span.End()

	log = telemetry.WithTrace(ctx, log).With(zap.String("op", op))

	opts := &redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis ping failed")

		log.Error("redis connect failed",
			zap.String("addr", addr),
			zap.Int("db", cfg.DB),
			zap.Error(err),
		)
		return nil, err
	}

	log.Info("redis connected",
		zap.String("addr", addr),
		zap.Int("db", cfg.DB),
	)

	return client, nil
}
