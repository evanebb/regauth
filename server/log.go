package server

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
)

type contextHandler struct {
	slog.Handler
}

func newContextHandler(handler slog.Handler) *contextHandler {
	return &contextHandler{Handler: handler}
}

// Handle overrides the default Handle method to add context values.
func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	requestID := middleware.GetReqID(ctx)
	if requestID != "" {
		r.AddAttrs(slog.String("requestId", requestID))
	}

	return h.Handler.Handle(ctx, r)
}
