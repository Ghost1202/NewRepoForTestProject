package resend

import (
	"net/http"

	"go.uber.org/zap"
)

type Client struct {
	httpClient     *http.Client
	baseURL        string
	apiKey         string
	from           string
	maxBodyBytes   int64
	convertToEmail *ConvertToEmail
	log            *zap.Logger
}

func NewClient(cfg Config, log *zap.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL:        cfg.BaseURL,
		apiKey:         cfg.APIKey,
		from:           cfg.From,
		maxBodyBytes:   cfg.MaxBodyBytes,
		convertToEmail: NewConvertToEmail(),
		log:            log,
	}
}
