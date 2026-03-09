package server

import (
	"errors"

	"github.com/turtlepavlo/stripe_integration/internal/service"
)

var (
	ErrPaymentAlreadyPaid = service.ErrPaymentAlreadyPaid
	ErrPaymentLinkExists  = service.ErrPaymentLinkExists

	ErrTicketIDsRequired  = service.ErrTicketIDsRequired
	ErrInvalidTicketID    = service.ErrInvalidTicketID
	ErrDuplicateTicketIDs = service.ErrDuplicateTicketIDs
	ErrInvalidTokenTTL    = service.ErrInvalidTokenTTL

	ErrRefundAlreadyProcessed = errors.New("refund already processed")
	ErrRefundLinkExists       = errors.New("active refund link already exists")
)
