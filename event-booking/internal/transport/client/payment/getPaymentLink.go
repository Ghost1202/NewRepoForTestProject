package client

import (
	"context"
	"time"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.uber.org/zap"
)

func (c *PaymentClient) GetPaymentLink(ctx context.Context, booking domain.PaymentInput) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	log := telemetry.WithTrace(ctx, c.log)

	log.Debug("sending grpc request to create payment link",
		zap.Int64("user_id", booking.UserID),
		zap.String("order_id", booking.OrderID),
		zap.Int64("amount", booking.Amount),
		zap.Int64("ticket_ids.count", int64(len(booking.TicketIDs))),
	)

	req := c.converter.ToCreatePaymentRequest(booking)
	resp, err := c.api.CreatePayment(ctx, req)
	if err != nil {
		log.Error("grpc call CreatePayment failed", zap.Error(err))
		return "", err
	}

	return resp.PaymentUrl, nil
}
