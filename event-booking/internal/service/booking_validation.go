package service

import (
	"github.com/badoux/checkmail"
	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
)

func (srv *BookingService) validateBookingInput(booking domain.Booking) (discount string, ids []int64, err error) {
	discount, err = selectDiscountType(booking)
	if err != nil {
		return "", nil, err
	}
	if err := validateUserEmail(booking.UserEmail, srv.cfg.EmailValidateHost); err != nil {
		return "", nil, err
	}
	ids, err = normalizeTicketIDs(booking.TicketIDs)
	if err != nil {
		return "", nil, err
	}

	return discount, ids, nil
}

func selectDiscountType(booking domain.Booking) (string, error) {
	switch {
	case booking.PromoCode == "" && booking.EarlyID == 0 && booking.BundleID == 0:
		return DiscountTypeNone, nil
	case booking.PromoCode != "" && booking.EarlyID == 0 && booking.BundleID == 0:
		return DiscountTypePromo, nil
	case booking.PromoCode == "" && booking.EarlyID != 0 && booking.BundleID == 0:
		return DiscountTypeEarly, nil
	case booking.PromoCode == "" && booking.EarlyID == 0 && booking.BundleID != 0:
		return DiscountTypeBundle, nil
	default:
		return "", ErrInvalidBookingInput
	}
}

func normalizeTicketIDs(ticketIDs []int64) ([]int64, error) {
	for i, id := range ticketIDs {
		if id <= 0 || (i > 0 && id == ticketIDs[i-1]) {
			return nil, ErrInvalidBookingInput
		}
	}
	return ticketIDs, nil
}

func validateUserEmail(email string, validateHost bool) error {
	if err := checkmail.ValidateFormat(email); err != nil {
		return ErrInvalidBookingInput
	}
	if !validateHost {
		return nil
	}
	if err := checkmail.ValidateHost(email); err != nil {
		return ErrInvalidBookingInput
	}
	return nil
}

func ensureTicketsValid(tickets []storage.TicketModel) (int64, error) {
	eventID := tickets[0].EventID
	for _, ticket := range tickets {
		if ticket.EventID != eventID || ticket.Status != StatusCreated {
			return 0, ErrInvalidBookingInput
		}
	}

	return eventID, nil
}
