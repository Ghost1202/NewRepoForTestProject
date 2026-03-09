package domain

type PromoCode struct {
	ID      int64
	EventID int64
	Code    string
	Type    string
	Value   float64
	Sector  string
}
