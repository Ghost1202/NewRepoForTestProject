package domain

import "time"

type User struct {
	ID        int64
	Name      string
	LastName  string
	Login     string
	Email     string
	Phone     string
	PassHash  []byte
	Role      string
	GoogleID  *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
