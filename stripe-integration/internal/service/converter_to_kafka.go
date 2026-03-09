package service

import (
	"time"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/producer"
)

type ToKafkaConvert struct{}

func NewToKafkaConvert() *ToKafkaConvert { return &ToKafkaConvert{} }

func (s *ToKafkaConvert) ToPaymentEvent(model domain.Payment) producer.PaymentConfirmedEvent {
	return producer.PaymentConfirmedEvent{
		OrderID:    model.OrderID.String(),
		Amount:     model.Amount,
		Currency:   model.Currency,
		PaymentID:  model.ExternalID,
		Status:     StatusPaid,
		OccurredAt: time.Now(),
	}
}

func (s *ToKafkaConvert) ToRefundEvent(refund domain.Refund) producer.TicketRefundedEvent {
	return producer.TicketRefundedEvent{
		UserID:     refund.UserID,
		EventID:    refund.EventID,
		TicketID:   refund.TicketID,
		Amount:     refund.Amount,
		Currency:   DefaultCurrencyUSD,
		Status:     StatusRefunded,
		OccurredAt: time.Now(),
	}
}

func (s *ToKafkaConvert) ToPaymentStatusEvent(payment domain.Payment, providerID, status string) producer.PaymentStatusChangedEvent {
	return producer.PaymentStatusChangedEvent{
		OrderID:           payment.OrderID.String(),
		UserID:            payment.UserID,
		Amount:            payment.Amount,
		Currency:          payment.Currency,
		Status:            status,
		ProviderPaymentID: providerID,
		OccurredAt:        time.Now(),
	}
}

func (s *ToKafkaConvert) ToTopupStatusEvent(input domain.TopupWalletInput, providerID, status string) producer.PaymentStatusChangedEvent {
	return producer.PaymentStatusChangedEvent{
		OrderID:           input.OrderID.String(),
		UserID:            input.UserID,
		Amount:            input.Amount,
		Currency:          input.Currency,
		Status:            status,
		ProviderPaymentID: providerID,
		OccurredAt:        time.Now(),
	}
}
