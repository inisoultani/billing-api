package http

import (
	"billing-api/internal/domain"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

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

func DecodeCursor(r *http.Request) (*domain.PaymentCursor, error) {
	encodedCursor := r.URL.Query().Get("cursor")
	if encodedCursor == "" {
		return nil, nil
	}

	decodedCursor, err := base64.RawURLEncoding.DecodeString(encodedCursor)
	if err != nil {
		return nil, err
	}

	var cursor domain.PaymentCursor
	if err := json.Unmarshal(decodedCursor, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}
