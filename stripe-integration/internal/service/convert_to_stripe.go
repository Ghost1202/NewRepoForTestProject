package service

import (
	"fmt"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	stripe "github.com/turtlepavlo/stripe_integration/pkg/stripe"
)

type ToStripeConvert struct{}

func NewToStripeConvert() *ToStripeConvert { return &ToStripeConvert{} }

func (s *ToStripeConvert) ToStripeParams(payment domain.CreatePaymentInput) stripe.CheckoutSessionParams {
	return stripe.CheckoutSessionParams{
		OrderID:   payment.OrderID.String(),
		UserID:    fmt.Sprintf("%d", payment.UserID),
		WalletID:  0,
		UserEmail: payment.UserEmail,
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		TTL:       payment.TokenTTL,
		TicketIDs: payment.TicketIDs,
	}
}

func (s *ToStripeConvert) ToTopupStripeParams(payment domain.TopupWalletInput) stripe.CheckoutSessionParams {
	return stripe.CheckoutSessionParams{
		OrderID:   payment.OrderID.String(),
		UserID:    fmt.Sprintf("%d", payment.UserID),
		WalletID:  payment.WalletID,
		UserEmail: "",
		Amount:    payment.Amount,
		Currency:  payment.Currency,
		TTL:       payment.TokenTTL,
		TicketIDs: nil,
	}
}

func (s *ToStripeConvert) ToRefundParams(refund domain.Refund) stripe.RefundParams {
	return stripe.RefundParams{
		UserID:      refund.UserID,
		EventID:     refund.EventID,
		TicketID:    refund.TicketID,
		Amount:      refund.Amount,
		Description: refund.Description,
	}
}
