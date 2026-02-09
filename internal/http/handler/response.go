package handler

import (
	"billing-api/internal/domain"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

type SubmitLoanResponse struct {
	LoanID              int64 `json:"loan_id"`
	WeeklyPaymentAmount int64 `json:"weekly_payment_amount"`
	TotalPayable        int64 `json:"total_payable"`
}

type DetailLoanResponse struct {
	LoanID              int64  `json:"loan_id"`
	TotalPayable        int64  `json:"total_payable"`
	WeeklyPaymentAmount int64  `json:"weekly_payment_amount"`
	TotalWeeks          int    `json:"total_weeks"`
	CreatedAt           string `json:"created_at"`
	IsDelinquent        bool   `json:"is_delinquent"`
}

type OutstandingResponse struct {
	LoanID      int64 `json:"loan_id"`
	Outstanding int64 `json:"outstanding"`
}

type SubmitPaymentResponse struct {
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

type ScheduleResponse struct {
	Sequence   int    `json:"sequence"`
	DueDate    string `json:"due_date"`
	Amount     int64  `json:"amount"`
	PaidAmount int64  `json:"paid_amount"`
	Status     string `json:"status"`
}

type ListScheduleResponse struct {
	Schedules  []ScheduleResponse `json:"schedules"`
	NextCursor *string            `json:"next_cursor,omitempty"`
}

func ToListPaymentResponse(payments []*domain.Payment, nextCursor *string) ListPaymentResponse {
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

// DecodeCursor generic function to decode any struct from a base64 URL query param
func DecodeCursor[T any](r *http.Request) (*T, error) {
	encodedCursor := r.URL.Query().Get("cursor")
	if encodedCursor == "" {
		return nil, nil
	}

	decodedCursor, err := base64.RawURLEncoding.DecodeString(encodedCursor)
	if err != nil {
		return nil, err
	}

	var cursor T
	if err := json.Unmarshal(decodedCursor, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}

// ToListScheduleResponse maps domain models to the API response format
func ToListScheduleResponse(schedules []*domain.LoanSchedule, nextCursor *string) ListScheduleResponse {
	list := make([]ScheduleResponse, len(schedules))
	for i, s := range schedules {
		list[i] = ScheduleResponse{
			Sequence:   s.Sequence,
			DueDate:    s.DueDate.Format("2006-01-02"), // Standard ISO date
			Amount:     s.Amount,
			PaidAmount: s.PaidAmount,
			Status:     s.Status,
		}
	}
	return ListScheduleResponse{
		Schedules:  list,
		NextCursor: nextCursor,
	}
}
