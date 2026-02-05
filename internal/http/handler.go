package http

import (
	"billing-api/internal/domain"
	"billing-api/internal/service"
	"encoding/base64"
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
	LoanID              int64 `json:"loan_id"`
	WeeklyPaymentAmount int64 `json:"weekly_payment_amount"`
	TotalPayable        int64 `json:"total_payable"`
}

type detailLoanResponse struct {
	LoanID              int64  `json:"loan_id"`
	TotalPayable        int64  `json:"total_payable"`
	WeeklyPaymentAmount int64  `json:"weekly_payment_amount"`
	TotalWeeks          int    `json:"total_weeks"`
	CreatedAt           string `json:"created_at"`
	IsDelinquent        bool   `json:"is_delinquent"`
}

type outstandingResponse struct {
	LoanID      int64 `json:"loan_id"`
	Outstanding int64 `json:"outstanding"`
}

type submitPaymentResponse struct {
	PaymentID int64 `json:"payment_id"`
}

type PaymentResponse struct {
	WeekNumber int    `json:"week_number"`
	Amount     int    `json:"amount"`
	PaidAt     string `json:"paid_at"`
}

type ListPaymentResponse struct {
	Payments   []PaymentResponse `json:"payments"`
	NextCursor *string           `json:"next_cursor,omitempty"`
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

	cursor, err := decodeCursor(r)
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

	encodedNextCursor, err := encodeCursor(nextCursor)
	if err != nil {
		log.Printf("Error encoding next cursor %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := toListPaymentResponse(payments, encodedNextCursor)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func toListPaymentResponse(payments []*domain.Payment, nextCursor *string) ListPaymentResponse {
	list := make([]PaymentResponse, len(payments))
	for i, p := range payments {
		list[i] = PaymentResponse{
			WeekNumber: p.WeekNumber,
			Amount:     int(p.Amount),
			PaidAt:     p.PaidAt.Format(time.RFC3339),
		}
	}
	return ListPaymentResponse{
		Payments:   list,
		NextCursor: nextCursor,
	}
}

func decodeCursor(r *http.Request) (*domain.PaymentCursor, error) {
	encodedCursor := r.URL.Query().Get("cursor")
	if encodedCursor == "" {
		return nil, nil
	}

	decodedCursor, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return nil, err
	}

	var cursor domain.PaymentCursor
	if err := json.Unmarshal(decodedCursor, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}

func encodeCursor(cursor *domain.PaymentCursor) (*string, error) {
	if cursor == nil {
		return nil, nil
	}

	data, err := json.Marshal(cursor)
	if err != nil {
		return nil, err
	}

	// use RawURLEncoding to avoid trailing '==' characters
	r := base64.RawURLEncoding.EncodeToString(data)
	return &r, nil
}
