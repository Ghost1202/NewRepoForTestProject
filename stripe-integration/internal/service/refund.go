package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (s *PaymentService) CreateRefund(ctx context.Context, refund domain.Refund) error {
	const op = "service.PaymentService.CreateRefund"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.Int64("user_id", refund.UserID),
		zap.Int64("ticket_id", refund.TicketID),
	)

	params := s.toStripe.ToRefundParams(refund)
	log.Debug("creating refund in stripe")
	if err := s.provider.CreateRefund(ctx, params); err != nil {
		log.Error("stripe refund creation failed", zap.Error(err))
		return err
	}
	log.Info("refund request created successfully")
	return nil
}

func (s *PaymentService) ConfirmRefund(ctx context.Context, externalID string, refund domain.Refund) error {
	const op = "service.PaymentService.ConfirmRefund"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.String("external_id", externalID),
		zap.Int64("ticket_id", refund.TicketID),
	)

	if externalID != "" {
		_, err := s.repo.GetPaymentByExternalID(ctx, externalID)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			log.Warn("payment not found by external_id; skipping db update")
		case err != nil:
			log.Error("GetPaymentByExternalID failed", zap.Error(err))
			return err
		default:
			updateExternalID, status := s.toStorage.ToSetRefundedStatus(externalID)
			rows, updErr := s.repo.UpdatePaymentStatusByExternalID(ctx, updateExternalID, status)
			if updErr != nil {
				log.Error("UpdatePaymentStatusByExternalID failed", zap.Error(updErr))
				return updErr
			}
			if rows == 0 {
				log.Warn("no rows updated for refund status", zap.String("external_id", externalID))
			}
		}
	}

	event := s.toKafka.ToRefundEvent(refund)
	log.Debug("publishing refund event to kafka")
	if err := s.producer.PublishRefundEvent(ctx, event); err != nil {
		log.Error("failed to publish refund event to kafka", zap.Error(err))
		return err
	}

	log.Info("refund confirmed and event published", zap.String("status", StatusRefunded))
	go func() {
		ctx := context.Background()
		if err := s.CheckWaitlistAndNotify(ctx, refund.EventID); err != nil {
			s.log.Error("async waitlist processing failed",
				zap.Int64("event_id", refund.EventID),
				zap.Error(err),
			)
		}
	}()

	return nil
}
