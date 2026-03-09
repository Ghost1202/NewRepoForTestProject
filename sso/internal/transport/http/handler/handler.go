package handler

import (
	"context"

	"github.com/turtlepavlo/sso/internal/domain"
	"go.uber.org/zap"
)

const (
	CredTypeEmail = "email"
	CredTypeLogin = "login"
	CredTypePhone = "phone"
)

// @title           SSO Service API
// @version         1.0
// @description     Simple SSO microservice: register, auth (JWT), protected login.
// @BasePath        /
//
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer <JWT>".
type AuthService interface {
	CreateUser(ctx context.Context, in domain.CreateUser) (int64, error)
	GetGoogleAuthURL(ctx context.Context) string
	LoginWithGoogle(ctx context.Context, code string) (domain.AuthSession, error)
	LoginByCredentials(ctx context.Context, in domain.AuthCredentials) (domain.AuthSession, error)
	SendOTP(ctx context.Context, destination, typ string) error
	ResetPhoneSendOTP(ctx context.Context, phone string) error
	ResetEmailSendOTP(ctx context.Context, email string) error
	ConfirmOTP(ctx context.Context, in domain.OTPConfirm) (domain.AuthSession, error)
	ConfirmPasswordReset(ctx context.Context, in domain.PasswordResetConfirm) (domain.AuthSession, error)

}

type Handler struct {
	authService   AuthService
	reqConverter  *ReqConverter
	respConverter *RespConverter
	log           *zap.Logger
	frontendURL   string
}

func New(authService AuthService, log *zap.Logger, frontendURL string) *Handler {
	return &Handler{
		authService:   authService,
		reqConverter:  NewReqConverter(),
		respConverter: NewRespConverter(),
		log: log.With(
			zap.String("layer", "transport"),
			zap.String("component", "handler"),
		),
		frontendURL: frontendURL,
	}
}
