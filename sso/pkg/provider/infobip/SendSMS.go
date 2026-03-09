package infobip

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/turtlepavlo/sso/internal/domain"
	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (c *Client) SendSMS(ctx context.Context, m domain.SMSMessage) error {
	const op = "Infobip.SendSMS"
	tracer := otel.Tracer("/provider/infobip")

	ctx, span := tracer.Start(ctx, op,
		trace.WithAttributes(
			attribute.String("messaging.system", "infobip"),
			attribute.String("http.method", http.MethodPost),
			attribute.String("user.phone", m.Phone),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, c.log).With(
		zap.String("op", op),
		zap.String("phone", m.Phone),
	)

	reqBody := c.convertToSMS.ToSMSRequest(c.senderID, m.Phone, m.Text)

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "marshal request failed")
		log.Error("marshal request failed", zap.Error(err))
		return err
	}

	url := "https://" + c.baseURL + "/sms/2/text/advanced"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "create request failed")
		log.Error("create request failed", zap.Error(err))
		return err
	}

	req.Header.Set("Authorization", "App "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "request failed")
		log.Error("request failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	const maxBodyLog = 2048
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Warn("failed to read infobip response body", zap.Error(readErr))
	} else {
		body := string(bodyBytes)
		if len(body) > maxBodyLog {
			body = body[:maxBodyLog] + "...(truncated)"
		}
		log.Info("infobip response", zap.Int("status", resp.StatusCode), zap.String("body", body))
	}

	if resp.StatusCode >= 400 {
		err = errors.New(resp.Status)
		span.RecordError(err)
		span.SetStatus(codes.Error, "infobip api error")
		log.Warn("infobip api returned error status", zap.Int("status", resp.StatusCode))
		return err
	}

	log.Info("sms sent", zap.Int("status", resp.StatusCode))
	return nil
}
