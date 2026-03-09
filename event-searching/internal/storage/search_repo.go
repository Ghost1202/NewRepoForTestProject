package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

type SearchRepo struct {
	redis    *redis.Client
	log      *zap.Logger
	cacheTTL time.Duration
}

func NewSearchRepo(rdb *redis.Client, logger *zap.Logger, cacheTTL time.Duration) *SearchRepo {
	return &SearchRepo{
		redis:    rdb,
		log:      logger,
		cacheTTL: cacheTTL,
	}
}

const (
	prefixPopular = "popular:ranking"
	prefixData    = "event:data:"
)

func (repo *SearchRepo) GetPopular(ctx context.Context, limit, offset int64) ([]EventModel, error) {
	const op = "repository.Query.GetPopular"
	tracer := otel.Tracer("repository")
	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	logger := telemetry.WithTrace(ctx, repo.log)

	ids, err := repo.redis.ZRevRange(ctx, prefixPopular, offset, offset+limit-1).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get ranking")
		logger.Error("redis ranking fetch failed", zap.Error(err))
		return nil, err
	}

	if len(ids) == 0 {
		return []EventModel{}, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = prefixData + id
	}

	jsonPayloads, err := repo.redis.MGet(ctx, keys...).Result()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to mget data")
		logger.Error("redis mget failed", zap.Error(err))
		return nil, err
	}

	events := make([]EventModel, 0, len(jsonPayloads))
	for i, payload := range jsonPayloads {
		if payload == nil {
			continue
		}

		strPayload, ok := payload.(string)
		if !ok {
			continue
		}

		var doc EventModel
		if err := json.Unmarshal([]byte(strPayload), &doc); err != nil {
			logger.Warn("failed to unmarshal cached event",
				zap.String("key", keys[i]),
				zap.Error(err),
			)
			continue
		}

		events = append(events, doc)
	}

	return events, nil
}

func (repo *SearchRepo) PromoteToPopular(ctx context.Context, events []EventModel) error {
	const op = "repository.Query.PromoteToPopular"
	tracer := otel.Tracer("repository")
	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	logger := telemetry.WithTrace(ctx, repo.log)

	if len(events) == 0 {
		return nil
	}

	pipe := repo.redis.Pipeline()
	for i := range events {
		event := events[i]
		pipe.ZIncrBy(ctx, prefixPopular, 1, fmt.Sprintf("%d", event.ID))

		payload, err := json.Marshal(event)
		if err == nil {
			pipe.Set(ctx, prefixData+fmt.Sprintf("%d", event.ID), payload, repo.cacheTTL)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "redis pipeline failed")
		logger.Error("failed to promote events in redis", zap.Error(err))
		return err
	}

	return nil
}
