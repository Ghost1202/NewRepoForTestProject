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

// AddEarly godoc
// @Summary      Add early promo to an existing event
// @Description  Adds early promo codes linked to an event.
// @Tags         promos
// @Accept       json
// @Produce      json
// @Param        request body handler.AddEarlyReq true "Early promo data"
// @Success      201 {object} handler.AddEarlyResp
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /promos/early [post]
// @Security     ApiKeyAuth
func (h *Handler) AddEarly(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("/transport/http/handler")
	ctx, span := tracer.Start(r.Context(), "Handler.AddEarly")
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)
	actor, ok := ctx.Value(middleware.PerformerContextKey{}).(domain.Performer)
	if !ok {
		log.Error("performer context missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request AddEarlyReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		log.Warn("decode failed", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	early := h.reqConv.ToDomainEarly(request.Early, request.EventID)
	earlyIDs, err := h.srv.AddEarly(ctx, actor, request.EventID, early)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service add early failed")
		log.Error("service add early failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := h.respConv.ToRespAddEarly(earlyIDs)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode failed")
		log.Error("encode response failed", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
