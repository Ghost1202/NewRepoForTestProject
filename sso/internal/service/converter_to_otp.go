package service

import (
	"crypto/rand"
	"math/big"
	"strconv"
)

type ConvToOTP struct {
	norm *ConvToNormalizer
	min  int64
	max  int64
}

func NewConvToOTP(norm *ConvToNormalizer, otpMin, otpMax int64) *ConvToOTP {
	return &ConvToOTP{norm: norm, min: otpMin, max: otpMax}
}

func (c *ConvToOTP) GenerateCode() (int64, error) {
	diff := c.max - c.min + 1
	if diff <= 0 {
		return 0, ErrInvalidCredentials
	}

	bg := big.NewInt(diff)
	n, err := rand.Int(rand.Reader, bg)
	if err != nil {
		return 0, err
	}

	return n.Int64() + c.min, nil
}

func (c *ConvToOTP) KeyLoginPhone(phone string) string {
	return otpKeyPrefixLoginPhone + c.norm.Phone(phone)
}

func (c *ConvToOTP) KeyResetPhone(phone string) string {
	return otpKeyPrefixResetPhone + c.norm.Phone(phone)
}

func (c *ConvToOTP) KeyResetEmail(email string) string {
	return otpKeyPrefixResetEmail + c.norm.Email(email)
}

func (c *ConvToOTP) TextSMS(code int64) string {
	return otpMessagePrefixSMS + strconv.FormatInt(code, 10)
}

func (c *ConvToOTP) TextEmail(code int64) string {
	return otpMessagePrefixEmail + strconv.FormatInt(code, 10)
}

func (c *ConvToOTP) EmailSubject() string {
	return otpEmailSubject
}
