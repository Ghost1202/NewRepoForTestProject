package resend

import "github.com/turtlepavlo/stripe_integration/internal/domain"

type ConverterToMSG struct{}

func NewConverterToMSG() *ConverterToMSG {
	return &ConverterToMSG{}
}

func (c *ConverterToMSG) ToSendEmailRequest(from string, msg domain.Notification) sendEmailRequest {
	return sendEmailRequest{
		From:    from,
		To:      []string{msg.To},
		Subject: msg.Subject,
		Html:    msg.Body,
		Text:    "",
	}
}
