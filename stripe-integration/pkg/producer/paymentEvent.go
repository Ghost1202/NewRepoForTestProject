package producer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
	"go.uber.org/zap"
)

func (p *Producer) PublishPaymentEvent(ctx context.Context, event PaymentConfirmedEvent) error {
	log := telemetry.WithTrace(ctx, p.logger)

	payload, err := json.Marshal(event)
	if err != nil {
		log.Error("failed to marshal payment event", zap.Error(err))
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.OrderID),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Error("failed to publish payment event to kafka",
			zap.Error(err),
			zap.String("order_id", event.OrderID),
		)
		return err
	}

	log.Info("payment event published to kafka",
		zap.String("order_id", event.OrderID),
		zap.String("status", event.Status),
	)

	return nil
}
