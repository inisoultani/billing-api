package domain

import (
	"billing-api/internal/infra/db/sqlc"
	"time"
)

type Payment struct {
	ID         int64
	WeekNumber int
	Amount     int64
	PaidAt     time.Time
}

func MapPayment(p sqlc.ListPaymentsByLoanIDRow) *Payment {
	return &Payment{
		ID:         p.ID,
		WeekNumber: int(p.WeekNumber),
		Amount:     p.Amount,
		PaidAt:     p.PaidAt.Time,
	}
}
