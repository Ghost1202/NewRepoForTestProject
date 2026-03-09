package resend

import (
	"net/http"

	"go.uber.org/zap"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	from       string
	log        *zap.Logger
	toMSG      *ConverterToMSG
}

func NewClient(cfg Config, log *zap.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		from:    cfg.From,
		log:     log,
		toMSG:   NewConverterToMSG(),
	}
}
