package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (srv *BookingService) BookTickets(ctx context.Context, booking domain.Booking) (string, error) {
	const op = "internal.service.BookTickets"
	tracer := otel.Tracer("/service/booking")

	rawTicketCount := int64(len(booking.TicketIDs))
	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.Int64("user_id", booking.UserID),
			attribute.Int64("ticket_ids.count", rawTicketCount),
			attribute.String("promo_code", booking.PromoCode),
			attribute.Int64("early_id", booking.EarlyID),
			attribute.Int64("bundle_id", booking.BundleID),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, srv.log)
	discountType, ticketIDs, err := srv.validateBookingInput(booking)
	if err != nil {
		span.SetStatus(codes.Error, "invalid booking input")
		return "", err
	}

	tickets, err := srv.tickets.GetTicketsByIDs(ctx, ticketIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("failed to fetch tickets by ids", zap.Error(err))
		return "", err
	}
	if len(tickets) != len(ticketIDs) {
		span.SetStatus(codes.Error, "ticket not found")
		return "", ErrTicketNotFound
	}

	eventID, err := ensureTicketsValid(tickets)
	if err != nil {
		span.SetStatus(codes.Error, "invalid ticket state")
		return "", err
	}

	ticketCount := int64(len(ticketIDs))
	lockedIDs, err := srv.lockTickets(ctx, log, booking.UserID, ticketIDs)
	if err != nil {
		span.SetStatus(codes.Error, "lock failed")
		return "", err
	}

	finalAmount, err := srv.calculateFinalAmount(ctx, discountType, booking, eventID, tickets)
	if err != nil {
		srv.unlockTickets(ctx, lockedIDs, log)
		span.SetStatus(codes.Error, "discount invalid")
		return "", err
	}

	orderID := uuid.NewString()
	paymentInput := srv.toStorage.ToPaymentMulti(booking.UserID, booking.UserEmail, orderID, ticketIDs, finalAmount)
	paymentURL, err := srv.payment.GetPaymentLink(ctx, paymentInput)
	if err != nil {
		srv.unlockTickets(ctx, lockedIDs, log)
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment client failed")
		log.Error("grpc payment call failed", zap.Error(err))
		return "", ErrBookingFailed
	}

	span.SetAttributes(
		attribute.Bool("booking.initiated", true),
		attribute.String("order.id", orderID),
		attribute.Int64("amount.final", finalAmount),
	)

	log.Info("booking successfully initiated",
		zap.Int64("ticket_ids.count", ticketCount),
		zap.String("payment_url", paymentURL),
	)

	return paymentURL, nil
}

func (srv *BookingService) lockTickets(ctx context.Context, log *zap.Logger, userID int64, ticketIDs []int64) ([]int64, error) {
	lockedIDs := make([]int64, 0, len(ticketIDs))
	for _, id := range ticketIDs {
		success, err := srv.cache.Lock(ctx, id, userID, srv.cfg.TicketLockTTL)
		if err != nil {
			srv.unlockTickets(ctx, lockedIDs, log)
			return nil, err
		}
		if !success {
			srv.unlockTickets(ctx, lockedIDs, log)
			return nil, ErrTicketAlreadyReserved
		}
		lockedIDs = append(lockedIDs, id)
	}
	return lockedIDs, nil
}

func (srv *BookingService) unlockTickets(ctx context.Context, ticketIDs []int64, log *zap.Logger) {
	if len(ticketIDs) == 0 {
		return
	}
	if err := srv.cache.Unlock(ctx, ticketIDs); err != nil {
		log.Error("failed to unlock tickets", zap.Error(err))
	}
}
