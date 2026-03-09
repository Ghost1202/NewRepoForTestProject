package service

import (
	"context"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (s *EventService) ConfirmFromWallet(ctx context.Context, ticketID, userID, eventID int64) error {
	const op = "EventService.ConfirmFromWallet"
	tracer := otel.Tracer("internal/service")
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("service.layer", "service"),
			attribute.Int64("ticket_id", ticketID),
			attribute.Int64("user_id", userID),
			attribute.Int64("event_id", eventID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log)
	if ticketID <= 0 || userID <= 0 || eventID <= 0 {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "invalid_input"),
		)
		span.SetStatus(codes.Ok, "skipped: invalid input")
		log.Warn("invalid wallet charge input",
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
		)
		return nil
	}

	meta, found, err := s.repo.GetTicketMeta(ctx, ticketID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "fetch ticket meta failed")
		log.Error("failed to fetch ticket meta for wallet charge",
			zap.Error(err),
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
		)
		return err
	}
	if !found {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "ticket_not_found"),
		)
		span.SetStatus(codes.Ok, "skipped: ticket not found")
		log.Warn("wallet charge skipped: ticket not found",
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
		)
		return nil
	}
	if meta.EventID != eventID {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "event_mismatch"),
		)
		span.SetStatus(codes.Ok, "skipped: event mismatch")
		log.Warn("wallet charge skipped: event mismatch",
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
			zap.Int64("ticket_event_id", meta.EventID),
		)
		return nil
	}
	if meta.Status != StatusCreated {
		if meta.Status == StatusSold && meta.UserID == userID {
			span.SetAttributes(
				attribute.Bool("skipped", true),
				attribute.String("skip_reason", "already_confirmed"),
			)
			span.SetStatus(codes.Ok, "skipped: already confirmed")
			log.Info("wallet charge already confirmed",
				zap.Int64("ticket_id", ticketID),
				zap.Int64("user_id", userID),
				zap.Int64("event_id", eventID),
			)
			return nil
		}
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "invalid_status"),
		)
		span.SetStatus(codes.Ok, "skipped: invalid status")
		log.Warn("wallet charge skipped: invalid ticket status",
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
			zap.String("status", meta.Status),
		)
		return nil
	}
	if meta.UserID != 0 && meta.UserID != userID {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "owner_mismatch"),
		)
		span.SetStatus(codes.Ok, "skipped: owner mismatch")
		log.Warn("wallet charge skipped: owner mismatch",
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
			zap.Int64("current_user_id", meta.UserID),
		)
		return nil
	}

	rows, err := s.repo.UpdateTicketStatusAndOwner(ctx, ticketID, StatusSold, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "wallet charge confirmation failed")
		log.Error("failed to confirm ticket payment from wallet",
			zap.Error(err),
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
		)
		return err
	}
	if rows == 0 {
		span.SetAttributes(
			attribute.Bool("skipped", true),
			attribute.String("skip_reason", "no_rows_affected"),
		)
		span.SetStatus(codes.Ok, "skipped: no rows affected")
		log.Warn("wallet charge confirmation skipped: no rows affected (state changed)",
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
		)
		return nil
	}

	log.Info("ticket payment confirmed from wallet",
		zap.Int64("ticket_id", ticketID),
		zap.Int64("user_id", userID),
		zap.Int64("event_id", eventID),
	)
	span.SetStatus(codes.Ok, "confirmed")

	return nil
}
