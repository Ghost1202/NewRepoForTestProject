package service

import (
	"context"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (s *BookingService) GetUserTickets(ctx context.Context, userID int64) ([]domain.Ticket, error) {
	const op = "BookingService.GetUserTickets"
	tracer := otel.Tracer("internal/service")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(attribute.Int64("user_id", userID)),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log)

	ticket, err := s.tickets.GetTicketsByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get tickets by user_id failed")
		log.Error("get tickets by user_id failed",
			zap.Error(err),
			zap.Int64("user_id", userID),
		)
		return nil, err
	}

	out := s.fromStorage.Tickets(ticket)
	return out, nil
}
