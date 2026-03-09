package service

import (
	"context"
	"errors"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
)

func (s *AuthService) LoginWithPhone(ctx context.Context, phone string, codeInput int64) (domain.AuthSession, error) {
	const op = "AuthService.LoginWithPhone"
	tracer := otel.Tracer("sso/service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", op))

	phone = s.convToNorm.Phone(phone)
	if phone == "" {
		span.SetStatus(codes.Error, "invalid phone")
		log.Info("login with phone rejected")
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	span.SetAttributes(attribute.String("user.phone", phone))
	key := s.convToOTP.KeyLoginPhone(phone)
	savedCode, err := s.cacheRepo.GetCode(ctx, key)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			span.SetStatus(codes.Error, "otp not found")
			log.Info("otp not found", zap.String("reason", ReasonOtpNotFound))
			return domain.AuthSession{}, ErrCodeNotFound
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "cache get failed")
		log.Error("cache get failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	if savedCode != codeInput {
		span.SetStatus(codes.Error, "otp invalid")
		log.Info("otp invalid", zap.String("reason", ReasonOtpInvalid))
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	if err := s.cacheRepo.DeleteCode(ctx, key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache delete failed")
		log.Error("cache delete failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	model, err := s.userRepo.GetByPhone(ctx, phone)

	var (
		user  *domain.User
		isNew bool
	)

	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "repo error")
			log.Error("get by phone failed", zap.Error(err))
			return domain.AuthSession{}, err
		}

		createIn := s.convToStorage.CreateUserByPhone(phone)
		id, createErr := s.userRepo.CreateUserByPhone(ctx, createIn)
		if createErr != nil {
			span.RecordError(createErr)
			span.SetStatus(codes.Error, "create by phone failed")
			log.Error("create by phone failed", zap.Error(createErr))
			return domain.AuthSession{}, createErr
		}

		user = s.convToDomain.ForPhone(id, phone)
		isNew = true
	} else {
		user = s.convFromStorage.ToDomainUser(model)
		isNew = false
	}

	if user == nil {
		span.SetStatus(codes.Error, "user nil")
		log.Error("user nil")
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	token, err := s.convToJWT.SignUser(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "token sign failed")
		log.Error("token sign failed", zap.Error(err), zap.Int64("user_id", user.ID))
		return domain.AuthSession{}, err
	}

	go func(ctx context.Context, id int64, newUser bool) {
		if newUser {
			ev := s.convToProto.RegisterEvent(id, "")
			if err := s.producer.PublishRegisterEvent(ctx, ev); err != nil {
				s.log.Error("publish register event failed", zap.Error(err), zap.Int64("user_id", id))
			}
			return
		}

		ev := s.convToProto.LoginEvent(id, authTypePhone)
		if err := s.producer.PublishLoginEvent(ctx, ev); err != nil {
			s.log.Error("publish login event failed", zap.Error(err), zap.Int64("user_id", id))
		}
	}(context.WithoutCancel(ctx), user.ID, isNew)

	span.SetAttributes(attribute.Int64("auth.user_id", user.ID))
	span.SetStatus(codes.Ok, "ok")
	log.Info("phone login ok", zap.Int64("user_id", user.ID))
	return s.convToSession.ToAuthSession(token, user), nil
}
