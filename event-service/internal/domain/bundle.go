package domain

type Bundle struct {
	ID       int64
	EventID  int64
	Sector   string
	Code     string
	BuyCount int64
	GetCount int64
}
