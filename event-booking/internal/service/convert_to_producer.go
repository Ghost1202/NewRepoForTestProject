package service

import (
	"time"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/producer"
)

type ToProducerConvert struct{}

func NewToProducerConvert() *ToProducerConvert { return &ToProducerConvert{} }

func (c *ToProducerConvert) TicketTransfer(in domain.TicketTransfer) producer.TicketTransfer {
	return producer.TicketTransfer{
		TicketID:   in.TicketID,
		FromUserID: in.FromUserID,
		ToUserID:   in.ToUserID,
		OccurredAt: time.Now(),
	}
}
