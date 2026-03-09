package service

import "strings"

type ConvToNormalizer struct{}

func NewConvToNormalizer() *ConvToNormalizer { return &ConvToNormalizer{} }

func (n *ConvToNormalizer) Email(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func (n *ConvToNormalizer) Login(v string) string {
	return strings.TrimSpace(v)
}

func (n *ConvToNormalizer) Phone(v string) string {
	return strings.TrimSpace(v)
}

func (n *ConvToNormalizer) Password(v string) string {
	return strings.TrimSpace(v)
}

func (n *ConvToNormalizer) Name(v string) string {
	return strings.TrimSpace(v)
}

func (n *ConvToNormalizer) DestinationByType(destination, typ string) string {
	t := strings.ToLower(strings.TrimSpace(typ))
	switch t {
	case "phone":
		return n.Phone(destination)
	case "email":
		return n.Email(destination)
	default:
		return strings.TrimSpace(destination)
	}
}
