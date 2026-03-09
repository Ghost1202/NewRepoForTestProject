package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/turtlepavlo/sso/internal/domain"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (c *Client) SendEmail(ctx context.Context, msg domain.EmailMessage) error {
	const op = "Resend.SendEmail"
	tracer := otel.Tracer("provider/resend")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("email.to", msg.To),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, c.log)

	reqBody := c.convertToEmail.ToSendEmailRequest(c.from, msg.To, msg.Subject, msg.Text)
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal request failed")
		log.Error("marshal resend request failed", zap.Error(err))
		return err
	}

	url := c.baseURL + "/emails"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create request failed")
		log.Error("create resend request failed", zap.Error(err))
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "http do failed")
		log.Error("resend request failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		err = errors.New(resp.Status)
		span.RecordError(err)
		span.SetStatus(codes.Error, "resend non-2xx")
		log.Error("resend non-2xx",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return err
	}

	log.Debug("email sent", zap.Int("status_code", resp.StatusCode))
	return nil
}
