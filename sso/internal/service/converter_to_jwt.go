package service

import (
	"strings"
	"time"

	"github.com/turtlepavlo/sso/internal/domain"
	jwtauth "github.com/turtlepavlo/sso/internal/lib/jwt"
)

type ConvToJWT struct {
	secret []byte
	ttl    time.Duration
}

func NewConvToJWT(secret []byte, ttl time.Duration) *ConvToJWT {
	return &ConvToJWT{
		secret: secret,
		ttl:    ttl,
	}
}

type jwtProvider struct {
	id   int64
	role string
}

func (p *jwtProvider) GetID() int64    { return p.id }
func (p *jwtProvider) GetRole() string { return p.role }

func (c *ConvToJWT) SignUser(u *domain.User) (string, error) {
	if u == nil {
		return "", ErrInvalidCredentials
	}

	if strings.TrimSpace(u.Role) == "" {
		return "", ErrInvalidRole
	}

	p := &jwtProvider{
		id:   u.ID,
		role: u.Role,
	}

	return jwtauth.NewToken(p, c.ttl, c.secret)
}
