package http

import (
	"billing-api/internal/domain"
	"encoding/base64"
	"encoding/json"
)

type submitLoanRequest struct {
	PrincipalAmount    int64   `json:"principal_amount"`
	AnnualInterestRate float64 `json:"annual_interest_rate"`
	TotalWeeks         int     `json:"total_weeks"`
	StartDate          string  `json:"start_date"` // YYYY-MM-DD
}

type submitPaymentRequest struct {
	Amount int64 `json:"amount"`
}

func EncodeCursor(cursor *domain.PaymentCursor) (*string, error) {
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
