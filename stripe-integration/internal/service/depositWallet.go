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

func (s *PaymentService) DepositWallet(ctx context.Context, input domain.TopupWalletInput) (string, time.Time, error) {
	const op = "service.PaymentService.DepositWallet"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.String("order_id", input.OrderID.String()),
		zap.Int64("user_id", input.UserID),
		zap.Int64("wallet_id", input.WalletID),
	)

	ttl, err := resolveTokenTTL(input.TokenTTL, s.cfg.DefaultTokenTTL)
	if err != nil {
		log.Warn("invalid token_ttl", zap.Error(err))
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(ttl)
	existing, err := s.repo.GetPaymentByOrderID(ctx, input.OrderID)
	switch {
	case err == nil:
		if existing.Status == StatusPaid {
			return "", time.Time{}, ErrPaymentAlreadyPaid
		}
		if existing.Status != StatusFailed {
			return "", time.Time{}, ErrPaymentLinkExists
		}
	case errors.Is(err, pgx.ErrNoRows):
		reservation := s.toStorage.ToTopupReservationParams(input)
		if err := s.repo.CreatePayment(ctx, reservation); err != nil {
			log.Error("failed to create topup payment reservation", zap.Error(err))
			return "", time.Time{}, err
		}
	default:
		log.Error("failed to check payment existence", zap.Error(err))
		return "", time.Time{}, err
	}

	stripePayment := input
	stripePayment.TokenTTL = ttl
	stripeParams := s.toStripe.ToTopupStripeParams(stripePayment)
	result, stripeErr := s.provider.CreateCheckoutSession(ctx, stripeParams)
	updatePayment := s.toStorage.ToFailedParams(input.OrderID)
	if stripeErr != nil {
		log.Error("failed to create stripe session for topup", zap.Error(stripeErr))
	} else {
		updatePayment = s.toStorage.ToCompleteParams(input.OrderID, result.ID)
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

	statusEvent := s.toKafka.ToTopupStatusEvent(input, result.ID, StatusPending)
	if err := s.producer.PublishPaymentStatusEvent(ctx, statusEvent); err != nil {
		log.Error("failed to publish payment status to kafka", zap.Error(err))

		updatePayment := s.toStorage.ToFailedParams(input.OrderID)
		if _, rollbackErr := s.repo.UpdatePaymentStatus(ctx, updatePayment); rollbackErr != nil {
			log.Error("CRITICAL: failed to rollback payment status", zap.Error(rollbackErr))
		}

		return "", time.Time{}, err
	}

	log.Info("topup payment link created successfully", zap.String("url", result.URL))
	return result.URL, expiresAt, nil
}
