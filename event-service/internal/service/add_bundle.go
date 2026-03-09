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

func (srv *EventService) AddBundle(ctx context.Context, actor domain.Performer, eventID int64, bundles []domain.Bundle) ([]int64, error) {
	tracer := otel.Tracer("/internal/service")
	ctx, span := tracer.Start(ctx, "EventService.AddBundle")
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)
	if actor.Role != Performer {
		err := errors.New("access denied")
		span.RecordError(err)
		span.SetStatus(codes.Error, "forbidden")

		log.Warn("access denied: only performers can add promos",
			zap.Int64("user_id", actor.ID),
			zap.String("user_role", actor.Role),
		)
		return nil, err
	}

	span.SetAttributes(
		attribute.Int64("event_id", eventID),
		attribute.Int("promos_count", len(bundles)),
	)

	ownerID, err := srv.repo.GetEventPerformerID(ctx, eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to fetch event owner")
		log.Error("failed to fetch event owner", zap.Error(err))
		return nil, err
	}

	if ownerID != actor.ID {
		err := errors.New("access denied: you do not own this event")
		span.RecordError(err)
		span.SetStatus(codes.Error, "forbidden")
		log.Warn("access denied: performer tried to add promos to alien event",
			zap.Int64("actor_id", actor.ID),
			zap.Int64("owner_id", ownerID),
		)
		return nil, err
	}

	repoBundles := srv.conv.ToStorageBundlesWithEvent(bundles, eventID)
	bundleIDs, err := srv.repo.AddBundle(ctx, repoBundles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to add bundle promos")
		log.Error("failed to add bundle promos", zap.Error(err))
		return nil, err
	}

	log.Info("bundle promos added successfully",
		zap.Int64("event_id", eventID),
		zap.Int("count", len(bundles)),
	)

	return bundleIDs, nil
}
