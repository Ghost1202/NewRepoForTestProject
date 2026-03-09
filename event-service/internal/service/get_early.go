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

func (srv *EventService) GetEarly(ctx context.Context, actor domain.Performer, filter domain.Filter) ([]domain.Early, error) {
	tracer := otel.Tracer("event-service/internal/service")
	ctx, span := tracer.Start(ctx, "EventService.GetEarly")
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

	storageEarly, err := srv.repo.GetEarly(ctx, filter.EventID, filter.Limit, filter.Offset)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get early promos")
		log.Error("failed to get early promos", zap.Error(err))
		return nil, err
	}

	log.Debug("early promos fetched", zap.Int("count", len(storageEarly)))

	return srv.conv.ToDomainEarly(storageEarly), nil
}
