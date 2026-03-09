package handler

import "github.com/turtlepavlo/event-booking/internal/domain"

type RequestConverter struct{}

func NewRequestConverter() *RequestConverter { return &RequestConverter{} }

func (c *RequestConverter) ToTicketTransfer(userID int64, req TicketTransferRequest) domain.TicketTransfer {
	return domain.TicketTransfer{
		FromUserID: userID,
		TicketID:   req.TicketID,
		ToUserID:   req.ToUserID,
	}
}

func (c *RequestConverter) ToRefund(userID int64, req RefundBookingRequest) domain.Refund {
	return domain.Refund{
		UserID:   userID,
		EventID:  req.EventID,
		TicketID: req.TicketID,
	}
}

func (c *RequestConverter) ToWaitlist(userID int64, req WaitlistRequest) domain.Waitlist {
	return domain.Waitlist{
		UserID:    userID,
		UserEmail: req.Email,
		EventID:   req.EventID,
	}
}

func (c *RequestConverter) ToBooking(userID int64, req CreateBookingRequest) domain.Booking {
	return domain.Booking{
		UserID:    userID,
		TicketIDs: req.TicketIDs,
		UserEmail: req.UserEmail,
		PromoCode: req.PromoCode,
		EarlyID:   req.EarlyID,
		BundleID:  req.BundleID,
	}
}
