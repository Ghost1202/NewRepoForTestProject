package domain

type Refund struct {
	UserID   int64
	EventID  int64
	TicketID int64
}

type RefundPayment struct {
	UserID   int64
	EventID  int64
	TicketID int64
	Amount   int64
}
