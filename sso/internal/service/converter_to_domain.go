package service

import "github.com/turtlepavlo/sso/internal/domain"

type ConvToDomain struct {
	norm *ConvToNormalizer
}

func NewConvToDomain(norm *ConvToNormalizer) *ConvToDomain {
	return &ConvToDomain{norm: norm}
}

func (f *ConvToDomain) ForPhone(id int64, phone string) *domain.User {
	return &domain.User{
		ID:    id,
		Phone: f.norm.Phone(phone),
		Role:  defaultUserRole,
	}
}
