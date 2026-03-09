package handler

import "github.com/turtlepavlo/sso/internal/domain"

type RespConverter struct {
}

func NewRespConverter() *RespConverter {
	return &RespConverter{}
}

func (c *RespConverter) ToLoginResponse(result domain.AuthSession) LoginResp {
	return LoginResp{
		AccessToken: result.AccessToken,
		UserID:      result.UserID,
		Role:        result.Role,
		Login:       result.Login,
	}
}

func (c *RespConverter) ToAuthResponse(user domain.User) AuthResp {
	return AuthResp{
		UserID: user.ID,
		Role:   user.Role,
	}
}

func (c *RespConverter) ToCreateUserResponse(userID int64) CreateUserResp {
	return CreateUserResp{
		ID: userID,
	}
}
