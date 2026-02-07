package middleware

import (
	"context"
	"net/http"
)

type contextKey string

const IdempotencyKey contextKey = "idempotency_key"

func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Idempotency-Key")
		ctx := context.WithValue(r.Context(), IdempotencyKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
