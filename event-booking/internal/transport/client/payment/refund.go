package client

import (
	"context"
	"time"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.uber.org/zap"
)

func (c *PaymentClient) CreateRefund(ctx context.Context, refund domain.RefundPayment) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log := telemetry.WithTrace(ctx, c.log)

	log.Debug("sending grpc request to create refund",
		zap.Int64("user_id", refund.UserID),
		zap.Int64("event_id", refund.EventID),
		zap.Int64("ticket_id", refund.TicketID),
		zap.Int64("amount", refund.Amount),
	)

	req := c.converter.ToCreateRefundRequest(refund)
	_, err := c.api.CreateRefund(ctx, req)
	if err != nil {
		log.Error("grpc call CreateRefund failed", zap.Error(err))
		return err
	}

	return nil
}
