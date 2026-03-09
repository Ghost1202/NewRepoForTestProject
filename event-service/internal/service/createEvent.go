package service

import (
	"context"
	"errors"
	"time"

	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (srv *EventService) CreateEvent(ctx context.Context, actor domain.Performer, event *domain.Event, sectors []domain.TicketSector, promos []domain.Promo, early []domain.Early, bundles []domain.Bundle) (int64, error) {
	tracer := otel.Tracer("/internal/service")
	ctx, span := tracer.Start(ctx, "EventService.CreateEvent")
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)
	if event == nil {
		err := errors.New("event is nil")
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid input")

		log.Warn("invalid input: event is nil",
			zap.Int64("user_id", actor.ID),
		)
		return 0, err
	}

	if actor.Role != Performer {
		err := errors.New("access denied")
		span.RecordError(err)
		span.SetStatus(codes.Error, "forbidden")

		log.Warn("access denied: only performers can create events",
			zap.Int64("user_id", actor.ID),
			zap.String("user_role", actor.Role),
		)
		return 0, err
	}

	if event.StartDate.Before(time.Now()) {
		err := errors.New("event date cannot be in the past")
		span.RecordError(err)
		span.SetStatus(codes.Error, "validation failed")

		log.Warn("validation failed: event date cannot be in the past",
			zap.Time("event_start_date", event.StartDate),
			zap.Int64("user_id", actor.ID),
		)
		return 0, err
	}

	promoCount := len(promos) + len(early) + len(bundles)
	event.PerformerID = actor.ID
	span.SetAttributes(
		attribute.String("event_name", event.Name),
		attribute.Int64("performer_id", event.PerformerID),
		attribute.Int64("venue_id", event.VenueID),
		attribute.Int("sectors_count", len(sectors)),
		attribute.Int("promos_count", promoCount),
	)

	repoEvent := srv.conv.toRepoEventModel(event)
	repoTickets := srv.conv.toRepoTickets(repoEvent.VenueID, sectors)
	repoPromos := srv.conv.ToStoragePromos(promos)
	repoEarly := srv.conv.ToStorageEarly(early)
	repoBundles := srv.conv.ToStorageBundles(bundles)
	id, err := srv.repo.CreateEvent(ctx, repoEvent, repoTickets, repoPromos, repoEarly, repoBundles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to create event with tickets")

		log.Error("failed to create event with tickets",
			zap.Error(err),
		)
		return 0, err
	}

	span.SetAttributes(attribute.Int64("event_id", id))

	log.Info("event created with tickets and promos",
		zap.Int64("event_id", id),
		zap.Int("tickets_count", len(repoTickets)),
		zap.Int("promos_count", promoCount),
	)

	return id, nil
}
