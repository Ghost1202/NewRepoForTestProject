package storage

import "time"

type Events struct {
	ID            int64     `gorm:"primaryKey;column:id;autoIncrement"`
	PerformerID   int64     `gorm:"column:performer_id;not null"`
	VenueID       int64     `gorm:"column:venue_id;not null"`
	Name          string    `gorm:"column:name;not null"`
	DateStart     time.Time `gorm:"column:date_start;not null"`
	SoldOut       bool      `gorm:"column:sold_out;default:false"`
	PostDate      time.Time `gorm:"column:post_date;not null"`
	SaleStartDate time.Time `gorm:"column:sale_start_date;not null"`
	MaxPriceCof   float64   `gorm:"column:max_price_cof;not null"`
	MinPriceCof   float64   `gorm:"column:min_price_cof;not null"`
}

type Tickets struct {
	ID         int64     `gorm:"primaryKey;column:id;autoIncrement"`
	EventID    int64     `gorm:"column:event_id;not null"`
	VenueID    int64     `gorm:"column:venue_id;not null"`
	SectorName string    `gorm:"column:sector_name;not null"`
	RowNumber  int64     `gorm:"column:row_number;not null"`
	SeatNumber int64     `gorm:"column:seat_number;not null"`
	Price      int64     `gorm:"column:price;not null"`
	Status     string    `gorm:"column:status;default:created"`
	UserID     int64     `gorm:"column:user_id;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

type TicketMeta struct {
	ID      int64  `gorm:"column:id"`
	EventID int64  `gorm:"column:event_id"`
	Status  string `gorm:"column:status"`
	UserID  int64  `gorm:"column:user_id"`
}

type row struct {
	PerformerID int64 `gorm:"column:performer_id"`
}

type TicketTransfer struct {
	TicketID   int64
	FromUserID int64
	ToUserID   int64
}

type Promo struct {
	ID         int64     `gorm:"primaryKey;column:id;autoIncrement"`
	EventID    int64     `gorm:"column:event_id;not null"`
	Code       string    `gorm:"column:code;not null"`
	Type       string    `gorm:"column:type;not null"`
	Value      float64   `gorm:"column:value;not null"`
	SectorName string    `gorm:"column:sector_name"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

type Early struct {
	ID         int64     `gorm:"primaryKey;column:id;autoIncrement"`
	EventID    int64     `gorm:"column:event_id;not null"`
	Code       string    `gorm:"column:code;not null"`
	Type       string    `gorm:"column:type;not null"`
	Value      float64   `gorm:"column:value;not null"`
	SectorName string    `gorm:"column:sector_name"`
	ValidUntil time.Time `gorm:"column:valid_until;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}

type Bundle struct {
	ID         int64     `gorm:"primaryKey;column:id;autoIncrement"`
	EventID    int64     `gorm:"column:event_id;not null"`
	Code       string    `gorm:"column:code;not null"`
	SectorName string    `gorm:"column:sector_name"`
	BuyCount   int64     `gorm:"column:bundle_buy_count;not null"`
	GetCount   int64     `gorm:"column:bundle_get_count;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime"`
}
