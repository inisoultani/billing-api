package http

import (
	"billing-api/internal/service"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

func (h *Handler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// Check if the error string contains our custom repo prefix
	// Or check if it's a context error
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "repo-timeout") {
		log.Printf("[TIMEOUT] %s %s: %v", r.Method, r.URL.Path, err)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("Database timeout: " + err.Error()))
		return
	}

	// handle other errors
	switch err {
	case service.ErrLoanNotFound:
		http.Error(w, "Loan not found", http.StatusNotFound)
	case service.ErrInvalidPayment:
		http.Error(w, "Invalid payment amount", http.StatusBadRequest)
	case service.ErrLoanAlreadyClosed:
		http.Error(w, "Loan already closed", http.StatusConflict)
	case service.ErrDuplicatePayment:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "payment already processed",
		})
	default:
		log.Printf("general server error - %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
