package google

import (
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Client struct {
	oauthCfg        *oauth2.Config
	userInfoURL     string
	maxBodyBytes    int64
	maxErrBodyBytes int64
	log             *zap.Logger
}

func NewClient(cfg Config, log *zap.Logger) *Client {
	return &Client{
		userInfoURL: cfg.UserInfoURL,
		oauthCfg: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				cfg.ScopeEmail,
				cfg.ScopeProfile,
			},
			Endpoint: google.Endpoint,
		},
		maxBodyBytes:    cfg.MaxBodyBytes,
		maxErrBodyBytes: cfg.MaxErrBodyBytes,
		log:             log,
	}
}
