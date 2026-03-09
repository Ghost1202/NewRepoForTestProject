package domain

type Booking struct {
	UserID    int64
	TicketIDs []int64
	UserEmail string
	PromoCode string
	EarlyID   int64
	BundleID  int64
}
