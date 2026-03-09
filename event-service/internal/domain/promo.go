package domain

type Promo struct {
	ID      int64
	EventID int64
	Sector  string
	Code    string
	Type    string
	Value   float64
}
