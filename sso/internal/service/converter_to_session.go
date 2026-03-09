package service

import "github.com/turtlepavlo/sso/internal/domain"

type ConvToSession struct{}

func NewConvToSession() *ConvToSession { return &ConvToSession{} }

func (c *ConvToSession) ToAuthSession(token string, u *domain.User) domain.AuthSession {
	if u == nil {
		return domain.AuthSession{AccessToken: token}
	}

	return domain.AuthSession{
		AccessToken: token,
		UserID:      u.ID,
		Role:        u.Role,
		Login:       u.Login,
	}
}
