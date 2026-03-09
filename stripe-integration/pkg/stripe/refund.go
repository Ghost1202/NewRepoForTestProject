package stripe

import (
	"context"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/refund"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
	"go.uber.org/zap"
)

func (c *Client) CreateRefund(ctx context.Context, params RefundParams) error {
	const op = "stripe.Client.CreateRefund"

	log := telemetry.WithTrace(ctx, c.log).With(
		zap.String("op", op),
		zap.String("payment_intent_id", params.PaymentIntentID),
		zap.Int64("user_id", params.UserID),
		zap.Int64("event_id", params.EventID),
		zap.Int64("ticket_id", params.TicketID),
		zap.Int64("amount", params.Amount),
	)

	stripeSDKParams := c.mapToRefundSDKParams(ctx, params)

	rc := refund.Client{
		B:   stripe.GetBackend(stripe.APIBackend),
		Key: c.cfg.APIKey,
	}

	_, err := rc.New(stripeSDKParams)
	if err != nil {
		log.Error("failed to create stripe refund", zap.Error(err))
		return err
	}

	log.Info("stripe refund created successfully")
	return nil
}
