package service

import "errors"

var (
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidToken       = errors.New("invalid token")
	ErrCodeNotFound       = errors.New("otp code not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)
