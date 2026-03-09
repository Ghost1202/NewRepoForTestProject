package service

import (
	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
)

type ConvFromStorage struct{}

func NewConvFromStorage() *ConvFromStorage { return &ConvFromStorage{} }

func (c *ConvFromStorage) ToDomainUser(model *storage.UserLoginModel) *domain.User {
	if model == nil {
		return nil
	}

	return &domain.User{
		ID:       model.UserID,
		Email:    model.Email,
		Phone:    model.Phone,
		Login:    model.Login,
		PassHash: model.PassHash,
		Role:     model.Role,
		GoogleID: model.GoogleID,
	}
}
