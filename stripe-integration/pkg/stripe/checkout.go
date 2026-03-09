package stripe

import (
	"context"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
	"go.uber.org/zap"
)

func (c *Client) CreateCheckoutSession(ctx context.Context, params CheckoutSessionParams) (*CheckoutSessionResult, error) {
	const op = "stripe.Client.CreateCheckoutSession"

	log := telemetry.WithTrace(ctx, c.log).With(
		zap.String("op", op),
		zap.String("order_id", params.OrderID),
	)

	stripeSDKParams := c.mapToSDKParams(params)

	sc := session.Client{
		B:   stripe.GetBackend(stripe.APIBackend),
		Key: c.cfg.APIKey,
	}

	sess, err := sc.New(stripeSDKParams)
	if err != nil {
		log.Error("failed to create stripe session", zap.Error(err))
		return nil, err
	}

	log.Debug("stripe session created successfully", zap.String("session_id", sess.ID))

	return &CheckoutSessionResult{
		URL: sess.URL,
		ID:  sess.ID,
	}, nil
}
