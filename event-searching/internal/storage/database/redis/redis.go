package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func New(ctx context.Context, cfg Config, log *zap.Logger) (*redis.Client, error) {
	const op = "storage.database.redis.New"
	tracer := otel.Tracer("event-searching/internal/storage/database/redis")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.address", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
			attribute.String("db.stitch_ttl", cfg.StitchTTL.String()),
		),
	)
	defer span.End()

	log = telemetry.WithTrace(ctx, log)

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to ping redis")

		log.Error("failed to connect to redis",
			zap.String("op", op),
			zap.Error(err),
			zap.String("host", cfg.Host),
		)
		return nil, err
	}

	log.Info("successfully connected to redis",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.Int("db", cfg.DB),
		zap.String("stitch_ttl", cfg.StitchTTL.String()),
	)

	return client, nil
}
