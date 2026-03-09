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

// GetBundle godoc
// @Summary      Get bundle promos
// @Description  Get list of bundle promos filtered by event_id with pagination.
// @Tags         promos
// @Accept       json
// @Produce      json
// @Param        request body handler.GetBundleReq true "Filter and Pagination"
// @Success      200 {object} handler.GetBundleRes
// @Failure      400 {string} string "Bad Request"
// @Failure      401 {string} string "Unauthorized"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /promos/bundle/search [post]
// @Security     ApiKeyAuth
func (h *Handler) GetBundle(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("/http/handler")
	ctx, span := tracer.Start(r.Context(), "Handler.GetBundle")
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)
	actor, ok := ctx.Value(middleware.PerformerContextKey{}).(domain.Performer)
	if !ok {
		log.Error("performer context missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request GetBundleReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		log.Warn("decode failed", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filter := h.reqConv.ToDomainFilter(request.EventID, request.Limit, request.Offset)
	bundles, err := h.srv.GetBundle(ctx, actor, filter)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service get bundle failed")
		log.Error("service get bundle failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := h.respConv.ToRespGetBundle(bundles)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode failed")
		log.Error("encode response failed", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
