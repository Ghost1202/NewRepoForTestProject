package server

import (
	"strings"

	authv1 "github.com/turtlepavlo/proto-contract/gen/go/sso/auth/v1"
	"github.com/turtlepavlo/sso/internal/domain"
)

func ToDomainCreateUser(req *authv1.RegisterRequest) domain.CreateUser {
	return domain.CreateUser{
		Login:    strings.TrimSpace(req.GetLogin()),
		Email:    strings.ToLower(strings.TrimSpace(req.GetEmail())),
		Password: strings.TrimSpace(req.GetPassword()),
		Name:     strings.TrimSpace(req.GetName()),
		LastName: strings.TrimSpace(req.GetLastName()),
	}
}

func toProtoUserFromLoginResult(result domain.AuthSession, identifier string, isEmail bool) *authv1.User {
	userMessage := &authv1.User{
		Id:       result.UserID,
		Login:    result.Login,
		Role:     result.Role,
		Email:    "",
		Name:     "",
		LastName: "",
	}

	if isEmail {
		userMessage.Email = strings.ToLower(strings.TrimSpace(identifier))
	}

	return userMessage
}

func toProtoUserFromDomain(domainUser *domain.User) *authv1.User {
	if domainUser == nil {
		return nil
	}

	return &authv1.User{
		Id:       domainUser.ID,
		Email:    domainUser.Email,
		Role:     domainUser.Role,
		Login:    domainUser.Login,
		Name:     domainUser.Name,
		LastName: domainUser.LastName,
	}
}
