package google

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	ClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	ClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
	RedirectURL  string `envconfig:"GOOGLE_REDIRECT_URL"`

	UserInfoURL  string `envconfig:"GOOGLE_USER_INFO_URL" default:"https://www.googleapis.com/oauth2/v2/userinfo"`
	ScopeEmail   string `envconfig:"GOOGLE_SCOPE_EMAIL" default:"https://www.googleapis.com/auth/userinfo.email"`
	ScopeProfile string `envconfig:"GOOGLE_SCOPE_PROFILE" default:"https://www.googleapis.com/auth/userinfo.profile"`

	Timeout time.Duration `envconfig:"GOOGLE_TIMEOUT" default:"5s"`

	MaxBodyBytes    int64 `envconfig:"GOOGLE_MAX_BODY_BYTES" default:"1048576"`
	MaxErrBodyBytes int64 `envconfig:"GOOGLE_MAX_ERR_BODY_BYTES" default:"4096"`
}

func (cfg Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.ClientID, validation.Required),
		validation.Field(&cfg.ClientSecret, validation.Required),
		validation.Field(&cfg.RedirectURL, validation.Required),
		validation.Field(&cfg.UserInfoURL, validation.Required),
		validation.Field(&cfg.ScopeEmail, validation.Required),
		validation.Field(&cfg.ScopeProfile, validation.Required),
	)
}
