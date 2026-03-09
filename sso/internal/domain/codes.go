package domain

type SMSMessage struct {
	Phone string
	Text  string
}

type EmailMessage struct {
	To      string
	Subject string
	Text    string
}
