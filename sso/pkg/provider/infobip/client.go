package infobip

import (
	"net/http"

	"go.uber.org/zap"
)

type Client struct {
	client       *http.Client
	baseURL      string
	apiKey       string
	senderID     string
	convertToSMS *ConvertToSMS
	log          *zap.Logger
}

func NewClient(cfg Config, log *zap.Logger) *Client {
	return &Client{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL:      cfg.BaseURL,
		apiKey:       cfg.APIKey,
		senderID:     cfg.SenderID,
		convertToSMS: NewConvertToSMS(),
		log:          log,
	}
}
