package payment

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.uber.org/zap"
)

func (c *Consumer) Start(ctx context.Context) {
	c.logger.Info("kafka consumer started", zap.String("topic", c.reader.Config().Topic))

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				c.logger.Info("kafka consumer stopped")
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
			c.logger.Error("failed to process message",
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
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(value, &envelope); err != nil {
		return err
	}

	if _, ok := envelope["order_id"]; ok {
		var event ConfirmedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}

		if event.Status != StatusSucceeded && event.Status != StatusPaid {
			c.logger.Debug("skipping payment event",
				zap.String("status", event.Status),
				zap.String("order_id", event.OrderID),
			)
			return nil
		}

		ticketID, err := strconv.ParseInt(event.OrderID, 10, 64)
		if err != nil {
			return err
		}

		log := telemetry.WithTrace(ctx, c.logger)
		log.Info("processing payment confirmation",
			zap.Int64("ticket_id", ticketID),
			zap.String("payment_id", event.PaymentID),
			zap.Int64("amount", event.Amount),
		)

		return c.service.ConfirmTicketPayment(ctx, ticketID)
	}

	if _, ok := envelope["ticket_id"]; ok {
		var event RefundedEvent
		if err := json.Unmarshal(value, &event); err != nil {
			return err
		}
		if event.Status != StatusRefund && event.Status != StatusRefunded {
			c.logger.Debug("skipping refund event",
				zap.String("status", event.Status),
				zap.Int64("ticket_id", event.TicketID),
			)
			return nil
		}

		log := telemetry.WithTrace(ctx, c.logger)
		log.Info("processing refund confirmation",
			zap.Int64("ticket_id", event.TicketID),
			zap.Int64("event_id", event.EventID),
			zap.Int64("user_id", event.UserID),
			zap.Int64("amount", event.Amount),
			zap.String("currency", event.Currency),
		)

		return c.service.ConfirmTicketRefund(ctx, event.TicketID)
	}

	c.logger.Debug("skipping unknown kafka event format")
	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
