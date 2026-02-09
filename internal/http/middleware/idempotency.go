package middleware

import (
	"context"
	"net/http"

	billingApiContextKey "billing-api/internal/contextkey"
)

func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Idempotency-Key")
		ctx := context.WithValue(r.Context(), billingApiContextKey.IdempotencyKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
