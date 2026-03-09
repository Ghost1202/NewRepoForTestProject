package infobip

type smsRequest struct {
	Messages []smsMessage `json:"messages"`
}

type smsMessage struct {
	Sender       string           `json:"from"`
	Destinations []smsDestination `json:"destinations"`
	Text         string           `json:"text"`
}

type smsDestination struct {
	Recipient string `json:"to"`
}
