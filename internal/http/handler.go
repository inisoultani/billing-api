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

func NewHandler(bs *service.BillingService) *Handler {
	return &Handler{
		billingService: bs,
	}
}

func (h *Handler) GetLoanByID(w http.ResponseWriter, r *http.Request) {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid loan id", http.StatusBadRequest)
	}
	loan, err := h.billingService.GetLoanByID(r.Context(), loanID)
	if err != nil {
		switch err {
		case service.ErrLoanNotFound:
			http.Error(w, "Loan not found", http.StatusNotFound)
		default:
			log.Printf("Error find loan %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// intentionally put isDelinquent as part of the loan detail rather than as a separated rest API
	// considering :
	// - avoiding complexity where frontend must call 2 endpoints
	// - avoiding adding more latency
	isDelinquent, err := h.billingService.IsDelinquent(r.Context(), loan.ID, time.Now())
	if err != nil {
		log.Printf("Failed to compute loan delinquency", err)
		http.Error(w, "Failed to compute loan delinquency", http.StatusInternalServerError)
		return
	}

	resp := detailLoanResponse{
		LoanID:              loan.ID,
		WeeklyPaymentAmount: loan.WeeklyPaymentAmount,
		TotalPayable:        loan.TotalPayableAmount,
		TotalWeeks:          loan.TotalWeeks,
		CreatedAt:           loan.CreatedAt.Format(time.RFC3339),
		IsDelinquent:        isDelinquent,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

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
		LoanID:              loan.ID,
		WeeklyPaymentAmount: loan.WeeklyPaymentAmount,
		TotalPayable:        loan.TotalPayableAmount,
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
		log.Printf("Error get outstanding %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	resp := outstandingResponse{
		LoanID:      loanID,
		Outstanding: outstanding,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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

func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid loan id", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		http.Error(w, "Invalid page limit number", http.StatusBadRequest)
		return
	}

	cursor, err := DecodeCursor(r)
	if err != nil {
		http.Error(w, "Invalid cursor format", http.StatusBadRequest)
		return
	}

	payments, nextCursor, err := h.billingService.ListPayments(r.Context(), loanID, limit, cursor)
	if err != nil {
		log.Printf("Error get list payments %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	encodedNextCursor, err := EncodeCursor(nextCursor)
	if err != nil {
		log.Printf("Error encoding next cursor %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := ToListPaymentResponse(payments, encodedNextCursor)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
