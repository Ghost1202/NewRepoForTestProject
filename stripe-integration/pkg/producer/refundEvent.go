package producer

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
	"go.uber.org/zap"
)

func (p *Producer) PublishRefundEvent(ctx context.Context, event TicketRefundedEvent) error {
	log := telemetry.WithTrace(ctx, p.logger)

	payload, err := json.Marshal(event)
	if err != nil {
		log.Error("failed to marshal refund event", zap.Error(err))
		return err
	}

	key := strconv.FormatInt(event.TicketID, 10)

	msg := kafka.Message{
		Key:   []byte(key),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Error("failed to publish refund event to kafka",
			zap.Error(err),
			zap.Int64("ticket_id", event.TicketID),
			zap.Int64("event_id", event.EventID),
			zap.Int64("user_id", event.UserID),
		)
		return err
	}

	log.Info("refund event published to kafka",
		zap.Int64("ticket_id", event.TicketID),
		zap.String("status", event.Status),
	)

	return nil
}
