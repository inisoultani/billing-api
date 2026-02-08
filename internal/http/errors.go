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

type AppError struct {
	Code    int    // we defind HTTP Status code here
	Message string // the message FE sees
	Err     error  // actual technical issue that captured
}

func (e *AppError) Error() string {
	return e.Message
}

// we'll reuse this within handler as default return object for bad request 400
func BadRequest(msg string, internalError error) error {
	return &AppError{
		Code:    http.StatusBadRequest,
		Message: msg,
		Err:     internalError,
	}
}

func (h *Handler) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	// handle predefined AppError type
	var appError *AppError
	if errors.As(err, &appError) {
		http.Error(w, appError.Message, appError.Code)
		return
	}

	// Check if the error string contains our custom repo prefix
	// Or check if it's a context error
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "repo-timeout") {
		log.Printf("[TIMEOUT] %s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "Database timeout", http.StatusGatewayTimeout)
		return
	}

	// handle other errors
	switch {
	case errors.Is(err, service.ErrLoanNotFound):
		http.Error(w, "Loan not found", http.StatusNotFound)
	case errors.Is(err, service.ErrInvalidPayment):
		http.Error(w, "Invalid payment amount", http.StatusBadRequest)
	case errors.Is(err, service.ErrLoanAlreadyClosed):
		http.Error(w, "Loan already closed", http.StatusConflict)
	case errors.Is(err, service.ErrDuplicatePayment):
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "payment already processed",
		})
	case errors.Is(err, service.ErrDelinquencyCheck):
		log.Printf("[LOGIC ERROR] %v", err)
		http.Error(w, "Failed to compute loan deliquency", http.StatusInternalServerError)
	default:
		log.Printf("[SERVER ERROR] %s %s - %v", r.Method, r.URL.Path, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
