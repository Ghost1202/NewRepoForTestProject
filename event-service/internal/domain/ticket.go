package domain

type TicketSector struct {
	ID          int64
	EventID     int64
	Name        string
	Type        string
	Price       int64
	RowsCount   int64
	SeatsPerRow int64
}
