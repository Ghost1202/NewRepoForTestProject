package storage

import "time"

type UserLoginModel struct {
	UserID   int64   `gorm:"column:id;primaryKey;autoIncrement"`
	Login    string  `gorm:"column:user_login"`
	Email    string  `gorm:"column:email"`
	Phone    string  `gorm:"column:phone"`
	PassHash []byte  `gorm:"column:pass_hash"`
	Role     string  `gorm:"column:role"`
	GoogleID *string `gorm:"column:google_id"`
}

type UserInfoModel struct {
	UserID    int64     `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:first_name"`
	LastName  string    `gorm:"column:last_name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

type VerificationCodeModel struct {
	Phone string        `redis:"phone"`
	Code  string        `redis:"code"`
	TTL   time.Duration `redis:"-"`
}

type SaveUserIn struct {
	Login UserLoginModel
	Info  UserInfoModel
}

type SaveGoogleUserIn struct {
	Login UserLoginModel
	Info  UserInfoModel
}

type UpdatePassHashByIDIn struct {
	UserID   int64
	PassHash []byte
}

type CreateUserByPhone struct {
	Phone string
	Role  string
}
