package service

import (
	"context"
	"errors"
	"strings"

	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (s *AuthService) ConfirmOTP(ctx context.Context, in domain.OTPConfirm) (domain.AuthSession, error) {
	const op = "AuthService.ConfirmOTP"
	tracer := otel.Tracer("service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.String("type", in.Type),
	)

	typ := strings.ToLower(strings.TrimSpace(in.Type))
	destination := s.convToNorm.DestinationByType(in.Destination, typ)
	if destination == "" {
		span.SetStatus(codes.Error, "destination empty")
		log.Info("confirm otp rejected")
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	if typ != typePhone {
		span.SetStatus(codes.Error, "unsupported otp type")
		log.Info("unsupported otp type")
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	key := s.convToOTP.KeyLoginPhone(destination)
	saved, err := s.cacheRepo.GetCode(ctx, key)
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

	if saved != in.Code {
		span.SetStatus(codes.Error, "otp invalid")
		log.Info("otp invalid",
			zap.Int64("expected", saved),
			zap.Int64("received", in.Code),
			zap.String("reason", ReasonOtpInvalid),
		)
		return domain.AuthSession{}, ErrInvalidCredentials
	}

	if err := s.cacheRepo.DeleteCode(ctx, key); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache delete failed")
		log.Error("cache delete failed", zap.Error(err))
	}

	model, err := s.userRepo.GetByPhone(ctx, destination)
	var user *domain.User
	var isNew bool

	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			span.RecordError(err)
			span.SetStatus(codes.Error, "repo error")
			log.Error("get by phone failed", zap.Error(err))
			return domain.AuthSession{}, err
		}

		createIn := s.convToStorage.CreateUserByPhone(destination)
		id, createErr := s.userRepo.CreateUserByPhone(ctx, createIn)
		if createErr != nil {
			span.RecordError(createErr)
			span.SetStatus(codes.Error, "create by phone failed")
			log.Error("create by phone failed", zap.Error(createErr))
			return domain.AuthSession{}, createErr
		}

		user = s.convToDomain.ForPhone(id, destination)
		isNew = true
	} else {
		user = s.convFromStorage.ToDomainUser(model)
		isNew = false
	}

	if user == nil {
		span.SetStatus(codes.Error, "user nil")
		log.Error("user nil conversion failed")
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
			ev := s.convToProto.RegisterEvent(id, authTypePhone)
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
	log.Info("otp confirmed", zap.Int64("user_id", user.ID))

	return s.convToSession.ToAuthSession(token, user), nil
}
