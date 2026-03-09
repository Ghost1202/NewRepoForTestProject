package service

import (
	"context"
	"errors"

	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (srv *EventService) GetBundle(ctx context.Context, actor domain.Performer, filter domain.Filter) ([]domain.Bundle, error) {
	tracer := otel.Tracer("event-service/internal/service")
	ctx, span := tracer.Start(ctx, "EventService.GetBundle")
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)
	if actor.Role != Performer {
		err := errors.New("access denied")
		span.RecordError(err)
		span.SetStatus(codes.Error, "forbidden")
		log.Warn("access denied", zap.Int64("user_id", actor.ID))
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64("filter_event_id", filter.EventID),
		attribute.Int64("limit", filter.Limit),
		attribute.Int64("offset", filter.Offset),
	)

	ownerID, err := srv.repo.GetEventPerformerID(ctx, filter.EventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch event owner")
		log.Error("failed to verify ownership", zap.Error(err))
		return nil, err
	}

	if ownerID != actor.ID {
		err := errors.New("access denied: not your event")
		span.RecordError(err)
		span.SetStatus(codes.Error, "forbidden")
		log.Warn("access denied: attempt to view promos of alien event",
			zap.Int64("actor_id", actor.ID),
			zap.Int64("event_id", filter.EventID),
		)
		return nil, err
	}

	storageBundles, err := srv.repo.GetBundle(ctx, filter.EventID, filter.Limit, filter.Offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get bundle promos")
		log.Error("failed to get bundle promos", zap.Error(err))
		return nil, err
	}

	log.Debug("bundle promos fetched", zap.Int("count", len(storageBundles)))

	return srv.conv.ToDomainBundles(storageBundles), nil
}
