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

func (s *BookingService) TransferTicket(ctx context.Context, transfer domain.TicketTransfer) error {
	const op = "BookingService.TransferTicket"
	tracer := otel.Tracer("internal/service")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.Int64("from_user_id", transfer.FromUserID),
			attribute.Int64("ticket_id", transfer.TicketID),
			attribute.Int64("to_user_id", transfer.ToUserID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log)

	if transfer.FromUserID <= 0 || transfer.TicketID <= 0 || transfer.ToUserID <= 0 || transfer.ToUserID == transfer.FromUserID {
		span.SetStatus(codes.Error, "invalid input")
		log.Warn("invalid ticket transfer input",
			zap.Int64("from_user_id", transfer.FromUserID),
			zap.Int64("ticket_id", transfer.TicketID),
			zap.Int64("to_user_id", transfer.ToUserID),
		)
		return ErrInvalidTicketTransfer
	}

	event := s.toProducer.TicketTransfer(transfer)
	if err := s.producer.TicketTransfer(ctx, event); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "producer publish failed")
		log.Error("ticket transfer publish failed",
			zap.Error(err),
			zap.Int64("ticket_id", transfer.TicketID),
			zap.Int64("from_user_id", transfer.FromUserID),
			zap.Int64("to_user_id", transfer.ToUserID),
		)
		return err
	}

	return nil
}
