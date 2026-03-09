package server

import (
	"time"

	"github.com/google/uuid"

	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
	"github.com/turtlepavlo/stripe_integration/internal/domain"
)

type RequestConverter struct{}

func NewRequestConverter() *RequestConverter { return &RequestConverter{} }

func (c *RequestConverter) ToCreatePaymentInput(req *paymentv1.CreatePaymentRequest) (domain.CreatePaymentInput, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return domain.CreatePaymentInput{}, err
	}

	var ttl time.Duration
	if req.GetTokenTtl() != nil {
		ttl = req.GetTokenTtl().AsDuration()
	}

	return domain.CreatePaymentInput{
		OrderID:     orderID,
		UserID:      req.GetUserId(),
		UserEmail:   req.GetUserEmail(),
		Amount:      req.GetAmount(),
		Currency:    req.GetCurrency(),
		Description: req.GetDescription(),
		TokenTTL:    ttl,
		TicketIDs:   req.GetTicketIds(),
	}, nil
}

func (c *RequestConverter) ToCreateRefundInput(req *paymentv1.CreateRefundRequest) domain.Refund {
	return domain.Refund{
		UserID:      req.GetUserId(),
		EventID:     req.GetEventId(),
		TicketID:    req.GetTicketId(),
		Amount:      req.GetAmount(),
		Description: req.GetDescription(),
	}
}

func (c *RequestConverter) ToTopupWalletInput(req *paymentv1.TopupWalletRequest) (domain.TopupWalletInput, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return domain.TopupWalletInput{}, err
	}

	var ttl time.Duration
	if req.GetTokenTtl() != nil {
		ttl = req.GetTokenTtl().AsDuration()
	}

	return domain.TopupWalletInput{
		OrderID:  orderID,
		UserID:   req.GetUserId(),
		WalletID: req.GetWalletId(),
		Amount:   req.GetAmount(),
		Currency: req.GetCurrency(),
		TokenTTL: ttl,
	}, nil
}

func (c *RequestConverter) ToWaitlistInput(req *paymentv1.WaitlistRequest) domain.ToWaitlist {
	return domain.ToWaitlist{
		UserID:    req.GetUserId(),
		EventID:   req.GetEventId(),
		UserEmail: req.GetUserEmail(),
	}
}
