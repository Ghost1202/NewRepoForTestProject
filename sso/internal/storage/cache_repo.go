package storage

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var ErrKeyNotFound = errors.New("key not found")

type CacheRepo struct {
	client     *redis.Client
	log        *zap.Logger
	defaultTTL time.Duration
}

func NewCacheRepo(client *redis.Client, log *zap.Logger, ttl time.Duration) *CacheRepo {
	if log == nil {
		log = zap.NewNop()
	}
	return &CacheRepo{
		client:     client,
		log:        log,
		defaultTTL: ttl,
	}
}

func (r *CacheRepo) SetCode(ctx context.Context, key string, value int64) error {
	const op = "CacheRepo.SetCode"
	tracer := otel.Tracer("database/redis")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("cache.key", key),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, r.log).With(zap.String("op", op))

	err := r.client.Set(ctx, key, value, r.defaultTTL).Err()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis set failed")
		log.Error("cache set failed",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (r *CacheRepo) GetCode(ctx context.Context, key string) (int64, error) {
	const op = "CacheRepo.GetCode"
	tracer := otel.Tracer("database/redis")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, r.log).With(zap.String("op", op))

	val, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			span.AddEvent("cache miss")
			return 0, ErrKeyNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "redis get failed")
		log.Error("cache get failed", zap.String("key", key), zap.Error(err))
		return 0, err
	}

	return val, nil
}

func (r *CacheRepo) DeleteCode(ctx context.Context, key string) error {
	const op = "CacheRepo.DeleteCode"
	tracer := otel.Tracer("database/redis")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.String("cache.key", key)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, r.log).With(zap.String("op", op))

	if err := r.client.Del(ctx, key).Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis del failed")
		log.Error("cache delete failed", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}
