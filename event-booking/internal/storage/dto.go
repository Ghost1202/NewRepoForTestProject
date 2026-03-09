package storage

import "time"

type TicketModel struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	EventID    int64     `gorm:"column:event_id"`
	VenueID    int64     `gorm:"column:venue_id"`
	SectorName string    `gorm:"column:sector_name"`
	RowNumber  int64     `gorm:"column:row_number"`
	SeatNumber int64     `gorm:"column:seat_number"`
	Price      int64     `gorm:"column:price"`
	Status     string    `gorm:"column:status"`
	UserID     int64     `gorm:"column:user_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (TicketModel) TableName() string {
	return "tickets"
}

type PaymentModel struct {
	TicketID int64  `json:"ticket_id"`
	UserID   int64  `json:"user_id"`
	Status   string `json:"status"`
}

type Promo struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	EventID    int64     `gorm:"column:event_id;not null"`
	Code       string    `gorm:"column:code;not null"`
	Type       string    `gorm:"column:type;not null"`
	Value      float64   `gorm:"column:value;not null"`
	SectorName string    `gorm:"column:sector_name"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (Promo) TableName() string {
	return "promo"
}

type Early struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	EventID    int64     `gorm:"column:event_id"`
	Code       string    `gorm:"column:code"`
	Type       string    `gorm:"column:type"`
	Value      float64   `gorm:"column:value"`
	SectorName string    `gorm:"column:sector_name"`
	ValidUntil time.Time `gorm:"column:valid_until"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (Early) TableName() string {
	return "early"
}

type Bundle struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	EventID        int64     `gorm:"column:event_id"`
	Code           string    `gorm:"column:code"`
	SectorName     string    `gorm:"column:sector_name"`
	BundleBuyCount int64     `gorm:"column:bundle_buy_count"`
	BundleGetCount int64     `gorm:"column:bundle_get_count"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (Bundle) TableName() string {
	return "bundle"
}
