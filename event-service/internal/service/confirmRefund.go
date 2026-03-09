package service

import (
	"context"
	"errors"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.uber.org/zap"
)

func (s *EventService) ConfirmTicketRefund(ctx context.Context, ticketID int64) error {
	const op = "service.EventService.ConfirmTicketRefund"
	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", op), zap.Int64("ticket_id", ticketID))

	log.Info("processing ticket refund confirmation")
	rows, err := s.repo.UpdateTicketStatusAndOwner(ctx, ticketID, StatusCreated, 0)
	if err != nil {
		log.Error("ConfirmTicketRefund failed", zap.Error(err))
		return err
	}
	if rows == 0 {
		err := errors.New("ticket not found")
		log.Error("ConfirmTicketRefund failed", zap.Error(err))
		return err
	}

	log.Info("ticket refund confirmed: reset to CREATED and user_id=0")
	return nil
}
