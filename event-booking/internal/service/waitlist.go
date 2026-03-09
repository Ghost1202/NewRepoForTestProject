package service

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
)

func (s *BookingService) AddToWaitlist(ctx context.Context, item domain.Waitlist) error {
	const op = "Service.AddToWaitlist"
	tracer := otel.Tracer("internal/service")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.Int64("user_id", item.UserID),
			attribute.Int64("event_id", item.EventID),
			attribute.String("user_email", item.UserEmail),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log)

	if err := s.payment.AddToWaitlist(ctx, item); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment client call failed")
		log.Error("failed to add to waitlist via payment client", zap.Error(err))
		return err
	}

	log.Info("waitlist request sent to payment service successfully")
	return nil
}
