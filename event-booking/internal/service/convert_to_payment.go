package service

import "github.com/turtlepavlo/event-booking/internal/domain"

type ToPaymentConvert struct{}

func NewToPaymentConvert() *ToPaymentConvert { return &ToPaymentConvert{} }

func (c *ToPaymentConvert) RefundPayment(ref domain.Refund, amount int64) domain.RefundPayment {
	return domain.RefundPayment{
		UserID:   ref.UserID,
		EventID:  ref.EventID,
		TicketID: ref.TicketID,
		Amount:   amount,
	}
}
