package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/turtlepavlo/stripe_integration/pkg/telemetry"
)

func (h *Handler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	log := telemetry.WithTrace(ctx, h.log)

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to read body")
		log.Error("failed to read request body", zap.Error(err))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.cfg.StripeWebhookSecret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid signature")
		log.Error("failed to verify stripe signature", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.String("stripe.event_id", event.ID),
		attribute.String("stripe.event_type", string(event.Type)),
	)

	switch event.Type {
	case CheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unmarshal checkout session failed")
			log.Error("failed to unmarshal stripe checkout session", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		orderIDStr := session.Metadata["order_id"]
		orderID, parseErr := uuid.Parse(orderIDStr)
		if parseErr != nil {
			span.RecordError(parseErr)
			span.SetStatus(codes.Error, "invalid metadata order_id")
			log.Error("invalid order_id in metadata",
				zap.Error(parseErr),
				zap.String("raw_id", orderIDStr),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.String("payment.order_id", orderID.String()),
			attribute.String("payment.session_id", session.ID),
		)

		walletID := int64(0)
		if rawWalletID, ok := session.Metadata["wallet_id"]; ok && rawWalletID != "" {
			parsedWalletID, err := strconv.ParseInt(rawWalletID, 10, 64)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "invalid metadata wallet_id")
				log.Error("invalid wallet_id in metadata",
					zap.Error(err),
					zap.String("raw_id", rawWalletID),
				)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			walletID = parsedWalletID
			span.SetAttributes(attribute.Int64("payment.wallet_id", walletID))
		}

		var confirmErr error
		if walletID > 0 {
			confirmErr = h.srv.ConfirmDeposit(ctx, orderID, session.ID)
		} else {
			confirmErr = h.srv.ConfirmPayment(ctx, orderID, session.ID)
		}

		if confirmErr != nil {
			span.RecordError(confirmErr)
			span.SetStatus(codes.Error, "confirm payment failed")
			log.Error("failed to confirm payment via webhook",
				zap.Error(confirmErr),
				zap.String("order_id", orderID.String()),
				zap.String("session_id", session.ID),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("payment confirmed successfully",
			zap.String("order_id", orderID.String()),
			zap.String("session_id", session.ID),
		)

		w.WriteHeader(http.StatusOK)
		return

	case RefundUpdated:
		var rf stripe.Refund
		if err := json.Unmarshal(event.Data.Raw, &rf); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "unmarshal refund failed")
			log.Error("failed to unmarshal stripe refund", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if rf.Status != stripe.RefundStatusSucceeded {
			log.Debug("ignored refund event (not succeeded)",
				zap.String("refund_id", rf.ID),
				zap.String("status", string(rf.Status)),
			)
			w.WriteHeader(http.StatusOK)
			return
		}

		meta := rf.Metadata

		userID, err := strconv.ParseInt(meta["user_id"], 10, 64)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid metadata user_id")
			log.Error("invalid metadata user_id",
				zap.String("value", meta["user_id"]),
				zap.Error(err),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		eventID, err := strconv.ParseInt(meta["event_id"], 10, 64)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid metadata event_id")
			log.Error("invalid metadata event_id",
				zap.String("value", meta["event_id"]),
				zap.Error(err),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ticketID, err := strconv.ParseInt(meta["ticket_id"], 10, 64)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid metadata ticket_id")
			log.Error("invalid metadata ticket_id",
				zap.String("value", meta["ticket_id"]),
				zap.Error(err),
			)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.String("refund.id", rf.ID),
			attribute.String("refund.status", string(rf.Status)),
			attribute.Int64("refund.user_id", userID),
			attribute.Int64("refund.event_id", eventID),
			attribute.Int64("refund.ticket_id", ticketID),
			attribute.Int64("refund.amount", rf.Amount),
		)

		refund := h.toDM.ToRefund(
			userID,
			eventID,
			ticketID,
			rf.Amount,
			meta["description"],
		)

		if err := h.srv.ConfirmRefund(ctx, rf.ID, refund); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "confirm refund failed")
			log.Error("failed to confirm refund via webhook",
				zap.Error(err),
				zap.String("refund_id", rf.ID),
				zap.Int64("user_id", userID),
				zap.Int64("event_id", eventID),
				zap.Int64("ticket_id", ticketID),
			)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Info("refund confirmed successfully",
			zap.String("refund_id", rf.ID),
			zap.Int64("user_id", userID),
			zap.Int64("event_id", eventID),
			zap.Int64("ticket_id", ticketID),
		)

		w.WriteHeader(http.StatusOK)
		return

	default:
		log.Debug("ignored stripe event type", zap.String("type", string(event.Type)))
		w.WriteHeader(http.StatusOK)
		return
	}
}
