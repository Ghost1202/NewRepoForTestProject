package service

import (
	"context"
	"strings"

	"github.com/turtlepavlo/sso/internal/domain"
)

func (s *AuthService) ConfirmPasswordReset(ctx context.Context, in domain.PasswordResetConfirm) (domain.AuthSession, error) {
	typ := strings.ToLower(strings.TrimSpace(in.Type))

	dst := s.convToNorm.DestinationByType(in.Destination, typ)
	if dst == "" || strings.TrimSpace(in.NewPassword) == "" {
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	switch typ {
	case typeEmail:
		return s.ResetEmailConfirm(ctx, dst, in.Code, in.NewPassword)

	case typePhone:
		return s.ResetPhoneConfirm(ctx, dst, in.Code, in.NewPassword)

	default:
		return domain.AuthSession{}, ErrInvalidCredentials
	}
}
