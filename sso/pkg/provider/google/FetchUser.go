package google

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

var tracer = otel.Tracer("/provider/google")

func (c *Client) AuthCodeURL(ctx context.Context, state string) string {
	const op = "Google.AuthCodeURL"

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("oauth.provider", "google"),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, c.log)
	url := c.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	log.Info("auth url generated")

	return url
}

func (c *Client) FetchUser(ctx context.Context, code string) (*UserInfo, error) {
	const op = "Google.FetchUser"

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("oauth.provider", "google"),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, c.log)
	token, err := c.oauthCfg.Exchange(ctx, code)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "exchange failed")
		log.Error("oauth exchange failed", zap.Error(err))
		return nil, err
	}

	client := c.oauthCfg.Client(ctx, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.userInfoURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	if resp.StatusCode != http.StatusOK {
		err := errors.New(resp.Status)
		span.RecordError(err)
		span.SetStatus(codes.Error, "userinfo non-200")
		log.Warn("userinfo returned non-200", zap.Int("status", resp.StatusCode))
		return nil, err
	}

	var userInfo UserInfo
	if err := json.NewDecoder(io.LimitReader(resp.Body, c.maxBodyBytes)).Decode(&userInfo); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode userinfo failed")
		log.Error("decode userinfo failed", zap.Error(err))
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.email", userInfo.Email),
		attribute.String("user.id", userInfo.ID),
	)
	log.Info("userinfo fetched")

	return &userInfo, nil
}
