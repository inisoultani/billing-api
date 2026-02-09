package handler

import (
	"billing-api/internal/domain"
	"billing-api/internal/service"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) MakeHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			// centralized error handling
			h.HandleError(w, r, err)
		}
	}
}

func (h *Handler) GetLoanByID(w http.ResponseWriter, r *http.Request) error {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		return BadRequest("Invalid loan id", err)
	}
	loan, err := h.billingService.GetLoanByID(r.Context(), loanID)
	if err != nil {
		return err
	}

	// intentionally put isDelinquent as part of the loan detail rather than as a separated rest API
	// considering :
	// - avoiding complexity where frontend must call 2 endpoints
	// - avoiding adding more latency
	isDelinquent, err := h.billingService.IsDelinquent(r.Context(), loan.ID, time.Now())
	if err != nil {
		return err
	}

	resp := DetailLoanResponse{
		LoanID:              loan.ID,
		WeeklyPaymentAmount: loan.WeeklyPaymentAmount,
		TotalPayable:        loan.TotalPayableAmount,
		TotalWeeks:          loan.TotalWeeks,
		CreatedAt:           loan.CreatedAt.Format(time.RFC3339),
		IsDelinquent:        isDelinquent,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(resp)
}

func (h *Handler) SubmitLoan(w http.ResponseWriter, r *http.Request) error {
	var req SubmitLoanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return BadRequest("Invalid request body", err)
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return BadRequest("Invalid start_date", err)
	}

	loan, err := h.billingService.SubmitLoan(r.Context(), service.SubmitLoanInput{
		PrincipalAmount:    req.PrincipalAmount,
		AnnualInterestRate: req.AnnualInterestRate,
		TotalWeeks:         req.TotalWeeks,
		StartDate:          startDate,
	})
	if err != nil {
		return err
	}

	resp := SubmitLoanResponse{
		LoanID:              loan.ID,
		WeeklyPaymentAmount: loan.WeeklyPaymentAmount,
		TotalPayable:        loan.TotalPayableAmount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetOutstanding(w http.ResponseWriter, r *http.Request) error {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		return BadRequest("Invalid loan ID", err)
	}

	outstanding, err := h.billingService.GetOutstanding(r.Context(), loanID)
	if err != nil {
		return err
	}

	resp := OutstandingResponse{
		LoanID:      loanID,
		Outstanding: outstanding,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(resp)
}

func (h *Handler) MakePayment(w http.ResponseWriter, r *http.Request) error {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)

	if err != nil {
		return BadRequest("Invalid loan ID", err)
	}

	var req SubmitPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// http.Error(w, "Invalid request body", http.StatusBadRequest)
		return BadRequest("Invalid body request", err)
	}

	// extract idempotency key
	idempotencyKey := GetIdempotencyKey(r.Context())
	if idempotencyKey == "" {
		return BadRequest("Request failed due to not providing X-Idempotency-Key", err)
	}

	// sample on implementing hybrid timeout management
	// the policy that cover entire service+repo process,
	// while we later set dedicated limit on each repo process within the service
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	id, err := h.billingService.SubmitPayment(ctx, service.SubmitPaymentInput{
		LoanID:         loanID,
		Amount:         req.Amount,
		PaidAt:         time.Now(),
		IdempotencyKey: idempotencyKey,
	})

	if err != nil {
		return err
	}

	resp := SubmitPaymentResponse{
		PaymentID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	return json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListPayments(w http.ResponseWriter, r *http.Request) error {
	loanIDStr := chi.URLParam(r, "loanID")
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		return BadRequest("Invalid loan ID", err)
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return BadRequest("Invalid page limit number", err)
	}
	if limit <= 0 || limit > h.config.PagingLimitMax {
		limit = h.config.PagingLimitDefault
	}

	cursor, err := DecodeCursor[domain.PaymentCursor](r)
	if err != nil {
		return BadRequest("Invalid payment cursor", err)
	}

	payments, nextCursor, err := h.billingService.ListPayments(r.Context(), loanID, limit, cursor)
	if err != nil {
		return err
	}

	encodedNextCursor, err := EncodeCursor(nextCursor)
	if err != nil {
		return InternalError("Error encoding next cursor", err)
	}

	resp := ToListPaymentResponse(payments, encodedNextCursor)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(resp)
}

func (h *Handler) ListSchedules(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	loanIDStr := chi.URLParam(r, "loanID") // Assuming you use chi router
	loanID, err := strconv.ParseInt(loanIDStr, 10, 64)
	if err != nil {
		return BadRequest("Invalid loan ID", err)
	}

	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return BadRequest("Invalid page limit number", err)
	}
	if limit <= 0 || limit > h.config.PagingLimitMax {
		limit = h.config.PagingLimitDefault
	}

	cursor, err := DecodeCursor[domain.ScheduleCursor](r)
	if err != nil {
		return BadRequest("Invalid schedule cursor", err)
	}

	schedules, nextCursorObj, err := h.billingService.ListSchedules(ctx, loanID, limit, cursor)
	if err != nil {
		return err
	}

	nextCursorStr, err := EncodeCursor(nextCursorObj)
	if err != nil {
		return InternalError("Error encoding next cursor", err)
	}
	response := ToListScheduleResponse(schedules, nextCursorStr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}
