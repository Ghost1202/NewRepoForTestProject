package jwt

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.uber.org/zap"
)

var ErrInvalidToken = errors.New("invalid token")

type Validator struct {
	secretKey []byte
	log       *zap.Logger
}

func NewJWTValidator(secret string, log *zap.Logger) *Validator {
	return &Validator{
		secretKey: []byte(secret),
		log:       log,
	}
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (validator *Validator) Validate(ctx context.Context, tokenString string) (*domain.Performer, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return validator.secretKey, nil
	})
	if err != nil {
		telemetry.WithTrace(ctx, validator.log).Warn("token parsing failed", zap.Error(err))
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		telemetry.WithTrace(ctx, validator.log).Warn("invalid token claims or token expired")
		return nil, ErrInvalidToken
	}

	return &domain.Performer{
		ID:   claims.UserID,
		Role: claims.Role,
	}, nil
}
