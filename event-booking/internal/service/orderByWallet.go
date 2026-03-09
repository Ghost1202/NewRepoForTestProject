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

func (srv *BookingService) OrderByWallet(ctx context.Context, booking domain.Booking) (domain.WalletChargeResult, error) {
	const op = "internal.service.OrderByWallet"
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
	if srv.wallet == nil {
		span.SetStatus(codes.Error, "wallet client not configured")
		log.Error("wallet client not configured")
		return domain.WalletChargeResult{}, ErrBookingFailed
	}

	discountType, ticketIDs, err := srv.validateBookingInput(booking)
	if err != nil {
		span.SetStatus(codes.Error, "invalid booking input")
		return domain.WalletChargeResult{}, err
	}

	tickets, err := srv.tickets.GetTicketsByIDs(ctx, ticketIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "db error")
		log.Error("failed to fetch tickets by ids", zap.Error(err))
		return domain.WalletChargeResult{}, err
	}
	if len(tickets) != len(ticketIDs) {
		span.SetStatus(codes.Error, "ticket not found")
		return domain.WalletChargeResult{}, ErrTicketNotFound
	}

	eventID, err := ensureTicketsValid(tickets)
	if err != nil {
		span.SetStatus(codes.Error, "invalid ticket state")
		return domain.WalletChargeResult{}, err
	}

	lockedIDs, err := srv.lockTickets(ctx, log, booking.UserID, ticketIDs)
	if err != nil {
		span.SetStatus(codes.Error, "lock failed")
		return domain.WalletChargeResult{}, err
	}

	finalAmount, err := srv.calculateFinalAmount(ctx, discountType, booking, eventID, tickets)
	if err != nil {
		srv.unlockTickets(ctx, lockedIDs, log)
		span.SetStatus(codes.Error, "discount invalid")
		return domain.WalletChargeResult{}, err
	}

	bookingID := uuid.NewString()
	charge := srv.toWallet.ChargeInput(bookingID, booking, eventID, ticketIDs, finalAmount)
	status, err := srv.wallet.ChargeWallet(ctx, charge)
	if err != nil {
		srv.unlockTickets(ctx, lockedIDs, log)
		span.RecordError(err)
		span.SetStatus(codes.Error, "wallet charge failed")
		log.Error("wallet charge failed", zap.Error(err))
		return domain.WalletChargeResult{}, ErrBookingFailed
	}

	if status != domain.WalletChargeStatusSuccess {
		srv.unlockTickets(ctx, lockedIDs, log)
	}

	span.SetAttributes(
		attribute.Bool("booking.initiated", status == domain.WalletChargeStatusSuccess),
		attribute.String("booking.id", bookingID),
		attribute.String("wallet.charge_status", string(status)),
		attribute.Int64("amount.final", finalAmount),
	)

	log.Info("wallet charge completed",
		zap.String("booking_id", bookingID),
		zap.String("status", string(status)),
		zap.Int64("amount", finalAmount),
	)

	return srv.toWallet.ChargeResult(bookingID, status), nil
}
