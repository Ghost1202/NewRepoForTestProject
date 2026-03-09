package domain

type PaymentInput struct {
	UserID    int64
	OrderID   string
	Amount    int64
	Currency  string
	UserEmail string
	TicketIDs []int64
}
