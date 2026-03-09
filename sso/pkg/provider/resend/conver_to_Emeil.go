package resend

type ConvertToEmail struct{}

func NewConvertToEmail() *ConvertToEmail {
	return &ConvertToEmail{}
}

func (c *ConvertToEmail) ToSendEmailRequest(from, to, subject, text string) sendEmailRequest {
	return sendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    text,
	}
}
