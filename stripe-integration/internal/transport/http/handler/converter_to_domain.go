package http

import "github.com/turtlepavlo/stripe_integration/internal/domain"

type convertToDomain struct{}

func newConvertToDomain() convertToDomain { return convertToDomain{} }

func (c convertToDomain) ToRefund(userID, eventID, ticketID, amount int64, description string) domain.Refund {
	return domain.Refund{
		UserID:      userID,
		EventID:     eventID,
		TicketID:    ticketID,
		Amount:      amount,
		Description: description,
	}
}
