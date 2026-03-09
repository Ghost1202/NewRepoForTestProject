package service

import (
	"context"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (srv *BookingService) GetTickets(ctx context.Context, eventID int64) ([]domain.Ticket, error) {
	const op = "internal.service.GetTickets"
	tracer := otel.Tracer("/service/booking")
	ctx, span := tracer.Start(ctx, op, trace.WithAttributes(attribute.Int64("event_id", eventID)))
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)

	models, err := srv.tickets.GetTicketsByEvent(ctx, eventID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("failed to fetch tickets", zap.Error(err))
		return nil, err
	}
	if len(models) == 0 {
		return []domain.Ticket{}, nil
	}

	keys := make([]string, len(models))
	for i, m := range models {
		keys[i] = storage.TicketLockKey(m.ID)
	}

	locks, err := srv.cache.MGet(ctx, keys)
	if err != nil {
		log.Warn("cache mget failed, showing all tickets", zap.Error(err))
		return srv.fromStorage.Tickets(models), nil
	}

	availableModels := make([]storage.TicketModel, 0, len(models))
	for i, lock := range locks {
		if lock == nil {
			availableModels = append(availableModels, models[i])
		}
	}

	span.SetAttributes(attribute.Int64("tickets.available", int64(len(availableModels))))
	return srv.fromStorage.Tickets(availableModels), nil
}
