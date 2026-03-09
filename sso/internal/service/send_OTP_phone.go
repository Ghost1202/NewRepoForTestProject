package service

import (
	"context"
	"fmt"
	"strings"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func (s *AuthService) SendOTP(ctx context.Context, destination, typ string) error {
	const op = "AuthService.SendOTP"
	tracer := otel.Tracer("service/auth")

	ctx, span := tracer.Start(ctx, op)
	defer span.End()

	log := telemetry.WithTrace(ctx, s.log).With(
		zap.String("op", op),
		zap.String("dest", destination),
		zap.String("type", typ),
	)

	typ = strings.ToLower(strings.TrimSpace(typ))
	destination = s.convToNorm.DestinationByType(destination, typ)
	if destination == "" {
		span.SetStatus(codes.Error, "invalid destination")
		log.Warn("invalid destination")
		return ErrInvalidCredentials
	}

	code, err := s.convToOTP.GenerateCode()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "otp generation failed")
		log.Error("otp generation failed", zap.Error(err))
		return err
	}

	var key string

	switch typ {
	case typePhone:
		key = s.convToOTP.KeyLoginPhone(destination)
		if err := s.cacheRepo.SetCode(ctx, key, code); err != nil {
			span.RecordError(err)
			log.Error("failed to save otp to cache", zap.Error(err))
			return err
		}

		msg := s.convToMessage.SMS(destination, s.convToOTP.TextSMS(code))
		if err := s.smsProvider.SendSMS(ctx, msg); err != nil {
			span.RecordError(err)
			log.Error("failed to send sms", zap.Error(err))
			return err
		}

	case typeEmail:
		key = "otp:login_email:" + destination
		if err := s.cacheRepo.SetCode(ctx, key, code); err != nil {
			span.RecordError(err)
			log.Error("failed to save otp to cache", zap.Error(err))
			return err
		}

		msg := s.convToMessage.Email(destination, s.convToOTP.EmailSubject(), s.convToOTP.TextEmail(code))
		if err := s.emailProvider.SendEmail(ctx, msg); err != nil {
			span.RecordError(err)
			log.Error("failed to send email", zap.Error(err))
			return err
		}

	default:
		span.SetStatus(codes.Error, "unsupported otp type")
		return fmt.Errorf("unsupported otp type: %s", typ)
	}

	log.Info("otp sent successfully", zap.String("destination", destination))
	return nil
}
