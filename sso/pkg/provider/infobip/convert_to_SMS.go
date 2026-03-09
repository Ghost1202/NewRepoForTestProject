package infobip

type ConvertToSMS struct {
}

func NewConvertToSMS() *ConvertToSMS {
	return &ConvertToSMS{}
}

func (c *ConvertToSMS) ToSMSRequest(sender, phone, text string) smsRequest {
	return smsRequest{
		Messages: []smsMessage{
			{
				Sender: sender,
				Destinations: []smsDestination{
					{Recipient: phone},
				},
				Text: text,
			},
		},
	}
}
