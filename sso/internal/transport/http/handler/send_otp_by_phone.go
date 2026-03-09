package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	telemetry "github.com/turtlepavlo/sso/pkg/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SendOTPByPhone godoc
// @Summary      Send OTP to phone
// @Description  Sends OTP code to phone number.
// @Tags         auth
// @Accept       json
// @Param        input  body      SendOTPByPhoneReq  true  "Phone"
// @Success      204   "No Content"
// @Failure      400   "Bad Request"
// @Failure      500   "Internal Server Error"
// @Router       /users/otp/phone/send [post]
func (hand *Handler) SendOTPByPhone(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.SendOTPByPhone"
	tr := otel.Tracer("http/handler")

	ctx, span := tr.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, hand.log).With(
		zap.String("op", op),
		zap.String("route", "/users/otp/phone/send"),
	)

	var req SendOTPByPhoneReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Phone = strings.TrimSpace(req.Phone)
	if req.Phone == "" {
		span.SetStatus(codes.Error, "missing required fields")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := hand.reqConverter.SendOTPByPhone(ctx, hand.authService, req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service send otp failed")
		log.Error("service send otp failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
