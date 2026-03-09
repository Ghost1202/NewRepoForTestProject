package service

import "errors"

var (
	ErrTicketAlreadyReserved  = errors.New("ticket already reserved")
	ErrReservationExpired     = errors.New("reservation expired")
	ErrOwnerMismatch          = errors.New("ownership mismatch")
	ErrBookingFailed          = errors.New("failed to initiate booking")
	ErrInvalidBookingInput    = errors.New("invalid booking input")
	ErrInvalidTicketTransfer  = errors.New("invalid ticket transfer")
	ErrTicketNotFound         = errors.New("ticket not found")
	ErrTicketNotOwned         = errors.New("ticket not owned by user")
	ErrBookingNotPaid         = errors.New("booking is not paid")
	ErrRefundAlreadyProcessed = errors.New("refund already processed")
)
