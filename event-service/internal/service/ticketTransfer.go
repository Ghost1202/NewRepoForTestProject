package service

import (
	"context"

	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (s *EventService) TransferTicket(ctx context.Context, transfer domain.TicketTransfer) error {
	const op = "EventService.TransferTicket"
	tracer := otel.Tracer("internal/service")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("service.layer", "service"),
			attribute.Int64("ticket_id", transfer.TicketID),
			attribute.Int64("from_user_id", transfer.FromUserID),
			attribute.Int64("to_user_id", transfer.ToUserID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log)
	if transfer.TicketID <= 0 || transfer.FromUserID <= 0 || transfer.ToUserID <= 0 {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "invalid_input"),
		)
		span.SetStatus(codes.Ok, "skipped: invalid input")
		log.Warn("invalid ticket transfer input",
			zap.Int64("ticket_id", transfer.TicketID),
			zap.Int64("from_user_id", transfer.FromUserID),
			zap.Int64("to_user_id", transfer.ToUserID),
		)
		return nil
	}

	if transfer.FromUserID == transfer.ToUserID {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "same_user"),
		)
		span.SetStatus(codes.Ok, "skipped: same user")
		log.Warn("skip ticket transfer to same user",
			zap.Int64("ticket_id", transfer.TicketID),
			zap.Int64("user_id", transfer.FromUserID),
		)
		return nil
	}

	rows, err := s.repo.UpdateTicketOwner(ctx, transfer.TicketID, transfer.FromUserID, transfer.ToUserID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo transfer failed")
		log.Error("ticket transfer failed",
			zap.Error(err),
			zap.Int64("ticket_id", transfer.TicketID),
			zap.Int64("from_user_id", transfer.FromUserID),
			zap.Int64("to_user_id", transfer.ToUserID),
		)
		return err
	}
	if rows == 0 {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "no_rows_affected"),
		)
		span.SetStatus(codes.Ok, "skipped: no rows affected")
		log.Warn("ticket transfer skipped: no rows affected",
			zap.Int64("ticket_id", transfer.TicketID),
			zap.Int64("from_user_id", transfer.FromUserID),
			zap.Int64("to_user_id", transfer.ToUserID),
		)
		return nil
	}

	log.Info("ticket transfer processed",
		zap.Int64("ticket_id", transfer.TicketID),
		zap.Int64("from_user_id", transfer.FromUserID),
		zap.Int64("to_user_id", transfer.ToUserID),
	)

	return nil
}
