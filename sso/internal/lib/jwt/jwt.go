package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	user "github.com/turtlepavlo/sso/internal/domain"
)

const (
	claimKeyUserID = "user_id"
	claimKeyRole   = "role"
	claimKeyExp    = "exp"
	claimKeyIat    = "iat"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrInvalidClaims = errors.New("invalid token claims")
)

type UserProvider interface {
	GetID() int64
	GetRole() string
}

func NewToken(userProvider UserProvider, duration time.Duration, secret []byte) (string, error) {
	const op = "jwt.NewToken"

	if len(secret) == 0 {
		return "", fmt.Errorf("%s: secret is empty", op)
	}
	if duration <= 0 {
		return "", fmt.Errorf("%s: duration must be positive", op)
	}

	expirationTime := time.Now().Add(duration)

	claims := jwtlib.MapClaims{
		claimKeyUserID: userProvider.GetID(),
		claimKeyRole:   userProvider.GetRole(),
		claimKeyExp:    expirationTime.Unix(),
		claimKeyIat:    time.Now().Unix(),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("%s: sign token: %w", op, err)
	}

	return tokenString, nil
}

func ParseToken(tokenString string, secret []byte) (jwtlib.MapClaims, error) {
	const op = "jwt.ParseToken"

	if tokenString == "" || len(secret) == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	parser := jwtlib.NewParser(jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}))

	token, err := parser.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwtlib.MapClaims)
	if !ok {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidClaims)
	}

	if _, hasExp := claims[claimKeyExp]; !hasExp {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidClaims)
	}
	if _, hasIat := claims[claimKeyIat]; !hasIat {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidClaims)
	}

	return claims, nil
}

func ParseUserClaims(claims jwtlib.MapClaims) (*user.User, error) {
	const op = "jwt.ParseUserClaims"

	userID, ok := toInt64(claims[claimKeyUserID])
	role, okRole := claims[claimKeyRole].(string)
	if !ok || !okRole || role == "" {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidClaims)
	}

	return &user.User{
		ID:   userID,
		Role: role,
	}, nil
}

func toInt64(value interface{}) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case string:
		parsed, err := strconv.ParseInt(typed, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
