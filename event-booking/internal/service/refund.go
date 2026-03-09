package service

import (
	"context"
	"errors"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *BookingService) RefundTicket(ctx context.Context, refund domain.Refund) error {
	const op = "service.BookingService.RefundTicket"

	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.Int64("user_id", refund.UserID),
		zap.Int64("event_id", refund.EventID),
		zap.Int64("ticket_id", refund.TicketID),
	)

	ticket, err := s.tickets.GetTicketByID(ctx, refund.TicketID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTicketNotFound
		}
		log.Error("GetTicketByID failed", zap.Error(err))
		return err
	}
	if ticket.UserID != refund.UserID {
		return ErrTicketNotOwned
	}
	if ticket.Status != StatusPaid {
		return ErrBookingNotPaid
	}

	if err := s.payment.CreateRefund(ctx, s.toPayment.RefundPayment(refund, ticket.Price)); err != nil {
		log.Error("CreateRefund failed", zap.Error(err))
		return err
	}

	return nil
}
