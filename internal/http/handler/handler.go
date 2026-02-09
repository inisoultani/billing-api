package handler

import (
	"billing-api/internal/contextkey"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
)

type LogContextHandler struct {
	slog.Handler
}

func (l *LogContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		r.AddAttrs(slog.String(string(contextkey.RequestIDKey), reqID))
	}
	return l.Handler.Handle(ctx, r)
}
