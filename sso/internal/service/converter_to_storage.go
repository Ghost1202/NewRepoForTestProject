package service

import (
	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
)

type ConvToStorage struct {
	norm *ConvToNormalizer
}

func NewConvToStorage(norm *ConvToNormalizer) *ConvToStorage {
	return &ConvToStorage{norm: norm}
}

func (c *ConvToStorage) FromCreateUser(in domain.CreateUser, passHash []byte) (*storage.UserLoginModel, *storage.UserInfoModel) {
	email := c.norm.Email(in.Email)
	login := c.norm.Login(in.Login)

	return &storage.UserLoginModel{
			Email:    email,
			Login:    login,
			Phone:    "",
			PassHash: passHash,
			Role:     defaultUserRole,
			GoogleID: nil,
		}, &storage.UserInfoModel{
			Name:     c.norm.Name(in.Name),
			LastName: c.norm.Name(in.LastName),
		}
}

func (c *ConvToStorage) CreateUserByPhone(phone string) storage.CreateUserByPhone {
	return storage.CreateUserByPhone{
		Phone: phone,
		Role:  defaultUserRole,
	}
}
