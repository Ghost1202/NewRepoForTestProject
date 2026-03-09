package service

import (
	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/internal/storage"
)

type ToStorageConvert struct{}

func NewToStorageConvert() *ToStorageConvert { return &ToStorageConvert{} }

func (c *ToStorageConvert) TicketModel(t *domain.Ticket) *storage.TicketModel {
	return &storage.TicketModel{
		ID:         t.ID,
		EventID:    t.EventID,
		VenueID:    t.VenueID,
		SectorName: t.SectorName,
		RowNumber:  t.RowNumber,
		SeatNumber: t.SeatNumber,
		Price:      t.Price,
		Status:     t.Status,
		UserID:     t.UserID,
		CreatedAt:  t.CreatedAt,
	}
}

func (c *ToStorageConvert) PaymentEvent(userID, _, ticketID int64) *storage.PaymentModel {
	return &storage.PaymentModel{
		TicketID: ticketID,
		UserID:   userID,
		Status:   StatusPaid,
	}
}

func (c *ToStorageConvert) ToPayment(userID int64, orderID string, ticket storage.TicketModel, finalAmount int64) domain.PaymentInput {
	return domain.PaymentInput{
		UserID:    userID,
		OrderID:   orderID,
		Amount:    finalAmount,
		Currency:  CurrencyUSD,
		UserEmail: "",
		TicketIDs: []int64{ticket.ID},
	}
}

func (c *ToStorageConvert) ToPaymentMulti(userID int64, userEmail, orderID string, ticketIDs []int64, finalAmount int64) domain.PaymentInput {
	return domain.PaymentInput{
		UserID:    userID,
		OrderID:   orderID,
		Amount:    finalAmount,
		Currency:  CurrencyUSD,
		UserEmail: userEmail,
		TicketIDs: ticketIDs,
	}
}
