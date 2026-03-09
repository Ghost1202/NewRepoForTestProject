package server

import (
	"time"

	paymentv1 "github.com/turtlepavlo/proto-contract/gen/go/payment/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ResponseConverter struct{}

func NewResponseConverter() *ResponseConverter { return &ResponseConverter{} }

func (c *ResponseConverter) ToCreatePaymentResponse(paymentURL string, expiresAt time.Time) *paymentv1.CreatePaymentResponse {
	resp := &paymentv1.CreatePaymentResponse{
		PaymentUrl: paymentURL,
	}
	if !expiresAt.IsZero() {
		resp.ExpiresAt = timestamppb.New(expiresAt)
	}
	return resp
}

func (c *ResponseConverter) ToTopupWalletResponse(paymentURL string, expiresAt time.Time) *paymentv1.TopupWalletResponse {
	resp := &paymentv1.TopupWalletResponse{
		PaymentUrl: paymentURL,
	}
	if !expiresAt.IsZero() {
		resp.ExpiresAt = timestamppb.New(expiresAt)
	}
	return resp
}
