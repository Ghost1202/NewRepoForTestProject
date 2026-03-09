package service

import (
	"context"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (s *AuthService) ResetEmailSendOTP(ctx context.Context, email string) error {
	const op = "AuthService.ResetEmailSendOTP"
	tracer := otel.Tracer("sso/service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()
	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", op))
	email = s.convToNorm.Email(email)
	if email == "" {
		span.SetStatus(codes.Error, "invalid email")
		log.Info("reset email send rejected")
		return ErrInvalidCredentials
	}

	span.SetAttributes(attribute.String("auth.email_norm", email))
	code, err := s.convToOTP.GenerateCode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "generate otp failed")
		log.Error("generate otp failed", zap.Error(err))
		return err
	}

	key := s.convToOTP.KeyResetEmail(email)
	if err = s.cacheRepo.SetCode(ctx, key, code); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "cache set failed")
		log.Error("cache set failed", zap.Error(err))
		return err
	}

	text := s.convToOTP.TextEmail(code)
	subject := s.convToOTP.EmailSubject()
	msg := s.convToMessage.Email(email, subject, text)
	if err = s.emailProvider.SendEmail(ctx, msg); err != nil {
		if err := s.cacheRepo.DeleteCode(ctx, key); err != nil {
			s.log.Warn("failed to delete otp from cache", zap.Error(err))
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "email provider failed")
		log.Error("email provider failed", zap.Error(err))
		return err
	}

	log.Info("reset email otp sent")
	return nil
}
