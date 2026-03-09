package jwt

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.uber.org/zap"
)

var ErrInvalidToken = errors.New("invalid token")

type Validator struct {
	secretKey []byte
	log       *zap.Logger
}

func NewValidator(secret string, log *zap.Logger) *Validator {
	return &Validator{
		secretKey: []byte(secret),
		log:       log,
	}
}

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func (v *Validator) Validate(ctx context.Context, tokenString string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return v.secretKey, nil
	})
	if err != nil {
		log := telemetry.WithTrace(ctx, v.log)
		log.Warn("token parsing failed", zap.Error(err))
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return &domain.User{ID: claims.UserID}, nil
}
