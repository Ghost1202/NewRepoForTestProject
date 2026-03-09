package service

import "github.com/turtlepavlo/sso/internal/domain"

type ConvToMessage struct {
	norm *ConvToNormalizer
}

func NewConvToMessage(norm *ConvToNormalizer) *ConvToMessage {
	return &ConvToMessage{norm: norm}
}

func (c *ConvToMessage) SMS(phone, text string) domain.SMSMessage {
	return domain.SMSMessage{
		Phone: c.norm.Phone(phone),
		Text:  text,
	}
}

func (c *ConvToMessage) Email(to, subject, text string) domain.EmailMessage {
	return domain.EmailMessage{
		To:      c.norm.Email(to),
		Subject: subject,
		Text:    text,
	}
}
