package domain

import (
	"billing-api/internal/infra/db/sqlc"
	"time"
)

type Loan struct {
	ID                  int64
	PrincipalAmount     int64
	TotalPayableAmount  int64
	WeeklyPaymentAmount int64
	TotalWeeks          int
	CreatedAt           time.Time
}

func MapLoan(l sqlc.Loan) *Loan {
	return &Loan{
		ID:                  l.ID,
		PrincipalAmount:     l.PrincipalAmount,
		TotalPayableAmount:  l.TotalPayableAmount,
		WeeklyPaymentAmount: l.WeeklyPaymentAmount,
		TotalWeeks:          int(l.TotalWeeks),
		CreatedAt:           l.CreatedAt.Time,
	}
}
