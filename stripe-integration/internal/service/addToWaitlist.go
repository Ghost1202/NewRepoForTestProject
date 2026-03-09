package service

import (
	"context"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
	"go.uber.org/zap"
)

func (s *PaymentService) AddUserToWaitlist(ctx context.Context, waitList domain.ToWaitlist) error {
	const op = "service.PaymentService.AddUserToWaitlist"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
	)
	log.Info("attempting to add user to waitlist",
		zap.Int64("event_id", waitList.EventID),
		zap.Int64("user_id", waitList.UserID),
		zap.String("email", waitList.UserEmail),
	)

	err := s.repo.AddToWaitlist(ctx, waitList)
	if err != nil {
		log.Error("failed to add user to waitlist",
			zap.Int64("event_id", waitList.EventID),
			zap.Int64("user_id", waitList.UserID),
			zap.Error(err),
		)
		return err
	}

	log.Info("user successfully added to waitlist",
		zap.Int64("event_id", waitList.EventID),
		zap.Int64("user_id", waitList.UserID),
	)

	return nil
}
