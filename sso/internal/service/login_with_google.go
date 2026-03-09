package service

import (
	"context"
	"errors"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
	"github.com/turtlepavlo/sso/pkg/provider/google"
)

const opLoginWithGoogle = "AuthService.LoginWithGoogle"

func (s *AuthService) LoginWithGoogle(ctx context.Context, code string) (domain.AuthSession, error) {
	tracer := otel.Tracer("sso/service/auth")
	ctx, span := tracer.Start(ctx, opLoginWithGoogle)
	defer span.End()
	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", opLoginWithGoogle))

	gUser, err := s.googleProvider.FetchUser(ctx, code)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "fetch google user failed")
		log.Error("fetch google user failed", zap.Error(err))
		return domain.AuthSession{}, err
	}

	email := s.convToNorm.Email(gUser.Email)
	span.SetAttributes(attribute.String("auth.email", email))

	model, err := s.userRepo.GetByEmail(ctx, email)
	user, err := s.resolveGoogleUser(ctx, gUser, email, model, err)
	if err != nil {
		return domain.AuthSession{}, err
	}

	token, err := s.convToJWT.SignUser(user)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "token sign failed")
		log.Error("token sign failed", zap.Error(err), zap.Int64("user_id", user.ID))
		return domain.AuthSession{}, err
	}

	go func(id int64) {
		ev := s.convToProto.LoginEvent(id, authTypeGoogle)
		if pubErr := s.producer.PublishLoginEvent(context.Background(), ev); pubErr != nil {
			s.log.Error("publish login event failed", zap.Error(pubErr), zap.Int64("user_id", id))
		}
	}(user.ID)

	span.SetAttributes(attribute.Int64("auth.user_id", user.ID))
	log.Info("google login ok", zap.Int64("user_id", user.ID), zap.String("role", user.Role))
	return s.convToSession.ToAuthSession(token, user), nil
}

func (s *AuthService) resolveGoogleUser(ctx context.Context, gUser *google.UserInfo, email string, model *storage.UserLoginModel, getErr error) (*domain.User, error) {
	span := trace.SpanFromContext(ctx)
	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", opLoginWithGoogle))

	switch {
	case getErr == nil:
		return s.googleUserFromModel(ctx, model, gUser)
	case errors.Is(getErr, storage.ErrUserNotFound):
		return s.createGoogleUser(ctx, gUser, email)
	default:
		span.RecordError(getErr)
		span.SetStatus(codes.Error, "repo error")
		log.Error("get by email failed", zap.Error(getErr))
		return nil, getErr
	}
}

func (s *AuthService) createGoogleUser(ctx context.Context, gUser *google.UserInfo, email string) (*domain.User, error) {
	span := trace.SpanFromContext(ctx)
	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", opLoginWithGoogle))

	login := s.convToGoogle.LoginFromEmail(email)
	loginModel, infoModel := s.convToGoogle.ToStorageModels(gUser, login)
	id, saveErr := s.userRepo.SaveUserGoogle(ctx, loginModel, infoModel)
	if saveErr != nil {
		span.RecordError(saveErr)
		span.SetStatus(codes.Error, "save google user failed")
		log.Error("save google user failed", zap.Error(saveErr))
		return nil, saveErr
	}

	user := s.convToGoogle.ToDomainUser(id, gUser, login, defaultUserRole)
	go func(id int64, em string) {
		ev := s.convToProto.RegisterEvent(id, em)
		if pubErr := s.producer.PublishRegisterEvent(context.Background(), ev); pubErr != nil {
			s.log.Error("publish register event failed", zap.Error(pubErr), zap.Int64("user_id", id))
		}
	}(user.ID, user.Email)

	return user, nil
}

func (s *AuthService) googleUserFromModel(ctx context.Context, model *storage.UserLoginModel, gUser *google.UserInfo) (*domain.User, error) {
	span := trace.SpanFromContext(ctx)
	log := telemetry.WithTrace(ctx, s.log).With(zap.String("op", opLoginWithGoogle))

	user := s.convFromStorage.ToDomainUser(model)
	if user == nil {
		span.SetStatus(codes.Error, "user model nil")
		return nil, ErrInvalidCredentials
	}

	if user.GoogleID == nil {
		setErr := s.userRepo.SetGoogleID(ctx, user.ID, gUser.ID)
		if setErr != nil {
			log.Error("set google id failed", zap.Error(setErr), zap.Int64("user_id", user.ID))
		} else {
			s.convToGoogle.AssignGoogleID(user, gUser)
		}
	}

	return user, nil
}
