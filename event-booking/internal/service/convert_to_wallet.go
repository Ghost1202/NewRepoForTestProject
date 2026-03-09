package service

import "github.com/turtlepavlo/event-booking/internal/domain"

type ToWalletConvert struct{}

func NewToWalletConvert() *ToWalletConvert { return &ToWalletConvert{} }

func (c *ToWalletConvert) ChargeInput(bookingID string, booking domain.Booking, eventID int64, ticketIDs []int64, finalAmount int64) domain.WalletCharge {
	return domain.WalletCharge{
		BookingID: bookingID,
		UserID:    booking.UserID,
		EventID:   eventID,
		TicketIDs: ticketIDs,
		Amount:    finalAmount,
		Currency:  CurrencyUSD,
	}
}

func (c *ToWalletConvert) ChargeResult(bookingID string, status domain.WalletChargeStatus) domain.WalletChargeResult {
	return domain.WalletChargeResult{
		BookingID: bookingID,
		Status:    status,
	}
}
