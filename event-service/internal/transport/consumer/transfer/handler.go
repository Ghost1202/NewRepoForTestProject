package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"go.uber.org/zap"
)

func (c *Consumer) Start(ctx context.Context) {
	c.log.Info("kafka transfer consumer started",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("group_id", c.reader.Config().GroupID),
	)

	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				c.log.Info("kafka transfer consumer stopped")
				return
			}
			c.log.Error("failed to fetch message from kafka", zap.Error(err))

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}

		if err := c.handleMessage(ctx, m.Value); err != nil {
			c.log.Error("failed to process transfer message",
				zap.Error(err),
				zap.String("key", string(m.Key)),
				zap.Int64("offset", m.Offset),
			)
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.log.Error("failed to commit message", zap.Error(err))
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, value []byte) error {
	var msg TicketTransfer
	if err := json.Unmarshal(value, &msg); err != nil {
		return fmt.Errorf("unmarshal ticket transfer event: %w", err)
	}

	log := telemetry.WithTrace(ctx, c.log)

	transfer := c.conv.TicketTransfer(msg)
	log.Info("processing ticket transfer",
		zap.Int64("ticket_id", transfer.TicketID),
		zap.Int64("from_user_id", transfer.FromUserID),
		zap.Int64("to_user_id", transfer.ToUserID),
	)

	return c.svc.TransferTicket(ctx, transfer)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
