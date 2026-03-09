package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (s *PaymentService) CheckWaitlistAndNotify(ctx context.Context, eventID int64) error {
	const op = "service.PaymentService.CheckWaitlistAndNotify"
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.Int64("event_id", eventID),
	)

	log.Info("checking waitlist for available spot")
	waiter, err := s.repo.GetNextWaiter(ctx, eventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug("waitlist is empty, no action needed")
			return nil
		}
		log.Error("failed to get next waiter", zap.Error(err))
		return err
	}

	log = log.With(
		zap.Int64("waiter_user_id", waiter.UserID),
		zap.String("waiter_email", waiter.UserEmail),
	)
	log.Info("found user in waitlist, sending notification")

	notification := domain.Notification{
		To:      waiter.UserEmail,
		Subject: "Good news! A ticket became available!",
		Body:    fmt.Sprintf("Hi! A ticket for event %d is now available. Go and grab it!", eventID),
	}

	if err := s.email.SendEmail(ctx, notification); err != nil {
		log.Error("failed to send email notification", zap.Error(err))
		return err
	}

	if err := s.repo.MarkAsNotified(ctx, waiter.ID); err != nil {
		log.Error("failed to mark waiter as notified", zap.Error(err))
		return err
	}

	log.Info("user notified successfully")
	return nil
}
