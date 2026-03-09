package service

import (
	"context"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"github.com/turtlepavlo/event-searching/internal/storage"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (srv *SearchService) SearchEventsFilter(ctx context.Context, filter domain.SearchFilter) ([]domain.Event, error) {
	const op = "SearchService.SearchEventsFilter"
	tracer := otel.Tracer("internal/service")
	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)

	storageFilter := srv.toStore.toStorageFilter(filter)
	docs, err := srv.eventRepo.Search(ctx, storageFilter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to search events")

		log.Error("failed to search events", zap.Error(err))
		return nil, err
	}

	if len(docs) > 0 {
		go func(ctx context.Context, foundDocs []storage.EventModel) {
			log := telemetry.WithTrace(ctx, srv.log)
			if err := srv.favouriteRepo.PromoteToPopular(ctx, foundDocs); err != nil {
				log.Warn("failed to promote search results to popular",
					zap.Error(err),
					zap.Int("count", len(foundDocs)),
				)
			}
		}(ctx, docs)
	}

	return srv.toDomain.toDomainEvents(docs), nil
}
