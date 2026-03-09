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

func (srv *EventService) UpdateEvent(ctx context.Context, actor domain.Performer, event *domain.Event) error {
	tracer := otel.Tracer("/internal/service")
	ctx, span := tracer.Start(ctx, "EventService.UpdateEvent")
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)

	span.SetAttributes(
		attribute.Int64("event_id", event.ID),
		attribute.Int64("actor_id", actor.ID),
	)

	ownerID, err := srv.repo.GetEventPerformerID(ctx, event.ID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to get event owner")
		log.Error("failed to get event owner", zap.Error(err))
		return err
	}

	if ownerID != actor.ID {
		log.Warn("access denied: user is not the owner",
			zap.Int64("user_id", actor.ID),
			zap.Int64("owner_id", ownerID),
			zap.Int64("event_id", event.ID),
		)

		err := errors.New("access denied")
		span.RecordError(err)
		span.SetStatus(codes.Error, "forbidden")
		return err
	}

	repoEvent := srv.conv.toRepoEventModel(event)

	if err := srv.repo.UpdateEvent(ctx, repoEvent); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to update event")
		log.Error("failed to update event", zap.Error(err))
		return err
	}

	log.Info("event updated successfully", zap.Int64("event_id", event.ID))
	span.SetStatus(codes.Ok, "updated")

	return nil
}
