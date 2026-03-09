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

// AddBundle godoc
// @Summary      Add bundle promo to an existing event
// @Description  Adds bundle promo codes linked to an event.
// @Tags         promos
// @Accept       json
// @Produce      json
// @Param        request body handler.AddBundleReq true "Bundle promo data"
// @Success      201 {object} handler.AddBundleResp
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /promos/bundle [post]
// @Security     ApiKeyAuth
func (h *Handler) AddBundle(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("/transport/http/handler")
	ctx, span := tracer.Start(r.Context(), "Handler.AddBundle")
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)
	actor, ok := ctx.Value(middleware.PerformerContextKey{}).(domain.Performer)
	if !ok {
		log.Error("performer context missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request AddBundleReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		log.Warn("decode failed", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bundles := h.reqConv.ToDomainBundles(request.Bundles, request.EventID)
	bundleIDs, err := h.srv.AddBundle(ctx, actor, request.EventID, bundles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service add bundle failed")
		log.Error("service add bundle failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := h.respConv.ToRespAddBundle(bundleIDs)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode failed")
		log.Error("encode response failed", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
