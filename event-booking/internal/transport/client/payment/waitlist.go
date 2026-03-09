package client

import (
	"context"
	"time"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.uber.org/zap"
)

func (c *PaymentClient) AddToWaitlist(ctx context.Context, item domain.Waitlist) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log := telemetry.WithTrace(ctx, c.log)

	log.Debug("sending grpc request to add user to waitlist",
		zap.Int64("user_id", item.UserID),
		zap.Int64("event_id", item.EventID),
		zap.String("email", item.UserEmail),
	)

	req := c.converter.ToWaitlistRequest(item)
	_, err := c.api.Waitlist(ctx, req)
	if err != nil {
		log.Error("grpc call Waitlist failed", zap.Error(err))
		return err
	}

	return nil
}
