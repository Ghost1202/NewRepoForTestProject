package service

import (
	"context"
	"errors"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (srv *EventService) ConfirmTicketPayment(ctx context.Context, ticketID int64) error {
	tracer := otel.Tracer("/internal/service")
	ctx, span := tracer.Start(ctx, "EventService.ConfirmTicketPayment")
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)
	span.SetAttributes(attribute.Int64("ticket_id", ticketID))

	log.Info("processing ticket payment confirmation", zap.Int64("ticket_id", ticketID))

	rows, err := srv.repo.UpdateTicketStatus(ctx, ticketID, StatusCreated, StatusSold)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment confirmation failed")

		log.Error("failed to confirm ticket payment",
			zap.Int64("ticket_id", ticketID),
			zap.Error(err),
		)
		return err
	}
	if rows == 0 {
		err := errors.New("no rows affected")
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment confirmation skipped")
		log.Warn("ticket payment confirmation skipped",
			zap.Int64("ticket_id", ticketID),
		)
		return err
	}

	log.Info("ticket payment confirmed successfully", zap.Int64("ticket_id", ticketID))
	span.SetStatus(codes.Ok, "confirmed")

	return nil
}
