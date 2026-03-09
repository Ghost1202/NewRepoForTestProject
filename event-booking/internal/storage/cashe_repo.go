package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const TicketLockPrefix = "ticket:lock:"

func TicketLockKey(ticketID int64) string {
	return fmt.Sprintf("%s%d", TicketLockPrefix, ticketID)
}

type CacheRepo struct {
	redis *redis.Client
	log   *zap.Logger
}

func NewCacheRepo(rdb *redis.Client, log *zap.Logger) *CacheRepo {
	return &CacheRepo{
		redis: rdb,
		log:   log,
	}
}

func (repo *CacheRepo) Lock(ctx context.Context, ticketID, userID int64, ttl time.Duration) (bool, error) {
	const op = "internal.storage.CacheRepo.Lock"
	tracer := otel.Tracer("event-booking/internal/storage/cache_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "SETNX"),
			attribute.Int64("ticket_id", ticketID),
			attribute.Int64("user_id", userID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	key := TicketLockKey(ticketID)

	success, err := repo.redis.SetNX(ctx, key, userID, ttl).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis setnx failed")
		log.Error("redis lock error",
			zap.Error(err),
			zap.Int64("ticket_id", ticketID),
		)
		return false, err
	}

	span.SetAttributes(attribute.Bool("lock.acquired", success))
	return success, nil
}

func (repo *CacheRepo) Unlock(ctx context.Context, ticketIDs []int64) error {
	const op = "internal.storage.CacheRepo.Unlock"
	tracer := otel.Tracer("event-booking/internal/storage/cache_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "DEL"),
			attribute.Int64("tickets.count", int64(len(ticketIDs))),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	if len(ticketIDs) == 0 {
		return nil
	}

	keys := make([]string, len(ticketIDs))
	for i := range ticketIDs {
		keys[i] = TicketLockKey(ticketIDs[i])
	}

	if _, err := repo.redis.Del(ctx, keys...).Result(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis del failed")
		log.Error("redis unlock error", zap.Error(err))
		return err
	}

	span.SetStatus(codes.Ok, "unlock success")
	return nil
}

func (repo *CacheRepo) MGet(ctx context.Context, keys []string) ([]interface{}, error) {
	const op = "internal.storage.CacheRepo.MGet"
	tracer := otel.Tracer("event-booking/internal/storage/cache_repo")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "MGET"),
			attribute.Int64("keys.count", int64(len(keys))),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, repo.log)

	values, err := repo.redis.MGet(ctx, keys...).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis mget failed")

		log.Error("redis mget execution error",
			zap.Error(err),
			zap.Strings("keys", keys),
		)
		return nil, err
	}

	span.SetStatus(codes.Ok, "mget success")
	return values, nil
}
