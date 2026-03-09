package service

import (
	"strings"

	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
	"github.com/turtlepavlo/sso/pkg/provider/google"
)

type ConvToGoogle struct {
	norm *ConvToNormalizer
}

func NewConvToGoogle(norm *ConvToNormalizer) *ConvToGoogle {
	return &ConvToGoogle{norm: norm}
}

func (c *ConvToGoogle) LoginFromEmail(email string) string {
	parts := strings.Split(c.norm.Email(email), "@")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func (c *ConvToGoogle) ToStorageModels(gUser *google.UserInfo, login string) (*storage.UserLoginModel, *storage.UserInfoModel) {
	if gUser == nil {
		return &storage.UserLoginModel{}, &storage.UserInfoModel{}
	}
	googleID := gUser.ID
	return &storage.UserLoginModel{
			Email:    c.norm.Email(gUser.Email),
			Login:    c.norm.Login(login),
			Phone:    "",
			PassHash: nil,
			Role:     defaultUserRole,
			GoogleID: &googleID,
		}, &storage.UserInfoModel{
			Name:     strings.TrimSpace(gUser.FirstName),
			LastName: strings.TrimSpace(gUser.LastName),
		}
}

func (c *ConvToGoogle) ToDomainUser(id int64, gUser *google.UserInfo, login, role string) *domain.User {
	if gUser == nil {
		return nil
	}
	googleID := gUser.ID
	return &domain.User{
		ID:       id,
		Email:    c.norm.Email(gUser.Email),
		Login:    c.norm.Login(login),
		Phone:    "",
		Role:     role,
		GoogleID: &googleID,
		Name:     strings.TrimSpace(gUser.FirstName),
		LastName: strings.TrimSpace(gUser.LastName),
	}
}

func (c *ConvToGoogle) AssignGoogleID(user *domain.User, gUser *google.UserInfo) {
	if user == nil || gUser == nil {
		return
	}
	googleID := gUser.ID
	user.GoogleID = &googleID
}
