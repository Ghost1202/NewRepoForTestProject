package stripe

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v76"
)

func (c *Client) mapToSDKParams(params CheckoutSessionParams) *stripe.CheckoutSessionParams {
	expiresAt := time.Now().Add(params.TTL)

	metadata := map[string]string{
		"order_id": params.OrderID,
		"user_id":  params.UserID,
	}
	if len(params.TicketIDs) > 0 {
		metadata["ticket_ids"] = formatTicketIDs(params.TicketIDs)
	}
	if params.WalletID > 0 {
		metadata["wallet_id"] = strconv.FormatInt(params.WalletID, 10)
		metadata["payment_flow"] = "wallet-topup"
	}

	sdkParams := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(fmt.Sprintf("%s?order_id=%s", c.cfg.SuccessURL, params.OrderID)),
		CancelURL:  stripe.String(c.cfg.CancelURL),
		ExpiresAt:  stripe.Int64(expiresAt.Unix()),
		Metadata:   metadata,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(params.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(fmt.Sprintf("Order #%s", params.OrderID)),
					},
					UnitAmount: stripe.Int64(params.Amount),
				},
				Quantity: stripe.Int64(1),
			},
		},
	}
	if params.UserEmail != "" {
		sdkParams.CustomerEmail = stripe.String(params.UserEmail)
	}

	return sdkParams
}

func formatTicketIDs(ids []int64) string {
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatInt(id, 10))
	}
	return strings.Join(parts, ",")
}

func (c *Client) mapToRefundSDKParams(ctx context.Context, p RefundParams) *stripe.RefundParams {
	rp := &stripe.RefundParams{
		PaymentIntent: stripe.String(p.PaymentIntentID),
		Reason:        stripe.String(string(stripe.RefundReasonRequestedByCustomer)),
		Metadata: map[string]string{
			"user_id":      strconv.FormatInt(p.UserID, 10),
			"event_id":     strconv.FormatInt(p.EventID, 10),
			"ticket_id":    strconv.FormatInt(p.TicketID, 10),
			"description":  p.Description,
			"payment_flow": "event-booking-refund",
		},
	}
	rp.Amount = stripe.Int64(p.Amount)
	rp.Context = ctx
	rp.IdempotencyKey = stripe.String("refund:ticket:" + strconv.FormatInt(p.TicketID, 10))

	return rp
}
