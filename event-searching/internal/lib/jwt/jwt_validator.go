package jwt

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/turtlepavlo/event-searching/internal/domain"
	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
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

func (v *Validator) Validate(ctx context.Context, tokenString string) (*domain.User, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return v.secretKey, nil
	})
	if err != nil {
		telemetry.WithTrace(ctx, v.log).Warn("token parsing failed", zap.Error(err))
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		telemetry.WithTrace(ctx, v.log).Warn("invalid token claims or token expired")
		return nil, ErrInvalidToken
	}

	if claims.UserID <= 0 {
		telemetry.WithTrace(ctx, v.log).Warn("invalid token claims: empty user_id")
		return nil, ErrInvalidToken
	}

	return &domain.User{
		ID:   claims.UserID,
		Role: claims.Role,
	}, nil
}
