package service

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (s *PaymentService) ConfirmDeposit(ctx context.Context, orderID uuid.UUID, sessionID string) error {
	const op = "service.PaymentService.ConfirmDeposit"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.String("order_id", orderID.String()),
		zap.String("session_id", sessionID),
	)

	payment, err := s.repo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		log.Error("failed to get payment for confirmation", zap.Error(err))
		return err
	}

	if payment.Status == StatusPaid {
		log.Info("payment already confirmed")
		return nil
	}

	updatePayment := s.toStorage.ToPaidParams(orderID, sessionID)
	if _, err := s.repo.UpdatePaymentStatus(ctx, updatePayment); err != nil {
		log.Error("failed to update payment status to PAID", zap.Error(err))
		return err
	}

	payment.Status = StatusPaid
	payment.ExternalID = sessionID
	statusEvent := s.toKafka.ToPaymentStatusEvent(payment, sessionID, StatusSucceeded)
	if err := s.producer.PublishPaymentStatusEvent(ctx, statusEvent); err != nil {
		log.Error("failed to publish payment status to kafka, reverting status to FAILED", zap.Error(err))

		updatePayment := s.toStorage.ToFailedParams(orderID)
		if _, rollbackErr := s.repo.UpdatePaymentStatus(ctx, updatePayment); rollbackErr != nil {
			log.Error("CRITICAL: failed to rollback payment status", zap.Error(rollbackErr))
		}

		return err
	}

	log.Info("topup confirmed and status event published",
		zap.Int64("user_id", payment.UserID),
		zap.String("status", StatusSucceeded),
	)

	return nil
}
