package wallet

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.uber.org/zap"
)

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("kafka wallet consumer started",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group_id", c.reader.Config().GroupID),
	)

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				c.logger.Info("kafka wallet consumer stopped")
				return
			}
			c.logger.Error("failed to fetch message from kafka", zap.Error(err))

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}

		if err := c.handleMessage(ctx, m.Value); err != nil {
			c.logger.Error("failed to process wallet charge message",
				zap.Error(err),
				zap.String("key", string(m.Key)),
				zap.Int64("offset", m.Offset),
			)
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.logger.Error("failed to commit message", zap.Error(err))
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, value []byte) error {
	var event ChargeCompletedEvent
	if err := json.Unmarshal(value, &event); err != nil {
		return err
	}

	log := telemetry.WithTrace(ctx, c.logger)
	if event.UserID <= 0 {
		log.Warn("skip wallet charge event: invalid user_id",
			zap.Int64("user_id", event.UserID),
		)
		return nil
	}

	if len(event.TicketIDs) == 0 {
		log.Warn("skip wallet charge event: empty ticket_ids",
			zap.String("booking_id", event.BookingID),
			zap.Int64("user_id", event.UserID),
			zap.Int64("event_id", event.EventID),
		)
		return nil
	}

	for _, ticketID := range event.TicketIDs {
		if ticketID <= 0 {
			log.Warn("skip wallet charge event: invalid ticket_id",
				zap.Int64("ticket_id", ticketID),
				zap.String("booking_id", event.BookingID),
				zap.Int64("user_id", event.UserID),
				zap.Int64("event_id", event.EventID),
			)
			continue
		}

		log.Info("processing wallet charge confirmation",
			zap.String("booking_id", event.BookingID),
			zap.Int64("ticket_id", ticketID),
			zap.Int64("user_id", event.UserID),
			zap.Int64("event_id", event.EventID),
			zap.Int64("amount", event.Amount),
			zap.Int64("final_amount", event.FinalAmount),
			zap.String("currency", event.Currency),
			zap.Time("occurred_at", event.OccurredAt),
		)

		if err := c.service.ConfirmFromWallet(ctx, ticketID, event.UserID, event.EventID); err != nil {
			log.Error("failed to confirm ticket payment from wallet charge",
				zap.Error(err),
				zap.String("booking_id", event.BookingID),
				zap.Int64("ticket_id", ticketID),
				zap.Int64("user_id", event.UserID),
				zap.Int64("event_id", event.EventID),
			)
			return err
		}
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
