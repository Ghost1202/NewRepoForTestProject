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

// CreateEvent godoc
// @Summary      Create a new event with tickets
// @Description  Creates an event and generates tickets for it using "constructor" and adds promos.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        event body handler.CreateEventReq true "Event creation data"
// @Success      201 {object} handler.CreateEventResp
// @Failure      400 {string} string "Bad Request"
// @Failure      500 {string} string "Internal Server Error"
// @Router       /events [post]
// @Security     ApiKeyAuth
func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("/transport/http/handler")
	ctx, span := tracer.Start(r.Context(), "Handler.Create")
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)

	actor, ok := ctx.Value(middleware.PerformerContextKey{}).(domain.Performer)
	if !ok {
		log.Error("performer context missing")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request CreateEventReq
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode failed")
		log.Warn("decode failed", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event := h.reqConv.ToDomainCreateEvent(request)
	sectors := h.reqConv.ToDomainTicketSectors(request.Constructor)
	promos := h.reqConv.ToDomainPromos(request.Promos, 0)
	early := h.reqConv.ToDomainEarly(request.Early, 0)
	bundles := h.reqConv.ToDomainBundles(request.Bundles, 0)
	eventID, err := h.srv.CreateEvent(ctx, actor, event, sectors, promos, early, bundles)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service create failed")
		log.Error("service create failed", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(CreateEventResp{EventID: eventID}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode failed")
		log.Error("encode create event response failed", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
