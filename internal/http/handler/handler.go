package handler

import (
	"billing-api/internal/config"
	"billing-api/internal/contextkey"
	"billing-api/internal/service"
	"context"
	"log/slog"
	"net/http"

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

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

type Handler struct {
	billingService *service.BillingService
	config         *config.Config
}

func NewHandler(bs *service.BillingService, cfg *config.Config) *Handler {
	return &Handler{
		billingService: bs,
		config:         cfg,
	}
}

// GetIdempotencyKey safely retrieves the key from context
func GetIdempotencyKey(ctx context.Context) string {
	val, ok := ctx.Value(contextkey.IdempotencyKey).(string)
	if !ok {
		return ""
	}
	return val
}
