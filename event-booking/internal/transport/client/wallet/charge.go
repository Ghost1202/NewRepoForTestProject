package wallet

import (
	"context"

	"github.com/turtlepavlo/event-booking/internal/domain"
	"github.com/turtlepavlo/event-booking/pkg/telemetry"
	"go.uber.org/zap"
)

func (c *Client) ChargeWallet(ctx context.Context, input domain.WalletCharge) (domain.WalletChargeStatus, error) {
	timeout := c.callTimeout

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	log := telemetry.WithTrace(ctx, c.log)

	log.Debug("sending grpc request to charge wallet",
		zap.String("booking_id", input.BookingID),
		zap.Int64("user_id", input.UserID),
		zap.Int64("event_id", input.EventID),
		zap.Int64("amount", input.Amount),
		zap.Int64("ticket_ids.count", int64(len(input.TicketIDs))),
	)

	req := c.converter.ToChargeRequest(input)
	resp, err := c.api.ChargeWallet(ctx, req)
	if err != nil {
		log.Error("grpc call ChargeWallet failed", zap.Error(err))
		return domain.WalletChargeStatusUnspecified, err
	}

	return c.converter.ToChargeStatus(resp.GetStatus()), nil
}
