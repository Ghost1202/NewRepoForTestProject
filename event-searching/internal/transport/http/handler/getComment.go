package handler

import (
	"encoding/json"
	"net/http"

	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// GetComments godoc
// @Summary      List comments
// @Description  List comments for an event. Rating is integer in tenths (0..50); client divides by 10.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        request  body      GetCommentsRequest   true  "list params"
// @Success      200      {object}  GetCommentsResponse  "comments page"
// @Failure      400      {string}  string              "Bad Request"
// @Failure      500      {string}  string              "Internal Server Error"
// @Router       /comments/list [post]
func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetComments"
	var tracer = otel.Tracer("http/handler")

	ctx, span := tracer.Start(r.Context(), op,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
		),
	)
	defer span.End()

	log := telemetry.WithTrace(ctx, h.log)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var req GetCommentsRequest
	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	span.SetAttributes(
		attribute.Int64("comment.event_id", req.EventID),
		attribute.Int64("paging.limit", req.Limit),
		attribute.Int64("paging.offset", req.Offset),
	)

	domainReq := h.reqConv.ToDomainListComments(req)
	items, err := h.srv.GetComments(ctx, domainReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service execution failed")
		log.Error("failed to get comments", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := h.respConv.ToCommentsResponse(req, items)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "encode response failed")
		log.Error("failed to encode response", zap.Error(err))
		return
	}
}
