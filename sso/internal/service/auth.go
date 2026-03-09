package service

import (
	"context"
	"errors"
	"strings"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
)

func (s *AuthService) LoginByCredentials(ctx context.Context, in domain.AuthCredentials) (domain.AuthSession, error) {
	isEmail := strings.Contains(in.Identifier, "@")

	return s.Login(ctx, in.Identifier, in.Password, isEmail)
}

func (s *AuthService) Login(ctx context.Context, identifier, password string, isEmail bool) (domain.AuthSession, error) {
	const op = "AuthService.Login"
	tracer := otel.Tracer("sso/service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()
	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.Bool("is_email", isEmail),
	)

	if isEmail {
		identifier = s.convToNorm.Email(identifier)
	} else {
		identifier = s.convToNorm.Login(identifier)
	}
	password = s.convToNorm.Password(password)
	span.SetAttributes(attribute.Bool("auth.is_email", isEmail))

	var model *storage.UserLoginModel
	var err error

	if isEmail {
		model, err = s.userRepo.GetByEmail(ctx, identifier)
	} else {
		model, err = s.userRepo.GetByLogin(ctx, identifier)
	}

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			span.SetStatus(codes.Error, "invalid credentials")
			log.Info("login rejected", zap.String("reason", ReasonInvalidCredentials))
			return domain.AuthSession{}, ErrInvalidCredentials
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "repo error")
		log.Error("get user failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	user := s.convFromStorage.ToDomainUser(model)
	if user == nil {
		span.SetStatus(codes.Error, "user model nil")
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	span.SetAttributes(attribute.Int64("auth.user_id", user.ID))

	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		span.SetStatus(codes.Error, "invalid credentials")
		log.Info("login rejected", zap.String("reason", ReasonInvalidCredentials), zap.Int64("user_id", user.ID))
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	token, err := s.convToJWT.SignUser(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "token sign failed")
		log.Error("token sign failed", zap.Error(err), zap.Int64("user_id", user.ID))
		return domain.AuthSession{}, err
	}

	go func(id int64) {
		ev := s.convToProto.LoginEvent(id, authTypePassword)
		if pubErr := s.producer.PublishLoginEvent(context.Background(), ev); pubErr != nil {
			s.log.Error("publish login event failed", zap.Error(pubErr), zap.Int64("user_id", id))
		}
	}(user.ID)

	log.Info("login ok", zap.Int64("user_id", user.ID), zap.String("role", user.Role))
	return s.convToSession.ToAuthSession(token, user), nil
}
