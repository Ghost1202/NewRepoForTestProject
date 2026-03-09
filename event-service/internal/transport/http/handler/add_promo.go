package handler

import (
	"encoding/json"
	"net/http"

	"github.com/turtlepavlo/event-service/internal/domain"
	"github.com/turtlepavlo/event-service/internal/lib/telemetry"
	"github.com/turtlepavlo/event-service/internal/transport/http/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// AddPromo godoc
// @Summary      Add promo to an existing event
// @Description  Adds promo codes linked to an event.
// @Tags         promos
// @Accept       json
// @Produce      json
// @Param        request body handler.AddPromoReq true "Promo data"
// @Success      201 {object} handler.AddPromoResp
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /promos/promo [post]
// @Security     ApiKeyAuth
func (h *Handler) AddPromo(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("/transport/http/handler")
	ctx, span := tracer.Start(r.Context(), "Handler.AddPromo")
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)
	actor, ok := ctx.Value(middleware.PerformerContextKey{}).(domain.Performer)
	if !ok {
		log.Error("performer context missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request AddPromoReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		log.Warn("decode failed", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	promos := h.reqConv.ToDomainPromos(request.Promos, request.EventID)
	promoIDs, err := h.srv.AddPromo(ctx, actor, request.EventID, promos)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service add promo failed")
		log.Error("service add promo failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := h.respConv.ToRespAddPromo(promoIDs)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode failed")
		log.Error("encode response failed", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
