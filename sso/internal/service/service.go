package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	eventv1 "github.com/turtlepavlo/proto-contract/gen/go/analytics/event/v1"
	"github.com/turtlepavlo/sso/internal/domain"
	"github.com/turtlepavlo/sso/internal/storage"
	"github.com/turtlepavlo/sso/pkg/provider/google"
)

const (
	authTypePassword = "password"
	authTypePhone    = "phone"
	authTypeGoogle   = "google"
	authTypeReset    = "password_reset"
	typePhone        = "phone"
	typeEmail        = "email"
	defaultUserRole  = "user"
)

const (
	ReasonEmailRequired    = "email_required"
	ReasonPasswordRequired = "password_required"
	ReasonLoginRequired    = "login_required"

	ReasonDuplicateEmail     = "duplicate_email"
	ReasonDuplicateLogin     = "duplicate_login"
	ReasonInvalidCredentials = "invalid_credentials" //nolint:gosec
	ReasonOtpInvalid         = "otp_invalid"
	ReasonOtpNotFound        = "otp not foutd"
)

const (
	otpKeyPrefixLoginPhone = "otp:login_phone:"
	otpKeyPrefixResetPhone = "otp:reset_phone:"
	otpKeyPrefixResetEmail = "otp:reset_email:"
	otpMessagePrefixSMS    = "Your verification code is: "
	otpMessagePrefixEmail  = "Your verification code is: "
	otpEmailSubject        = "Verification code"
)

type UserRepo interface {
	Save(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) error
	GetByEmail(ctx context.Context, email string) (*storage.UserLoginModel, error)
	GetByID(ctx context.Context, id int64) (*storage.UserLoginModel, error)
	GetByLogin(ctx context.Context, login string) (*storage.UserLoginModel, error)
	SaveUserGoogle(ctx context.Context, login *storage.UserLoginModel, info *storage.UserInfoModel) (int64, error)
	SetGoogleID(ctx context.Context, userID int64, googleID string) error
	GetByPhone(ctx context.Context, phone string) (*storage.UserLoginModel, error)
	CreateUserByPhone(ctx context.Context, in storage.CreateUserByPhone) (int64, error)
	GetByGoogleID(ctx context.Context, googleID string) (*storage.UserLoginModel, error)
	UpdatePassHash(ctx context.Context, userID int64, passHash []byte) error
}

type CacheRepo interface {
	SetCode(ctx context.Context, key string, value int64) error
	GetCode(ctx context.Context, key string) (int64, error)
	DeleteCode(ctx context.Context, key string) error
}

type Producer interface {
	PublishLoginEvent(ctx context.Context, event *eventv1.LoginEvent) error
	PublishRegisterEvent(ctx context.Context, event *eventv1.RegisterEvent) error
}

type OAuthProvider interface {
	AuthCodeURL(ctx context.Context, state string) string
	FetchUser(ctx context.Context, code string) (*google.UserInfo, error)
}

type SMSProvider interface {
	SendSMS(ctx context.Context, msg domain.SMSMessage) error
}

type EmailProvider interface {
	SendEmail(ctx context.Context, msg domain.EmailMessage) error
}

type AuthService struct {
	log       *zap.Logger
	userRepo  UserRepo
	cacheRepo CacheRepo
	producer  Producer

	googleProvider OAuthProvider
	smsProvider    SMSProvider
	emailProvider  EmailProvider

	convToNorm    *ConvToNormalizer
	convToOTP     *ConvToOTP
	convToMessage *ConvToMessage

	convToStorage   *ConvToStorage
	convFromStorage *ConvFromStorage
	convToDomain    *ConvToDomain

	convToGoogle  *ConvToGoogle
	convToJWT     *ConvToJWT
	convToProto   *ConvToProto
	convToSession *ConvToSession
}

func New(
	userRepo UserRepo,
	cacheRepo CacheRepo,
	log *zap.Logger,
	jwtSecretKey []byte,
	tokenTTL time.Duration,
	producer Producer,
	googleProvider OAuthProvider,
	smsProvider SMSProvider,
	emailProvider EmailProvider,
	otpMin, otpMax int64,
) *AuthService {
	normalizer := NewConvToNormalizer()

	return &AuthService{
		log:             log,
		userRepo:        userRepo,
		cacheRepo:       cacheRepo,
		producer:        producer,
		googleProvider:  googleProvider,
		smsProvider:     smsProvider,
		emailProvider:   emailProvider,
		convToNorm:      normalizer,
		convToOTP:       NewConvToOTP(normalizer, otpMin, otpMax),
		convToMessage:   NewConvToMessage(normalizer),
		convToStorage:   NewConvToStorage(normalizer),
		convFromStorage: NewConvFromStorage(),
		convToDomain:    NewConvToDomain(normalizer),
		convToGoogle:    NewConvToGoogle(normalizer),
		convToJWT:       NewConvToJWT(jwtSecretKey, tokenTTL),
		convToProto:     NewConvToProto(),
		convToSession:   NewConvToSession(),
	}
}
