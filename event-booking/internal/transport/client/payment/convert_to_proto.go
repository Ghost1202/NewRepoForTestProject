package client

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/turtlepavlo/event-booking/internal/domain"
	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
)

type protoConverter struct {
	tokenTTL time.Duration
}

func newProtoConverter(tokenTTL time.Duration) protoConverter {
	return protoConverter{
		tokenTTL: tokenTTL,
	}
}

func (c protoConverter) ToCreatePaymentRequest(input domain.PaymentInput) *paymentv1.CreatePaymentRequest {
	return &paymentv1.CreatePaymentRequest{
		OrderId:     input.OrderID,
		UserId:      input.UserID,
		Amount:      input.Amount,
		Currency:    input.Currency,
		UserEmail:   input.UserEmail,
		Description: fmt.Sprintf(OrderDescriptionFormat, input.OrderID),
		TokenTtl:    durationpb.New(c.tokenTTL),
		TicketIds:   input.TicketIDs,
	}
}

func (c protoConverter) ToCreateRefundRequest(input domain.RefundPayment) *paymentv1.CreateRefundRequest {
	return &paymentv1.CreateRefundRequest{
		UserId:      input.UserID,
		EventId:     input.EventID,
		TicketId:    input.TicketID,
		Amount:      input.Amount,
		Description: RefundDescription,
		TokenTtl:    durationpb.New(c.tokenTTL),
	}
}

func (c protoConverter) ToWaitlistRequest(input domain.Waitlist) *paymentv1.WaitlistRequest {
	return &paymentv1.WaitlistRequest{
		UserId:    input.UserID,
		EventId:   input.EventID,
		UserEmail: input.UserEmail,
	}
}
