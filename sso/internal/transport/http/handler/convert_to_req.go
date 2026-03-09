package handler

import (
	"context"

	"github.com/turtlepavlo/sso/internal/domain"
)

type ReqConverter struct{}

func NewReqConverter() *ReqConverter { return &ReqConverter{} }

func (c *ReqConverter) CreateUser(ctx context.Context, svc AuthService, req CreateUserReq) (int64, error) {
	return svc.CreateUser(ctx, domain.CreateUser{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
		LastName: req.LastName,
		Login:    req.Login,
	})
}

func (c *ReqConverter) Login(ctx context.Context, svc AuthService, req LoginReq) (domain.AuthSession, error) {
	typ := CredTypeLogin
	if req.IsEmail {
		typ = CredTypeEmail
	}

	return svc.LoginByCredentials(ctx, domain.AuthCredentials{
		Identifier: req.Identifier,
		Password:   req.Password,
		Type:       typ,
	})
}

func (c *ReqConverter) LoginByPhone(ctx context.Context, svc AuthService, req LoginByPhoneReq) (domain.AuthSession, error) {
	return svc.LoginByCredentials(ctx, domain.AuthCredentials{
		Identifier: req.Phone,
		Password:   req.Password,
		Type:       CredTypePhone,
	})
}

func (c *ReqConverter) SendOTPByPhone(ctx context.Context, svc AuthService, req SendOTPByPhoneReq) error {
	return svc.SendOTP(ctx, req.Phone, CredTypePhone)
}

func (c *ReqConverter) ConfirmOTPByPhone(ctx context.Context, svc AuthService, req ConfirmOTPByPhoneReq) (domain.AuthSession, error) {
	return svc.ConfirmOTP(ctx, domain.OTPConfirm{
		Destination: req.Phone,
		Type:        CredTypePhone,
		Code:        req.Code,
	})
}

func (c *ReqConverter) AuthByPhone(ctx context.Context, svc AuthService, req LoginByPhoneReq) (domain.AuthSession, error) {
	return svc.LoginByCredentials(ctx, domain.AuthCredentials{
		Identifier: req.Phone,
		Password:   req.Password,
		Type:       CredTypePhone,
	})
}

func (c *ReqConverter) ResetPhone(ctx context.Context, svc AuthService, req ResetPhoneReq) error {
	return svc.ResetPhoneSendOTP(ctx, req.Phone)
}

func (c *ReqConverter) ResetPhoneConfirm(ctx context.Context, svc AuthService, req ResetPhoneConfirmReq) (domain.AuthSession, error) {
	return svc.ConfirmPasswordReset(ctx, domain.PasswordResetConfirm{
		Destination: req.Phone,
		Type:        CredTypePhone,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	})
}

func (c *ReqConverter) ResetEmail(ctx context.Context, svc AuthService, req ResetEmailReq) error {
	return svc.ResetEmailSendOTP(ctx, req.Email)
}

func (c *ReqConverter) ResetEmailConfirm(ctx context.Context, svc AuthService, req ResetEmailConfirmReq) (domain.AuthSession, error) {
	return svc.ConfirmPasswordReset(ctx, domain.PasswordResetConfirm{
		Destination: req.Email,
		Type:        CredTypeEmail,
		Code:        req.Code,
		NewPassword: req.NewPassword,
	})
}
