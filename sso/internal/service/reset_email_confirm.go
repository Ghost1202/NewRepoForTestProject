package service

import (
	"context"
	"errors"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
)

func (s *AuthService) ResetEmailConfirm(ctx context.Context, email string, code int64, newPassword string) (domain.AuthSession, error) {
	const op = "AuthService.ResetEmailConfirm"
	tracer := otel.Tracer("sso/service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", op))
	email = s.convToNorm.Email(email)
	newPassword = s.convToNorm.Password(newPassword)

	if email == "" || newPassword == "" {
		span.SetStatus(codes.Error, "invalid input")
		log.Info("reset email confirm rejected")
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	span.SetAttributes(attribute.String("auth.email_norm", email))
	key := s.convToOTP.KeyResetEmail(email)
	saved, err := s.cacheRepo.GetCode(ctx, key)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			span.SetStatus(codes.Error, "otp not found")
			log.Info("otp not found")
			return domain.AuthSession{}, ErrCodeNotFound
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache get failed")
		log.Error("cache get failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	if saved != code {
		span.SetStatus(codes.Error, "otp invalid")
		log.Info("otp invalid")
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	if err := s.cacheRepo.DeleteCode(ctx, key); err != nil {
		s.log.Warn("failed to delete otp from cache", zap.Error(err))
	}

	model, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			span.SetStatus(codes.Error, "user not found")
			log.Info("user not found")
			return domain.AuthSession{}, ErrInvalidCredentials
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo error")
		log.Error("get user failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "bcrypt failed")
		log.Error("hash password failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	if err = s.userRepo.UpdatePassHash(ctx, model.UserID, passHash); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "update password failed")
		log.Error("update password failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	user := s.convFromStorage.ToDomainUser(model)
	user.PassHash = passHash
	token, err := s.convToJWT.SignUser(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "token sign failed")
		log.Error("token sign failed", zap.Error(err), zap.Int64("user_id", user.ID))
		return domain.AuthSession{}, err
	}

	go func(id int64) {
		ev := s.convToProto.LoginEvent(id, authTypeReset)
		if pubErr := s.producer.PublishLoginEvent(context.Background(), ev); pubErr != nil {
			s.log.Error("publish login event failed", zap.Error(pubErr), zap.Int64("user_id", id))
		}
	}(user.ID)

	span.SetAttributes(attribute.Int64("auth.user_id", user.ID))
	log.Info("reset email ok", zap.Int64("user_id", user.ID))
	return s.convToSession.ToAuthSession(token, user), nil
}
