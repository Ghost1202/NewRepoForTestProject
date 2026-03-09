package service

import (
	"context"
	"errors"

	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) CreateUser(ctx context.Context, in domain.CreateUser) (int64, error) {
	const op = "AuthService.CreateUser"
	tracer := otel.Tracer("sso/service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", op))

	email := s.convToNorm.Email(in.Email)
	login := s.convToNorm.Login(in.Login)
	password := s.convToNorm.Password(in.Password)
	span.SetAttributes(
		attribute.String("auth.email_norm", email),
		attribute.String("auth.login", login),
	)

	if email == "" {
		span.SetStatus(codes.Error, "email required")
		log.Info("create user rejected", zap.String("reason", ReasonEmailRequired))
		return 0, ErrInvalidCredentials
	}
	if password == "" {
		span.SetStatus(codes.Error, "password required")
		log.Info("create user rejected", zap.String("reason", ReasonPasswordRequired))
		return 0, ErrInvalidCredentials
	}
	if login == "" {
		span.SetStatus(codes.Error, "login required")
		log.Info("create user rejected", zap.String("reason", ReasonLoginRequired))
		return 0, ErrInvalidCredentials
	}

	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		span.SetStatus(codes.Error, "duplicate email")
		log.Info("create user rejected", zap.String("reason", ReasonDuplicateEmail))
		return 0, ErrUserAlreadyExists
	}
	if !errors.Is(err, storage.ErrUserNotFound) {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo error")
		log.Error("get by email failed", zap.Error(err))
		return 0, err
	}

	_, err = s.userRepo.GetByLogin(ctx, login)
	if err == nil {
		span.SetStatus(codes.Error, "duplicate login")
		log.Info("create user rejected", zap.String("reason", ReasonDuplicateLogin))
		return 0, ErrUserAlreadyExists
	}
	if !errors.Is(err, storage.ErrUserNotFound) {
		span.RecordError(err)
		span.SetStatus(codes.Error, "repo error")
		log.Error("get by login failed", zap.Error(err))
		return 0, err
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "bcrypt failed")
		log.Error("hash password failed", zap.Error(err))
		return 0, err
	}

	in.Email = email
	in.Login = login
	loginModel, infoModel := s.convToStorage.FromCreateUser(in, passHash)
	if err = s.userRepo.Save(ctx, loginModel, infoModel); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "save failed")
		log.Error("save user failed", zap.Error(err))
		return 0, err
	}

	userID := loginModel.UserID
	go func(id int64, em string) {
		ev := s.convToProto.RegisterEvent(id, em)
		if pubErr := s.producer.PublishRegisterEvent(context.Background(), ev); pubErr != nil {
			s.log.Error("publish register event failed", zap.Error(pubErr), zap.Int64("user_id", id))
		}
	}(userID, loginModel.Email)

	span.SetAttributes(attribute.Int64("auth.user_id", userID))
	log.Info("user created", zap.Int64("user_id", userID))

	return userID, nil
}
