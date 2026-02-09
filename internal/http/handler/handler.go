package handler

import (
	"billing-api/internal/config"
	"billing-api/internal/contextkey"
	"billing-api/internal/service"
	"context"
	"net/http"
)

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

func (h *Handler) MakeHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			// centralized error handling
			h.HandleError(w, r, err)
		}
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
