package transfer

import "github.com/turtlepavlo/event-service/internal/domain"

type ToDomainConvert struct{}

func NewToDomainConvert() *ToDomainConvert { return &ToDomainConvert{} }

func (c *ToDomainConvert) TicketTransfer(msg TicketTransfer) domain.TicketTransfer {
	return domain.TicketTransfer{
		TicketID:   msg.TicketID,
		FromUserID: msg.FromUserID,
		ToUserID:   msg.ToUserID,
	}
}
