package http

import (
	"billing-api/internal/service"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	billingService *service.BillingService
}

type outstandingResponse struct {
	LoanID      int64 `json:"loan_id"`
	Outstanding int64 `json:"outstanding"`
}

func NewHandler(bs *service.BillingService) *Handler {
	return &Handler{
		billingService: bs,
	}
}

func (h *Handler) GetOutstanding(w http.ResponseWriter, r *http.Request) {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid loan id", http.StatusBadRequest)
	}

	outstanding, err := h.billingService.GetOutstanding(r.Context(), loanID)
	if err != nil {
		if err == service.ErrLoanNotFound {
			http.Error(w, "Loan not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	resp := outstandingResponse{
		LoanID:      loanID,
		Outstanding: outstanding,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) MakePayment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("api not implemented yet"))
}
