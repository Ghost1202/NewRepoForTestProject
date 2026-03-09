package stripe

import (
	"go.uber.org/zap"
)

type Client struct {
	cfg Config
	log *zap.Logger
}

func New(cfg Config, logger *zap.Logger) *Client {
	return &Client{
		cfg: cfg,
		log: logger,
	}
}
