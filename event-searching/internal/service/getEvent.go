package service

import (
	"context"
	"time"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"github.com/turtlepavlo/event-searching/internal/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (srv *SearchService) GetPopularEvents(ctx context.Context, limit, offset int64) ([]domain.Event, error) {
	const op = "SearchService.GetPopularEvents"
	tracer := otel.Tracer("/internal/service")
	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)

	docs, err := srv.favouriteRepo.GetPopular(ctx, limit, offset)
	if err != nil {
		log.Warn("redis fetch failed, falling back to elasticsearch", zap.Error(err))
	}

	if len(docs) == 0 {
		log.Info("popular events cache miss: warming up from elasticsearch",
			zap.Int64("limit", limit),
		)

		storageFilter := srv.toStore.toPopularFilter(limit, offset)
		docs, err = srv.eventRepo.Search(ctx, storageFilter)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to fetch from elastic after redis miss")

			log.Error("elastic search failed during warmup", zap.Error(err))
			return nil, err
		}

		if len(docs) > 0 {
			go func(ctx context.Context, warmupDocs []storage.EventModel) {
				ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				log := telemetry.WithTrace(ctx, srv.log)

				if pErr := srv.favouriteRepo.PromoteToPopular(ctx, warmupDocs); pErr != nil {
					log.Warn("failed to warm up popular events in redis", zap.Error(pErr))
				} else {
					log.Debug("redis cache warmed up successfully", zap.Int("count", len(warmupDocs)))
				}
			}(ctx, docs)
		}
	}

	events := srv.toDomain.toDomainEvents(docs)
	return events, nil
}
