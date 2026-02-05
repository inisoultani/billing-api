package http

import (
	"billing-api/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	billingService *service.BillingService
}

type submitLoanRequest struct {
	PrincipalAmount    int64   `json:"principal_amount"`
	AnnualInterestRate float64 `json:"annual_interest_rate"`
	TotalWeeks         int     `json:"total_weeks"`
	StartDate          string  `json:"start_date"` // YYYY-MM-DD
}

type submitPaymentRequest struct {
	Amount int64 `json:"amount"`
}

type submitLoanResponse struct {
	LoanID        int64 `json:"loan_id"`
	WeeklyPayment int64 `json:"weekly_payment"`
	TotalPayable  int64 `json:"total_payable"`
}

type outstandingResponse struct {
	LoanID      int64 `json:"loan_id"`
	Outstanding int64 `json:"outstanding"`
}

type submitPaymentResponse struct {
	PaymentID int64 `json:"payment_id"`
}

func NewHandler(bs *service.BillingService) *Handler {
	return &Handler{
		billingService: bs,
	}
}

func (h *Handler) SubmitLoan(w http.ResponseWriter, r *http.Request) {
	var req submitLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, "Invalid start_date", http.StatusBadRequest)
		return
	}

	loan, err := h.billingService.SubmitLoan(r.Context(), service.SubmitLoanInput{
		PrincipalAmount:    req.PrincipalAmount,
		AnnualInterestRate: req.AnnualInterestRate,
		TotalWeeks:         req.TotalWeeks,
		StartDate:          startDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := submitLoanResponse{
		LoanID:        loan.ID,
		WeeklyPayment: loan.WeeklyPaymentAmount,
		TotalPayable:  loan.TotalPayableAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	resp := outstandingResponse{
		LoanID:      loanID,
		Outstanding: outstanding,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) MakePayment(w http.ResponseWriter, r *http.Request) {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)

	if err != nil {
		http.Error(w, "Invalid loan id", http.StatusBadRequest)
		return
	}

	var req submitPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.billingService.SubmitPayment(r.Context(), service.SubmitPaymentInput{
		LoanID: loanID,
		Amount: req.Amount,
		PaidAt: time.Now(),
	})

	if err != nil {
		switch err {
		case service.ErrLoanNotFound:
			http.Error(w, "Loan not found", http.StatusNotFound)
		case service.ErrInvalidPayment:
			http.Error(w, "Invalid payment amount", http.StatusBadRequest)
		case service.ErrLoanAlreadyClosed:
			http.Error(w, "Loan already closed", http.StatusConflict)
		default:
			log.Printf("Error submit payment %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := submitPaymentResponse{
		PaymentID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
