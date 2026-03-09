package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/internal/domain"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (c *Client) SendEmail(ctx context.Context, msg domain.Notification) error {
	const op = "ResendProvider.SendEmail"
	tracer := otel.Tracer("provider/resend")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("email.to", msg.To),
			attribute.String("email.subject", msg.Subject),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, c.log)

	reqBody := c.toMSG.ToSendEmailRequest(c.from, msg)
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal failed")
		log.Error("failed to marshal request", zap.Error(err))
		return err
	}

	url := fmt.Sprintf("%s/emails", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		span.RecordError(err)
		log.Error("failed to create http request", zap.Error(err))
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "http call failed")
		log.Error("failed to execute http request", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode >= 400 {
		err = fmt.Errorf("resend api error: status %d", resp.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, "api error")
		log.Error("resend api returned error",
			zap.Int("status", resp.StatusCode),
			zap.String("to", msg.To),
		)
		return err
	}

	log.Debug("email sent successfully", zap.String("to", msg.To))
	return nil
}
