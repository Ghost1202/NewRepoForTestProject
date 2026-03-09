package service

import "errors"

var (
	ErrPaymentAlreadyPaid     = errors.New("payment already processed")
	ErrPaymentNotFound        = errors.New("payment not found")
	ErrPaymentLinkExists      = errors.New("payment link already exists")
	ErrProducerNoRefundEvents = "producer does not support refund events"
	ErrTicketIDsRequired      = errors.New("ticket_ids are required")
	ErrInvalidTicketID        = errors.New("ticket_ids must be positive")
	ErrDuplicateTicketIDs     = errors.New("ticket_ids must be unique")
	ErrInvalidTokenTTL        = errors.New("token_ttl must be non-negative")
)
