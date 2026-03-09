package service

import (
	"context"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (s *AuthService) ResetPhoneSendOTP(ctx context.Context, phone string) error {
	const op = "AuthService.ResetPhoneSendOTP"
	tracer := otel.Tracer("/service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", op))

	phone = s.convToNorm.Phone(phone)
	if phone == "" {
		span.SetStatus(codes.Error, "invalid phone")
		log.Info("reset phone send rejected")
		return ErrInvalidCredentials
	}

	span.SetAttributes(attribute.String("user.phone", phone))
	code, err := s.convToOTP.GenerateCode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "generate otp failed")
		log.Error("generate otp failed", zap.Error(err))
		return err
	}

	key := s.convToOTP.KeyResetPhone(phone)
	if err = s.cacheRepo.SetCode(ctx, key, code); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache set failed")
		log.Error("cache set failed", zap.Error(err))
		return err
	}

	text := s.convToOTP.TextSMS(code)
	sms := s.convToMessage.SMS(phone, text)
	if err = s.smsProvider.SendSMS(ctx, sms); err != nil {
		if err := s.cacheRepo.DeleteCode(ctx, key); err != nil {
			s.log.Warn("failed to delete otp from cache", zap.Error(err))
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "sms provider failed")
		log.Error("sms provider failed", zap.Error(err))
		return err
	}

	log.Info("reset otp sent")
	return nil
}
