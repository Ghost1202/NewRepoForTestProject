package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (s *PaymentService) CreatePaymentLink(ctx context.Context, payment domain.CreatePaymentInput) (string, time.Time, error) {
	const op = "service.CreatePaymentLink"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.String("order_id", payment.OrderID.String()),
		zap.Int64("user_id", payment.UserID),
	)

	if err := validateTicketIDs(payment.TicketIDs); err != nil {
		log.Warn("invalid ticket_ids", zap.Error(err))
		return "", time.Time{}, err
	}

	ttl, err := resolveTokenTTL(payment.TokenTTL, s.cfg.DefaultTokenTTL)
	if err != nil {
		log.Warn("invalid token_ttl", zap.Error(err))
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(ttl)
	existing, err := s.repo.GetPaymentByOrderID(ctx, payment.OrderID)
	switch {
	case err == nil:
		if existing.Status == StatusPaid {
			return "", time.Time{}, ErrPaymentAlreadyPaid
		}
		if existing.Status != StatusFailed {
			return "", time.Time{}, ErrPaymentLinkExists
		}
	case errors.Is(err, pgx.ErrNoRows):
		reservation := s.toStorage.ToReservationParams(payment)
		if err := s.repo.CreatePayment(ctx, reservation); err != nil {
			log.Error("failed to create payment reservation", zap.Error(err))
			return "", time.Time{}, err
		}
	default:
		log.Error("failed to check payment existence", zap.Error(err))
		return "", time.Time{}, err
	}

	stripePayment := payment
	stripePayment.TokenTTL = ttl
	stripeParams := s.toStripe.ToStripeParams(stripePayment)
	result, stripeErr := s.provider.CreateCheckoutSession(ctx, stripeParams)
	updatePayment := s.toStorage.ToFailedParams(payment.OrderID)
	if stripeErr != nil {
		log.Error("failed to create stripe session", zap.Error(stripeErr))
	} else {
		updatePayment = s.toStorage.ToCompleteParams(payment.OrderID, result.ID)
	}

	if _, dbErr := s.repo.UpdatePaymentStatus(ctx, updatePayment); dbErr != nil {
		if stripeErr != nil {
			log.Error("failed to update payment status to failed (cleanup)", zap.Error(dbErr))
			return "", time.Time{}, stripeErr
		}

		log.Error("failed to update payment status to pending", zap.Error(dbErr))
		return "", time.Time{}, dbErr
	}

	if stripeErr != nil {
		return "", time.Time{}, stripeErr
	}

	log.Info("payment link created successfully", zap.String("url", result.URL))
	return result.URL, expiresAt, nil
}

func validateTicketIDs(ticketIDs []int64) error {
	if len(ticketIDs) == 0 {
		return ErrTicketIDsRequired
	}

	seen := make(map[int64]struct{}, len(ticketIDs))
	for _, id := range ticketIDs {
		if id <= 0 {
			return ErrInvalidTicketID
		}
		if _, ok := seen[id]; ok {
			return ErrDuplicateTicketIDs
		}
		seen[id] = struct{}{}
	}

	return nil
}

func resolveTokenTTL(ttl, fallback time.Duration) (time.Duration, error) {
	if ttl < 0 {
		return 0, ErrInvalidTokenTTL
	}
	if ttl == 0 {
		return fallback, nil
	}
	return ttl, nil
}
