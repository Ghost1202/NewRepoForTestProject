package handler

import (
	"encoding/json"
	"net/http"

	"github.com/turtlepavlo/event-searching/internal/lib/telemetry"
	mw "github.com/turtlepavlo/event-searching/internal/transport/http/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SetComment godoc
// @Summary      Create comment
// @Description  Create a comment for an event. Requires Bearer JWT. Rating is an integer in tenths (0..50); client divides by 10.
// @Tags         comments
// @Accept       json
// @Produce      json
// @Param        request  body      SetCommentRequest  true  "comment payload"
// @Success      204      {string}  string            "No Content"
// @Failure      400      {string}  string            "Bad Request"
// @Failure      401      {string}  string            "Unauthorized"
// @Failure      500      {string}  string            "Internal Server Error"
// @Security     BearerAuth
// @Router       /comments [post]
func (h *Handler) SetComment(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.SetComment"
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
	var req SetCommentRequest
	if err := dec.Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "decode body failed")
		log.Error("failed to decode request body", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, _ := mw.UserFromContext(ctx)
	span.SetAttributes(
		attribute.Int64("user_id", user.ID),
		attribute.String("user_role", user.Role),
		attribute.Int64("comment.event_id", req.EventID),
		attribute.Int64("comment.rating_tenths", req.Rating),
	)

	comment := h.reqConv.ToDomainUpsertComment(req, user.ID)
	if err := h.srv.SetComment(ctx, comment); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "service execution failed")
		log.Error("failed to create comment", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
