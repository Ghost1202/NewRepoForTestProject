package producer

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.uber.org/zap"
)

func (p *Producer) TicketTransfer(ctx context.Context, event TicketTransfer) error {
	log := telemetry.WithTrace(ctx, p.logger)

	payload, err := json.Marshal(event)
	if err != nil {
		log.Error("failed to marshal ticket transfer event", zap.Error(err))
		return err
	}

	msg := kafka.Message{
		Key:   []byte(strconv.FormatInt(event.TicketID, 10)),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Error("failed to publish ticket transfer event to kafka",
			zap.Error(err),
			zap.Int64("ticket_id", event.TicketID),
			zap.Int64("from_user_id", event.FromUserID),
			zap.Int64("to_user_id", event.ToUserID),
		)
		return err
	}

	log.Info("ticket transfer event published to kafka",
		zap.Int64("ticket_id", event.TicketID),
		zap.Int64("from_user_id", event.FromUserID),
		zap.Int64("to_user_id", event.ToUserID),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
